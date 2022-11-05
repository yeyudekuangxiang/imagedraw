package imagedraw

import (
	"image/draw"
)

type FillItem interface {
	draw(img draw.Image) (draw.Image, error)
}

func fill(dst draw.Image, items ...FillItem) (draw.Image, error) {
	var err error
	for _, item := range items {
		dst, err = item.draw(dst)
		if err != nil {
			return nil, err
		}
	}
	return dst, nil
}
