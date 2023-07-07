package encoder

import (
	"github.com/speps/go-hashids"
)

func EncodeURL(url string) (string, error) {
	hd := hashids.NewData()
	hd.Salt = url
	hd.MinLength = 5
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}
	e, err := h.Encode([]int{45, 434, 1313, 99})
	if err != nil {
		return "", err
	}

	return e, nil
}
