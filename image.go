package imagedraw

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

//hsv颜色模型
type Hsv struct {
	h float64
	s float64
	v float64
	a uint8 //用于暂存rgba中的a值
}

//HSV转RGBA
func (hsv Hsv) ToRGBA() color.RGBA {
	return hsv2RGBA(hsv.h, hsv.s, hsv.v, hsv.a)
}

//色度  -100到100 0不变
func (hsv *Hsv) Hue(h float64) {
	if h < -100 {
		h = -100
	}
	if h > 100 {
		h = 100
	}

	hsv.h *= 1 + h/100
	if hsv.h > 360 {
		hsv.h = 360
	}
}

//饱和度  -100到100 0不变
func (hsv *Hsv) Saturation(s float64) {
	if s < -100 {
		s = -100
	}
	if s > 100 {
		s = 100
	}
	hsv.s *= 1 + s/100
	if hsv.s > 1 {
		hsv.s = 1
	}
}

//明度  -100到100 0不变
func (hsv *Hsv) Value(v float64) {
	if v < -100 {
		v = 100
	}
	if v > 100 {
		v = 100
	}
	hsv.v *= 1 + v/100
	if hsv.v > 1 {
		hsv.v = 1
	}
}

//color.Color转HSV对象
func Color2HSV(c color.Color) Hsv {
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	r := rgba.R
	g := rgba.G
	b := rgba.B

	var h, s, v float64
	r1 := float64(r) / 255
	g1 := float64(g) / 255
	b1 := float64(b) / 255
	maxc := max(r1, g1, b1)
	minc := min(r1, g1, b1)
	v = maxc
	if minc == maxc {
		return Hsv{
			0.0, 0.0, v, rgba.A,
		}
	}
	s = (maxc - minc) / maxc
	rc := (maxc - r1) / (maxc - minc)
	gc := (maxc - g1) / (maxc - minc)
	bc := (maxc - b1) / (maxc - minc)
	if r1 == maxc {
		h = bc - gc
	} else if g1 == maxc {
		h = 2.0 + rc - bc
	} else {
		h = 4.0 + gc - rc
	}
	h0 := h / 6.0
	h = h0 - float64(int(h0))
	return Hsv{
		h * 360, s, v, rgba.A,
	}
}

//hsv转RGBA
func hsv2RGBA(h, s, v float64, alpha uint8) color.RGBA {
	h /= 360
	if s == 0.0 {
		return color.RGBA{
			R: uint8(v * 255),
			G: uint8(v * 255),
			B: uint8(v * 255),
			A: alpha,
		}
	}
	i := int(h * 6.0)
	f := (h * 6.0) - float64(i)
	p := v * (1.0 - s)
	q := v * (1.0 - s*f)
	t := v * (1.0 - s*(1.0-f))
	i %= 6
	switch i {
	case 0:
		return color.RGBA{
			R: uint8(q * 255), G: uint8(v * 255), B: uint8(p * 255), A: alpha,
		}
	case 1:
		return color.RGBA{
			R: uint8(q * 255), G: uint8(v * 255), B: uint8(p * 255), A: alpha,
		}

	case 2:
		return color.RGBA{
			R: uint8(p * 255), G: uint8(v * 255), B: uint8(t * 255), A: alpha,
		}
	case 3:
		return color.RGBA{
			R: uint8(p * 255), G: uint8(q * 255), B: uint8(v * 255), A: alpha,
		}
	case 4:
		return color.RGBA{
			R: uint8(t * 255), G: uint8(p * 255), B: uint8(v * 255), A: alpha,
		}
	case 5:
		return color.RGBA{
			R: uint8(v * 255), G: uint8(p * 255), B: uint8(q * 255), A: alpha,
		}
	default:
		return color.RGBA{
			R: 0, G: 0, B: 0, A: alpha,
		}
	}
}
func max(items ...float64) float64 {
	if len(items) == 0 {
		return 0
	}
	maxItem := items[0]
	for _, item := range items {
		if item > maxItem {
			maxItem = item
		}
	}
	return maxItem
}
func min(items ...float64) float64 {
	if len(items) == 0 {
		return 0
	}
	minItem := items[0]
	for _, item := range items {
		if item < minItem {
			minItem = item
		}
	}
	return minItem
}

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
	return ellipse(img, x, y, r, r)
	/*circle := image.NewRGBA(image.Rect(0, 0, 2*r, 2*r))
	for w := 0; w < 2*r; w++ {
		for h := 0; h < 2*r; h++ {
			x1 := x - r + w
			y1 := y - r + h

			if math.Pow(float64(w-r), 2)+math.Pow(float64(h-r), 2) <= math.Pow(float64(r),2) {
				circle.Set(w, h, img.At(x1, y1))
			}
		}
	}
	return circle*/
}

