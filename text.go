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

//带有字符串长度的结构体
type SplitText struct {
	//字符串
	Str string
	//字符串的长度 单位px
	Width float64
}

//将一个字符串根据字体和最大长度 分割成行
func splitText(face font.Face, s string, maxWidth float64) []SplitText {
	runeList := []rune(s)
	stList := make([]SplitText, 0)
	st := SplitText{}
	for i, r := range runeList {
		//返回字体宽度
		advance, _ := face.GlyphAdvance(r)
		w := float64(advance>>6) + float64(advance&(1<<6-1))*0.01
		if (st.Width + w) >= maxWidth {
			stList = append(stList, st)
			st = SplitText{
				Str:   string(r),
				Width: w,
			}
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

//像素转换成磅
func pxToPoint(px int, dpi int) float64 {
	return float64(px * dpi / 72)
}

type Text struct {
	c          *freetype.Context
	fontSize   int
	dpi        int
	textAlign  string
	area       image.Rectangle
	maxLineNum int
	outStr     string
	font       *truetype.Font
	color      color.RGBA
	lineHeight int
	s          string
}

func NewText(s string) *Text {
	return &Text{
		c:         freetype.NewContext(),
		fontSize:  24,
		dpi:       72,
		textAlign: "left",
		font:      DefaultFont,
		color:     color.RGBA{A: 255},
		s:         s,
	}
}

//设置字体对齐方式 left right center 默认left
func (t *Text) SetTextAlign(textAlign string) *Text {
	if textAlign != "left" && textAlign != "right" && textAlign != "center" {
		textAlign = "left"
	}
	t.textAlign = textAlign
	return t
}

//设置文本放置的区域 默认整个图片
func (t *Text) SetArea(pt image.Rectangle) *Text {
	t.area = pt
	return t
}

//设置字体大小 单位像素 默认24px
func (t *Text) SetFontSize(px int) *Text {
	t.fontSize = px
	return t
}

//设置dpi 默认72
func (t *Text) SetDpi(dpi int) *Text {
	t.dpi = dpi
	return t
}

//设置文本最大行数 默认不限制但超出区域返回后的行会截取掉
func (t *Text) SetMaxLineNum(num int) *Text {
	t.maxLineNum = num
	return t
}

//设置文本超出后提示符 如`...` 默认空字符串
func (t *Text) SetOutStr(str string) *Text {
	t.outStr = str
	return t
}

//设置字体 默认字体思源黑体
func (t *Text) SetFont(font *truetype.Font) *Text {
	t.font = font
	return t
}

//设置行高 默认字体高度
func (t *Text) SetLineHeight(h int) *Text {
	t.lineHeight = h
	return t
}

//设置字体颜色 默认
func (t *Text) SetColor(rgba color.RGBA) *Text {
	t.color = rgba
	return t
}

//黑色
func (t *Text) SetText(s string) *Text {
	t.s = s
	return t
}

//实现FillItem接口
func (t *Text) draw(dst draw.Image) draw.Image {
	//设置要绘制的图像
	t.c.SetDst(dst)
	//设置字体颜色???
	t.c.SetSrc(image.NewUniform(t.color))
	//设置字体大小
	t.c.SetFontSize(pxToPoint(t.fontSize, t.dpi))
	//设置字体
	t.c.SetFont(t.font)
	//设置dpi
	t.c.SetDPI(float64(t.dpi))
	//设置绘制返回
	t.c.SetClip(dst.Bounds())

	//用于计算字体长度
	face := truetype.NewFace(t.font, &truetype.Options{
		Size: float64(t.fontSize),
		DPI:  float64(t.dpi),
	})

	//用于存储绘制的最大长度
	var maxWidth float64

	//绘制区域
	area := t.area
	if area.Max.X == 0 && area.Max.Y == 0 {
		area.Max = dst.Bounds().Max
	}
	maxWidth = float64(area.Max.X - area.Min.X)

	//将一个字符串按最大绘制宽度分割成多行字体
	splitTextList := t.dealOut(face, splitText(face, t.s, maxWidth))

	//行高
	lineHeight := t.lineHeight
	if lineHeight <= 0 {
		lineHeight = t.fontSize
	}

	//绘制字体
	for i, text := range splitTextList {
		//计算相对于绘制区域开始绘制位置
		var startX float64
		switch t.textAlign {
		case "left":
			startX = 0
		case "center":
			startX = (maxWidth - text.Width) / 2
		case "right":
			startX = maxWidth - text.Width
		}
		//获取字体相对于原点的绘制位置
		bounds, _, _ := face.GlyphBounds([]rune(text.Str)[0])
		//计算偏移量
		deviation := lineHeight/2 - bounds.Max.Y.Round() + (bounds.Max.Y.Round()-bounds.Min.Y.Round())/2
		_, _ = t.c.DrawString(text.Str, freetype.Pt(area.Min.X+int(startX), area.Min.Y+i*lineHeight+deviation))
	}
	return dst
}

//处理超出展示
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
