package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"fdats/internal/api"
	"fdats/internal/brand"
	"fdats/internal/domain"
	"fdats/internal/repo"
	"fdats/internal/service"
)

// App is the Wails binding surface. Every method delegates to the service
// layer; no business logic lives here.
type App struct {
	ctx       context.Context
	db        *sql.DB
	service   *service.FDService
	initError string
	authed    bool
}

// NewApp creates the application shell. Initialisation happens in startup so
// the window can open even if data loading needs a moment.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	db, err := repo.Open(dataDir())
	if err != nil {
		a.initError = "Could not open the local database: " + err.Error()
		return
	}
	a.db = db

	svc, err := service.NewFDService(db)
	if err != nil {
		a.initError = "Could not initialise the application: " + err.Error()
		return
	}
	a.service = svc

	// Seed the default login on first run and keep the recovery file in sync.
	username, code, err := svc.EnsureAuthSeeded()
	if err != nil {
		a.initError = "Could not initialise the login: " + err.Error()
		return
	}
	writeRecoveryFile(dataDir(), username, code)
}

// dataDir resolves the folder holding the local database. The
// FD_ATS_DATA_DIR environment variable overrides the default location,
// which also makes backups and tests straightforward.
func dataDir() string {
	if d := os.Getenv("FD_ATS_DATA_DIR"); d != "" {
		return d
	}
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		base = "."
	}
	return filepath.Join(base, "FixedDepositManagement")
}

// recoveryFileName is written next to the database so a forgotten login can
// always be recovered.
const recoveryFileName = "FD_ATS_RECOVERY.txt"

// writeRecoveryFile keeps FD_ATS_RECOVERY.txt in sync with the credentials.
func writeRecoveryFile(dir, username, code string) {
	content := fmt.Sprintf(
		"Fixed Deposit Management - Login Recovery\n"+
			"========================================\n\n"+
			"Username:      %s\n"+
			"Recovery code: %s\n\n"+
			"If you forget your username or password, open the application,\n"+
			"click \"Forgot password?\" on the sign-in screen and enter this\n"+
			"recovery code to set new credentials.\n\n"+
			"Keep this file safe. It is updated whenever the credentials change.\n",
		username, code,
	)
	_ = os.WriteFile(filepath.Join(dir, recoveryFileName), []byte(content), 0o600)
}

// ready gates every data call behind both a working backend and a signed-in
// session.
func (a *App) ready() error {
	if a.service != nil && a.authed {
		return nil
	}
	if a.service == nil {
		return domain.ErrInternal
	}
	return domain.ErrNotAuthenticated
}

// SystemStatus reports whether the backend finished initialising and whether
// a signed-in session survived a frontend reload.
func (a *App) SystemStatus() api.SystemStatus {
	return api.SystemStatus{Ready: a.service != nil, Error: a.initError, Authed: a.authed}
}

// AppInfo returns the branding shown on the About page and footer (SRS §31).
func (a *App) AppInfo() brand.Info {
	return brand.Current()
}

// PreviewFD calculates an FD without saving it (SRS §10).
func (a *App) PreviewFD(req api.PreviewRequest) (api.Calculation, error) {
	if err := a.ready(); err != nil {
		return api.Calculation{}, err
	}
	return a.service.Preview(req)
}

// CreateFD validates, calculates and saves a new FD (SRS §10).
func (a *App) CreateFD(req api.PreviewRequest) (api.FD, error) {
	if err := a.ready(); err != nil {
		return api.FD{}, err
	}
	return a.service.Create(req)
}

// EditFD updates an active FD's details and recomputes derived amounts.
// Closed FDs cannot be edited.
func (a *App) EditFD(req api.EditFDRequest) (api.FD, error) {
	if err := a.ready(); err != nil {
		return api.FD{}, err
	}
	return a.service.EditFD(req)
}

// ListFDs returns one page of the dashboard FD list (SRS §9.2).
func (a *App) ListFDs(req api.ListRequest) (api.ListResponse, error) {
	if err := a.ready(); err != nil {
		return api.ListResponse{}, err
	}
	return a.service.List(req)
}

