// Command dummydb generates a restore-compatible SQLite backup file seeded
// with 50 fixed-deposit records. It reuses the production service layer so
// interest maths, closure types and renewal chains match the real app.
//
// Usage:
//
//	go run ./tools/dummydb [-out testdata/dummy_fd_50.db]
//
// The output is written with VACUUM INTO (same as the app's Backup feature),
// so Settings → Restore accepts it directly.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fdats/internal/api"
	"fdats/internal/domain"
	"fdats/internal/repo"
	"fdats/internal/service"
)

type spec struct {
	name         string
	number       string
	principal    int64
	startDaysAgo int
	tenureDays   int
}

// Deterministic 47-row base dataset (plus 3 renewal-created FDs = 50 total).
// startDaysAgo/tenureDays cycles spread near/far maturities and a few
// already-matured-but-unclosed deposits.
var baseSpecs = buildSpecs()

func buildSpecs() []spec {
	names := []string{
		"Asha Patil", "Ravi Kumar", "Meera Joshi", "Suresh Patil", "Anita Deshmukh",
		"Vikram Rao", "Sneha Kulkarni", "Amit Sharma", "Pooja Mehta", "Rahul Verma",
		"Divya Nair", "Karan Singh", "Lakshmi Iyer", "Manoj Gupta", "Neha Sharma",
		"Om Prakash", "Priya Menon", "Qureshi Ahmed", "Ramesh Chavan", "Sunita Rao",
		"Tarun Bhatia", "Usha Jadhav", "Vinod Kapse", "Wamanrao Jagtap", "Xavier D'Souza",
		"Yogesh Pawar", "Zoya Khan", "Arun Nakhate", "Bhavna Shinde", "Chetan More",
		"Dipali Sonawane", "Ekta Bhosale", "Farhan Shaikh", "Girish Talware", "Hema Wagh",
		"Ishaan Khot", "Jyoti Salunkhe", "Kailas Bhosale", "Lalita Deshmukh", "Mahesh Nikam",
		"Namrata Gaikwad", "Omkar Jagtap", "Pallavi Sawant", "Rohit Salunke", "Shalini Parab",
		"Tushar Thorat", "Vaishali Mane",
	}
	tenures := []int{180, 365, 365, 730, 730, 1095, 395, 270, 545, 910, 200, 450}
	starts := []int{5, 25, 80, 150, 240, 320, 400, 520, 640, 780, 900, 1050}
	principals := []int64{20000, 50000, 75000, 100000, 150000, 200000, 250000, 300000, 400000, 500000, 750000, 1000000}

	specs := make([]spec, 0, len(names))
	for i, name := range names {
		specs = append(specs, spec{
			name:         name,
			number:       fmt.Sprintf("C-%04d", 2001+i),
			principal:    principals[i%len(principals)],
			startDaysAgo: starts[i%len(starts)],
			tenureDays:   tenures[i%len(tenures)],
		})
	}
	return specs
}

func main() {
	out := flag.String("out", filepath.Join("testdata", "dummy_fd_50.db"), "output backup .db path")
	flag.Parse()

	tmpDir, err := os.MkdirTemp("", "dummydb-*")
	if err != nil {
		fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := repo.Open(tmpDir)
	if err != nil {
		fatal(err)
	}
	defer db.Close()

	svc, err := service.NewFDService(db)
	if err != nil {
		fatal(err)
	}

	now := time.Now().UTC()
	today := domain.FormatDate(now)

	created := make(map[string]api.FD, len(baseSpecs))
	for _, sp := range baseSpecs {
		fd, err := svc.Create(api.PreviewRequest{
			CustomerName:   sp.name,
			CustomerNumber: sp.number,
			Principal:      sp.principal,
			StartDate:      domain.FormatDate(now.AddDate(0, 0, -sp.startDaysAgo)),
			TenureDays:     sp.tenureDays,
		})
		if err != nil {
			fatal(fmt.Errorf("create FD for %s: %w", sp.number, err))
		}
		created[sp.number] = fd
	}

	// Three renewals on still-active deposits (principal + interest).
	for _, num := range []string{"C-2003", "C-2005", "C-2010"} {
		if _, err := svc.Renew(api.RenewRequest{
			FDNumber:   created[num].FDNumber,
			Mode:       domain.RenewPrincipalPlusInterest,
			TenureDays: 730,
			Remark:     "Renewed for another two years",
		}); err != nil {
			fatal(fmt.Errorf("renew %s: %w", num, err))
		}
	}

	// Three premature closures (held part of the tenure).
	for _, num := range []string{"C-2002", "C-2004", "C-2006"} {
		if _, err := svc.Close(api.CloseRequest{
			FDNumber:    created[num].FDNumber,
			ClosureDate: today,
			Remark:      "Closed early for a family expense",
		}); err != nil {
			fatal(fmt.Errorf("premature close %s: %w", num, err))
		}
	}

	// Three closures at maturity on already-matured deposits.
	for _, num := range []string{"C-2007", "C-2009", "C-2011"} {
		closure := created[num].MaturityDate
		if _, err := svc.Close(api.CloseRequest{
			FDNumber:    created[num].FDNumber,
			ClosureDate: closure,
			Remark:      "Matured deposit paid out",
		}); err != nil {
			fatal(fmt.Errorf("matured close %s: %w", num, err))
		}
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fatal(err)
	}
	if err := repo.BackupTo(db, *out); err != nil {
		fatal(fmt.Errorf("backup to %s: %w", *out, err))
	}

	abs, _ := filepath.Abs(*out)
	fmt.Printf("Wrote %s\n", abs)
	fmt.Printf("FD rows: 47 created + 3 renewed = 50 (9 closed, 41 active)\n")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "dummydb:", err)
	os.Exit(1)
}
