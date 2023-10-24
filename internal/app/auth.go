package app

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
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

func getAuthMiddleware(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			needToCreateSign, err := isNeedToCreateSign(request)
			if err != nil {
				logger.Error(fmt.Sprintf("Error while check sign: %s", err.Error()))
			}

			if needToCreateSign {
				UserUUID := GenerateUserUUID()
				logger.Info(fmt.Sprintf("Create new UUID: %s", UserUUID))
				sign, err := GenerateSign(UserUUID)
				if err != nil {
					logger.Error(fmt.Sprintf("Cannot generate signed user UUID: %s", err.Error()))
				} else {
					authCookie := http.Cookie{
						Name:  LoginCookieName,
						Value: sign,
					}

					request.AddCookie(&authCookie)
					http.SetCookie(writer, &authCookie)
				}
			} else {
				cc, _ := request.Cookie(LoginCookieName)
				tk, _ := FetchUUID(cc.Value)
				logger.Info(fmt.Sprintf("fetch already exists UUID: %s", tk))
			}

			next.ServeHTTP(writer, request)
		})
	}
}
