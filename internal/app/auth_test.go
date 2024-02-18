package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestGenerateUserUUID(t *testing.T) {
	uuid := GenerateUserUUID()
	rg := regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")
	assert.True(t, rg.Match([]byte(uuid)))
}

func TestGenerateSignAndFetch(t *testing.T) {
	uuid := GenerateUserUUID()

	sign, err := generateSign(uuid)
	assert.NoError(t, err)

	fetchedUUID, err := FetchUUID(sign)
	require.NoError(t, err)
	assert.Equal(t, uuid, string(fetchedUUID))

	fetchedUUID, err = FetchUUID("someShitString")

	assert.Error(t, err)
	assert.Equal(t, "", string(fetchedUUID))
}

func TestIsNeedToCreateSign(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/anymethod", nil)

	assert.True(t, isNeedToCreateSign(request))

	cookie := http.Cookie{
		Name:  LoginCookieName,
		Value: "someShit",
	}
	request.AddCookie(&cookie)

	assert.True(t, isNeedToCreateSign(request))

	validSign, _ := generateSign(GenerateUserUUID())
	cookie = http.Cookie{
		Name:  LoginCookieName,
		Value: validSign,
	}

	request = httptest.NewRequest(http.MethodGet, "/anymethod", nil)
	request.AddCookie(&cookie)

	assert.False(t, isNeedToCreateSign(request))
}

func TestAuthorize(t *testing.T) {
	tests := []struct {
		name string
		sign string
		uuid string
	}{
		{
			name: "Valid UUID",
			uuid: GenerateUserUUID(),
		},
		{
			name: "Invalid sign",
			sign: "someinvalidsign",
		},
		{
			name: "No auth data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "https://test.com", nil)
			w := httptest.NewRecorder()

			var cookie *http.Cookie
			if tt.sign != "" {
				cookie = &http.Cookie{
					Name:  LoginCookieName,
					Value: tt.sign,
				}
			} else if tt.uuid != "" {
				validSign, err := generateSign(tt.uuid)
				require.NoError(t, err)
				cookie = &http.Cookie{
					Name:  LoginCookieName,
					Value: validSign,
				}
			}

			if cookie != nil {
				r.AddCookie(cookie)
			}

			uuid, err := authorize(w, r)
			require.NoError(t, err)
			if tt.uuid != "" && tt.sign == "" {
				assert.Equal(t, tt.uuid, string(uuid))
			} else {
				assert.NotEmpty(t, string(uuid))
			}
		})
	}
}

func TestGetAuthMiddleware(t *testing.T) {
	userUUID := GenerateUserUUID()
	validSign, _ := generateSign(userUUID)
	tests := []struct {
		name string

		cookie       *http.Cookie
		userUUID     string
		isAuthorized bool
	}{
		{
			name: "No cookie",

			cookie:       nil,
			isAuthorized: false,
		},
		{
			name: "Valid cookie",

			cookie: &http.Cookie{
				Name:  LoginCookieName,
				Value: validSign,
			},
			userUUID:     userUUID,
			isAuthorized: true,
		},
		{
			name: "Invalid cookie",

			cookie: &http.Cookie{
				Name:  LoginCookieName,
				Value: "someshitsign",
			},
			isAuthorized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				require.True(t, tt.isAuthorized)
				userUUIDCtxValue := request.Context().Value(ContextUserUUIDKey)
				assert.Equal(t, UserUUID(tt.userUUID), userUUIDCtxValue)
				writer.WriteHeader(http.StatusOK)
			})

			handlerToTest := GetAuthMiddleware(zap.NewNop())(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/some/authorized/page", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			recorder := httptest.NewRecorder()
			handlerToTest.ServeHTTP(recorder, req)

			if tt.isAuthorized {
				assert.Equal(t, http.StatusOK, recorder.Code)
			} else {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			}
		})
	}
}

func TestGetRegisterMiddleware(t *testing.T) {

	userUUID := GenerateUserUUID()
	sign, _ := generateSign(userUUID)

	tests := []struct {
		name string

		cookie   *http.Cookie
		userUUID string
	}{
		{
			name: "User have UUID",

			cookie: &http.Cookie{
				Name:  LoginCookieName,
				Value: sign,
			},
			userUUID: userUUID,
		},
		{
			name: "Invalid sign",
			cookie: &http.Cookie{
				Name:  LoginCookieName,
				Value: "someshitsign",
			},
		},
		{
			name: "No cookie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				userUUIDCtxValue := request.Context().Value(ContextUserUUIDKey)
				if tt.userUUID != "" {
					assert.Equal(t, UserUUID(tt.userUUID), userUUIDCtxValue)
				} else {
					assert.NotEmpty(t, userUUIDCtxValue)
				}

				writer.WriteHeader(http.StatusOK)
			})

			handlerToTest := GetRegisterMiddleware(zap.NewNop())(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/some/authorized/page", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			recorder := httptest.NewRecorder()
			handlerToTest.ServeHTTP(recorder, req)

			require.Equal(t, http.StatusOK, recorder.Code)
		})
	}
}
