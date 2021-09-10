package encryption_test

import (
	"bytes"
	"testing"

	"github.com/enorith/framework/encryption"
)

func TestEncrypt(t *testing.T) {
	ec := encryption.NewAesEncrypter([]byte("somerandomkey!!!"))

	data := []byte("secret string!!!!")
	v, e := ec.Encrypt(data)
	if e != nil {
		t.Fatal(e)
	}
	t.Log("encrypted:", string(v))

	v2, e2 := ec.Decrypt([]byte(v))
	if e2 != nil {
		t.Fatal(e2)
	}
	t.Log("decrypted:", string(v2))

	if !bytes.Equal(data, v2) {
		t.Errorf("%s != %s", data, v2)
		t.Fail()
	}
}
