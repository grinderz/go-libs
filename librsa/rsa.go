package librsa

import (
	"crypto/rsa"
	"encoding/binary"
	"errors"
	"math"
	"math/big"
)

// ErrPKCS1DelimiterNotFound is returned by PublicDecrypt when the decrypted
// block carries no PKCS#1 v1.5 `0xff 0x00` padding delimiter — e.g. the data
// was signed with a different key or is corrupted.
var ErrPKCS1DelimiterNotFound = errors.New("pkcs1 v1.5 padding delimiter not found")

// ErrExponentOutOfRange is returned by ParsePublicKey when the exponent bytes
// exceed the 4-byte big-endian uint32 range produced by MarshalPublicKey;
// big.Int.Int64 is undefined beyond int64 and int() truncates on 32-bit
// platforms, so oversized input must not be converted silently.
var ErrExponentOutOfRange = errors.New("rsa exponent out of range")

const exponentSize = 4 // big-endian uint32

func ParsePublicKey(ns, es []byte) (*rsa.PublicKey, error) {
	modulus := new(big.Int)
	modulus.SetBytes(ns)

	exponent := new(big.Int)
	exponent.SetBytes(es)

	if !exponent.IsInt64() || exponent.Int64() > math.MaxInt32 {
		return nil, ErrExponentOutOfRange
	}

	return &rsa.PublicKey{N: modulus, E: int(exponent.Int64())}, nil
}

func MarshalPublicKey(pubKey *rsa.PublicKey) ([]byte, []byte) {
	n := pubKey.N.Bytes()
	e := make([]byte, exponentSize)

	binary.BigEndian.PutUint32(e, uint32(pubKey.E)) //nolint:gosec

	return n, e
}

func ParsePrivateKey(pubKey *rsa.PublicKey, ds []byte) *rsa.PrivateKey {
	d := new(big.Int)
	d.SetBytes(ds)

	return &rsa.PrivateKey{PublicKey: *pubKey, D: d} //nolint:exhaustruct
}

func MarshalPrivateKey(privatekey *rsa.PrivateKey) []byte {
	return privatekey.D.Bytes()
}

// https://www.openssl.org/docs/man1.1.0/crypto/RSA_public_decrypt.html

func PublicDecrypt(pubKey *rsa.PublicKey, data []byte) ([]byte, error) {
	c := new(big.Int) //nolint:varnamelen
	m := new(big.Int) //nolint:varnamelen

	m.SetBytes(data)

	e := big.NewInt(int64(pubKey.E))

	c.Exp(m, e, pubKey.N)

	out := c.Bytes()
	step := 2

	for ind := step; ind+1 < len(out); ind++ {
		if out[ind] == 0xff && out[ind+1] == 0 {
			return out[ind+step:], nil
		}
	}

	return nil, ErrPKCS1DelimiterNotFound
}