func antiAliasing2(img draw.Image, dx int) draw.Image {
	img = resize(img, img.Bounds().Max.X*dx, img.Bounds().Max.Y*dx, BilinearInterpolation)
	newImage := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y))
	for x := 0; x < img.Bounds().Max.X*dx; x += dx {
		for y := 0; y < img.Bounds().Max.Y*dx; y += dx {
			var r, g, b, a uint32
			for ddx := 0; ddx < dx; ddx++ {
				for ddy := 0; ddy < dx; ddy++ {
					r1, g1, b1, a1 := img.At(x+ddx, y+ddy).RGBA()
					r += r1
					g += g1
					b += b1
					a += a1
				}
			}
			aaa := uint32(math.Pow(float64(dx), 2))
			c := color.NRGBA64{
				R: uint16(r / aaa),
				G: uint16(g / aaa),
				B: uint16(b / aaa),
				A: uint16(a / aaa),
			}
			newImage.Set(x/dx, y/dx, c)
		}
	}
	return newImage
}

//抗锯齿???
func antiAliasing(img draw.Image, dx int) draw.Image {
	newImage := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y))
	resizeImage := resize(img, img.Bounds().Max.X*dx, img.Bounds().Max.Y*dx, BilinearInterpolation)
	for x := 0; x < resizeImage.Bounds().Max.X; x += dx {
		for y := 0; y < resizeImage.Bounds().Max.Y; y += dx {
			var r, g, b, a uint32
			for ddx := 0; ddx < dx; ddx++ {
				for ddy := 0; ddy < dx; ddy++ {
					r1, g1, b1, a1 := resizeImage.At(x+ddx, y+ddy).RGBA()
					r += r1
					g += g1
					b += b1
					a += a1
				}
			}
			aaa := uint32(math.Pow(float64(dx), 2))
			c := color.NRGBA64{
				R: uint16(r / aaa),
				G: uint16(g / aaa),
				B: uint16(b / aaa),
				A: uint16(a / aaa),
			}
			newImage.Set(x/dx, y/dx, c)
		}
	}
	return newImage
}

//抗锯齿???
func (i *Image) AntiAliasing(dx int) *Image {
	return NewImage(antiAliasing(i.img, dx))
}

