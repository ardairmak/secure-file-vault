package ui

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"secure-file-vault/db"
	"secure-file-vault/vault"

	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var currentVault *vault.Vault
var vaultKey []byte

func RunApp(dbPath string) {
    myApp := app.New()
    myWindow := myApp.NewWindow("Secure File Vault")

    dbConn, err := db.InitDB(dbPath)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize the database: %v", err))
    }
    defer dbConn.Close()

    err = db.CreateVaultTable(dbConn)
    if err != nil {
        panic(fmt.Sprintf("Failed to create vaults table: %v", err))
    }

    err = db.CreateUsersTable(dbConn)
    if err != nil {
        panic(fmt.Sprintf("Failed to create users table: %v", err))
    }

    myWindow.SetContent(makeLoginRegisterScreen(dbConn, myWindow))
    myWindow.Resize(fyne.NewSize(400, 400))
    myWindow.ShowAndRun()
}

func makeLoginRegisterScreen(dbConn *sql.DB, myWindow fyne.Window) fyne.CanvasObject {
        loginButton := widget.NewButton("Login", func() {
        myWindow.SetContent(makeLoginScreen(dbConn, myWindow))
    })

    registerButton := widget.NewButton("Register", func() {
        myWindow.SetContent(makeRegisterScreen(dbConn, myWindow))
    })

    return container.NewVBox(
        loginButton,
        registerButton,
    )
}

func makeLoginScreen(dbConn *sql.DB, myWindow fyne.Window) fyne.CanvasObject {

    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Enter username...")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Enter password...")

    loginButton := widget.NewButton("Login", func() {
        username := usernameEntry.Text
        password := passwordEntry.Text

        vaultPath, err := db.AuthenticateUser(dbConn,username,password)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title: "Error",
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
        currentVault = vlt
        vaultKey = key
        mainScreen := makeMainScreen(dbConn,myWindow,vaultPath)
        myWindow.SetContent(mainScreen)
    })

    return container.NewVBox(
        widget.NewLabel("Username"),
        usernameEntry,
        widget.NewLabel("Password"),
        passwordEntry,
        loginButton,
        widget.NewButton("Back",func (){
            myWindow.SetContent(makeLoginRegisterScreen(dbConn,myWindow))
        }),
    )
}

func makeRegisterScreen(dbConn *sql.DB, myWindow fyne.Window) fyne.CanvasObject {

    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Enter username...")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Enter password...")

    vaultPathEntry := widget.NewEntry()
    vaultPathEntry.SetPlaceHolder("Enter vault path...")

    vaultPathEntry.SetText("")

    selectPathButton := widget.NewButton("Select Custom Vault Path", func() {
        dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
            if uri != nil {
                vaultPathEntry.SetText(uri.Path() + "/vault.dat")
            }
        }, myWindow)
    })

    registerButton := widget.NewButton("Register", func() {

        username := usernameEntry.Text
        password := passwordEntry.Text
        vaultPath := vaultPathEntry.Text

        if vaultPath == "" {
            homeDir, err := os.UserHomeDir()
            if err != nil {
                fyne.CurrentApp().SendNotification(&fyne.Notification{
                    Title:   "Error",
                    Content: err.Error(),
                })
                return
        }
        vaultPath = filepath.Join(homeDir,username, "vault.dat")
    }

        vaultDir := filepath.Dir(vaultPath)
        err := os.MkdirAll(vaultDir, os.ModePerm)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: fmt.Sprintf("Failed to create directory: %v", err),
            })
            return
        } 
        

        err = db.RegisterUser(dbConn, username, password, vaultPath)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
        })
        return
        }

        _, err = vault.CreateVault(vaultPath, password)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: fmt.Sprintf("Failed to create vault: %v", err),
            })
            return
        }

        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Success",
            Content: "User registered successfully",
        })
        myWindow.SetContent(makeLoginRegisterScreen(dbConn, myWindow))
    })
    return container.NewVBox(
        widget.NewLabel("Username"),
        usernameEntry,
        widget.NewLabel("Password"),
        passwordEntry,
        widget.NewLabel("Vault Path"),
        vaultPathEntry,
        selectPathButton,
        registerButton,
        widget.NewButton("Back", func() {
            myWindow.SetContent(makeLoginRegisterScreen(dbConn, myWindow))
        }),
    )
}

