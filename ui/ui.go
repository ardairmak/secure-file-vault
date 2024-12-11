package ui

import (
	"database/sql"
	"fmt"
	"image/color"
	"path/filepath"
	"secure-file-vault/db"
	"secure-file-vault/vault"
	"time"

	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var currentVault *vault.Vault
var vaultKey []byte

func RunApp(dbPath string) {
    
    currentDir, err := os.Getwd()
    if err != nil {
        panic(fmt.Sprintf("Failed to get current directory: %v", err))
    }

    dbConn, err := db.InitDB(dbPath)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize the database: %v", err))
    }
    defer dbConn.Close()

    myApp := app.New()
    myWindow := myApp.NewWindow("Secure File Vault")

    iconPath := filepath.Join("./assets", "logo.png")
    appIcon := canvas.NewImageFromFile(iconPath)
    myApp.SetIcon(appIcon.Resource)

    var frames []*canvas.Image
    for i := 1; i <= 42; i++ {
        framePath := filepath.Join(currentDir,"/assets/gif", fmt.Sprintf("frame_apngframe%d.png", i))
        frame := canvas.NewImageFromFile(framePath)
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
            time.Sleep(45 * time.Millisecond) // Adjust the delay as needed
        }
		// Set the main window content to the login/register screen
		myWindow.SetContent(makeLoginScreen(dbConn, myWindow))        
    }()

    err = db.CreateVaultTable(dbConn)
    if err != nil {
        panic(fmt.Sprintf("Failed to create vaults table: %v", err))
    }

    err = db.CreateUsersTable(dbConn)
    if err != nil {
        panic(fmt.Sprintf("Failed to create users table: %v", err))
    }

    myApp.Run()
}

func makeLoginScreen(dbConn *sql.DB, myWindow fyne.Window) fyne.CanvasObject {

    logo := canvas.NewImageFromFile("./assets/logoText.png")
    logo.SetMinSize(fyne.NewSize(300, 200))
    logo.FillMode = canvas.ImageFillContain

    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Username")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Password")

    loginButton := widget.NewButton("Login", func() {
        username := usernameEntry.Text
        password := passwordEntry.Text

        vaultPath := filepath.Join("vaults", username, "vault.dat")

        vaultPathFromDB, err := db.AuthenticateUser(dbConn,username,password)
        if err != nil {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title: "Error",
                Content: err.Error(),
            })
            return
        }

        if vaultPathFromDB != vaultPath {
            fyne.CurrentApp().SendNotification(&fyne.Notification{
                Title: "Error",
                Content: "Invalid vault path",
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
        mainScreen := makeMainScreen(dbConn,myWindow,vaultPath,username)
        myWindow.SetContent(mainScreen)
    })
    loginButton.Resize(fyne.NewSize(200,40))

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

func checkIfPathExists(path string) bool {
    _, err := os.Stat(path)
    return !os.IsNotExist(err)
}

func makeRegisterScreen(dbConn *sql.DB, myWindow fyne.Window) fyne.CanvasObject {

    logo := canvas.NewImageFromFile("./assets/logoText.png")
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

                vaultPath := filepath.Join(uri.Path(),usernameEntry.Text+"_vault.dat")
                
                if checkIfPathExists(vaultPath) {
                    fyne.CurrentApp().SendNotification(&fyne.Notification{
                        Title:   "Error",
                        Content: "The selected vault path already exists. Please choose a different path.",
                    })
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
            //make default path inside the vaults dir
            vaultPath = filepath.Join("vaults", username, "vault.dat")
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

func makeMainScreen(dbConn *sql.DB, myWindow fyne.Window, vaultPath, username string) fyne.CanvasObject {

    logo := canvas.NewImageFromFile("./assets/logoText.png")
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
    ))
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

                progressBar := widget.NewProgressBarInfinite()
                progressDialog := dialog.NewCustomWithoutButtons("Extracting Files", progressBar, filesWindow)
                progressDialog.Show()

                go func() {
                    defer progressDialog.Hide()

                    for _, fileItem := range *selectedItems {
                    outputPath := filepath.Join(outputDir, fileItem.Name)
                    data, err := currentVault.ExtractFile(fileItem.Name, vaultKey,outputPath)
                    // if multiple files are selected and one of them fails to extract, rest are not extracted!!
                    if err != nil {
                        fyne.CurrentApp().SendNotification(&fyne.Notification{
                            Title:   "Error",
                            Content: err.Error(),
                        })
                        //return
                        continue
                    }

                    err = os.WriteFile(outputPath, data, 0644)
                    if err != nil {
                        fyne.CurrentApp().SendNotification(&fyne.Notification{
                            Title:   "Error",
                            Content: err.Error(),
                        })
                        //return
                        continue 
                    }
                }
                fyne.CurrentApp().SendNotification(&fyne.Notification{
                    Title:   "Success",
                    Content: "Files extracted successfully",
                })

                *selectedItems = []vault.FileEntry{}

                fileList.Refresh()
                filesWindow.Content().Refresh()
                }()
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
    filesWindow.CenterOnScreen()
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
