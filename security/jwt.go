package security

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v5"
)

var secretKey []byte = []byte(os.Getenv("JWT_KEY"))

func GenerateJWT(username string) (string, error) {

	signer, err := jwt.NewSignerHS(jwt.HS256, secretKey)
	if err != nil {
		return "", err
	}

	claims := &jwt.RegisteredClaims{
		Audience:  []string{"admin"},
		ID:        "asdf",
		Subject:   strings.Clone(username),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 100)),
	}

	builder := jwt.NewBuilder(signer)

	token, err := builder.Build(claims)
	if err != nil {
		return "", err
	}

	return token.String(), nil
}

func ValidateJWT(tokenStr, username string) (bool, error) {
	verifier, err := jwt.NewVerifierHS(jwt.HS256, secretKey)
	if err != nil {
		return false, err
	}

	tokenBytes := []byte(tokenStr)
	newToken, err := jwt.Parse(tokenBytes, verifier)
	if err != nil {
		return false, err
	}

	err = verifier.Verify(newToken)
	if err != nil {
		return false, err
	}

	var newClaims jwt.RegisteredClaims
	errClaims := json.Unmarshal(newToken.Claims(), &newClaims)
	if errClaims != nil {
		return false, err
	}

	if newClaims.IsSubject(username) {
		return true, nil
	}

	return false, nil
}
