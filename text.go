package imagedraw

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"math"
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
	//字符串的高度 单位px
	Height     float64
	RealWidth  float64
	RealHeight float64
	//此行文本相对原点位置 最小y点
	MinY float64
	//此行文本相对原点位置 最大y点
	MaxY float64
}

func Int26ToFloat(d fixed.Int26_6) float64 {
	return float64(d>>6) + float64(d&(1<<6-1))*0.01
}

//将一个字符串根据字体和最大长度 分割成行
func splitText(face font.Face, s string, maxWidth float64) []SplitText {
	runeList := []rune(s)
	stList := make([]SplitText, 0)
	st := SplitText{
		MaxY: float64(1 - 1<<16),
		MinY: float64(1<<16 - 1),
	}
	for i, r := range runeList {
		//返回字体高度、宽度
		bounds, advance, _ := face.GlyphBounds(r)

		maxY := float64(bounds.Max.Y.Ceil())
		minY := float64(bounds.Min.Y.Ceil())
		width := float64(advance.Ceil())

		if (st.Width + width) > maxWidth {
			stList = append(stList, st)
			st = SplitText{
				Str:    string(r),
				Width:  width,
				Height: maxY - minY,
				MaxY:   maxY,
				MinY:   minY,
			}
		} else {
			st.Width += width
			st.RealWidth += Int26ToFloat(advance)
			st.Str += string(r)
			if st.MaxY < maxY {
				st.MaxY = maxY
			}
			if st.MinY > minY {
				st.MinY = minY
			}
			st.Height = st.MaxY - st.MinY
			st.RealHeight = st.MaxY - st.MinY
		}

		if i == len(runeList)-1 {
			stList = append(stList, st)
		}
	}
	return stList
}

//计算字符串的长度
func str2SplitText(face font.Face, str string) SplitText {
	runeList := []rune(str)
	st := SplitText{
		MaxY: float64(1 - 1<<16),
		MinY: float64(1<<16 - 1),
		Str:  str,
	}
	for _, r := range runeList {
		bounds, advance, _ := face.GlyphBounds(r)
		st.Width += float64(advance.Ceil())
		st.RealWidth += Int26ToFloat(advance)
		if maxY := float64(bounds.Max.Y.Ceil()); maxY > st.MaxY {
			st.MaxY = maxY
		}
		if minY := float64(bounds.Min.Y.Ceil()); minY < st.MinY {
			st.MinY = minY
		}
	}
	st.Height = st.MaxY - st.MinY
	st.RealHeight = st.MaxY - st.MinY
	return st
}

//处理多行超出提示
func dealLineOut(face font.Face, list []SplitText, outStr, outStrPosition string, maxLineNum int) []SplitText {
	if maxLineNum <= 0 || len(list) <= maxLineNum {
		return list
	}
	if len(outStr) == 0 {
		return list[:maxLineNum]
	}

	outStrRune := []rune(outStr)
	outStrWidth := 0.00
	MaxY := float64(1 - 1<<16)
	MinY := float64(1<<16 - 1)
	for _, s := range outStrRune {
		bounds, advance, _ := face.GlyphBounds(s)
		outStrWidth += float64(advance.Ceil())
		if maxY := float64(bounds.Max.Y.Ceil()); maxY > MaxY {
			MaxY = maxY
		}
		if minY := float64(bounds.Min.Y.Ceil()); minY < MinY {
			MinY = minY
		}
	}

	list2 := list[:maxLineNum]
	lastLine := list2[maxLineNum-1]
	lastRuneList := []rune(lastLine.Str)
	total := 0.00
	if outStrPosition == "left" {
		for i := len(lastRuneList) - 1; i >= 0; i-- {
			advance, _ := face.GlyphAdvance(lastRuneList[i])
			total += float64(advance.Ceil())
			if total >= outStrWidth {
				lastRuneList = append(lastRuneList[:i], outStrRune...)
				list2[maxLineNum-1].Str = string(lastRuneList)

				list2[maxLineNum-1].Width = list2[maxLineNum-1].Width - total + outStrWidth
				if list2[maxLineNum-1].MaxY < MaxY {
					list2[maxLineNum-1].MaxY = MaxY
				}
				if list2[maxLineNum-1].MinY > MinY {
					list2[maxLineNum-1].MinY = MinY
				}
				break
			}
		}
	} else {
		for i := len(lastRuneList) - 1; i >= 0; i-- {
			advance, _ := face.GlyphAdvance(lastRuneList[i])
			total += float64(advance.Ceil())
			if total >= outStrWidth {
				lastRuneList = append(lastRuneList[:i], outStrRune...)
				list2[maxLineNum-1].Str = string(lastRuneList)

				list2[maxLineNum-1].Width = list2[maxLineNum-1].Width - total + outStrWidth
				if list2[maxLineNum-1].MaxY < MaxY {
					list2[maxLineNum-1].MaxY = MaxY
				}
				if list2[maxLineNum-1].MinY > MinY {
					list2[maxLineNum-1].MinY = MinY
				}
				break
			}
		}
	}
	return list2
}

