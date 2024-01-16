package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestGenerateUserUUID(t *testing.T) {
	uuid := generateUserUUID()
	rg := regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")
	assert.True(t, rg.Match([]byte(uuid)))
}

func TestGenerateSignAndFetch(t *testing.T) {
	uuid := generateUserUUID()

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

	validSign, _ := generateSign(generateUserUUID())
	cookie = http.Cookie{
		Name:  LoginCookieName,
		Value: validSign,
	}

	request = httptest.NewRequest(http.MethodGet, "/anymethod", nil)
	request.AddCookie(&cookie)

	assert.False(t, isNeedToCreateSign(request))
}
