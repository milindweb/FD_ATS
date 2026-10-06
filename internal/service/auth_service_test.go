package service_test

import (
	"errors"
	"testing"

	"fdats/internal/domain"
)

func TestEnsureAuthSeedsDefaultAdmin(t *testing.T) {
	svc := newTestService(t)

	username, code, err := svc.EnsureAuthSeeded()
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if username != "Admin" {
		t.Errorf("username=%q, want Admin", username)
	}
	if code == "" || len(code) != len("FD-XXXX-XXXX") {
		t.Errorf("code=%q, want FD-XXXX-XXXX format", code)
	}

	// Second call must not rotate anything.
	again, code2, err := svc.EnsureAuthSeeded()
	if err != nil {
		t.Fatalf("seed again: %v", err)
	}
	if again != username || code2 != code {
		t.Errorf("seed not idempotent: %q/%q vs %q/%q", again, code2, username, code)
	}
}

func TestLoginWithDefaultCredentials(t *testing.T) {
	svc := newTestService(t)
	if _, _, err := svc.EnsureAuthSeeded(); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := svc.Login("Admin", "0000"); err != nil {
		t.Errorf("default login: %v", err)
	}
	if err := svc.Login("admin", "0000"); err != nil {
		t.Errorf("case-insensitive username: %v", err)
	}
	if err := svc.Login("Admin", "wrong"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("wrong password err=%v, want ErrInvalidCredentials", err)
	}
	if err := svc.Login("Nobody", "0000"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("wrong username err=%v, want ErrInvalidCredentials", err)
	}
	if err := svc.Login("", ""); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("empty creds err=%v, want ErrInvalidCredentials", err)
	}
}

func TestChangeCredentials(t *testing.T) {
	svc := newTestService(t)
	_, oldCode, err := svc.EnsureAuthSeeded()
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, _, err := svc.ChangeCredentials("wrong", "NewUser", "1234"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("wrong current password err=%v, want ErrInvalidCredentials", err)
	}
	if _, _, err := svc.ChangeCredentials("0000", "", "12"); !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("short password err=%v, want ErrPasswordTooShort", err)
	}
	if _, _, err := svc.ChangeCredentials("0000", "", ""); !errors.Is(err, domain.ErrNothingToChange) {
		t.Errorf("nothing to change err=%v, want ErrNothingToChange", err)
	}

	username, newCode, err := svc.ChangeCredentials("0000", "Manager", "5678")
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if username != "Manager" {
		t.Errorf("username=%q, want Manager", username)
	}
	if newCode == oldCode {
		t.Error("recovery code was not rotated")
	}

	if err := svc.Login("Manager", "5678"); err != nil {
		t.Errorf("login with new credentials: %v", err)
	}
	if err := svc.Login("Admin", "0000"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("old credentials err=%v, want ErrInvalidCredentials", err)
	}
}

func TestResetCredentials(t *testing.T) {
	svc := newTestService(t)
	_, code, err := svc.EnsureAuthSeeded()
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, _, err := svc.ResetCredentials("FD-BAD1-BAD2", "X", "1234"); !errors.Is(err, domain.ErrInvalidRecoveryCode) {
		t.Errorf("bad code err=%v, want ErrInvalidRecoveryCode", err)
	}
	if _, _, err := svc.ResetCredentials(code, "", "1234"); !errors.Is(err, domain.ErrUsernameRequired) {
		t.Errorf("empty username err=%v, want ErrUsernameRequired", err)
	}
	if _, _, err := svc.ResetCredentials(code, "X", "12"); !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("short password err=%v, want ErrPasswordTooShort", err)
	}

	username, newCode, err := svc.ResetCredentials(code, "Recovered", "9999")
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if username != "Recovered" || newCode == code {
		t.Errorf("reset returned %q/%q, want Recovered + rotated code", username, newCode)
	}
	if err := svc.Login("Recovered", "9999"); err != nil {
		t.Errorf("login after reset: %v", err)
	}
	// The used code must not work twice.
	if _, _, err := svc.ResetCredentials(code, "Hacker", "1234"); !errors.Is(err, domain.ErrInvalidRecoveryCode) {
		t.Errorf("reused code err=%v, want ErrInvalidRecoveryCode", err)
	}
}
