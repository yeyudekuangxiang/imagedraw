package main

import (
	"github.com/yeyudekuangxiang/imagedraw"
	"image"
	"image/draw"
	"log"
)

func main() {
	base, err := imagedraw.LoadImage("./ccc.jpg")
	if err != nil {
		log.Fatal(err)
	}

	circle := base.Circle(500, 500, 300).Saturation(100)
	err = circle.SaveAs("./circle.png")
	if err != nil {
		log.Fatal("保存图片失败", err)
	}

	cut := base.Cut(0, 0, 500, 500).Opacity(1).SetOp(draw.Over)
	err = cut.SaveAs("./cut.png")
	if err != nil {
		log.Fatal("保存图片失败", err)
	}

	ellipse := base.Ellipse(500, 500, 500, 400).Opacity(30)
	err = ellipse.SaveAs("./ellipse.png")
	if err != nil {
		log.Fatal("保存图片失败", err)
	}

	circle.SetArea(image.Rect(1000, 1000, 500, 500))
	circle.SetOp(draw.Over)
	cut.SetArea(image.Rect(100, 300, 2000, 2000))

	fillText := imagedraw.NewText("测试文字啊啊啊")
	fillText.SetFontSize(200)

	base = base.Fill(cut, circle, fillText)

	cut.SetArea(image.Rect(100, 300, 400, 700))
	base = base.Fill(base, cut)

	err = base.SaveAs("./base2.png")
	if err != nil {
		log.Fatal("保存图片失败", err)
	}
}