//截取椭圆
func ellipse(img image.Image, x, y, w, h int) draw.Image {
	ellipse := image.NewRGBA(image.Rect(0, 0, 2*w, 2*h))
	if w > h {
		f := math.Sqrt(float64(w*w - h*h))
		f1 := float64(x) - f //f1 y
		f2 := float64(x) + f //f2 y

		for x1 := 0; x1 < 2*w; x1++ {
			for y1 := 0; y1 < 2*h; y1++ {
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
		for x1 := 0; x1 < 2*w; x1++ {
			for y1 := 0; y1 < 2*h; y1++ {
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
	for dx := x; dx < x1; dx++ {
		for dy := y; dy < y1; dy++ {
			cutImage.Set(dx-x, dy-y, img.At(dx, dy))
		}
	}
	return cutImage
}

//将img image.Image转化为draw.Image
func convertImage(img image.Image) draw.Image {
	newImg := image.NewRGBA(img.Bounds())
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			newImg.Set(x, y, img.At(x, y))
		}
	}
	return newImg
}

//将图片保存在本地
func saveAs(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return saveWriter(img, filepath.Ext(path)[1:], file)
}
func saveWriter(img image.Image, ext string, writer io.Writer) error {
	switch ext {
	case "png":
		return png.Encode(writer, img)
	case "jpg":
		return jpeg.Encode(writer, img, nil)
	}
	return errors.New("ext not support")
}

type ResizeType int

const (
	//三次卷积插值缩放
	CubicConvolution ResizeType = iota
	//双线性插值缩放
	BilinearInterpolation ResizeType = iota
)

//调整图片尺寸
func resize(img image.Image, w, h int, resizeType ResizeType) draw.Image {
	switch resizeType {
	case CubicConvolution:
		return cubicConvolution(img, w, h)
	case BilinearInterpolation:
		return bilinearInterpolation(img, w, h)
	}
	return nil
}

//双线性插值
func bilinearInterpolation(img image.Image, w, h int) draw.Image {
	newImage := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			xf := float64(x*img.Bounds().Max.X) / float64(w)
			yf := float64(y*img.Bounds().Max.Y) / float64(h)
			i := math.Floor(xf)
			j := math.Floor(yf)
			u := xf - i
			v := yf - j
			r1, g1, b1, a1 := img.At(int(i), int(j)).RGBA()
			r2, g2, b2, a2 := img.At(int(i), int(j)+1).RGBA()
			r3, g3, b3, a3 := img.At(int(i)+1, int(j)).RGBA()
			r4, g4, b4, a4 := img.At(int(i)+1, int(j)+1).RGBA()
			r := (1-u)*(1-v)*float64(r1) + (1-u)*v*float64(r2) + u*(1-v)*float64(r3) + u*v*float64(r4)
			g := (1-u)*(1-v)*float64(g1) + (1-u)*v*float64(g2) + u*(1-v)*float64(g3) + u*v*float64(g4)
			b := (1-u)*(1-v)*float64(b1) + (1-u)*v*float64(b2) + u*(1-v)*float64(b3) + u*v*float64(b4)
			a := (1-u)*(1-v)*float64(a1) + (1-u)*v*float64(a2) + u*(1-v)*float64(a3) + u*v*float64(a4)
			newImage.Set(x, y, color.RGBA64{
				R: uint16(r), G: uint16(g), B: uint16(b), A: uint16(a),
			})
		}
	}
	return newImage
}

//三次卷积插值
func cubicConvolution(img image.Image, w, h int) draw.Image {
	newImage := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			xf := float64(x*img.Bounds().Max.X) / float64(w)
			yf := float64(y*img.Bounds().Max.Y) / float64(h)
			i := math.Floor(xf)
			j := math.Floor(yf)
			u := xf - i
			v := yf - j
			A := [][]float64{
				{
					cubicConvolutionS(1 + v),
					cubicConvolutionS(v),
					cubicConvolutionS(1 - v),
					cubicConvolutionS(2 - v),
				},
			}
			C := [][]float64{
				{cubicConvolutionS(1 + u)},
				{cubicConvolutionS(u)},
				{cubicConvolutionS(1 - u)},
				{cubicConvolutionS(2 - u)},
			}
			BR := make([][]float64, 0)
			BG := make([][]float64, 0)
			BB := make([][]float64, 0)
			BA := make([][]float64, 0)
			for m := -1; m < 3; m++ {
				rb := make([]float64, 0)
				gb := make([]float64, 0)
				bb := make([]float64, 0)
				ab := make([]float64, 0)
				for n := -1; n < 3; n++ {
					r, g, b, a := img.At(int(i)+m, int(j)+n).RGBA()
					rb = append(rb, float64(r))
					gb = append(gb, float64(g))
					bb = append(bb, float64(b))
					ab = append(ab, float64(a))
				}
				BR = append(BR, rb)
				BG = append(BG, gb)
				BB = append(BB, bb)
				BA = append(BA, ab)
			}
			r := matrixMultiplication(matrixMultiplication(A, BR), C)
			g := matrixMultiplication(matrixMultiplication(A, BG), C)
			b := matrixMultiplication(matrixMultiplication(A, BB), C)
			a := matrixMultiplication(matrixMultiplication(A, BA), C)

			if len(r) > 0 && len(r[0]) > 0 && len(g) > 0 && len(g[0]) > 0 && len(b) > 0 && len(b[0]) > 0 && len(a) > 0 && len(a[0]) > 0 {
				newImage.Set(x, y, color.RGBA64{
					R: floatToUnit16(r[0][0]), G: floatToUnit16(g[0][0]), B: floatToUnit16(b[0][0]), A: floatToUnit16(a[0][0]),
				})
			}
		}
	}
	return newImage
}

