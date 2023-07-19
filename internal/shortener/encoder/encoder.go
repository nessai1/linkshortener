package encoder

import (
	"github.com/speps/go-hashids"
)

const (
	HashMinLength = 5
)

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
