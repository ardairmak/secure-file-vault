package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)


func DeriveKey(password string, salt []byte) ([]byte, error) {
    return scrypt.Key([]byte(password), salt, 16384, 8, 1, 32)
}

func GenerateSalt() ([]byte,error){
    salt := make([]byte,16)
    _, err := rand.Read(salt)
    return salt, err
}

func EncryptFileName(key []byte, fileName string) (string, error) {
    encryptedFileName, err := EncryptData(key, []byte(fileName))
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(encryptedFileName), nil
}

func DecryptFileName(key []byte, encryptedFileName string) (string, error) {
    encryptedFileNameBytes, err := base64.StdEncoding.DecodeString(encryptedFileName)
    if err != nil {
        return "", err
    }
    decryptedFileName, err := DecryptData(key, encryptedFileNameBytes)
    if err != nil {
        return "", err
    }
    return string(decryptedFileName), nil
}

func EncryptData(key, plaintext []byte) ([]byte, error) {
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    ciphertext := make([]byte, aes.BlockSize + len(plaintext))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }

    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

    return ciphertext, nil
}

func DecryptData(key, ciphertext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    if len(ciphertext) < aes.BlockSize {
        return nil, fmt.Errorf("ciphertext too short")
    }
    iv := ciphertext[:aes.BlockSize]
    ciphertext = ciphertext[aes.BlockSize:]

    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(ciphertext, ciphertext)
    return ciphertext, nil
}

func HashData(data []byte) string {
    hash := sha256.Sum256(data)
    return base64.StdEncoding.EncodeToString(hash[:])
}
