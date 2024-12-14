package ui

import (
	"fmt"
	"secure-file-vault/db"
	"secure-file-vault/vault"

	"fyne.io/fyne/v2/app"
)

var currentVault *vault.Vault
var vaultKey []byte

func RunApp(dbPath string) {
	dbConn, err := db.InitDB(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize the database: %v", err))
	}
	defer dbConn.Close()

	myApp := app.New()
	myWindow := myApp.NewWindow("Secure File Vault")

	showAnimation(myWindow, func() {
		myWindow.SetContent(makeLoginScreen(dbConn, myWindow))
	})

	if err := db.CreateVaultTable(dbConn); err != nil {
		panic(fmt.Sprintf("Failed to create vaults table: %v", err))
	}
	if err := db.CreateUsersTable(dbConn); err != nil {
		panic(fmt.Sprintf("Failed to create users table: %v", err))
	}

	myApp.Run()
}
