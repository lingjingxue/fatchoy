// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	mathrand "math/rand"
	"testing"
	"time"
)

func init() {
	mathrand.Seed(time.Now().UnixNano())
}

// Run command below to generate test key files:
// 	openssl genrsa -out rsa_prikey.pem 1024
// 	openssl rsa -in rsa_prikey.pem -pubout -out rsa_pubkey.pem

func TestRSADecrypt(t *testing.T) {
	prikey, err := LoadRSAPrivateKey(RSATestPrivateKey)
	if err != nil {
		t.Fatalf("load private key: %v", err)
	}
	pubkey, err := LoadRSAPublicKey(RSATestPublicKey)
	if err != nil {
		t.Fatalf("load public key: %v", err)
	}
	var maxSize = MaxEncryptSize(pubkey)
	var data = []byte("a quick brown fox jumps over the lazy dog")
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubkey, data, nil)
	if err != nil {
		t.Fatalf("RSAEncrypt: %v, %d/%d", err, len(data), maxSize)
	}
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, prikey, encrypted, nil)
	if err != nil {
		t.Fatalf("RSADecrypt: %v", err)
	}
	if !bytes.Equal(data, decrypted) {
		t.Fatalf("data not equal after encryption/decription")
	}
	t.Logf("RSA encryption OK")
}

func randBytes(length int) []byte {
	if length <= 0 {
		return nil
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		ch := uint8(mathrand.Int31() % 0xFF)
		result[i] = ch
	}
	return result
}

func cloneBytes(data []byte) []byte {
	newdata := make([]byte, len(data))
	copy(newdata, data)
	return newdata
}
