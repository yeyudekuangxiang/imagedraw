package imagedraw

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/yeyudekuangxiang/imagedraw/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"math"
	"net/http"
)

var (
	SiYuanHeiYi     = func() IDrawString { return NewTTFDraw(mustLoadTTFBytes(fonts.SiYuanHeiTiTTF())) }
	SiYuanHeiYiBold = func() IDrawString { return NewOTFDraw(mustLoadOTFBytes(fonts.SiYuanHeiTiOTFBold())) }
)

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
func pxToPoint(px float64, dpi float64) float64 {
	return px * dpi / 72
}

type Text struct {
	d              IDrawString
	fontSize       int
	dpi            int
	textAlign      string
	area           image.Rectangle
	maxLineNum     int
	outStr         string
	outStrPosition string
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
		d:              SiYuanHeiYi(),
		fontSize:       24,
		dpi:            72,
		textAlign:      "left",
		color:          color.RGBA{A: 255},
		s:              s,
		autoLine:       true,
		outStrPosition: "right",
		overHidden:     true,
	}
}
func NewLineText(linesText []string) *Text {
	return &Text{
		d:         SiYuanHeiYi(),
		fontSize:  24,
		dpi:       72,
		textAlign: "left",
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
func (t *Text) SetFont(drawString IDrawString) *Text {
	t.d = drawString
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
func (t *Text) draw(dst draw.Image) (draw.Image, error) {
	t.d.SetColor(t.color)
	//设置字体大小
	t.d.SetSize(float64(t.fontSize))
	t.d.SetDpi(float64(t.dpi))
	face, err := t.d.Face()
	if err != nil {
		return nil, err
	}

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
		splitTextList, err = t.dealText(face, maxWidth, maxHeight, lineHeight)
	} else {
		//将一个字符串按最大绘制宽度分割成多行字体
		splitTextList, err = t.dealLineText(face, maxWidth, maxHeight, lineHeight)
	}
	if err != nil {
		return nil, err
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
		t.d.SetDot(freetype.Pt(area.Min.X+int(startX), area.Min.Y+i*lineHeight+deviation))
		err = t.d.DrawString(text.Str, dst)
		if err != nil {
			return nil, err
		}
	}
	return dst, nil
}
func (t *Text) Copy() *Text {
	return &Text{
		d:              t.d,
		fontSize:       t.fontSize,
		dpi:            t.dpi,
		textAlign:      t.textAlign,
		area:           t.area,
		maxLineNum:     t.maxLineNum,
		outStr:         t.outStr,
		outStrPosition: t.outStrPosition,
		color:          t.color,
		lineHeight:     t.lineHeight,
		s:              t.s,
		lines:          t.lines,
		autoLine:       t.autoLine,
		overHidden:     t.overHidden,
	}
}

type CalcTextResult struct {
	LineHeight    int
	MaxWidth      float64
	SplitTextList []SplitText
	Height        int
	Width         int
	RealWidth     int
}

func (t *Text) Deal() (*CalcTextResult, error) {

	//用于计算字体长度
	face, err := t.d.Face()
	if err != nil {
		return nil, err
	}
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
		splitTextList, err = t.dealText(face, maxWidth, maxHeight, lineHeight)
	} else {
		//将一个字符串按最大绘制宽度分割成多行字体
		splitTextList, err = t.dealLineText(face, maxWidth, maxHeight, lineHeight)
	}
	if err != nil {
		return nil, err
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

	return &CalcTextResult{
		LineHeight:    lineHeight,
		MaxWidth:      maxWidth,
		SplitTextList: splitTextList,
		Height:        height,
		Width:         int(width),
		RealWidth:     int(math.Ceil(realWidth)),
	}, nil
}

//转换自己设置的多行文本
func (t *Text) dealLineText(face font.Face, maxWidth, maxHeight float64, lineHeight int) ([]SplitText, error) {
	face, err := t.d.Face()
	if err != nil {
		return nil, err
	}

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
	return list, nil
}

//转换字符串
func (t Text) dealText(face font.Face, maxWidth, maxHeight float64, lineHeight int) ([]SplitText, error) {
	face, err := t.d.Face()
	if err != nil {
		return nil, err
	}
	if t.autoLine == false {
		return []SplitText{dealSingleOut(face, str2SplitText(face, t.s), t.outStr, t.outStrPosition, maxWidth)}, nil
	}
	maxLineNum := t.maxLineNum
	if t.overHidden && int(maxHeight/float64(lineHeight)) < maxLineNum {
		maxLineNum = int(maxHeight / float64(lineHeight))
	}
	return dealLineOut(face, splitText(face, t.s, maxWidth), t.outStr, t.outStrPosition, maxLineNum), nil
}

func (t *Text) Width() (int, error) {
	result, err := t.Deal()
	if err != nil {
		return 0, err
	}
	return result.Width, nil
}
func (t *Text) Height() (int, error) {
	result, err := t.Deal()
	if err != nil {
		return 0, err
	}
	return result.Height, nil
}

type IDrawString interface {
	SetSize(size float64)
	SetDpi(dpi float64)
	SetDot(p fixed.Point26_6)
	SetColor(c color.RGBA)
	DrawString(s string, dst draw.Image) error
	Face() (font.Face, error)
}
type OTFDraw struct {
	font     *opentype.Font
	color    color.RGBA
	fontSize float64
	dpi      float64
	dot      fixed.Point26_6
}

func (o *OTFDraw) SetColor(c color.RGBA) {
	o.color = c
}

func NewOTFDraw(font *opentype.Font) *OTFDraw {
	return &OTFDraw{
		font: font,
	}
}
func NewOTFDrawFromFile(path string) *OTFDraw {
	return &OTFDraw{
		font: MustLoadOTF(path),
	}
}
func NewOTFDrawFromHttp(url string) *OTFDraw {
	return &OTFDraw{
		font: MustLoadOTFHttp(url),
	}
}

func (o *OTFDraw) Face() (font.Face, error) {
	return o.face()
}
func (o *OTFDraw) face() (font.Face, error) {
	return opentype.NewFace(o.font, &opentype.FaceOptions{
		Size:    o.fontSize,
		DPI:     o.dpi,
		Hinting: font.HintingNone,
	})
}
func (o *OTFDraw) SetSize(size float64) {
	o.fontSize = size
}
func (o *OTFDraw) SetDpi(dpi float64) {
	o.dpi = dpi
}
func (o *OTFDraw) SetDot(p fixed.Point26_6) {
	o.dot = p
}
func (o *OTFDraw) DrawString(s string, dst draw.Image) error {
	face, err := o.face()
	if err != nil {
		return err
	}
	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(o.color),
		Face: face,
		Dot:  o.dot,
	}
	d.DrawString(s)
	return nil
}

type TTFDraw struct {
	font     *truetype.Font
	color    color.RGBA
	fontSize float64
	dpi      float64
	dot      fixed.Point26_6
}

func NewTTFDrawFromFile(path string) *TTFDraw {
	return &TTFDraw{
		font: MustLoadTTF(path),
	}
}
func NewTTFDrawFromHttp(url string) *TTFDraw {
	return &TTFDraw{
		font: MustLoadTTFHttp(url),
	}
}
func NewTTFDraw(font *truetype.Font) *TTFDraw {
	return &TTFDraw{
		font: font,
	}
}
func (t *TTFDraw) SetColor(c color.RGBA) {
	t.color = c
}
func (t *TTFDraw) SetSize(size float64) {
	t.fontSize = size
}
func (t *TTFDraw) SetDpi(dpi float64) {
	t.dpi = dpi
}
func (t *TTFDraw) SetDot(p fixed.Point26_6) {
	t.dot = p
}
func (t *TTFDraw) face() font.Face {
	return truetype.NewFace(t.font, &truetype.Options{
		Size:    t.fontSize,
		DPI:     t.dpi,
		Hinting: font.HintingNone,
	})
}
func (t *TTFDraw) DrawString(s string, dst draw.Image) error {
	ctx := t.context()
	ctx.SetDst(dst)
	ctx.SetClip(dst.Bounds())
	_, err := ctx.DrawString(s, t.dot)
	return err
}
func (t *TTFDraw) context() *freetype.Context {
	ctx := freetype.NewContext()
	//设置要绘制的图像
	ctx.SetSrc(image.NewUniform(t.color))
	//设置字体大小
	ctx.SetFontSize(pxToPoint(t.fontSize, t.dpi))
	//设置字体
	ctx.SetFont(t.font)
	//设置dpi
	ctx.SetDPI(t.dpi)
	return ctx
}
func (t *TTFDraw) Face() (font.Face, error) {
	return t.face(), nil
}

func LoadTTF(path string) (*truetype.Font, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parseFont, err := freetype.ParseFont(data)
	if err != nil {
		return nil, err
	}
	return parseFont, nil
}
func LoadTTFHttp(url string) (*truetype.Font, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	parseFont, err := freetype.ParseFont(data)
	if err != nil {
		return nil, err
	}
	return parseFont, nil
}
func MustLoadTTF(path string) *truetype.Font {
	f, err := LoadTTF(path)
	if err != nil {
		panic(f)
	}
	return f
}
func MustLoadTTFHttp(url string) *truetype.Font {
	f, err := LoadTTFHttp(url)
	if err != nil {
		panic(f)
	}
	return f
}
func mustLoadTTFBytes(data []byte) *truetype.Font {
	parseFont, err := freetype.ParseFont(data)
	if err != nil {
		panic(err)
	}
	return parseFont
}

func LoadOTF(path string) (*opentype.Font, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parseFont, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return parseFont, nil
}
func LoadOTFHttp(url string) (*opentype.Font, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	parseFont, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return parseFont, nil
}
func MustLoadOTF(path string) *opentype.Font {
	f, err := LoadOTF(path)
	if err != nil {
		panic(f)
	}
	return f
}
func MustLoadOTFHttp(url string) *opentype.Font {
	f, err := LoadOTFHttp(url)
	if err != nil {
		panic(f)
	}
	return f
}
func mustLoadOTFBytes(data []byte) *opentype.Font {
	parseFont, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}
	return parseFont
}
