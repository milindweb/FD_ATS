package repo

import (
	"database/sql"
	"errors"
	"time"
)

// AuthRecord is the single login credential row (id = 1). The password is
// stored as a SHA-256 hash; the recovery code supports a forgotten login.
type AuthRecord struct {
	Username     string
	PasswordHash string
	RecoveryCode string
	UpdatedAt    string
}

// AuthRepo reads and writes the credential row.
type AuthRepo struct {
	db *sql.DB
}

// NewAuthRepo binds the repository to an open database.
func NewAuthRepo(db *sql.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

// Get returns the credential row, or nil when none has been created yet.
func (r *AuthRepo) Get() (*AuthRecord, error) {
	var rec AuthRecord
	err := r.db.QueryRow(
		"SELECT username, password_hash, recovery_code, updated_at FROM auth WHERE id = 1",
	).Scan(&rec.Username, &rec.PasswordHash, &rec.RecoveryCode, &rec.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// Save upserts the single credential row.
func (r *AuthRepo) Save(rec AuthRecord) error {
	if rec.UpdatedAt == "" {
		rec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := r.db.Exec(
		`INSERT INTO auth (id, username, password_hash, recovery_code, updated_at)
		 VALUES (1, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   username = excluded.username,
		   password_hash = excluded.password_hash,
		   recovery_code = excluded.recovery_code,
		   updated_at = excluded.updated_at`,
		rec.Username, rec.PasswordHash, rec.RecoveryCode, rec.UpdatedAt,
	)
	return err
}
