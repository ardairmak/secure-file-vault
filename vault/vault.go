package vault

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

type Vault struct {
	Salt    string      `json:"salt"`
	KeyHash string      `json:"key_hash"`
	Files   []FileEntry `json:"files"`
}

func CreateVault(vaultPath, password string) (*Vault, error) {

	dir := filepath.Dir(vaultPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	salt, err := GenerateSalt()
	if err != nil {
		return nil, err
	}
	key, err := DeriveKey(password, salt)
	if err != nil {
		return nil, err
	}
	keyHash := sha256.Sum256(key)
	vault := &Vault{
		Salt:    base64.StdEncoding.EncodeToString(salt),
		KeyHash: base64.StdEncoding.EncodeToString(keyHash[:]),
		Files:   []FileEntry{},
	}

	file, err := os.Create(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault file: %v", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to encode vault: %v", err)
	}

	return vault, nil
}

func OpenVault(vaultPath, password string) (*Vault, []byte, error) {
	file, err := os.Open(vaultPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var vault Vault
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&vault)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode vault: %v", err)
	}

	salt, err := base64.StdEncoding.DecodeString(vault.Salt)
	if err != nil {
		return nil, nil, err
	}

	key, err := DeriveKey(password, salt)
	if err != nil {
		return nil, nil, err
	}

	keyHash := sha256.Sum256(key)
	if base64.StdEncoding.EncodeToString(keyHash[:]) != vault.KeyHash {
		return nil, nil, fmt.Errorf("invalid password")
	}
	return &vault, key, nil
}

func (vault *Vault) AddFile(filePath, vaultPath string, data []byte, key []byte) error {
	//deny trying to add vault to itself
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	absVaultPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	if absFilePath == absVaultPath {
		return fmt.Errorf("vault file cannot be added to itself")
	}

	//check for duplicate file
	baseFileName := filepath.Base(filePath)
	for _, file := range vault.Files {
		decryptedFileName, err := DecryptFileName(key, file.Name)
		if err != nil {
			return fmt.Errorf("failed to decrypt filename: %v", err)
		}

		if decryptedFileName == baseFileName {
			return fmt.Errorf("file already exists")
		}
	}

	encryptedData, err := EncryptData(key, data)
	if err != nil {
		return err
	}
	encryptedFileName, err := EncryptFileName(key, baseFileName)
	if err != nil {
		return err
	}

	fileHash := HashData(data)

	FileEntry := FileEntry{
		Name: encryptedFileName,
		Hash: fileHash,
		Data: encryptedData,
	}

	vault.Files = append(vault.Files, FileEntry)
	return nil
}

func (vault *Vault) Save(vaultPath string) error {
	file, err := os.Create(vaultPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(vault)
	if err != nil {
		return err
	}

	return nil
}

func (vault *Vault) ListFiles(key []byte) ([]FileEntry, error) {
	var decryptedFiles []FileEntry
	for _, file := range vault.Files {
		decryptedFileName, err := DecryptFileName(key, file.Name)
		if err != nil {
			return nil, err
		}
		decryptedFiles = append(decryptedFiles, FileEntry{
			Name: decryptedFileName,
			Hash: file.Hash,
			Data: file.Data,
		})
	}
	return decryptedFiles, nil
}

func (vault *Vault) RemoveFile(filePath string, key []byte) error {
	for i, file := range vault.Files {

		decryptedFileName, err := DecryptFileName(key, file.Name)
		if err != nil {
			return err
		}

		if decryptedFileName == filePath {
			vault.Files = append(vault.Files[:i], vault.Files[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("file not found")
}

func (vault *Vault) ExtractFile(filePath string, key []byte, outputPath string) ([]byte, error) {
	for _, file := range vault.Files {
		decryptedFileName, err := DecryptFileName(key, file.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt filename: %v", err)
		}

		if decryptedFileName == filePath {
			decryptedData, err := DecryptData(key, file.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt data: %v", err)
			}

			// Verify file integrity
			decryptedFileHash := HashData(decryptedData)
			if decryptedFileHash != file.Hash {
				return nil, fmt.Errorf("file integrity check failed for %s", filePath)
			}
			//check if the file with the same name already exists in the extract path
			if _, err := os.Stat(outputPath); err == nil {
				return nil, fmt.Errorf("file already exists at the extract path")
			}

			err = os.WriteFile(outputPath, decryptedData, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to write extracted file: %v", err)
			}
			return decryptedData, nil
		}
	}
	return nil, fmt.Errorf("file not found")
}

func (vault *Vault) UpdateFile(fileName string, key []byte, newData []byte) error {
	for i, file := range vault.Files {
		decryptedFileName, err := DecryptFileName(key, file.Name)
		if err != nil {
			return fmt.Errorf("failed to decrypt filename: %v", err)
		}

		if decryptedFileName == fileName {
			encryptedData, err := EncryptData(key, newData)
			if err != nil {
				return fmt.Errorf("failed to encrypt data: %v", err)
			}

			fileHash := HashData(newData)

			vault.Files[i] = FileEntry{
				Name: file.Name, // Keep the original encrypted name
				Hash: fileHash,
				Data: encryptedData,
			}

			return nil
		}
	}
	return fmt.Errorf("file not found: %s", fileName)
}
