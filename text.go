package imagedraw

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
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
