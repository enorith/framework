package encryption

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
)

type Encrypter interface {
	Decrypt(data []byte) ([]byte, error)
	Encrypt(data []byte) ([]byte, error)
}

type AesEncrypter struct {
	key []byte
}

func (e *AesEncrypter) Decrypt(data []byte) ([]byte, error) {
	if len(data) < 1 {
		return nil, nil
	}

	d := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(data))
	dest, err := ioutil.ReadAll(d)
	if err != nil {
		return nil, err
	}
	dc, err := DecryptAes256Ecb(dest, e.key)
	return dc, err
}

func (e *AesEncrypter) Encrypt(data []byte) ([]byte, error) {
	if len(data) < 1 {
		return nil, nil
	}

	ec, err := EncryptAes256Ecb(data, e.key)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(ec)))
	base64.StdEncoding.Encode(buf, ec)
	return buf, nil
}

func NewAesEncrypter(key []byte) *AesEncrypter {
	return &AesEncrypter{key: AesNormalKey(key)}
}

// AesNormalKey key must be 16 24 32 bytes
func AesNormalKey(key []byte) []byte {
	l := len(key)
	if l < 16 {
		key = append(key, []byte("aA1bB2cC3dD4eE5fF6gG7")...)
	}

	if 16 <= l && l < 24 {
		return key[:16]
	} else if 24 <= l && l < 32 {
		return key[:24]
	} else {
		return key[:32]
	}
}
