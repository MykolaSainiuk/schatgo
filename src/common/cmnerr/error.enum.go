package cmnerr

import "errors"

var (
	ErrUniqueViolation     = errors.New("cannot insert duplicate")
	ErrNotFoundEntity      = errors.New("cannot found such entity")
	ErrPasswordMismatch    = errors.New("password mismatch")
	ErrInvalidToken        = errors.New("token is invalid")
	ErrExpiredToken        = errors.New("token has expired")
	ErrHashGeneration      = errors.New("hash failed to generate")
	ErrHashMismatch        = errors.New("hash mismatch")
	ErrGenerateAccessToken = errors.New("cannot access token")
	ErrServer              = errors.New("server error")
)
