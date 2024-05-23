# API Documentation

**i9chat** is a chat application server API built entirely with WebSocket technology. This document provides a conprehensive usage instructions for client applications.

## Important Notes

Before we dive in, separated by horizontal lines are informations we need to keep in mind.

**⚠ Note:** All connections must be explicitly closed because they remain open to allow retries without needing to establish a new connection in case of errors.

---
"Sent" and "Received" data implies, "the data to send to the server" and "the expected response", respectively.

---
The are two **formats of received data** :

- The first format indicates a **success response**:

  ```json
    {
      "statusCode": 200,
      "body": {...}
    }
  ```

- The second format indicates an **error response** (if an error has occured):

  ```json
    {
      "statusCode": 4xx,
      "error": "reason for error"
    }
  ```

## API Endpoints

All communication is implemented with WebSocket, therefore, all URLs are of the `ws://` protocol. Futhermore, all sent/received data are of the JSON data type.

The endpoint usage instructions follow a defined pattern:

1. A description of the endpoint and its usage.
2. The endpoint URL.
3. The sent header, if any. Most likely, it'd be the `Authorization` header.
4. A description of each type of sent data and its corresponding received data, accompanied by examples.
5. Final instructions on the endpoint, if any.

The application's endpoints are organized into three categories: user authentication, user management, and chat operations.

### User Authentication

This category includes endpoints handling all aspects of user authentication.

- Signup
  - [Request new account](#sign-up-request-new-account---step-1)
  - [Verify email](#sign-up-verify-email---step-2)
  - [Register user](#sign-up-register-user---step-3)
- [Signin](#sign-in)

### User Management

This category includes endpoints handling all aspects of user management.

### Chat Operations

## Endpoints and Usage

Creating an account (Sign up) is divided into three steps. Each step is uniquely identified by a URL and *must be accessed in order*.

### Sign up: Request New Account - (Step 1)

Here, the client submits their email and a verification code is sent to it.

**URL:** `ws://localhost:8000/api/auth/signup/request_new_account`

**Sent Data:**

```json
{
  "email": "example_user@gmail.com"
}
```

**Received Data:** (Success)

The `signupSessionJwt` is a session token to establish a session for the signup process, just like a session cookie.

```json
{
  "statusCode": 200,
  "body": {
    "msg": "A 6-digit verification code has been sent to
    example_user@gmail.com",
    "signupSessionJwt": `${jwt}`
  }
}
```

**Received Data:** (Error)

```json
{
  "statusCode": 422,
  "error": "An account with this email already exists"
}
```

Finally, close the open connection and navigate the user to the email verification page.

### Sign up: Verify Email - (Step 2)

Here, the client provides the verification code sent to the email in the previous step. Any attempt to access these steps out of order will result in an error.

**URL:** `ws://localhost:8000/api/auth/signup/verify_email`

**Sent Header:** `Authorization: Bearer ${signupSessionJwt}`

**Sent Data:**

```json
{
  "code": 123456
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body": {
    "msg": "Your email example_user@gmail.com has been verified"
  }
}
```

**Received Data:** (Error)

```json
{
  "statusCode": 422,
  "error": "Incorrect verification code. Check your email or 
  re-submit your email",
  "error": "Verification code has expired. Re-submit your email"
}
// Note: The duplicated `error` keys here are just to provide the two possible errors you can get.
```

### Sign up: Register User - (Step 3)

Here, the client provides the user's account information to be used for registration.

**URL:** `ws://localhost:8000/api/auth/signup/register_user`

**Sent Header:** `Authorization: Bearer ${signupSessionJwt}`

**Sent Data:**

Note: `geolocation` data should not be provided explicitly by the user, rather it should be provided somehow by the client application.

```json
{
  "username": "abc",
  "password": "blablabla",
  "geolocation": "2, 5, 5"
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body": {
    "msg": "Signup success!",
    "user": {
      "id": 1,
      "username": "abc",
      ...
    },
    "authJwt": `${authJwt}`
  }
}
```

**Received Data:** (Error)

```json
{
  "statusCode": 422,
  // if an existing user name is used
  "error": "Username unavailable", 
}
```

### Sign in
