package librsa

import (
	"crypto/rsa"
	"encoding/binary"
	"errors"
	"math"
	"math/big"
)

// ErrPKCS1InvalidPadding is returned by PublicDecrypt when the decrypted
// block is not a valid PKCS#1 v1.5 type 1 block (`00 01 ff..ff 00 data`) —
// e.g. the data was signed with a different key or is corrupted.
var ErrPKCS1InvalidPadding = errors.New("invalid pkcs1 v1.5 padding")

// ErrExponentOutOfRange is returned by ParsePublicKey when the exponent is
// below 2 or above math.MaxInt32, the same bounds crypto/rsa enforces.
// big.Int.Int64 is undefined beyond int64 and int() truncates on 32-bit
// platforms, so oversized input must not be converted silently.
var ErrExponentOutOfRange = errors.New("rsa exponent out of range")

// ErrInvalidPublicKey is returned by PublicDecrypt for a key without a
// positive modulus or with an exponent below 2: big.Int.Exp with a zero
// modulus computes the full, unreduced power.
var ErrInvalidPublicKey = errors.New("invalid rsa public key")

// ErrDataOutOfRange is returned by PublicDecrypt when the data, read as a
// big-endian integer, is not smaller than the modulus.
var ErrDataOutOfRange = errors.New("rsa data out of range")

const (
	exponentSize  = 4 // big-endian uint32
	minExponent   = 2
	minPaddingLen = 8 // PKCS#1 v1.5 requires at least 8 padding bytes
)

func ParsePublicKey(ns, es []byte) (*rsa.PublicKey, error) {
	modulus := new(big.Int)
	modulus.SetBytes(ns)

	exponent := new(big.Int)
	exponent.SetBytes(es)

	if !exponent.IsInt64() || exponent.Int64() < minExponent || exponent.Int64() > math.MaxInt32 {
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
	if pubKey == nil || pubKey.N == nil || pubKey.N.Sign() <= 0 || pubKey.E < minExponent {
		return nil, ErrInvalidPublicKey
	}

	m := new(big.Int).SetBytes(data)
	if m.Cmp(pubKey.N) >= 0 {
		return nil, ErrDataOutOfRange
	}

	c := new(big.Int).Exp(m, big.NewInt(int64(pubKey.E)), pubKey.N)

	// Left-pad to the modulus size: c.Bytes() drops the leading 0x00.
	out := c.FillBytes(make([]byte, pubKey.Size()))

	// Block layout: 00 01 PS 00 D, where PS is at least 8 bytes of 0xff.
	if len(out) < 2+minPaddingLen+1 || out[0] != 0x00 || out[1] != 0x01 {
		return nil, ErrPKCS1InvalidPadding
	}

	sep := 2
	for sep < len(out) && out[sep] == 0xff {
		sep++
	}

	if sep == len(out) || out[sep] != 0x00 || sep-2 < minPaddingLen {
		return nil, ErrPKCS1InvalidPadding
	}

	return out[sep+1:], nil
}
