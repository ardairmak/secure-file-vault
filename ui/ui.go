package ui

import (
	"secure-file-vault/db"
	"secure-file-vault/vault"

	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func RunApp(dbPath string) {
    myApp := app.New()
    myWindow := myApp.NewWindow("Secure File Vault")

    dbConn, err := db.InitDB(dbPath)
    if err != nil {
        panic(err)
    }
    defer dbConn.Close()

    err = db.CreateVaultTable(dbConn)
    if err != nil {
        panic(err)
    }

    vaultPathEntry := widget.NewEntry()
    vaultPathEntry.SetPlaceHolder("Enter vault path...")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Enter password...")

    fileEntry := widget.NewEntry()
    fileEntry.SetPlaceHolder("Enter file path...")

    outputPathEntry := widget.NewEntry()
    outputPathEntry.SetPlaceHolder("Enter output path...")

    createVaultButton := widget.NewButton("Create Vault", func() {
        password := passwordEntry.Text
        vaultPath := vaultPathEntry.Text
        vlt, err := vault.CreateVault(vaultPath, password)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }
        err = db.AddVault(dbConn, vaultPath, vlt.Salt, vlt.KeyHash)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }
        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Success",
            Content: "Vault created successfully",
        })
    })

    addFileButton := widget.NewButton("Add File", func() {
        vaultPath := vaultPathEntry.Text
        password := passwordEntry.Text
        filePath := fileEntry.Text

        _,_,err := db.GetVault(dbConn, vaultPath)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        vlt, key, err := vault.OpenVault(vaultPath, password)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        data, err := os.ReadFile(filePath)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        err = vlt.AddFile(filePath, data, key)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        err = vlt.Save(vaultPath)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Success",
            Content: "File added successfully",
        })
    })

    extractFileButton := widget.NewButton("Extract File", func() {
        vaultPath := vaultPathEntry.Text
        password := passwordEntry.Text
        filePath := fileEntry.Text
        outputPath := outputPathEntry.Text

        vlt, key, err := vault.OpenVault(vaultPath, password)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        data, err := vlt.ExtractFile(filePath, key)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        err = os.WriteFile(outputPath, data, 0644)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Success",
            Content: "File extracted successfully",
        })
    })

    form := container.NewVBox(
        widget.NewLabel("Vault Path"),
        vaultPathEntry,
        widget.NewLabel("Password"),
        passwordEntry,
        widget.NewLabel("File Path"),
        fileEntry,
        widget.NewLabel("Output Path"),
        outputPathEntry,
        createVaultButton,
        addFileButton,
        extractFileButton,
    )

    myWindow.SetContent(form)
    myWindow.Resize(fyne.NewSize(400, 400))
    myWindow.ShowAndRun()
}

