package ui

import (
	"database/sql"
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func makeMainScreen(dbConn *sql.DB, myWindow fyne.Window, vaultPath, username string) fyne.CanvasObject {

	logo := canvas.NewImageFromResource(Resources["logoText_png"])
	logo.SetMinSize(fyne.NewSize(150, 200))
	logo.FillMode = canvas.ImageFillContain

	vaultStatus := canvas.NewText("Vault Status: Unlocked", color.RGBA{R: 0, G: 128, B: 0, A: 255})
	vaultStatus.TextStyle = fyne.TextStyle{Bold: true}
	vaultStatusContainer := container.NewHBox(container.NewPadded(vaultStatus))

	usernameLabel := widget.NewLabelWithStyle(fmt.Sprintf("Logged in as: %s", username), fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	usernameLabel.Alignment = fyne.TextAlignTrailing
	usernameLabelContainer := container.NewHBox(layout.NewSpacer(), usernameLabel)

	fileEntry := widget.NewEntry()
	fileEntry.SetPlaceHolder("Enter file path...")

	selectFileButton := widget.NewButton("Select File", func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err == nil && uri != nil {
				fileEntry.SetText(uri.URI().Path())
			}
		}, myWindow)
	})

	addFileButton := widget.NewButton("Add File", func() {
		filePath := fileEntry.Text
		data, err := os.ReadFile(filePath)
		if err != nil {
			showErrorNotification(err.Error())
			return
		}

		err = currentVault.AddFile(filepath.Base(filePath), data, vaultKey)
		if err != nil {
			showErrorNotification(err.Error())
			return
		}

		err = currentVault.Save(vaultPath)
		if err != nil {
			showErrorNotification(err.Error())
			return
		}

		showSuccessNotification("File added successfully")
	})

	viewFilesButton := widget.NewButton("View Files", func() {
		showFilesWindow(vaultPath)
	})

	logoutButton := widget.NewButton("Logout", func() {
		currentVault = nil
		vaultKey = nil
		loginScreen := makeLoginScreen(dbConn, myWindow)
		myWindow.SetContent(loginScreen)
	})

	inputContainer := container.NewVBox(
		fileEntry,
		container.NewGridWithColumns(2,
			selectFileButton,
			addFileButton,
		),
	)

	buttonContainer := container.NewVBox(
		viewFilesButton,
		logoutButton,
	)

	form := container.NewVBox(
		logo,
		inputContainer,
		buttonContainer,
	)

	header := container.NewGridWithColumns(3,
		vaultStatusContainer,
		layout.NewSpacer(),
		usernameLabelContainer,
	)

	return container.NewVBox(
		header,
		container.NewGridWithColumns(3,
			layout.NewSpacer(),
			form,
			layout.NewSpacer(),
		),
	)
}
