package ui

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"secure-file-vault/db"
	"secure-file-vault/vault"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func makeRegisterScreen(dbConn *sql.DB, myWindow fyne.Window) fyne.CanvasObject {

	logo := canvas.NewImageFromResource(Resources["logoText_png"])
	logo.SetMinSize(fyne.NewSize(300, 200))
	logo.FillMode = canvas.ImageFillContain

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	vaultPathEntry := widget.NewEntry()
	vaultPathEntry.SetPlaceHolder("Vault Path (empty for default)")
	vaultPathEntry.SetText("")

	selectPathButton := widget.NewButton("Select Custom Vault Path", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				vaultPath := filepath.Join(uri.Path(), usernameEntry.Text+"_vault.dat")

				if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
					showErrorNotification("Vault already exists")
					return
				}

				vaultPathEntry.SetText(vaultPath)
			}
		}, myWindow)
	})

	registerButton := widget.NewButton("Register", func() {
		username := usernameEntry.Text
		password := passwordEntry.Text
		vaultPath := vaultPathEntry.Text

		if vaultPath == "" {
			vaultPath = filepath.Join("vaults", username, "vault.dat")
		}

		if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
			showErrorNotification("Vault already exists")
			return
		}

		vaultDir := filepath.Dir(vaultPath)
		err := os.MkdirAll(vaultDir, os.ModePerm)
		if err != nil {
			showErrorNotification(fmt.Sprintf("Failed to create directory: %v", err))
			return
		}

		err = db.RegisterUser(dbConn, username, password, vaultPath)
		if err != nil {
			showErrorNotification(fmt.Sprintf("Failed to register user: %v", err))
			return
		}

		_, err = vault.CreateVault(vaultPath, password)
		if err != nil {
			showErrorNotification(fmt.Sprintf("Failed to create vault: %v", err))
			return
		}
		vlt, key, err := vault.OpenVault(vaultPath, password)
		if err != nil {
			showErrorNotification(fmt.Sprintf("Failed to open vault: %v", err))
			return
		}

		currentVault = vlt
		vaultKey = key

		showSuccessNotification("User registered successfully")
		myWindow.SetContent(makeMainScreen(dbConn, myWindow, vaultPath, username))
	})

	loginLink := widget.NewHyperlink("Already a member? Login here.", nil)
	loginLink.OnTapped = func() {
		myWindow.SetContent(makeLoginScreen(dbConn, myWindow))
	}

	inputContainer := container.NewVBox(
		container.NewPadded(usernameEntry),
		container.NewPadded(passwordEntry),
		container.NewPadded(vaultPathEntry),
		container.NewPadded(selectPathButton),
		container.NewPadded(registerButton),
	)

	return container.NewGridWithColumns(3,
		layout.NewSpacer(),
		container.NewVBox(logo, inputContainer, loginLink),
		layout.NewSpacer(),
	)
}
