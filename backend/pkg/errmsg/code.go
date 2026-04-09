package errmsg

const (
	CodeSuccess = 200
	CodeError   = 500

	// User errors
	UserNotExist         = 1001
	UserPasswordError    = 1002
	UserTokenNotExist    = 1003
	UserTokenInvalid     = 1004
	UserTokenExpired     = 1005
	UserPermissionDenied = 1006
	UserAccountDisabled  = 1007

	// Email errors
	EmailFormatError     = 2001
	EmailLimit60sError   = 2002
	EmailLimit1hError    = 2003
	EmailLimit24hError   = 2004
	EmailSendFailedError = 2005

	// Registration errors
	UserAlreadyExistsError = 3001
	UserUsernameError      = 3002
	UserPasswordFormatErr  = 3003
	UserCodeError          = 3004
	UserCodeNotExist       = 3005
	UserCodeLimitError     = 3006
	UserSignatureError     = 3007

	// Login errors
	LoginLimit1mError  = 4001
	LoginLimit1hError  = 4002
	LoginLimit24hError = 4003
)
