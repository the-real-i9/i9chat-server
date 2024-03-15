package apptypes

type SignupSessionData struct {
	SessionId string `json:"sessionId"`
	Email     string `json:"email"`
}

type JWTUserData struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
}
