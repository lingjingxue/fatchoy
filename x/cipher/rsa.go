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

func LoadRSAPublicKeyFile(pemFile string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	return LoadRSAPublicKey(data)
}

// 解析公钥文件
func LoadRSAPublicKey(data []byte) (*rsa.PublicKey, error) {
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

func LoadRSAPrivateKeyFile(pemFile string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	return LoadRSAPrivateKey(data)
}

// 解析私钥文件
func LoadRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
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

var RSATestPublicKey = []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC0wugAtwSpvDgiYKi6GC5390KY
Qy4bAC2jBO13zVW5aQ83WPUHyvhXnj1N1xujGHMJyNGwEYA9voxmPxyYn83D4cRM
Bga/GaJtLzbJwakpFMaEzUtIq8bCgPSTXtxuUx+spw6G/yl6MxO9O+RhScDrQPmp
jvB4Z/u0Dl5tdwJPqQIDAQAB
-----END PUBLIC KEY-----`)

var RSATestPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQC0wugAtwSpvDgiYKi6GC5390KYQy4bAC2jBO13zVW5aQ83WPUH
yvhXnj1N1xujGHMJyNGwEYA9voxmPxyYn83D4cRMBga/GaJtLzbJwakpFMaEzUtI
q8bCgPSTXtxuUx+spw6G/yl6MxO9O+RhScDrQPmpjvB4Z/u0Dl5tdwJPqQIDAQAB
AoGATEj9NGAIvcFLR2bXjkHqSoK1PiEL8iUvHV9VAHxNs0PdQhRuxG0qRX/oi1M+
vKPy2KxBojagkm46PmRgIyE96rkI94boLKfctuMVsqg22GQDtcvBuSVrYPNfgDLw
1EbzQihFqgxO/QYnuakn7GAE4N9x1R5gAQr7Wy00aekhHkkCQQDlZgzAYAyFtZA4
A6NOGGPVM8/FLYwUZVyb9jh1uXJiOEj1j7p5bJhUrXRRduJ+Z2t4OP993OprTV86
slO/QVkvAkEAybkBY2JIK+nDxdxCEmbMcQRolTL/l/MQayBF0lbOVHb5svDdpWbm
q9Y6PwfVK8jbp8bJWYovDJ2wEQF3d0R+pwJAH+wzmhHDrFe32hOnhhaezeyH3UiZ
Vb1FRe7drIRCBqkOfh2iNYOHL0F0DmIc4rpBmllUNI+pj4UU23Y1cUgGwQJALBhq
+0Siria9iuTo9IjQK+xgyCyLvrV9Y018tcwP8lrHnpwUd3GU/v8nYFvf92BC09wa
a55PRpy5vh3p9YJdhQJBAKtEEaC8EB7ghXvvo1O+MJotd+EqO330JsLTUf0GcsjI
zV/yJu951ELuzMZTfemh6l8stjjDYlRvZVPbjwrZP8g=
-----END RSA PRIVATE KEY-----`)
