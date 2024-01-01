package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// LoginCookieName название куки в которой хранится подписанный пользовательский UUID
const LoginCookieName = "LINKSHORTER_USER"

// ContextUserUUIDKey ключ контекста запроса в которой хранится пользовательский UUID
const ContextUserUUIDKey ContextAuthKey = "UserUUIDKey"

const (
	signSecret = "unkTpTVMUcHQmgWXADBpcueGlcBAVXHAD2zUAuQCzD0MOKYhcg6Cvjrarl9RMmDUXRZuQz36S8Hs0Ak3OgkQy8vweiYtF2NaVV3qZLDvKYd75zaU1InkwRUEHUj01gkbSItyLh5V2eLO7lHAmpTYQ7N0CjOElRKeTIe23HEC4rAfDAavOLKATqrMKJnCzQvLNSaMPhzXpo9MzbHHfbPImn6tmVQiK9h63tKSQx3Dz0Mj2A8NHef3cvCEHC"
	tokenTTL   = time.Hour * 1
)

type (
	ContextAuthKey string
	UserUUID       string
)

type claims struct {
	jwt.RegisteredClaims
	UserUUID string
}

func generateUserUUID() string {
	return uuid.New().String()
}

func generateSign(UUID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		},
		UserUUID: UUID,
	})

	tokenString, err := token.SignedString([]byte(signSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// FetchUUID get user UUID from signed string, or get error if sign is wrong
func FetchUUID(sign string) (UserUUID, error) {
	claims := &claims{}
	_, err := jwt.ParseWithClaims(sign, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(signSecret), nil
	})

	if err != nil {
		return "", err
	}

	return UserUUID(claims.UserUUID), nil
}

// authorize user if uuid expired or does not exist
func authorize(writer http.ResponseWriter, request *http.Request) (UserUUID, error) {
	needToCreateSign, err := isNeedToCreateSign(request)
	if err != nil {
		return "", fmt.Errorf("error while check sign: %s", err.Error())
	}

	if needToCreateSign {
		userUUID := generateUserUUID()
		sign, err := generateSign(userUUID)
		if err != nil {
			return "", fmt.Errorf("cannot generate signed user UUID: %s", err.Error())
		} else {
			authCookie := http.Cookie{
				Name:  LoginCookieName,
				Value: sign,
			}

			request.AddCookie(&authCookie)
			http.SetCookie(writer, &authCookie)
			return UserUUID(userUUID), nil
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

// GetAuthMiddleware является middleware который добавляет поддержку авторизации UUID пользователя, делающего запрос на сервис
func GetAuthMiddleware(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			cookie, err := request.Cookie(LoginCookieName)
			if errors.Is(err, http.ErrNoCookie) {
				logger.Debug("User sends request without login cookies")
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			userUUID, err := FetchUUID(cookie.Value)
			if err != nil {
				logger.Info("User sends invalid cookies", zap.Error(err))
				c := &http.Cookie{
					Name:   LoginCookieName,
					Value:  "",
					MaxAge: -1,
				}
				http.SetCookie(writer, c)
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(request.Context(), ContextUserUUIDKey, userUUID)
			request = request.WithContext(ctx)

			logger.Debug("User successful authorized", zap.String("User UUID", string(userUUID)))
			next.ServeHTTP(writer, request)
		})
	}
}

// GetAuthMiddleware является middleware который добавляет поддержку регистрации нового UUID пользователя, делающего запрос на сервис
func GetRegisterMiddleware(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			userUUID, err := authorize(writer, request)
			if err != nil {
				logger.Error("Error while register user", zap.Error(err))
			}

			ctx := context.WithValue(request.Context(), ContextUserUUIDKey, userUUID)
			request = request.WithContext(ctx)

			next.ServeHTTP(writer, request)
		})
	}
}
