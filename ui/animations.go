package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func showAnimation(myWindow fyne.Window, nextScreen func()) {
	var frames []*canvas.Image
	for i := 1; i <= 42; i++ {
		resourceName := fmt.Sprintf("frame_apngframe%d_png", i)
		frameResource := Resources[resourceName]
		if frameResource == nil {
			panic(fmt.Sprintf("Resource %s not found", resourceName))
		}

		frame := canvas.NewImageFromResource(frameResource)
		frame.FillMode = canvas.ImageFillContain
		frames = append(frames, frame)
	}

	animationContainer := container.NewStack(frames[0])
	myWindow.SetContent(animationContainer)
	myWindow.Resize(fyne.NewSize(900, 600))
	myWindow.CenterOnScreen()
	myWindow.Show()

	go func() {
		for _, frame := range frames {
			animationContainer.Objects = []fyne.CanvasObject{frame}
			animationContainer.Refresh()
			time.Sleep(45 * time.Millisecond)
		}
		nextScreen()
	}()
}
