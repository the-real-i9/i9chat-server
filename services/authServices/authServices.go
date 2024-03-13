package authservices

func RequestNewAccount(email string) (jwtToken string, err error) {
	// check if email already exists. if yes, send error

	// generate 6-digit code and send to email

	// generate verification code and expiration

	// generate a JWT token that holds the "email" data

	// store the email(varchar), verfCode(int), and vefified(bool) in an ongoing_registration table

	// return the jwtToken

	return "", nil
}
