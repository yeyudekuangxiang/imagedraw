package imagedraw

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"image"
	"image/color"
	"image/draw"
)

type FillItem interface {
	draw(img draw.Image) draw.Image
}

func Fill(dst draw.Image, items ...FillItem) image.Image {
	for _, item := range items {
		dst = item.draw(dst)
	}
	return dst
}

type FillText struct {
	fontSize   int
	dpi        int
	textAlign  string
	area       *image.Rectangle
	maxLineNum int
	outStr     string
	font       *truetype.Font
	color      color.RGBA
	c          *freetype.Context
	lineHeight int
	s          string
}

func NewFillText(s string) *FillText {
	return &FillText{
		c:        freetype.NewContext(),
		color:    color.RGBA{A: 255},
		fontSize: 24,
		font:     DefaultFont,
		dpi:      72,
		s:        s,
	}
}
func (t *FillText) SetTextAlign(textAlign string) {
	t.textAlign = textAlign
}
func (t *FillText) SetArea(pt image.Rectangle) {
	t.area = &pt
}
func (t *FillText) SetFontSize(px int) {
	t.fontSize = px
}
func (t *FillText) SetDpi(dpi int) {
	t.dpi = dpi
}
func (t *FillText) SetMaxLineNum(num int) {
	t.maxLineNum = num
}
func (t *FillText) SetOutStr(str string) {
	t.outStr = str
}
func (t *FillText) SetFont(font *truetype.Font) {
	t.font = font
}
func (t *FillText) LineHeight(h int) {
	t.lineHeight = h
}
func (t *FillText) SetColor(rgba color.RGBA) {
	t.color = rgba
}
func (t *FillText) SetText(s string) {
	t.s = s
}
func (t *FillText) draw(dst draw.Image) draw.Image {
	t.c.SetDst(dst)
	t.c.SetSrc(image.NewUniform(t.color))
	t.c.SetFontSize(pxToPoint(t.fontSize, t.dpi))
	t.c.SetFont(t.font)
	t.c.SetDPI(float64(t.dpi))
	t.c.SetClip(dst.Bounds())
	face := truetype.NewFace(t.font, &truetype.Options{
		Size: float64(t.fontSize),
		DPI:  float64(t.dpi),
	})
	var maxWidth float64
	var area image.Rectangle
	if t.area == nil {
		area = dst.Bounds()
	} else {
		area = *t.area
	}
	maxWidth = float64(area.Max.X - area.Min.X)
	splitTextList := t.dealOut(face, splitText(face, t.s, maxWidth))
	lineHeight := t.lineHeight
	if lineHeight <= 0 {
		lineHeight = t.fontSize
	}
	for i, text := range splitTextList {
		startX := (maxWidth - text.Width) / 2
		_, _ = t.c.DrawString(text.Str, freetype.Pt(area.Min.X+int(startX), area.Min.Y+(i+1)*lineHeight))
	}
	return dst
}
func (t FillText) dealOut(face font.Face, list []SplitText) []SplitText {
	if t.maxLineNum <= 0 || len(list) <= t.maxLineNum {
		return list
	}
	if len(t.outStr) == 0 {
		return list[:t.maxLineNum]
	}

	list2 := list[:t.maxLineNum]

	lastLine := list2[t.maxLineNum-1]
	lastRuneList := []rune(lastLine.Str)
	total := 0.00

	outStrRune := []rune(t.outStr)
	outStrWidth := 0.00
	for _, s := range outStrRune {
		advance, _ := face.GlyphAdvance(s)
		outStrWidth += float64(int(advance)>>6) + float64(int(advance)&(1<<6-1))*0.01
	}

	for i := len(lastRuneList) - 1; i >= 0; i-- {
		advance, _ := face.GlyphAdvance(lastRuneList[i])
		total += float64(int(advance)>>6) + float64(int(advance)&(1<<6-1))*0.01
		if total >= outStrWidth {
			lastRuneList = append(lastRuneList[:i], outStrRune...)
			list2[t.maxLineNum-1].Str = string(lastRuneList)

			list2[t.maxLineNum-1].Width = list2[t.maxLineNum-1].Width - total + outStrWidth
			break
		}
	}
	return list2
}

func NewFillImage(img image.Image) *FillImage {
	return &FillImage{
		img: img,
		op:  draw.Src,
	}
}

type FillImage struct {
	area image.Rectangle
	img  image.Image
	op   draw.Op
}

func (i *FillImage) GetImage() image.Image {
	return i.img
}
func (i *FillImage) SetImage(img image.Image) {
	i.img = img
}
func (i *FillImage) SetArea(pt image.Rectangle) {
	i.area = pt
}
func (i *FillImage) SetOp(op draw.Op) {
	i.op = op
}
func (i *FillImage) draw(dst draw.Image) draw.Image {
	resizeImage := Resize(i.img, i.area.Max.X-i.area.Min.X, i.area.Max.Y-i.area.Min.Y)
	draw.Draw(dst, i.area, resizeImage, image.Pt(0, 0), i.op)
	return dst
}
func (i *FillImage) Circle(x, y, r int) *FillImage {
	circle := Circle(i.img, x, y, r)
	newImage := FillImage{}
	newImage.SetImage(circle)
	return &newImage
}
func (i *FillImage) Cut(x, y, x1, y1 int) *FillImage {
	cutImage := Cut(i.img, x, y, x1, y1)
	newImage := FillImage{}
	newImage.SetImage(cutImage)
	return &newImage
}
func (i FillImage) Resize(w, h int) *FillImage {
	resizeImage := Resize(i.img, w, h)
	newImage := FillImage{}
	newImage.SetImage(resizeImage)
	return &newImage
}
