package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"fdats/internal/domain"
	"fdats/internal/repo"
)

// minPasswordLength is the shortest accepted login password.
const minPasswordLength = 4

// defaultUsername/defaultPassword seed a brand-new database.
const (
	defaultUsername = "Admin"
	defaultPassword = "0000"
)

// hashPassword returns the SHA-256 hex digest of the password. Plaintext is
// never stored.
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

// newRecoveryCode mints a code like FD-7K2Q-9X4M for the forgot-password flow.
func newRecoveryCode() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate recovery code: %w", err)
	}
	return fmt.Sprintf("FD-%02X%02X-%02X%02X", b[0], b[1], b[2], b[3]), nil
}

// EnsureAuthSeeded creates the default Admin/0000 credentials the first time
// the application runs and always returns the current username and recovery
// code so the caller can keep the recovery file in sync.
func (s *FDService) EnsureAuthSeeded() (string, string, error) {
	rec, err := s.auth.Get()
	if err != nil {
		return "", "", err
	}
	if rec == nil {
		code, err := newRecoveryCode()
		if err != nil {
			return "", "", err
		}
		rec = &repo.AuthRecord{
			Username:     defaultUsername,
			PasswordHash: hashPassword(defaultPassword),
			RecoveryCode: code,
			UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
		}
		if err := s.auth.Save(*rec); err != nil {
			return "", "", err
		}
	}
	return rec.Username, rec.RecoveryCode, nil
}

// Login verifies the credentials. The username comparison is
// case-insensitive; the password must match exactly.
func (s *FDService) Login(username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return domain.ErrInvalidCredentials
	}
	rec, err := s.auth.Get()
	if err != nil {
		return err
	}
	if rec == nil ||
		!strings.EqualFold(rec.Username, username) ||
		rec.PasswordHash != hashPassword(password) {
		return domain.ErrInvalidCredentials
	}
	return nil
}

// AuthInfo returns the current username and recovery code (Settings).
func (s *FDService) AuthInfo() (string, string, error) {
	rec, err := s.auth.Get()
	if err != nil {
		return "", "", err
	}
	if rec == nil {
		return "", "", domain.ErrNotAuthenticated
	}
	return rec.Username, rec.RecoveryCode, nil
}

// ChangeCredentials validates the current password and applies a new
// username and/or password (empty fields keep their current values). The
// recovery code is regenerated so the old code no longer works.
func (s *FDService) ChangeCredentials(currentPassword, newUsername, newPassword string) (string, string, error) {
	if strings.TrimSpace(currentPassword) == "" {
		return "", "", domain.ErrCurrentPasswordRequired
	}
	newUsername = strings.TrimSpace(newUsername)
	if newPassword == "" && newUsername == "" {
		return "", "", domain.ErrNothingToChange
	}
	if newPassword != "" && len(newPassword) < minPasswordLength {
		return "", "", domain.ErrPasswordTooShort
	}

	rec, err := s.auth.Get()
	if err != nil {
		return "", "", err
	}
	if rec == nil || rec.PasswordHash != hashPassword(currentPassword) {
		return "", "", domain.ErrInvalidCredentials
	}

	if newUsername != "" {
		rec.Username = newUsername
	}
	if newPassword != "" {
		rec.PasswordHash = hashPassword(newPassword)
	}
	code, err := newRecoveryCode()
	if err != nil {
		return "", "", err
	}
	rec.RecoveryCode = code
	rec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := s.auth.Save(*rec); err != nil {
		return "", "", err
	}
	return rec.Username, rec.RecoveryCode, nil
}

// ResetCredentials replaces the login after the recovery code is verified.
// A fresh recovery code is issued and returned.
func (s *FDService) ResetCredentials(recoveryCode, newUsername, newPassword string) (string, string, error) {
	newUsername = strings.TrimSpace(newUsername)
	if newUsername == "" {
		return "", "", domain.ErrUsernameRequired
	}
	if len(newPassword) < minPasswordLength {
		return "", "", domain.ErrPasswordTooShort
	}

	rec, err := s.auth.Get()
	if err != nil {
		return "", "", err
	}
	code := strings.ToUpper(strings.TrimSpace(recoveryCode))
	if rec == nil || code == "" || rec.RecoveryCode != code {
		return "", "", domain.ErrInvalidRecoveryCode
	}

	newCode, err := newRecoveryCode()
	if err != nil {
		return "", "", err
	}
	rec.Username = newUsername
	rec.PasswordHash = hashPassword(newPassword)
	rec.RecoveryCode = newCode
	rec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := s.auth.Save(*rec); err != nil {
		return "", "", err
	}
	return rec.Username, rec.RecoveryCode, nil
}
