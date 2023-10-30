package app

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const LoginCookieName = "LINKSHORTER_USER"
const SignSecret = "unkTpTVMUcHQmgWXADBpcueGlcBAVXHAD2zUAuQCzD0MOKYhcg6Cvjrarl9RMmDUXRZuQz36S8Hs0Ak3OgkQy8vweiYtF2NaVV3qZLDvKYd75zaU1InkwRUEHUj01gkbSItyLh5V2eLO7lHAmpTYQ7N0CjOElRKeTIe23HEC4rAfDAavOLKATqrMKJnCzQvLNSaMPhzXpo9MzbHHfbPImn6tmVQiK9h63tKSQx3Dz0Mj2A8NHef3cvCEHC"
const TokenTTL = time.Hour * 1

type Claims struct {
	jwt.RegisteredClaims
	UserUUID string
}

func GenerateUserUUID() string {
	return uuid.New().String()
}

func GenerateSign(UUID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
		},
		UserUUID: UUID,
	})

	tokenString, err := token.SignedString([]byte(SignSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func FetchUUID(sign string) (string, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(sign, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SignSecret), nil
	})

	if err != nil {
		return "", err
	}

	return claims.UserUUID, nil
}

// Authorize user if uuid expired or does not exist
func Authorize(writer http.ResponseWriter, request *http.Request) (string, error) {
	needToCreateSign, err := isNeedToCreateSign(request)
	if err != nil {
		return "", fmt.Errorf("error while check sign: %s", err.Error())
	}

	if needToCreateSign {
		UserUUID := GenerateUserUUID()
		sign, err := GenerateSign(UserUUID)
		if err != nil {
			return "", fmt.Errorf("cannot generate signed user UUID: %s", err.Error())
		} else {
			authCookie := http.Cookie{
				Name:  LoginCookieName,
				Value: sign,
			}

			request.AddCookie(&authCookie)
			http.SetCookie(writer, &authCookie)
			return UserUUID, nil
		}
	} else {
		cc, _ := request.Cookie(LoginCookieName)
		tk, _ := FetchUUID(cc.Value)
		return tk, nil
	}
}

func isNeedToCreateSign(request *http.Request) (bool, error) {
	cookie, err := request.Cookie(LoginCookieName)
	if err != nil && errors.Is(err, http.ErrNoCookie) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	_, err = FetchUUID(cookie.Value)
	if err == nil {
		return false, nil
	}

	return true, nil
}
