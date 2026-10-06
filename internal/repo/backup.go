package repo

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
)

// sqliteHeader is the magic prefix present in every SQLite database file.
const sqliteHeader = "SQLite format 3\x00"

// DBPath returns the database file location inside dataDir.
func DBPath(dataDir string) string {
	return filepath.Join(dataDir, "fd_management.db")
}

// BackupTo writes a consistent snapshot of db to dest using SQLite's
// VACUUM INTO, so it is safe while the application is running. An existing
// destination file is replaced (the save dialog already asked for consent).
func BackupTo(db *sql.DB, dest string) error {
	if _, err := os.Stat(dest); err == nil {
		if err := os.Remove(dest); err != nil {
			return err
		}
	}
	_, err := db.Exec("VACUUM INTO ?", dest)
	return err
}

// IsBackupFile reports whether path points at a readable SQLite database,
// identified by the file's magic header.
func IsBackupFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, len(sqliteHeader))
	n, err := io.ReadFull(f, buf)
	return err == nil && n == len(sqliteHeader) && string(buf) == sqliteHeader
}

// CopyFile copies src to dest, replacing dest if it exists.
func CopyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