func makeMainScreen(dbConn *sql.DB, myWindow fyne.Window, vaultPath string) fyne.CanvasObject {

    vaultStatus := widget.NewLabel("Vault Unlocked")

    fileEntry := widget.NewEntry()
    fileEntry.SetPlaceHolder("Enter file path...")

    outputPathEntry := widget.NewEntry()
    outputPathEntry.SetPlaceHolder("Enter output path...")

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
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return 
        }

        err = currentVault.AddFile(filepath.Base(filePath), data, vaultKey)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return 
        }

        err = currentVault.Save(vaultPath)
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

    viewFilesButton := widget.NewButton("View Files", func() {
        showFilesWindow(vaultPath)
    })


    logoutButton := widget.NewButton("Logout", func() {
        currentVault = nil
        vaultKey = nil
        loginScreen := makeLoginScreen(dbConn, myWindow)
        myWindow.SetContent(loginScreen)
    })

    form := container.NewVBox(
        vaultStatus,
        selectFileButton,
        widget.NewLabel("File Path"),
        fileEntry,
        widget.NewLabel("Output Path"),
        outputPathEntry,
        addFileButton,
        widget.NewLabel("Files in Vault"),
        viewFilesButton,
        logoutButton,
    )

    return form
}

func showFilesWindow(vaultPath string) {
    filesWindow := fyne.CurrentApp().NewWindow("Files in Vault")

    fileList, selectedItems := makeFileListContent()

    extractButton := widget.NewButton("Extract", func() {
        if(len(*selectedItems) == 0){
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: "No file selected for extraction",
        })
        return
    }
        dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
            if err == nil && uri != nil {
                outputDir := uri.Path()

                for _, fileItem := range *selectedItems {
                    outputPath := filepath.Join(outputDir, fileItem.Name)
                    data, err := currentVault.ExtractFile(fileItem.Name, vaultKey,outputPath)
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
                }

                fyne.CurrentApp().SendNotification(&fyne.Notification{
                    Title:   "Success",
                    Content: "Files extracted successfully",
                })

                *selectedItems = []vault.FileEntry{}

                fileList.Refresh()
                filesWindow.Content().Refresh()
            }
        }, filesWindow)
    })

    removeButton := widget.NewButton("Remove", func() {
        if(len(*selectedItems) == 0){
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: "No file selected for removal",
        })
        return
    }
    
        for _, fileItem := range *selectedItems {
            err := currentVault.RemoveFile(fileItem.Name)
            if err != nil {
                fyne.CurrentApp().SendNotification(&fyne.Notification{
                    Title:   "Error",
                    Content: err.Error(),
                })
                return 
            }
        }

        err := currentVault.Save(vaultPath)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title:   "Error",
                Content: err.Error(),
            })
            return
        }

        fyne.CurrentApp().SendNotification(&fyne.Notification{
            Title:   "Success",
            Content: "Files removed successfully",
        })

        *selectedItems = []vault.FileEntry{}

        fileList.Refresh()
        filesWindow.Content().Refresh()
    })

    filesContainer := container.NewBorder(nil,container.NewVBox(extractButton,removeButton),nil,nil,fileList)

    filesWindow.SetContent(filesContainer)
    filesWindow.Resize(fyne.NewSize(500,400))
    filesWindow.Show()
}

func makeFileListContent() (fyne.CanvasObject, *[]vault.FileEntry) {
    
    selectedItems := &[]vault.FileEntry{}

    fileList := widget.NewList(
        func() int {
            return len(currentVault.ListFiles())
        },
        func() fyne.CanvasObject {
            return container.NewHBox(
                widget.NewCheck("", nil),
                widget.NewLabel(""),
            )
        },

        func(i widget.ListItemID, o fyne.CanvasObject) {
            files := currentVault.ListFiles()
            fileItem := files[i]

            check := o.(*fyne.Container).Objects[0].(*widget.Check)
            label := o.(*fyne.Container).Objects[1].(*widget.Label)

            label.SetText(fileItem.Name)
            check.SetChecked(false)

            check.OnChanged = func(checked bool) {
                if checked {
                    *selectedItems = append(*selectedItems, fileItem)
                } else {
                    
                    for index, item := range *selectedItems {
                        if item.Name == fileItem.Name {
                            *selectedItems = append((*selectedItems)[:index], (*selectedItems)[index+1:]...)
                            break
                        }
                    }
                }
            }
        },
    )

    return fileList, selectedItems
}