// GetFD returns one FD with its history (SRS §19, §20).
func (a *App) GetFD(fdNumber string) (api.FDDetail, error) {
	if err := a.ready(); err != nil {
		return api.FDDetail{}, err
	}
	return a.service.Get(fdNumber)
}

// DashboardStats returns the dashboard KPIs (SRS §9.1).
func (a *App) DashboardStats() (api.DashboardStats, error) {
	if err := a.ready(); err != nil {
		return api.DashboardStats{}, err
	}
	return a.service.Dashboard()
}

// UpcomingMaturities lists FDs maturing in the requested window (SRS §28).
func (a *App) UpcomingMaturities(req api.UpcomingRequest) ([]api.UpcomingFD, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.service.Upcoming(req)
}

// RenewFD renews an FD into a new FD number (SRS §21, §22).
func (a *App) RenewFD(req api.RenewRequest) (api.RenewResult, error) {
	if err := a.ready(); err != nil {
		return api.RenewResult{}, err
	}
	return a.service.Renew(req)
}

// PreviewClosure calculates the payable amount before confirming a closure
// (SRS §26).
func (a *App) PreviewClosure(req api.CloseRequest) (api.ClosurePreview, error) {
	if err := a.ready(); err != nil {
		return api.ClosurePreview{}, err
	}
	return a.service.PreviewClosure(req)
}

// CloseFD closes an FD (SRS §23–§26).
func (a *App) CloseFD(req api.CloseRequest) (api.FD, error) {
	if err := a.ready(); err != nil {
		return api.FD{}, err
	}
	return a.service.Close(req)
}

// ReopenFD returns a closed FD to ACTIVE status with a recorded audit reason.
func (a *App) ReopenFD(req api.ReopenRequest) (api.FD, error) {
	if err := a.ready(); err != nil {
		return api.FD{}, err
	}
	return a.service.Reopen(req)
}

// ReverseRenewal withdraws the renewed FD and reopens the previous one.
func (a *App) ReverseRenewal(req api.ReverseRenewalRequest) (api.FD, error) {
	if err := a.ready(); err != nil {
		return api.FD{}, err
	}
	return a.service.ReverseRenewal(req)
}

// GetRateSlabs returns the interest rate configuration (SRS §13, §30).
func (a *App) GetRateSlabs() ([]domain.RateSlab, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.service.RateSlabs()
}

// SaveRateSlabs replaces the interest rate configuration (SRS §30).
func (a *App) SaveRateSlabs(req api.SaveSlabsRequest) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.service.SaveRateSlabs(req.Slabs)
}

// ExportReport writes the selected report to an .xlsx file (SRS §29).
func (a *App) ExportReport(req api.ReportRequest) (api.ReportResult, error) {
	if err := a.ready(); err != nil {
		return api.ReportResult{}, err
	}
	return a.service.ExportReport(req)
}

// PickReportPath opens the native save dialog so the user chooses where the
// report is written. An empty string means the dialog was cancelled.
func (a *App) PickReportPath(defaultName string) (string, error) {
	if a.ctx == nil {
		return "", domain.ErrInternal
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Report",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		return "", domain.ErrInternal
	}
	return path, nil
}

// PreviewReport returns the headings and first rows of a report so the user
// can review it before exporting.
func (a *App) PreviewReport(req api.ReportRequest) (api.ReportPreview, error) {
	if err := a.ready(); err != nil {
		return api.ReportPreview{}, err
	}
	return a.service.PreviewReport(req)
}

// MaturityChart returns active FD maturities bucketed by month (next 12
// months) for the dashboard chart.
func (a *App) MaturityChart() ([]api.MaturityBucket, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.service.MaturityChart()
}

// PickBackupPath opens the native save dialog for the database backup.
// An empty string means the dialog was cancelled.
func (a *App) PickBackupPath() (string, error) {
	if a.ctx == nil {
		return "", domain.ErrInternal
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Backup",
		DefaultFilename: "fd_backup_" + time.Now().Format("2006-01-02") + ".db",
		Filters: []runtime.FileFilter{
			{DisplayName: "Database Files (*.db)", Pattern: "*.db"},
		},
	})
	if err != nil {
		return "", domain.ErrInternal
	}
	return path, nil
}

