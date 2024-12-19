package ui

import (
	"database/sql"
	"path/filepath"
	"secure-file-vault/db"
	"secure-file-vault/vault"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func makeLoginScreen(dbConn *sql.DB, myWindow fyne.Window) fyne.CanvasObject {

	logo := canvas.NewImageFromResource(Resources["logoText_png"])
	logo.SetMinSize(fyne.NewSize(300, 200))
	logo.FillMode = canvas.ImageFillContain

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginButton := widget.NewButton("Login", func() {
		username := usernameEntry.Text
		password := passwordEntry.Text

		err := db.AuthenticateUser(dbConn, username, password)
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: err.Error(),
			})
			return
		}

		vaultPath := filepath.Join("vaults", username, "vault.dat")

		vlt, key, err := vault.OpenVault(vaultPath, password)
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: err.Error(),
			})
			return
		}
		currentVault = vlt
		vaultKey = key
		mainScreen := makeMainScreen(dbConn, myWindow, username)
		myWindow.SetContent(mainScreen)
	})
	loginButton.Resize(fyne.NewSize(200, 40))

	registerLink := widget.NewHyperlink("Not a member? Register here", nil)
	registerLink.OnTapped = func() {
		myWindow.SetContent(makeRegisterScreen(dbConn, myWindow))
	}

	inputContainer := container.NewVBox(
		container.NewPadded(usernameEntry),
		container.NewPadded(passwordEntry),
		container.NewPadded(loginButton),
	)

	return container.NewGridWithColumns(3,
		layout.NewSpacer(),
		container.NewVBox(logo, inputContainer, registerLink),
		layout.NewSpacer(),
	)
}
