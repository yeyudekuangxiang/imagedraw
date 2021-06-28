package imagedraw

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"net/http"
)

//默认字体
var DefaultFont *truetype.Font

func init() {
	//加载字体
	resp, err := http.Get("https://cdn.laoyaojing.net/standard/resource/siyuanheiti.ttf")
	if err != nil {
		log.Println("加载字体失败", err)
		return
	}
	defer resp.Body.Close()

	fontData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("加载字体失败", err)
		return
	}
	parseFont, err := freetype.ParseFont(fontData)
	if err != nil {
		log.Println("加载字体失败", err)
	}
	DefaultFont = parseFont
}

type SplitText struct {
	Str   string
	Width float64
}

func splitText(face font.Face, s string, maxWidth float64) []SplitText {
	runeList := []rune(s)
	st := SplitText{}
	stList := make([]SplitText, 0)
	for i, r := range runeList {
		advance, _ := face.GlyphAdvance(r)
		w := float64(int(advance)>>6) + float64(int(advance)&(1<<6-1))*0.01
		if (st.Width + w) >= maxWidth {
			stList = append(stList, st)
			st = SplitText{}
		} else if i == len(runeList)-1 {
			st.Width += w
			st.Str += string(r)
			stList = append(stList, st)
			st = SplitText{}
		} else {
			st.Width += w
			st.Str += string(r)
		}
	}
	return stList
}
func pxToPoint(px int, dpi int) float64 {
	return float64(px * dpi / 72)
}

type Text struct {
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

func NewText(s string) *Text {
	return &Text{
		c:        freetype.NewContext(),
		color:    color.RGBA{A: 255},
		fontSize: 24,
		font:     DefaultFont,
		dpi:      72,
		s:        s,
	}
}
func (t *Text) SetTextAlign(textAlign string) *Text {
	t.textAlign = textAlign
	return t
}
func (t *Text) SetArea(pt image.Rectangle) *Text {
	t.area = &pt
	return t
}
func (t *Text) SetFontSize(px int) *Text {
	t.fontSize = px
	return t
}
func (t *Text) SetDpi(dpi int) *Text {
	t.dpi = dpi
	return t
}
func (t *Text) SetMaxLineNum(num int) *Text {
	t.maxLineNum = num
	return t
}
func (t *Text) SetOutStr(str string) *Text {
	t.outStr = str
	return t
}
func (t *Text) SetFont(font *truetype.Font) *Text {
	t.font = font
	return t
}
func (t *Text) LineHeight(h int) *Text {
	t.lineHeight = h
	return t
}
func (t *Text) SetColor(rgba color.RGBA) *Text {
	t.color = rgba
	return t
}
func (t *Text) SetText(s string) *Text {
	t.s = s
	return t
}
func (t *Text) draw(dst draw.Image) draw.Image {
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
func (t Text) dealOut(face font.Face, list []SplitText) []SplitText {
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
