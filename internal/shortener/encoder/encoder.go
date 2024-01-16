package encoder

import (
	"github.com/speps/go-hashids"
)

// HashMinLength минимальная длина хеша от ссылки
const HashMinLength = 5

// EncodeURL получает из переданной ссылки хеш длинной >= HashMinLength
func EncodeURL(url string) (string, error) {
	hd := hashids.NewData()
	hd.Salt = url
	hd.MinLength = HashMinLength
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}
	e, err := h.Encode([]int{1, 2, 42, 3})
	if err != nil {
		return "", err
	}

	return e, nil
}