// PickRestorePath opens the native open dialog for choosing a backup file.
// An empty string means the dialog was cancelled.
func (a *App) PickRestorePath() (string, error) {
	if a.ctx == nil {
		return "", domain.ErrInternal
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose Backup File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Database Files (*.db)", Pattern: "*.db"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", domain.ErrInternal
	}
	return path, nil
}

// BackupDatabase writes a consistent snapshot of the database to dest and
// returns the final path (a .db suffix is added when missing).
func (a *App) BackupDatabase(dest string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return "", domain.ErrBackupPathRequired
	}
	if !strings.HasSuffix(strings.ToLower(dest), ".db") {
		dest += ".db"
	}
	if err := repo.BackupTo(a.db, dest); err != nil {
		return "", fmt.Errorf("Could not create the backup: %v", err)
	}
	return dest, nil
}

// RestoreDatabase replaces the live database with the backup at src and
// reopens the data layer so the UI immediately shows the restored data.
// A safety copy of the current database is kept until the restore succeeds.
func (a *App) RestoreDatabase(src string) error {
	src = strings.TrimSpace(src)
	if src == "" {
		return domain.ErrRestorePathRequired
	}
	if !repo.IsBackupFile(src) {
		return domain.ErrInvalidBackup
	}
	dbPath := repo.DBPath(dataDir())
	srcAbs, srcErr := filepath.Abs(src)
	dstAbs, dstErr := filepath.Abs(dbPath)
	if srcErr == nil && dstErr == nil && srcAbs == dstAbs {
		return domain.ErrRestoreSameFile
	}

	safety := dbPath + ".pre-restore"
	_ = os.Remove(safety)
	if err := repo.CopyFile(dbPath, safety); err != nil {
		return fmt.Errorf("Could not back up the current database before restoring: %v", err)
	}

	reopen := func() error {
		db, err := repo.Open(dataDir())
		if err != nil {
			return err
		}
		svc, err := service.NewFDService(db)
		if err != nil {
			_ = db.Close()
			return err
		}
		a.db, a.service = db, svc
		return nil
	}
	rollback := func(cause error) error {
		_ = repo.CopyFile(safety, dbPath)
		_ = os.Remove(safety)
		if err := reopen(); err != nil {
			a.service = nil
		}
		return cause
	}

	if a.db != nil {
		_ = a.db.Close()
	}
	_ = os.Remove(dbPath + "-wal")
	_ = os.Remove(dbPath + "-shm")

	if err := repo.CopyFile(src, dbPath); err != nil {
		return rollback(fmt.Errorf("Could not copy the backup into place: %v", err))
	}
	if err := reopen(); err != nil {
		return rollback(fmt.Errorf("Could not reopen the restored database: %v", err))
	}
	_ = os.Remove(safety)
	return nil
}

// LoadSampleData seeds a temporary demo dataset into an empty database and
// returns the number of FDs stored afterwards. TEMPORARY — removed with the
// Settings button once real data exists.
func (a *App) LoadSampleData() (int, error) {
	if err := a.ready(); err != nil {
		return 0, err
	}
	return a.service.LoadSampleData()
}

// Login verifies the credentials and opens the session. Unlike ready(), it
// only requires a working backend — it is the way in.
func (a *App) Login(req api.LoginRequest) error {
	if a.service == nil {
		return domain.ErrInternal
	}
	if err := a.service.Login(req.Username, req.Password); err != nil {
		return err
	}
	a.authed = true
	return nil
}

// Logout closes the session. Idempotent.
func (a *App) Logout() {
	a.authed = false
}

// GetAuthInfo returns the signed-in username and recovery code for Settings.
func (a *App) GetAuthInfo() (api.AuthInfo, error) {
	if err := a.ready(); err != nil {
		return api.AuthInfo{}, err
	}
	username, code, err := a.service.AuthInfo()
	if err != nil {
		return api.AuthInfo{}, err
	}
	return api.AuthInfo{Username: username, RecoveryCode: code}, nil
}

