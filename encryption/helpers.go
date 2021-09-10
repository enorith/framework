package encryption

import (
	"bytes"
	"crypto/aes"
)

func DecryptAes256Ecb(data, key []byte) ([]byte, error) {
	if len(data) < 1 {
		return nil, nil
	}

	cipher, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	decrypted := make([]byte, len(data))
	size := aes.BlockSize

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Decrypt(decrypted[bs:be], data[bs:be])
	}

	return PKCS7UnPadding(decrypted), nil
}

func EncryptAes256Ecb(data, key []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	data = PKCS7Padding(data, block.BlockSize())
	ciphertext := make([]byte, len(data))
	size := block.BlockSize()
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Encrypt(ciphertext[bs:be], data[bs:be])
	}
	return ciphertext, nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
