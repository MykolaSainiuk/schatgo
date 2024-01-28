package jwthelper

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
)

type jwtCustomClaims struct {
	types.TokenPayload
	jwt.StandardClaims
}
type jwtDataStruct struct {
	secretKey []byte
	issuer    string
	expr      int
}

//nolint:gochecknoglobals // quick way
var JwtServerData *jwtDataStruct

func InitJwtData() {
	JwtServerData = &jwtDataStruct{
		secretKey: getSecretKey(),
		issuer:    "schatgo",
		expr:      getAuthTokenExpr(),
	}
}

func GenerateToken(userID string, userName string) (string, error) {
	claims := &jwtCustomClaims{
		types.TokenPayload{
			UserID:   userID,
			UserName: userName,
		},
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(int32(JwtServerData.expr))).Unix(),
			Issuer:    JwtServerData.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(JwtServerData.secretKey)
	if err != nil {
		slog.Error("failed to generate auth token")
		return "", err
	}

	return tokenStr, nil
}

// func ValidateToken(encodedToken string) (*jwt.Token, error) {
// 	return jwt.Parse(encodedToken, keyFunc)
// }

func keyFunc(token *jwt.Token) (interface{}, error) {
	_, isValid := token.Method.(*jwt.SigningMethodHMAC)
	if !isValid {
		return nil, cmnerr.ErrInvalidToken
	}
	return JwtServerData.secretKey, nil
}

func VerifyToken(encodedToken string) (*types.TokenPayload, error) {
	jwtToken, err := jwt.ParseWithClaims(encodedToken, &jwtCustomClaims{}, keyFunc)
	if err != nil {
		var vErr *jwt.ValidationError
		ok := errors.As(err, &vErr)
		if ok && errors.Is(vErr.Inner, cmnerr.ErrExpiredToken) {
			return nil, cmnerr.ErrExpiredToken
		}
		return nil, cmnerr.ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*jwtCustomClaims)
	if !ok {
		return nil, cmnerr.ErrInvalidToken
	}

	return &payload.TokenPayload, nil
}

func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		slog.Error("No auth secret for access token")
		os.Exit(1)
		return nil
	}
	return []byte(secret)
}

func getAuthTokenExpr() int {
	expr := os.Getenv("ACCESS_TOKEN_EXPIRATION_SECONDS")
	if expr == "" {
		expr = "24"
	}
	n, err := strconv.Atoi(expr)
	if err != nil {
		slog.Warn("No expiration time for access token")
		return DefaultAccessTokenLifetimeHrsInSeconds
	}
	return n
}

const DefaultAccessTokenLifetimeHrsInSeconds int = 2 * 60 * 60