// ChangeCredentials updates the login from Settings and refreshes the
// recovery file.
func (a *App) ChangeCredentials(req api.ChangeCredentialsRequest) (api.AuthInfo, error) {
	if err := a.ready(); err != nil {
		return api.AuthInfo{}, err
	}
	username, code, err := a.service.ChangeCredentials(req.CurrentPassword, req.NewUsername, req.NewPassword)
	if err != nil {
		return api.AuthInfo{}, err
	}
	writeRecoveryFile(dataDir(), username, code)
	return api.AuthInfo{Username: username, RecoveryCode: code}, nil
}

// ChangeUsername replaces the username from its own Settings card and
// refreshes the recovery file (SRS §30.2).
func (a *App) ChangeUsername(req api.ChangeUsernameRequest) (api.AuthInfo, error) {
	if err := a.ready(); err != nil {
		return api.AuthInfo{}, err
	}
	username, code, err := a.service.ChangeUsername(req.CurrentPassword, req.NewUsername)
	if err != nil {
		return api.AuthInfo{}, err
	}
	writeRecoveryFile(dataDir(), username, code)
	return api.AuthInfo{Username: username, RecoveryCode: code}, nil
}

// ChangePassword replaces the password from its own Settings card and
// refreshes the recovery file (SRS §30.2).
func (a *App) ChangePassword(req api.ChangePasswordRequest) (api.AuthInfo, error) {
	if err := a.ready(); err != nil {
		return api.AuthInfo{}, err
	}
	username, code, err := a.service.ChangePassword(req.CurrentPassword, req.NewPassword)
	if err != nil {
		return api.AuthInfo{}, err
	}
	writeRecoveryFile(dataDir(), username, code)
	return api.AuthInfo{Username: username, RecoveryCode: code}, nil
}

// ResetCredentials recovers a forgotten login using the recovery code and
// refreshes the recovery file. Available before signing in.
func (a *App) ResetCredentials(req api.ResetCredentialsRequest) (api.AuthInfo, error) {
	if a.service == nil {
		return api.AuthInfo{}, domain.ErrInternal
	}
	username, code, err := a.service.ResetCredentials(req.RecoveryCode, req.NewUsername, req.NewPassword)
	if err != nil {
		return api.AuthInfo{}, err
	}
	writeRecoveryFile(dataDir(), username, code)
	return api.AuthInfo{Username: username, RecoveryCode: code}, nil
}

// ListMembers returns one page of the member list with FD aggregates
// (SRS §51.1, §51.2).
func (a *App) ListMembers(req api.MemberListRequest) (api.MemberListResponse, error) {
	if err := a.ready(); err != nil {
		return api.MemberListResponse{}, err
	}
	return a.service.Members().List(req)
}

// GetMemberProfile returns a member with every linked FD (SRS §51.3).
func (a *App) GetMemberProfile(id int64) (api.MemberProfile, error) {
	if err := a.ready(); err != nil {
		return api.MemberProfile{}, err
	}
	return a.service.Members().GetProfile(id)
}

// SaveMember creates or updates a member record (SRS §50).
func (a *App) SaveMember(req api.SaveMemberRequest) (api.Member, error) {
	if err := a.ready(); err != nil {
		return api.Member{}, err
	}
	return a.service.Members().Save(req)
}

// PreviewMemberImport validates an Excel member file without saving anything
// and returns the import summary (SRS §52.3).
func (a *App) PreviewMemberImport(path string) (api.MemberImportPreview, error) {
	if err := a.ready(); err != nil {
		return api.MemberImportPreview{}, err
	}
	return a.service.Members().PreviewImport(path)
}

// CommitMemberImport re-validates and imports the member records from an
// Excel file, skipping existing GEN Nos. (SRS §52.4).
func (a *App) CommitMemberImport(path string) (api.MemberImportResult, error) {
	if err := a.ready(); err != nil {
		return api.MemberImportResult{}, err
	}
	return a.service.Members().CommitImport(path)
}

// PickMemberImportPath opens the native open dialog for choosing the Excel
// member file. An empty string means the dialog was cancelled.
func (a *App) PickMemberImportPath() (string, error) {
	if a.ctx == nil {
		return "", domain.ErrInternal
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose Member Excel File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", domain.ErrInternal
	}
	return path, nil
}
