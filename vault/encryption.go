package vault

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
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
    plaintext = pkcs7Pad(plaintext, aes.BlockSize)
    ciphertext := make([]byte, aes.BlockSize + len(plaintext))
    iv := ciphertext[:aes.BlockSize]

    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }

    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

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

    if len(ciphertext)%aes.BlockSize != 0 {
        return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
    }

    mode := cipher.NewCBCDecrypter(block, iv)
    mode.CryptBlocks(ciphertext, ciphertext)

    plaintext, err := pkcs7Unpad(ciphertext, aes.BlockSize)
    if err != nil {
        return nil, fmt.Errorf("failed to unpad ciphertext: %v", err)
    }

    return plaintext, nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
    padding := blockSize - len(data)%blockSize
    padtext := bytes.Repeat([]byte{byte(padding)}, padding)
    return append(data, padtext...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
    if len(data) == 0 || len(data)%blockSize != 0 {
        return nil, errors.New("invalid padding size")
    }

    padding := data[len(data)-1]
    padLen := int(padding)

    if padLen == 0 || padLen > blockSize {
        return nil, fmt.Errorf("invalid padding 1: padLen=%d, blockSize=%d, dataLen=%d", padLen, blockSize, len(data))
    }

    padtext := data[len(data)-padLen:]
    for i, v := range padtext {
        if v != padding {
            return nil, fmt.Errorf("invalid padding 2: padtext[%d]=%d, expected=%d", i, v, padding)
        }
    }

    return data[:len(data)-padLen], nil
}

func HashData(data []byte) string {
    hash := sha256.Sum256(data)
    return base64.StdEncoding.EncodeToString(hash[:])
}
