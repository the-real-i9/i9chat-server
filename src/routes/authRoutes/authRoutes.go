package authRoutes

import (
	"i9chat/src/controllers/authControllers/passwordResetControllers"
	"i9chat/src/controllers/authControllers/signinControllers"
	"i9chat/src/controllers/authControllers/signupControllers"
	authSess "i9chat/src/middlewares/sessionMiddlewares"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Post("/signup/request_new_account", signupControllers.RequestNewAccount)

	router.Post(
		"/signup/verify_email",
		authSess.SignupSession,
		signupControllers.VerifyEmail,
	)

	router.Post(
		"/signup/register_user",
		authSess.SignupSession,
		signupControllers.RegisterUser,
	)

	router.Post("/signin", signinControllers.Signin)

	router.Post(
		"/forgot_password/request_password_reset",
		passwordResetControllers.RequestPasswordReset,
	)

	router.Post(
		"/forgot_password/confirm_email",
		authSess.PasswordResetSession,
		passwordResetControllers.ConfirmEmail,
	)

	router.Post(
		"/forgot_password/reset_password",
		authSess.PasswordResetSession,
		passwordResetControllers.ResetPassword,
	)
}
