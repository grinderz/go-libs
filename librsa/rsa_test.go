package librsa_test

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/grinderz/go-libs/librsa"
)

func TestRsa(t *testing.T) {
	t.Parallel()

	plain1 := "test"

	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	checkError(t, err)
	checkError(t, key1.Validate())

	enc1, err := rsa.SignPKCS1v15(nil, key1, crypto.Hash(0), []byte(plain1))
	checkError(t, err)

	dec1, err := librsa.PublicDecrypt(&key1.PublicKey, enc1)
	checkError(t, err)

	if plain1 != string(dec1) {
		t.Fatal("plain1 != decrypt(enc1)")
	}

	dBytes := librsa.MarshalPrivateKey(key1)
	n, e := librsa.MarshalPublicKey(&key1.PublicKey)

	pubKey2, err := librsa.ParsePublicKey(n, e)
	checkError(t, err)

	key2 := librsa.ParsePrivateKey(pubKey2, dBytes)

	enc2, err := rsa.SignPKCS1v15(nil, key2, crypto.Hash(0), []byte(plain1))
	checkError(t, err)

	if !bytes.Equal(enc1, enc2) {
		t.Fatal("enc1 != enc2")
	}

	dec2, err := librsa.PublicDecrypt(&key2.PublicKey, enc2)
	checkError(t, err)

	if plain1 != string(dec2) {
		t.Fatal("plain1 != decrypt(enc2)")
	}
}

func TestParsePublicKeyOversizedExponent(t *testing.T) {
	t.Parallel()

	n := []byte{0x01, 0x02, 0x03}
	e := bytes.Repeat([]byte{0xff}, 9)

	if _, err := librsa.ParsePublicKey(n, e); !errors.Is(err, librsa.ErrExponentOutOfRange) {
		t.Fatalf("expected ErrExponentOutOfRange, got %v", err)
	}
}

func TestPublicDecryptWrongKey(t *testing.T) {
	t.Parallel()

	plain := "test"

	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	checkError(t, err)

	key2, err := rsa.GenerateKey(rand.Reader, 2048)
	checkError(t, err)

	enc, err := rsa.SignPKCS1v15(nil, key1, crypto.Hash(0), []byte(plain))
	checkError(t, err)

	if _, err := librsa.PublicDecrypt(&key2.PublicKey, enc); err == nil {
		t.Fatal("decrypt with wrong key must fail")
	}
}

func checkError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
