package imagedraw

import (
	"image/draw"
)

type FillItem interface {
	draw(img draw.Image) draw.Image
}

func fill(dst draw.Image, items ...FillItem) draw.Image {
	for _, item := range items {
		dst = item.draw(dst)
	}
	return dst
}
