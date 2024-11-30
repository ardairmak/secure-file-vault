package vault

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type Vault struct {
    Salt string          `json:"salt"`
    KeyHash string       `json:"key_hash"`
    Files []FileEntry    `json:"files"`
}


func CreateVault(vaultPath, password string) (*Vault, error) {
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
        Salt: base64.StdEncoding.EncodeToString(salt),
        KeyHash: base64.StdEncoding.EncodeToString(keyHash[:]),
        Files: []FileEntry{},
    }
    
    vaultData, err := json.Marshal(vault)
    if err != nil {
        return nil, err
    }
    if err := os.WriteFile(vaultPath, vaultData, 0644); err != nil {
        return nil, err
    }

    return vault, nil
}

func OpenVault(vaultPath, password string) (*Vault,[]byte, error) {
    vaultData, err := os.ReadFile(vaultPath)
    if err != nil {
        return nil, nil, err
    }
    var vault Vault
    if err := json.Unmarshal(vaultData, &vault); err != nil {
        return nil, nil, err
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
        return nil, nil, fmt.Errorf("Invalid password")
    }
    return &vault, key, nil
}


func (vault *Vault) AddFile(filename string, data []byte, key []byte) error {
    encryptedData, err := EncryptData(key,data)
    if err != nil {
        return err
    }
    fileHash := HashData(data)
    vault.Files = append(vault.Files, FileEntry{
        Name: filename,
        Hash: fileHash,
        Data: encryptedData,
    })
    return nil
}

func (vault *Vault) Save(vaultPath string) error {
    vaultData, err := json.Marshal(vault)
    if err != nil {
        return err
    }
    return os.WriteFile(vaultPath, vaultData, 0644)
}

func (vault *Vault) ExtractFile(filename string, key []byte) ([]byte, error) {
    for _, file := range vault.Files {
        if file.Name == filename {
            decryptedData, err := DecryptData(key, file.Data)
            if err != nil {
                return nil, err
            }
            if HashData(decryptedData) != file.Hash {
                return nil, fmt.Errorf("File data corrupted")
            }
            return decryptedData, nil

        }
    }
    return nil, fmt.Errorf("File not found")
}
