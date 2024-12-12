package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"secure-file-vault/vault"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func watchFileForChanges(filePath, vaultPath string, key []byte, vault *vault.Vault) {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Failed to get initial file stats: %v\n", err)
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		currentStat, err := os.Stat(filePath)
		if err != nil {
			return
		}

		if currentStat.ModTime().After(initialStat.ModTime()) {
			showUpdateFileDialog(filePath, func(update bool) {
				if update {
					handleFileChange(filePath, vaultPath, key, vault)
				}
			})
			return
		}
	}
}

func showUpdateFileDialog(filePath string, callback func(bool)) {
	dialog.ShowConfirm(
		"File Modified",
		fmt.Sprintf("The file %s has been modified. Would you like to update it in the vault?", filePath),
		callback,
		fyne.CurrentApp().Driver().AllWindows()[0],
	)
}

func handleFileChange(filePath, vaultPath string, key []byte, vault *vault.Vault) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	fileName := filepath.Base(filePath)
	if err := vault.UpdateFile(fileName, key, data); err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	if err := vault.Save(vaultPath); err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	dialog.ShowInformation("Success", "File updated successfully",
		fyne.CurrentApp().Driver().AllWindows()[0])
}
