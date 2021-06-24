package main

import (
	"github.com/yeyudekuangxiang/imagedraw"
	"image"
	"log"
)

func main() {
	base, err := imagedraw.LoadImage("./ccc.jpg")
	if err != nil {
		log.Fatal("加载图片失败", err)
	}

	fillImage := imagedraw.NewFillImage(base)
	circle := fillImage.Circle(500, 500, 300)
	err = imagedraw.SaveAs(circle.GetImage(), "./circle.png")
	if err != nil {
		log.Fatal("保存图片失败", err)
	}
	cut := fillImage.Cut(0, 0, 500, 500)
	err = imagedraw.SaveAs(cut.GetImage(), "./cut.png")
	if err != nil {
		log.Fatal("保存图片失败", err)
	}

	circle.SetArea(image.Rect(1000, 1000, 500, 500))
	cut.SetArea(image.Rect(100, 300, 2000, 2000))

	fillText := imagedraw.NewFillText("测试文字啊啊啊")
	fillText.SetFontSize(200)

	base2 := imagedraw.Fill(base, cut, circle, fillText)
	err = imagedraw.SaveAs(base2, "./base2.png")
	if err != nil {
		log.Fatal("保存图片失败", err)
	}
}