//三次卷积插值采样公式
func cubicConvolutionS(x float64) float64 {
	//当a取不同值时可以用来逼近不同的样条函数（常用值-0.5, -0.75）
	a := -1.0
	x = math.Abs(x)
	if x >= 2 {
		return 0
	} else if x >= 1 && x < 2 {
		return -4*a + 8*a*x - 5*a*math.Pow(x, 2) + a*math.Pow(x, 3)
	} else {
		return 1 - (a+3)*math.Pow(x, 2) + (a+2)*math.Pow(x, 3)
	}
}
func floatToUnit16(a float64) uint16 {
	return uint16(a)
}

//矩阵相乘
func matrixMultiplication(a [][]float64, b [][]float64) [][]float64 {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	//a的行数
	rowA := len(a)
	//a的列数
	colA := len(a[0])
	//b的行数
	rowB := len(b)
	//b的列数
	colB := len(b[0])

	if colA != rowB {
		return nil
	}

	mm := make([][]float64, 0)

	//循环a的行
	for i := 0; i < rowA; i++ {
		//检测a矩阵每一行的列数是否都是colA
		if len(a[i]) != colA {
			return nil
		}
		rowList := make([]float64, 0)
		//循环b的列
		for n := 0; n < colB; n++ {
			sum := 0.0
			//a[i][k] * b[k][n]
			for k := 0; k < colA; k++ {
				//检测b的列
				if len(b[k]) != colB {
					return nil
				}
				sum += a[i][k] * b[k][n]
			}
			rowList = append(rowList, sum)
		}
		mm = append(mm, rowList)
	}
	return mm
}

//设置不透明度 0-100 100为完全不透明 0为完全透明
func opacity(img image.Image, transparency uint32) draw.Image {
	if transparency < 0 {
		transparency = 0
	}
	if transparency > 100 {
		transparency = 100
	}
	opacity := image.NewRGBA(img.Bounds())
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			r, g, b, a := img.At(x, y).RGBA()
			opacity.Set(x, y, opacity.ColorModel().Convert(color.NRGBA64{
				R: uint16(r),
				G: uint16(g),
				B: uint16(b),
				A: uint16(a * transparency / 100),
			}))
		}
	}
	return opacity
}

//色度 -100到100 0不变
func hue(img image.Image, h float64) draw.Image {
	hue := image.NewRGBA(img.Bounds())
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			hsv := Color2HSV(img.At(x, y))
			hsv.Hue(h)
			hue.Set(x, y, hsv.ToRGBA())
		}
	}
	return hue
}

//饱和度  -100到100 0不变
func saturation(img image.Image, s float64) draw.Image {
	saturation := image.NewRGBA(img.Bounds())
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			hsv := Color2HSV(img.At(x, y))
			hsv.Saturation(s)
			saturation.Set(x, y, hsv.ToRGBA())
		}
	}
	return saturation
}

//亮度  -100到100 0不变
func brightness(img image.Image, v float64) draw.Image {
	brightness := image.NewRGBA(img.Bounds())
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			hsv := Color2HSV(img.At(x, y))
			hsv.Value(v)
			brightness.Set(x, y, hsv.ToRGBA())
		}
	}
	return brightness
}

