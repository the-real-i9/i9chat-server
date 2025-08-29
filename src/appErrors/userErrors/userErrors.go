package userErrors

const (
	EmailAlreadyExists   string = "uERR_4000" // an account with this email already exists
	IncorrectVerfCode    string = "uERR_4001" // incorrect verification code! check or re-submit your email
	VerfCodeExpired      string = "uERR_4002" // verification code expired! re-submit your email
	UsernameUnavailable  string = "uERR_4003" // username unavailable
	NonExistingUser      string = "uERR_4004" // user with the email doesn't exist
	IncorrectResetToken  string = "uERR_4005" // incorrect passwd reset token! check or re-submit your email
	ResetTokenExpired    string = "uERR_4006" // reset token expired! re-submit your email
	IncorrectCredentials string = "uERR_4007" // incorrect credentials
	MediaUploadTimedOut  string = "uERR_4008" // media upload timed out
)
