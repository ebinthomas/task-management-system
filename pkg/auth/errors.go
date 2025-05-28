package auth

import "errors"

// Common errors
var (
	ErrNoAuthHeader       = errors.New("no authorization header")
	ErrInvalidAuthType    = errors.New("invalid authorization type")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInsufficientRole   = errors.New("insufficient role")
	ErrInvalidSignature   = errors.New("invalid token signature")
	ErrExpiredToken       = errors.New("token has expired")
	ErrInvalidIssuer      = errors.New("invalid token issuer")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found in context")
	ErrUnauthorizedRole   = errors.New("user role not authorized for this action")
	ErrResourceNotOwned   = errors.New("user does not own this resource")
) 