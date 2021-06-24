package imagedraw

import (
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

func LoadImage(path string) (draw.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	return ConvertImage(img), nil
}
func LoadImageFromUrl(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//解码背景图
	img, _, err := image.Decode(resp.Body)
	return img, nil
}
func Circle(img image.Image, x, y, r int) image.Image {
	circle := image.NewRGBA(image.Rect(0, 0, 2*r, 2*r))
	for w := 0; w <= 2*r; w++ {
		for h := 0; h <= 2*r; h++ {
			x1 := x - r + w
			y1 := y - r + h
			if (math.Pow(float64(w-r), 2) + math.Pow(float64(h-r), 2)) <= math.Pow(float64(r), 2) {
				circle.Set(w, h, img.At(x1, y1))
			}
		}
	}
	return circle
}
func Cut(img image.Image, x, y, x1, y1 int) image.Image {
	cutImage := image.NewRGBA(image.Rect(0, 0, x1-x, y1-y))
	for dx := x; dx <= x1; dx++ {
		for dy := y; dy <= y1; dy++ {
			cutImage.Set(dx-x, dy-y, img.At(dx, dy))
		}
	}
	return cutImage
}
func ConvertImage(img image.Image) draw.Image {
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, dst.Bounds(), img, image.Pt(0, 0), draw.Src)
	return dst
}
func SaveAs(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	switch filepath.Ext(path) {
	case ".png":
		return png.Encode(file, img)
	case ".jpg":
		return jpeg.Encode(file, img, nil)
	}
	return errors.New("ext not support")
}
func Resize(img image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	bounds := img.Bounds()
	imgW := bounds.Max.X - bounds.Min.X
	imgH := bounds.Max.Y - bounds.Min.Y
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			x1 := int(math.Round(float64((x * imgW) / w)))
			y1 := int(math.Round(float64((y * imgH) / h)))
			dst.Set(x, y, img.At(bounds.Min.X+x1, bounds.Min.Y+y1))
		}
	}
	return dst
}
