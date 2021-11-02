package main

import (
	"github.com/yeyudekuangxiang/imagedraw"
	"log"
)

func main() {
	//从本地文件读取图片
	loadImage, err := imagedraw.LoadImage("./image.jpg")
	if err != nil {
		log.Fatal(err)
	}

	//尺寸调整
	resizeImage := loadImage.Resize(192, 108)
	//图片剪切
	cutImage := loadImage.Cut(800, 200, 200, 200)
	//圆形剪切
	circleImage := loadImage.Circle(850, 300, 100)
	//椭圆剪切
	ellipseImage := loadImage.Ellipse(850, 300, 75, 100)
	//色调、饱和度、亮度、不透明度调整
	colorImage := loadImage.Cut(800, 200, 200, 200).Hue(50).Saturation(30).Brightness(100).Opacity(50)

	//图片绘制
	resizeImage.SetArea(0, 0, resizeImage.Width(), resizeImage.Height())
	cutImage.SetArea(0, 200, cutImage.Width(), cutImage.Height())
	circleImage.SetArea(0, 400, circleImage.Width(), circleImage.Height())
	ellipseImage.SetArea(0, 600, ellipseImage.Width(), ellipseImage.Height())
	colorImage.SetArea(0, 800, colorImage.Width(), colorImage.Height())
	loadImage.Fill(resizeImage, cutImage, circleImage, ellipseImage, colorImage)

	//文字自动分行
	fillText := imagedraw.NewText("夜雨寄北(李商隐)君问归期未有期，巴山夜雨涨秋池。何当共剪西窗烛，却话巴山夜雨时。")
	fillText.SetFontSize(60)
	fillText.SetLineHeight(90)
	fillText.SetTextAlign("left")
	fillText.SetArea(1300, 0, 500, 1000)
	loadImage.Fill(fillText)

	//自定义分行
	fillText2 := imagedraw.NewLineText([]string{
		"夜雨寄北",
		"李商隐",
		"君问归期未有期",
		"巴山夜雨涨秋池",
		"何当共剪西窗烛",
		"却话巴山夜雨时",
	})
	fillText2.SetMaxLineNum(3)
	fillText2.SetTextAlign("right")
	fillText2.SetFontSize(60)
	fillText2.SetLineHeight(90)
	fillText2.SetArea(1300, 500, 500, 1000)
	loadImage.Fill(fillText2)

	err = loadImage.SaveAs("image2.png")
	if err != nil {
		log.Fatal("保存图片失败", err)
	}
}
