package core_errors

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrConflict             = errors.New("conflict")
	ErrInvalidToken         = errors.New("invalid token")
	ErrUnauthorized         = errors.New("Unauthorized")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrAccessForbidden      = errors.New("access forbidden")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrEmptyFile            = errors.New("empty file")
	ErrInternalError        = errors.New("internal error")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
)