//处理单行文本超出提示
func dealSingleOut(face font.Face, str SplitText, outStr, outStrPosition string, maxWidth float64) SplitText {
	if str.Width <= maxWidth {
		return str
	}

	outStrRune := []rune(outStr)
	outStrWidth := 0.00
	MaxY := float64(1 - 1<<16)
	MinY := float64(1<<16 - 1)
	for _, s := range outStrRune {
		bounds, advance, _ := face.GlyphBounds(s)
		outStrWidth += float64(advance.Ceil())
		if maxY := float64(bounds.Max.Y.Ceil()); maxY > MaxY {
			MaxY = maxY
		}
		if minY := float64(bounds.Min.Y.Ceil()); minY < MinY {
			MinY = minY
		}
	}

	strRuneList := []rune(str.Str)
	total := 0.00
	if outStrPosition == "left" {
		for i := len(strRuneList) - 1; i >= 0; i-- {
			advance, _ := face.GlyphAdvance(strRuneList[i])
			w := float64(advance.Ceil())

			if total+w+outStrWidth > maxWidth {
				if str.MaxY > MaxY {
					MaxY = str.MaxY
				}
				if str.MinY < MinY {
					MinY = str.MinY
				}
				return SplitText{
					Str:   outStr + string(strRuneList[i+1:]),
					Width: total + outStrWidth,
					MinY:  MinY,
					MaxY:  MaxY,
				}
			}
			total += w
		}
	}
	for i := 0; i < len(strRuneList); i++ {
		advance, _ := face.GlyphAdvance(strRuneList[i])
		w := float64(advance.Ceil())

		if total+w+outStrWidth > maxWidth {
			if str.MaxY > MaxY {
				MaxY = str.MaxY
			}
			if str.MinY < MinY {
				MinY = str.MinY
			}
			return SplitText{
				Str:   string(strRuneList[:i]) + outStr,
				Width: total + outStrWidth,
				MinY:  MinY,
				MaxY:  MaxY,
			}
		}
		total += w
	}
	return str
}

//像素转换成磅
func pxToPoint(px int, dpi int) float64 {
	return float64(px * dpi / 72)
}

type Text struct {
	c              *freetype.Context
	fontSize       int
	dpi            int
	textAlign      string
	area           image.Rectangle
	maxLineNum     int
	outStr         string
	outStrPosition string
	font           *truetype.Font
	color          color.RGBA
	lineHeight     int
	overHidden     bool
	s              string
	//是否自动分行
	autoLine bool
	lines    []string
}

