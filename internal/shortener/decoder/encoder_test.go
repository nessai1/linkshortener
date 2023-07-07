package encoder

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEncodeURL(t *testing.T) {
	tests := []struct {
		name         string
		firstUrl     string
		secondUrl    string
		hasCollision bool
	}{
		{
			name:         "Two different urls",
			firstUrl:     "https://ya.ru",
			secondUrl:    "https://vk.com",
			hasCollision: false,
		},
		{
			name:         "Two equal urls",
			firstUrl:     "https://ya.ru",
			secondUrl:    "https://ya.ru",
			hasCollision: true,
		},
		{
			name:         "Two urls with little difference",
			firstUrl:     "https://ya.ru",
			secondUrl:    "https://ya.ru/",
			hasCollision: false,
		},
		{
			name:         "Another two equal urls",
			firstUrl:     "https://vkontakte.ru",
			secondUrl:    "https://vkontakte.ru",
			hasCollision: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstToken, err := EncodeURL(tt.firstUrl)
			if err != nil {
				require.NoErrorf(t, err, "First encode of %s has error %s", tt.firstUrl, err.Error())
			}

			secondToken, err := EncodeURL(tt.secondUrl)
			if err != nil {
				require.NoErrorf(t, err, "Second encode of %s has error %s", tt.secondUrl, err.Error())
			}

			assert.Equal(t, tt.hasCollision, firstToken == secondToken, "Encode of %s and %s has collisions (%s)", tt.firstUrl, tt.secondUrl, secondToken)
		})
	}
}
