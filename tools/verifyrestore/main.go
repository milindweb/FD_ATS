package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fdats/internal/api"
	"fdats/internal/domain"
	"fdats/internal/repo"
	"fdats/internal/service"
)

func main() {
	src := os.Args[1]
	if !repo.IsBackupFile(src) {
		fmt.Println("FAIL: IsBackupFile = false")
		os.Exit(1)
	}
	fmt.Println("OK: SQLite header valid (passes restore check)")

	tmp, err := os.MkdirTemp("", "restore-verify-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)

	dst := repo.DBPath(tmp)
	if err := repo.CopyFile(src, dst); err != nil {
		panic(err)
	}

	db, err := repo.Open(tmp)
	if err != nil {
		fmt.Println("FAIL: repo.Open:", err)
		os.Exit(1)
	}
	defer db.Close()

	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM fds`).Scan(&total); err != nil {
		panic(err)
	}
	var active, closed int
	if err := db.QueryRow(`SELECT COUNT(*) FROM fds WHERE status='ACTIVE'`).Scan(&active); err != nil {
		panic(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM fds WHERE status='CLOSED'`).Scan(&closed); err != nil {
		panic(err)
	}
	var renewed, premature, matured int
	db.QueryRow(`SELECT COUNT(*) FROM fds WHERE closure_type='RENEWED'`).Scan(&renewed)
	db.QueryRow(`SELECT COUNT(*) FROM fds WHERE closure_type='PREMATURE'`).Scan(&premature)
	db.QueryRow(`SELECT COUNT(*) FROM fds WHERE closure_type='MATURED'`).Scan(&matured)
	var history int
	db.QueryRow(`SELECT COUNT(*) FROM fd_history`).Scan(&history)
	var slabs int
	db.QueryRow(`SELECT COUNT(*) FROM rate_slabs`).Scan(&slabs)
	var linked int
	db.QueryRow(`SELECT COUNT(*) FROM fds WHERE renewed_from IS NOT NULL`).Scan(&linked)

	svc, err := service.NewFDService(db)
	if err != nil {
		fmt.Println("FAIL: NewFDService:", err)
		os.Exit(1)
	}
	resp, err := svc.List(api.ListRequest{Filter: domain.FilterAll, Page: 1, PageSize: 100})
	if err != nil {
		fmt.Println("FAIL: service.List:", err)
		os.Exit(1)
	}

	fmt.Printf("FDs total=%d active=%d closed=%d\n", total, active, closed)
	fmt.Printf("closure types: renewed=%d premature=%d matured=%d\n", renewed, premature, matured)
	fmt.Printf("renewed_from links=%d history rows=%d rate_slabs=%d\n", linked, history, slabs)
	fmt.Printf("service.List total=%d (first page items=%d)\n", resp.Total, len(resp.Items))

	if total != 50 || resp.Total != 50 {
		fmt.Println("FAIL: expected 50 FDs")
		os.Exit(1)
	}
	if active+closed != total || renewed+premature+matured != closed {
		fmt.Println("FAIL: status/closure counts inconsistent")
		os.Exit(1)
	}
	if linked != renewed || history == 0 || slabs != 4 {
		fmt.Println("FAIL: history/slab/link checks failed")
		os.Exit(1)
	}
	abs, _ := filepath.Abs(src)
	fmt.Println("PASS: restore round-trip OK for", abs)
}
