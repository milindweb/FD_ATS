// Command blankdb writes a restore-compatible empty database backup:
// full app schema via repo.Migrate, default rate slabs, zero FDs.
//
// Usage:
//
//	go run ./tools/blankdb [-out blank_fd.db]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"fdats/internal/repo"
	"fdats/internal/service"
)

func main() {
	out := flag.String("out", "blank_fd.db", "output backup .db path")
	flag.Parse()

	tmpDir, err := os.MkdirTemp("", "blankdb-*")
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

	// Default login: Admin / 0000 (same path the app uses on first run).
	username, code, err := svc.EnsureAuthSeeded()
	if err != nil {
		fatal(err)
	}
	if err := svc.Login("Admin", "0000"); err != nil {
		fatal(fmt.Errorf("default Admin/0000 login check: %w", err))
	}

	var fds int
	if err := db.QueryRow(`SELECT COUNT(*) FROM fds`).Scan(&fds); err != nil {
		fatal(err)
	}
	if fds != 0 {
		fatal(fmt.Errorf("expected 0 FDs, got %d", fds))
	}
	var authRows int
	if err := db.QueryRow(`SELECT COUNT(*) FROM auth`).Scan(&authRows); err != nil {
		fatal(err)
	}
	if authRows != 1 {
		fatal(fmt.Errorf("expected 1 auth row, got %d", authRows))
	}

	if dir := filepath.Dir(*out); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fatal(err)
		}
	}
	if err := repo.BackupTo(db, *out); err != nil {
		fatal(fmt.Errorf("backup to %s: %w", *out, err))
	}
	if !repo.IsBackupFile(*out) {
		fatal(fmt.Errorf("output failed IsBackupFile check: %s", *out))
	}

	abs, _ := filepath.Abs(*out)
	fmt.Printf("Wrote %s\n", abs)
	fmt.Printf("Empty FDs + rate slabs; login Admin / 0000 (user=%s)\n", username)
	fmt.Printf("Recovery code: %s (app rewrites FD_ATS_RECOVERY.txt on open)\n", code)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "blankdb:", err)
	os.Exit(1)
}
