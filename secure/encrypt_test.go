// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package secure

import (
	"bytes"
	"crypto/rsa"
	"testing"

	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

func createRSAKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	prikey, err := cipher.LoadRSAPrivateKey(cipher.RSATestPrivateKey)
	if err != nil {
		t.Fatalf("%v", err)
	}
	pubkey, err := cipher.LoadRSAPublicKey(cipher.RSATestPublicKey)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return prikey, pubkey
}

func TestCryptor(t *testing.T) {
	var method = "aes-192"
	var message = "a quick brown fox jumps over the lazy dog"
	var encrypted = []byte(message)
	var decrypted = make([]byte, len(encrypted))

	prikey, pubkey := createRSAKeyPair(t)

	encrypt, err := CreateCryptor(method)
	if err != nil {
		t.Fatalf("%v", err)
	}
	encrypted = encrypt.Encrypt(encrypted)
	encryptedKey, encryptedIV, err := EncryptCryptor(encrypt, pubkey)
	if err != nil {
		t.Fatalf("%v", err)
	}

	decrypt, err := DecryptCryptor(method, encryptedKey, encryptedIV, prikey)
	if err != nil {
		t.Fatalf("%v", err)
	}
	decrypted = decrypt.Decrypt(encrypted)

	if !bytes.Equal(encrypted, decrypted) {
		t.Logf("not equal after decryption")
	}
}
