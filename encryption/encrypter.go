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
	return &AesEncrypter{key: key}
}
