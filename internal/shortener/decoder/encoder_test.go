package encoder

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEncodeURL(t *testing.T) {
	tests := []struct {
		name         string
		firstURL     string
		secondURL    string
		hasCollision bool
	}{
		{
			name:         "Two different urls",
			firstURL:     "https://ya.ru",
			secondURL:    "https://vk.com",
			hasCollision: false,
		},
		{
			name:         "Two equal urls",
			firstURL:     "https://ya.ru",
			secondURL:    "https://ya.ru",
			hasCollision: true,
		},
		{
			name:         "Two urls with little difference",
			firstURL:     "https://ya.ru",
			secondURL:    "https://ya.ru/",
			hasCollision: false,
		},
		{
			name:         "Another two equal urls",
			firstURL:     "https://vkontakte.ru",
			secondURL:    "https://vkontakte.ru",
			hasCollision: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstToken, err := EncodeURL(tt.firstURL)
			if err != nil {
				require.NoErrorf(t, err, "First encode of %s has error %s", tt.firstURL, err.Error())
			}

			secondToken, err := EncodeURL(tt.secondURL)
			if err != nil {
				require.NoErrorf(t, err, "Second encode of %s has error %s", tt.secondURL, err.Error())
			}

			assert.Equal(t, tt.hasCollision, firstToken == secondToken, "Encode of %s and %s has collisions (%s)", tt.firstURL, tt.secondURL, secondToken)
		})
	}
}
