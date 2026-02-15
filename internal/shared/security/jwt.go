package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid"
	ulidutil "github.com/vocbl/shared/utils/ulid"
)

func GenerateRefreshJWT(userID ulid.ULID, duration time.Duration, secretKey []byte) (string, ulid.ULID, error) {
	id := ulidutil.NewULID()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      id.String(),
		"userID":  userID.String(),
		"expires": time.Now().UTC().Add(duration).Unix(),
	})

	signedToken, err := token.SignedString(secretKey)

	if err != nil {
		return "", ulid.ULID{}, err
	}

	return signedToken, id, nil
}

func GenerateAccessJWT(userID ulid.ULID, duration time.Duration, secretKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":  userID.String(),
		"expires": time.Now().UTC().Add(duration).Unix(),
	})

	return token.SignedString(secretKey)
}

// returns tokenID, userID and an error
func VerifyRefreshJWT(token string, secretKey []byte) (ulid.ULID, ulid.ULID, error) {
	userID, claims, err := verifyJWT(token, secretKey)
	if err != nil {
		return ulid.ULID{}, ulid.ULID{}, err
	}

	tokenID, err := ulid.Parse(claims["id"].(string))
	if err != nil {
		return ulid.ULID{}, ulid.ULID{}, errors.New("invalid token")
	}

	return tokenID, userID, nil
}

func verifyJWT(token string, secretKey []byte) (ulid.ULID, jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return secretKey, nil
	})

	if err != nil {
		return ulid.ULID{}, nil, err
	}

	if !parsedToken.Valid {
		return ulid.ULID{}, nil, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return ulid.ULID{}, nil, errors.New("invalid token")
	}

	expires, ok := claims["expires"].(float64)
	if !ok {
		return ulid.ULID{}, nil, errors.New("invalid token")
	}

	if time.Unix(int64(expires), 0).Before(time.Now()) {
		return ulid.ULID{}, nil, errors.New("token has expired")
	}

	userID, err := ulid.Parse(claims["userID"].(string))
	if err != nil {
		return ulid.ULID{}, nil, errors.New("invalid token")
	}

	return userID, claims, nil
}
