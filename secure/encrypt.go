// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package secure

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	"github.com/pkg/errors"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

const (
	EncryptKeyLen = 32
	EncryptIVLen  = 32
)

var errInvalidEncrypt = errors.New("invalid encryption param")

// 创建加密器
func CreateCryptor(method string) (cipher.BlockCryptor, error) {
	var key = make([]byte, EncryptKeyLen)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	var iv = make([]byte, EncryptIVLen)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	encrypt := cipher.NewCrypt(method, key, iv)
	return encrypt, nil
}

func EncryptCryptor(crypt cipher.BlockCryptor, pubKey *rsa.PublicKey) (key, iv []byte, err error) {
	if key, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, crypt.Key(), nil); err != nil {
		return
	}
	if iv, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, crypt.IV(), nil); err != nil {
		return
	}
	return
}

// 创建解密器
func DecryptCryptor(method string, encryptedKey, encryptedIV []byte, priKey *rsa.PrivateKey) (cipher.BlockCryptor, error) {
	if method == "" {
		return nil, nil
	}
	if len(encryptedKey) == 0 || len(encryptedIV) == 0 {
		return nil, errInvalidEncrypt
	}
	key, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priKey, encryptedKey, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	iv, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priKey, encryptedIV, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	decrypt := cipher.NewCrypt(method, key, iv)
	return decrypt, nil
}

//
func SignEncryptSignature(method string, encrypt cipher.BlockCryptor, priKey *rsa.PrivateKey) ([]byte, error) {
	if method == "" {
		return nil, nil
	}
	hash := sha256.New()
	hash.Write([]byte(method))
	hash.Write(encrypt.Key())
	hash.Write(encrypt.IV())
	var digest = hash.Sum(nil)
	return rsa.SignPSS(rand.Reader, priKey, crypto.SHA256, digest, nil)
}

//
func VerifyEncryptSignature(method string, signature []byte, encrypt cipher.BlockCryptor, pubKey *rsa.PublicKey) error {
	if method == "" {
		return nil
	}
	hash := sha256.New()
	hash.Write([]byte(method))
	hash.Write(encrypt.Key())
	hash.Write(encrypt.IV())
	var digest = hash.Sum(nil)
	return rsa.VerifyPSS(pubKey, crypto.SHA256, digest, signature, nil)
}
