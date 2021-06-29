package imagedraw

import (
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

// 从本地路径读取图片 返回draw.Image
func loadImage(path string) (draw.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return convertImage(img), nil
}

// 从网络路径读取图片 返回draw.Image
func loadImageFromUrl(url string) (draw.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//解码背景图
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return convertImage(img), nil
}

// 从reader中读取图片 返回draw.Image
func loadImageFromReader(reader io.Reader) (draw.Image, error) {
	//解码背景图
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return convertImage(img), nil
}

// 截取圆形
func circle(img image.Image, x, y, r int) draw.Image {
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

//截取椭圆
func ellipse(img image.Image, x, y, w, h int) draw.Image {
	ellipse := image.NewRGBA(image.Rect(0, 0, 2*w, 2*h))
	if w > h {
		f := math.Sqrt(float64(w*w - h*h))
		f1 := float64(x) - f //f1 y
		f2 := float64(x) + f //f2 y

		//fmt.Println(f1,f2)
		for x1 := 0; x1 <= 2*w; x1++ {
			for y1 := 0; y1 <= 2*h; y1++ {
				px := x - w + x1
				py := y - h + y1
				length := math.Sqrt(math.Pow(float64(px)-f1, 2)+math.Pow(float64(py-y), 2)) + math.Sqrt(math.Pow(float64(px)-f2, 2)+math.Pow(float64(py-y), 2))
				if length <= float64(2*w) {
					ellipse.Set(x1, y1, img.At(px, py))
				}
			}
		}
	} else {
		f := math.Sqrt(float64(h*h - w*w))
		f1 := float64(y) - f //x f1
		f2 := float64(y) + f //x f2
		for x1 := 0; x1 <= 2*w; x1++ {
			for y1 := 0; y1 <= 2*h; y1++ {
				px := x - w + x1
				py := y - h + y1
				length := math.Sqrt(math.Pow(float64(px-x), 2)+math.Pow(float64(py)-f1, 2)) + math.Sqrt(math.Pow(float64(px-x), 2)+math.Pow(float64(py)-f2, 2))
				if length <= float64(2*h) {
					ellipse.Set(x1, y1, img.At(px, py))
				}
			}
		}
	}
	return ellipse
}

//截取方形
func cut(img image.Image, x, y, x1, y1 int) draw.Image {
	cutImage := image.NewRGBA(image.Rect(0, 0, x1-x, y1-y))
	for dx := x; dx <= x1; dx++ {
		for dy := y; dy <= y1; dy++ {
			cutImage.Set(dx-x, dy-y, img.At(dx, dy))
		}
	}
	return cutImage
}

//将img image.Image转化为draw.Image
func convertImage(img image.Image) draw.Image {
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, dst.Bounds(), img, image.Pt(0, 0), draw.Src)
	return dst
}

//将图片保存在本地
func saveAs(img image.Image, path string) error {
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

//调整图片尺寸
func resize(img image.Image, w, h int) draw.Image {
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

func LoadImage(path string) (*Image, error) {
	img, err := loadImage(path)
	if err != nil {
		return nil, err
	}
	return NewImage(img), nil
}
func LoadImageFromUrl(url string) (*Image, error) {
	img, err := loadImageFromUrl(url)
	if err != nil {
		return nil, err
	}
	return NewImage(img), nil
}
func LoadImageFromReader(reader io.Reader) (*Image, error) {
	img, err := loadImageFromReader(reader)
	if err != nil {
		return nil, err
	}
	return NewImage(img), nil
}

//创建一个图片操作对象
func NewImage(img draw.Image) *Image {
	return &Image{
		img: img,
		op:  draw.Src,
	}
}

//创建一个空白的图片
func NewBaseImage(width, height int) *Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	return NewImage(img)
}

//图片操作对象
type Image struct {
	area image.Rectangle
	img  draw.Image
	op   draw.Op
}

//设置本图片覆盖到另外一张图片的尺寸
func (i *Image) SetArea(pt image.Rectangle) *Image {
	i.area = pt
	return i
}

//设置图片填充方式
func (i *Image) SetOp(op draw.Op) *Image {
	i.op = op
	return i
}
func (i *Image) draw(dst draw.Image) draw.Image {
	resizeImage := resize(i.img, i.area.Max.X-i.area.Min.X, i.area.Max.Y-i.area.Min.Y)
	draw.Draw(dst, i.area, resizeImage, image.Pt(0, 0), i.op)
	return dst
}

//截取圆形并且返回一个新的对象
func (i *Image) Circle(x, y, r int) *Image {
	return NewImage(circle(i.img, x, y, r))
}

//剪切图片并且返回一个新的对象
func (i *Image) Cut(x, y, x1, y1 int) *Image {
	return NewImage(cut(i.img, x, y, x1, y1))
}

//调整图片大小并且返回一个新的对象
func (i Image) Resize(w, h int) *Image {
	return NewImage(resize(i.img, w, h))
}

//将其他元素填充进本图片
func (i *Image) Fill(item ...FillItem) *Image {
	return NewImage(fill(i.img, item...))
}

//将图片保存在本地
func (i *Image) SaveAs(path string) error {
	return saveAs(i.img, path)
}

//截取椭圆并返回一个新的对象 x,y中心点位置 w横半轴长度 h竖半轴长度
func (i *Image) Ellipse(x, y, w, h int) *Image {
	return NewImage(ellipse(i.img, x, y, w, h))
}