func NewText(s string) *Text {
	return &Text{
		c:              freetype.NewContext(),
		fontSize:       24,
		dpi:            72,
		textAlign:      "left",
		font:           DefaultFont,
		color:          color.RGBA{A: 255},
		s:              s,
		autoLine:       true,
		outStrPosition: "right",
		overHidden:     true,
	}
}
func NewLineText(linesText []string) *Text {
	return &Text{
		c:         freetype.NewContext(),
		fontSize:  24,
		dpi:       72,
		textAlign: "left",
		font:      DefaultFont,
		color:     color.RGBA{A: 255},
		lines:     linesText,
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

//设置文本放置的区域 默认整个图片 (x,y)起点坐标 w宽度 h长度
func (t *Text) SetArea(x, y, w, h int) *Text {
	t.area = image.Rect(x, y, x+w, y+h)
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

//设置字符串省略的位置 left左边 right右边 默认右边
func (t *Text) SetOutStrPosition(position string) *Text {
	if position == "left" || position == "right" {
		t.outStrPosition = position
	} else {
		t.outStrPosition = "right"
	}
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

//设置字体颜色 默认黑色
func (t *Text) SetColor(rgba color.RGBA) *Text {
	t.color = rgba
	return t
}

//设置字符串
func (t *Text) SetText(s string) *Text {
	t.s = s
	t.lines = nil
	return t
}

//设置字符串文本是否自动分行 autoLine文本是否自动分行 false不自动分行 true自动分行 默认true
func (t *Text) SetAutoLine(autoLine bool) *Text {
	t.autoLine = autoLine
	return t
}

//自定义多行字符串 每行字符串超出后都会按照 outStr处理
func (t *Text) SetLineText(lines []string) {
	t.lines = lines
	t.s = ""
}

func (t *Text) SetOverHidden(overHidden bool) *Text {
	t.overHidden = overHidden
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

	//行高
	lineHeight := t.lineHeight
	if lineHeight <= 0 {
		lineHeight = t.fontSize
	}

	//绘制区域
	area := t.area
	if area.Max.X == 0 && area.Max.Y == 0 {
		area.Max = dst.Bounds().Max
	}
	maxWidth := float64(area.Max.X - area.Min.X)
	maxHeight := float64(area.Max.Y - area.Min.Y)

	var splitTextList []SplitText
	if t.s != "" {
		//将一个字符串按最大绘制宽度分割成多行字体
		splitTextList = t.dealText(face, maxWidth, maxHeight, lineHeight)
	} else {
		//将一个字符串按最大绘制宽度分割成多行字体
		splitTextList = t.dealLineText(face, maxWidth, maxHeight, lineHeight)
	}

	//绘制字体
	for i, text := range splitTextList {
		//计算相对于绘制区域开始绘制位置
		var startX float64
		switch t.textAlign {
		case "left":
			startX = 0
		case "center":
			startX = (maxWidth - text.RealWidth) / 2
		case "right":
			startX = maxWidth - text.RealWidth
		}
		//计算偏移量
		deviation := int(float64(lineHeight)/2 - text.MaxY + (text.MaxY-text.MinY)/2)
		_, _ = t.c.DrawString(text.Str, freetype.Pt(area.Min.X+int(startX), area.Min.Y+i*lineHeight+deviation))
	}
	return dst
}
func (t *Text) Copy() *Text {
	return &Text{
		c:              t.c,
		fontSize:       t.fontSize,
		dpi:            t.dpi,
		textAlign:      t.textAlign,
		area:           t.area,
		maxLineNum:     t.maxLineNum,
		outStr:         t.outStr,
		outStrPosition: t.outStrPosition,
		font:           t.font,
		color:          t.color,
		lineHeight:     t.lineHeight,
		s:              t.s,
		lines:          t.lines,
		autoLine:       t.autoLine,
		overHidden:     t.overHidden,
	}
}
func (t *Text) Deal() struct {
	LineHeight    int
	MaxWidth      float64
	SplitTextList []SplitText
	Height        int
	Width         int
	RealWidth     int
} {

	//用于计算字体长度
	face := truetype.NewFace(t.font, &truetype.Options{
		Size: float64(t.fontSize),
		DPI:  float64(t.dpi),
	})

	//行高
	lineHeight := t.lineHeight
	if lineHeight <= 0 {
		lineHeight = t.fontSize
	}

	//用于存储绘制的最大长度
	var maxWidth float64

	//绘制区域
	area := t.area
	maxWidth = float64(area.Max.X - area.Min.X)
	maxHeight := float64(area.Max.Y - area.Min.Y)

	var splitTextList []SplitText
	if t.s != "" {
		//将一个字符串按最大绘制宽度分割成多行字体
		splitTextList = t.dealText(face, maxWidth, maxHeight, lineHeight)
	} else {
		//将一个字符串按最大绘制宽度分割成多行字体
		splitTextList = t.dealLineText(face, maxWidth, maxHeight, lineHeight)
	}

	width := 0.00
	realWidth := 0.00
	height := 0
	for _, item := range splitTextList {
		if width < item.Width {
			width = item.Width
		}
		if realWidth < item.RealWidth {
			realWidth = item.RealWidth
		}
		height += lineHeight
	}

	return struct {
		LineHeight    int
		MaxWidth      float64
		SplitTextList []SplitText
		Height        int
		Width         int
		RealWidth     int
	}{
		LineHeight:    lineHeight,
		MaxWidth:      maxWidth,
		SplitTextList: splitTextList,
		Height:        height,
		Width:         int(width),
		RealWidth:     int(math.Ceil(realWidth)),
	}
}

//转换自己设置的多行文本
func (t *Text) dealLineText(face font.Face, maxWidth, maxHeight float64, lineHeight int) []SplitText {
	list := make([]SplitText, 0)
	var lines []string

	maxLineNum := t.maxLineNum
	if t.overHidden && int(maxHeight/float64(lineHeight)) < maxLineNum {
		maxLineNum = int(maxHeight / float64(lineHeight))
	}

	if len(t.lines) > maxLineNum {
		lines = t.lines[:maxLineNum]
	}

	for _, line := range lines {
		list = append(list, dealSingleOut(face, str2SplitText(face, line), t.outStr, t.outStrPosition, maxWidth))
	}
	return list
}

//转换字符串
func (t Text) dealText(face font.Face, maxWidth, maxHeight float64, lineHeight int) []SplitText {
	if t.autoLine == false {
		return []SplitText{dealSingleOut(face, str2SplitText(face, t.s), t.outStr, t.outStrPosition, maxWidth)}
	}
	maxLineNum := t.maxLineNum
	if t.overHidden && int(maxHeight/float64(lineHeight)) < maxLineNum {
		maxLineNum = int(maxHeight / float64(lineHeight))
	}
	return dealLineOut(face, splitText(face, t.s, maxWidth), t.outStr, t.outStrPosition, maxLineNum)
}

func (t *Text) Width() int {
	return t.Deal().Width
}
func (t *Text) Height() int {
	return t.Deal().Height
}
