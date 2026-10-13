package librsa_test

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"math/big"
	"testing"

	"github.com/grinderz/go-libs/librsa"
)

func TestRsa(t *testing.T) {
	t.Parallel()

	plain1 := plainText

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

	plain := plainText

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

func TestParsePublicKeySmallExponent(t *testing.T) {
	t.Parallel()

	n := []byte{0x01, 0x02, 0x03}

	for _, e := range [][]byte{nil, {0x00, 0x00, 0x00, 0x01}} {
		if _, err := librsa.ParsePublicKey(n, e); !errors.Is(err, librsa.ErrExponentOutOfRange) {
			t.Fatalf("e=%v: expected ErrExponentOutOfRange, got %v", e, err)
		}
	}
}

func TestPublicDecryptInvalidKey(t *testing.T) {
	t.Parallel()

	keys := []*rsa.PublicKey{
		nil,
		{N: nil, E: 65537},
		{N: big.NewInt(0), E: 65537},
		{N: big.NewInt(3233), E: 1},
	}

	for _, key := range keys {
		if _, err := librsa.PublicDecrypt(key, []byte{0x01}); !errors.Is(err, librsa.ErrInvalidPublicKey) {
			t.Fatalf("key=%v: expected ErrInvalidPublicKey, got %v", key, err)
		}
	}
}

func TestPublicDecryptDataOutOfRange(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	checkError(t, err)

	data := key.N.Bytes()

	if _, err := librsa.PublicDecrypt(&key.PublicKey, data); !errors.Is(err, librsa.ErrDataOutOfRange) {
		t.Fatalf("expected ErrDataOutOfRange, got %v", err)
	}
}

func TestPublicDecryptRejectsNonType1Block(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	checkError(t, err)

	size := key.Size()

	// Raw RSA private operation over a hand-built block: the public operation
	// then yields exactly this block back.
	sign := func(block []byte) []byte {
		return new(big.Int).Exp(new(big.Int).SetBytes(block), key.D, key.N).FillBytes(make([]byte, size))
	}

	valid := func() []byte {
		block := bytes.Repeat([]byte{0xff}, size)
		block[0], block[1] = 0x00, 0x01
		block[size-len(plainText)-1] = 0x00
		copy(block[size-len(plainText):], plainText)

		return block
	}

	got, err := librsa.PublicDecrypt(&key.PublicKey, sign(valid()))
	checkError(t, err)

	if string(got) != plainText {
		t.Fatalf("got %q, want %q", got, plainText)
	}

	wrongType := valid()
	wrongType[1] = 0x02

	shortPadding := valid()
	shortPadding[2+minPaddingLen-1] = 0x00 // separator after 7 bytes of 0xff

	noSeparator := bytes.Repeat([]byte{0xff}, size)
	noSeparator[0], noSeparator[1] = 0x00, 0x01

	badFiller := valid()
	badFiller[5] = 0x7f

	for name, block := range map[string][]byte{
		"wrong block type": wrongType,
		"short padding":    shortPadding,
		"no separator":     noSeparator,
		"non-0xff filler":  badFiller,
	} {
		if _, err := librsa.PublicDecrypt(&key.PublicKey, sign(block)); !errors.Is(err, librsa.ErrPKCS1InvalidPadding) {
			t.Errorf("%s: expected ErrPKCS1InvalidPadding, got %v", name, err)
		}
	}
}

const (
	minPaddingLen = 8
	plainText     = "test"
)

func checkError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
