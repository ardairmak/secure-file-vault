package ui

import (
	"os"
	"path/filepath"
	"secure-file-vault/vault"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showFilesWindow(vaultPath string) {
	filesWindow := fyne.CurrentApp().NewWindow("Files in Vault")
	fileList, selectedItems := makeFileListContent()

	extractButton := widget.NewButton("Extract", func() {
		if len(*selectedItems) == 0 {
			showErrorNotification("No file selected for extraction")
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
						data, err := currentVault.ExtractFile(fileItem.Name, vaultKey, outputPath)
						if err != nil {
							showErrorNotification(err.Error())
							continue
						}

						err = os.WriteFile(outputPath, data, 0644)
						if err != nil {
							showErrorNotification(err.Error())
							continue
						}

						go watchFileForChanges(outputPath, vaultPath, vaultKey, currentVault)
					}
					showSuccessNotification("Files extracted successfully")

					*selectedItems = []vault.FileEntry{}
					fileList.Refresh()
					filesWindow.Content().Refresh()
				}()
			}
		}, filesWindow)
	})

	removeButton := widget.NewButton("Remove", func() {
		if len(*selectedItems) == 0 {
			showErrorNotification("No file selected for removal")
			return
		}

		for _, fileItem := range *selectedItems {
			err := currentVault.RemoveFile(fileItem.Name, vaultKey)
			if err != nil {
				showErrorNotification(err.Error())
				return
			}
		}

		err := currentVault.Save(vaultPath)
		if err != nil {
			showErrorNotification(err.Error())
			return
		}

		showSuccessNotification("Files removed successfully")
		*selectedItems = []vault.FileEntry{}

		fileList.Refresh()
		filesWindow.Content().Refresh()
	})

	filesContainer := container.NewBorder(nil, container.NewVBox(extractButton, removeButton), nil, nil, fileList)
	filesWindow.SetContent(filesContainer)
	filesWindow.Resize(fyne.NewSize(500, 400))
	filesWindow.CenterOnScreen()
	filesWindow.Show()
}

func makeFileListContent() (*widget.List, *[]vault.FileEntry) {
	selectedItems := &[]vault.FileEntry{}

	fileList := widget.NewList(
		func() int {
			files, err := currentVault.ListFiles(vaultKey)
			if err != nil {
				showErrorNotification(err.Error())
				return 0
			}
			return len(files)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", nil),
				widget.NewLabel(""),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			files, err := currentVault.ListFiles(vaultKey)
			if err != nil {
				showErrorNotification(err.Error())
				return
			}
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
