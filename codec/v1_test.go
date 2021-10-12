// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"bytes"
	"testing"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/secure"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
	"gopkg.in/qchencc/fatchoy.v1/x/strutil"
)

func isEqualPacket(t *testing.T, a, b fatchoy.IPacket) bool {
	if a.Command() != b.Command() || (a.Seq() != b.Seq()) {
		return false
	}
	data1, _ := a.EncodeBodyToBytes()
	data2, _ := b.EncodeBodyToBytes()
	if len(data1) > 0 && len(data2) > 0 {
		if !bytes.Equal(data1, data2) {
			println("msg a md5sum", Md5Sum(data1))
			println("msg b md5sum", Md5Sum(data2))
			t.Fatalf("packet not equal, %v != %v", a, b)
			return false
		}
	}
	return true
}

func newTestPacket(bodyLen int) fatchoy.IPacket {
	var pkt testPacket
	pkt.flag = 0
	pkt.command = 1234
	pkt.seq = 2012
	if bodyLen > 0 {
		s := strutil.RandString(bodyLen)
		pkt.SetBodyBytes([]byte(s))
	}
	return &pkt
}

func testProtoCodec(t *testing.T, size int, msgSent fatchoy.IPacket) {
	var encoded bytes.Buffer
	encrypt, _ := secure.CreateCryptor("aes-192")
	decrypt := cipher.NewCrypt("aes-192", encrypt.Key(), encrypt.IV())
	// 如果加密方式是原地加密，会导致packet的body是加密后的内容
	clone := append([]byte(nil), msgSent.BodyAsBytes()...)
	if _, err := Marshal(&encoded, msgSent, encrypt, VersionV1); err != nil {
		t.Fatalf("Encode with size %d: %v", size, err)
	}
	msgSent.SetBodyBytes(nil)
	var msgRecv testPacket
	if _, err := Unmarshal(&encoded, &msgRecv, decrypt); err != nil {
		t.Fatalf("Decode with size %d: %v", size, err)
	}
	msgSent.SetBodyBytes(clone)
	if !isEqualPacket(t, msgSent, &msgRecv) {
		t.Fatalf("Encode and Decode not equal: %d\n%v\n%v", size, msgSent, msgRecv)
	}
}

func TestCodecEncode(t *testing.T) {
	var sizeList = []int{404, 1012, 2014, 4018, 8012, 40487, 1024 * 31} //
	for _, n := range sizeList {
		var pkt = newTestPacket(n)
		testProtoCodec(t, n, pkt)
	}
}

func BenchmarkCodecMarshal(b *testing.B) {
	b.StopTimer()
	var size = 4096
	b.Logf("benchmark with message size %d\n", size)
	var msg = newTestPacket(int(size))
	b.StartTimer()

	var buf bytes.Buffer
	if _, err := Marshal(&buf, msg, nil, VersionV1); err != nil {
		b.Logf("Encode: %v", err)
	}
	var msg2 testPacket
	if _, err := Marshal(&buf, &msg2, nil, VersionV1); err != nil {
		b.Logf("Decode: %v", err)
	}
}
