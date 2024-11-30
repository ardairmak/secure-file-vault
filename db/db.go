package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    return db,nil
}

func CreateVaultTable(db *sql.DB) error {
    createTableSQL := `CREATE TABLE IF NOT EXISTS vaults (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "path" TEXT NOT NULL,
    "salt" TEXT NOT NULL,
    "key_hash" TEXT NOT NULL
    );`
    _, err := db.Exec(createTableSQL)
    return err
}

func AddVault(db *sql.DB, vaultPath, salt, keyHash string) error {
    insertSQL := `INSERT INTO vaults (path, salt, key_hash) VALUES (?, ?, ?);`
    _, err := db.Exec(insertSQL, vaultPath, salt, keyHash)
    return err
}

func GetVault(db *sql.DB, vaultPath string) (string, string, error) {
    querySQL := `SELECT salt, key_hash FROM vaults WHERE path = ?;`
    row := db.QueryRow(querySQL, vaultPath)
    var salt, keyHash string
    err := row.Scan(&salt, &keyHash)
    return salt, keyHash, err
}