//从本地读取图片
func LoadImage(path string) (*Image, error) {
	img, err := loadImage(path)
	if err != nil {
		return nil, err
	}
	return NewImage(img), nil
}

//从http链接读取图片
func LoadImageFromUrl(url string) (*Image, error) {
	img, err := loadImageFromUrl(url)
	if err != nil {
		return nil, err
	}
	return NewImage(img), nil
}

//从reader中读取图片
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
		op:  draw.Over,
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

//设置绘制到另外一张图片上时 在另外一张图片上的范围 x,y开始坐标 w宽度 h长度
func (i *Image) SetArea(x, y, w, h int) *Image {
	i.area = image.Rect(x, y, x+w, y+h)
	return i
}

//设置绘制到另外一张图片上时的填充方式
func (i *Image) SetOp(op draw.Op) *Image {
	i.op = op
	return i
}

//实现FillItem接口
func (i *Image) draw(dst draw.Image) (draw.Image, error) {
	//resizeImage := resize(i.img, i.area.Max.X-i.area.Min.X, i.area.Max.Y-i.area.Min.Y, BilinearInterpolation)
	draw.Draw(dst, i.area, i.img, image.Pt(0, 0), i.op)
	return dst, nil
}

//截取圆形并且返回一个新的对象 (x,y)原点坐标 r圆半径长度
func (i *Image) Circle(x, y, r int) *Image {
	return NewImage(circle(i.img, x, y, r))
}

//剪切图片并且返回一个新的对象 (x,y)剪切开始坐标点 w剪切宽度 h剪切长度
func (i *Image) Cut(x, y, w, h int) *Image {
	return NewImage(cut(i.img, x, y, x+w, y+h))
}

//调整图片大小并且返回一个新的对象 w宽度 h高度
func (i Image) Resize(w, h int, resizeType ...ResizeType) *Image {
	if len(resizeType) == 0 {
		resizeType = append(resizeType, BilinearInterpolation)
	}
	return NewImage(resize(i.img, w, h, resizeType[0]))
}

//将其他元素填充进本图片
func (i *Image) Fill(item ...FillItem) (*Image, error) {
	img, err := fill(i.img, item...)
	if err != nil {
		return nil, err
	}
	i.img = img
	return i, nil
}

//将图片保存在本地
func (i *Image) SaveAs(path string) error {
	return saveAs(i.img, path)
}

//截取椭圆并返回一个新的对象 (x,y)中心点位置 w横半轴长度 h竖半轴长度
func (i *Image) Ellipse(x, y, w, h int) *Image {
	return NewImage(ellipse(i.img, x, y, w, h))
}

//设置不透明度 0-100 100为完全不透明 0为完全透明
func (i *Image) Opacity(transparency uint32) *Image {
	return NewImage(opacity(i.img, transparency))
}

//色度 -100到100 0不变
func (i *Image) Hue(h float64) *Image {
	return NewImage(hue(i.img, h))
}

//饱和度  -100到100 0不变
func (i *Image) Saturation(s float64) *Image {
	return NewImage(saturation(i.img, s))
}

//亮度  -100到100 0不变
func (i *Image) Brightness(v float64) *Image {
	return NewImage(brightness(i.img, v))
}

//返回图片宽度
func (i *Image) Width() int {
	return i.img.Bounds().Max.X - i.img.Bounds().Min.X
}

//返回图片长度
func (i *Image) Height() int {
	return i.img.Bounds().Max.Y - i.img.Bounds().Min.Y
}

//返回 draw.image对象
func (i *Image) Image() draw.Image {
	return i.img
}

//将图片数据写进io.Writer  ext 图片格式 支持png jpg
func (i *Image) Encode(writer io.Writer, ext string) error {
	return saveWriter(i.img, ext, writer)
}

//创建一个副本
func (i *Image) Copy() *Image {
	copyImg := *i
	copyImg.img = convertImage(i.img)
	return &copyImg
}
