package domain

import "errors"

// User-facing validation messages (SRS §40). Defined once so every layer
// reports exactly the same text.
var (
	ErrCustomerNameRequired   = errors.New("Please enter the customer/member name.")
	ErrCustomerNumberRequired = errors.New("Please enter the customer/member number.")
	ErrInvalidAmount          = errors.New("Please enter a valid deposit amount.")
	ErrStartDateRequired      = errors.New("Please select a start date.")
	ErrInvalidTenure          = errors.New("Please select a valid tenure.")
	ErrNoRateForTenure        = errors.New("No interest rate is configured for this tenure.")
	ErrInvalidDate            = errors.New("Please enter a valid date.")

	ErrFDNotFound             = errors.New("FD record not found.")
	ErrAlreadyClosed          = errors.New("This FD is already closed.")
	ErrClosedCannotRenew      = errors.New("This FD cannot be renewed because it is already closed.")
	ErrClosureBeforeStart     = errors.New("Closure date cannot be before the FD start date.")
	ErrClosureDateRequired    = errors.New("Please select a closure date.")
	ErrClosureRemarkRequired  = errors.New("Please enter a closure remark.")
	ErrRenewBeforeMaturity    = errors.New("Renewal cannot start before the FD maturity date.")
	ErrRenewNotMatured        = errors.New("This FD has not matured yet. Renewal becomes available on the maturity date.")
	ErrEditClosed             = errors.New("Closed FDs cannot be edited.")
	ErrInvalidRenewalMode     = errors.New("Please select a valid renewal option.")
	ErrNotClosed              = errors.New("This FD is not closed.")
	ErrReversalReasonRequired = errors.New("Please enter a reason for this reversal.")
	ErrRenewedCannotReopen    = errors.New("This FD was closed by a renewal. Reverse the renewal instead.")
	ErrNotRenewed             = errors.New("This FD was not closed by renewal.")
	ErrRenewalAlreadyUsed     = errors.New("The renewed FD is no longer active, so the renewal cannot be reversed.")
	ErrInvalidFilter          = errors.New("Please select a valid filter.")
	ErrInvalidReport          = errors.New("Please select a valid report.")
	ErrInvalidSlabConfig      = errors.New("Interest rate slabs are invalid. Check that ranges do not overlap and rates are positive.")
	ErrExportPathRequired     = errors.New("Please choose a location to save the report.")

	ErrBackupPathRequired  = errors.New("Please choose where to save the backup.")
	ErrRestorePathRequired = errors.New("Please select a backup file to restore.")
	ErrInvalidBackup       = errors.New("The selected file is not a valid backup.")
	ErrRestoreSameFile     = errors.New("Please choose a backup file other than the current database.")
	ErrSampleDataNotEmpty  = errors.New("Sample data can only be loaded into an empty database.")

	ErrGENRequired        = errors.New("Please enter the GEN No.")
	ErrMemberNameRequired = errors.New("Please enter the member name.")
	ErrDuplicateGEN       = errors.New("A member with this GEN No. already exists.")
	ErrMemberNotFound     = errors.New("Member record not found.")
	ErrMemberRequired     = errors.New("Please select an existing member.")
	ErrImportPathRequired = errors.New("Please choose an Excel file to import.")
	ErrNoImportableRows   = errors.New("No valid member records were found in the file.")
	ErrTemplatePathRequired = errors.New("Please choose a location to save the template.")

	ErrNotAuthenticated        = errors.New("Please sign in to continue.")
	ErrInvalidCredentials      = errors.New("Invalid username or password.")
	ErrUsernameRequired        = errors.New("Username is required.")
	ErrPasswordTooShort        = errors.New("Password must be at least 4 characters.")
	ErrCurrentPasswordRequired = errors.New("Enter your current password.")
	ErrPasswordsDontMatch      = errors.New("Passwords do not match.")
	ErrInvalidRecoveryCode     = errors.New("Invalid recovery code.")
	ErrNothingToChange         = errors.New("Enter a new username or password.")

	// ErrInternal deliberately hides technical detail from normal users
	// (SRS §40); the underlying cause is available via SystemStatus.
	ErrInternal = errors.New("Something went wrong. Please try again.")
)
