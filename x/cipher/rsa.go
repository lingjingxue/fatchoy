// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

// 解析公钥文件
func LoadRSAPublicKey(pemFile string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("incorrect public key file")
	}
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected key type %s", block.Type)
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key.(*rsa.PublicKey), nil
}

// 解析私钥文件
func LoadRSAPrivateKey(pemFile string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("incorrect private key file")
	}
	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected key type %s", block.Type)
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// 最大加密内容大小
func MaxEncryptSize(pubkey *rsa.PublicKey) int {
	var k = pubkey.Size()
	var hash = sha256.New()
	return k - 2*hash.Size() - 2
}
