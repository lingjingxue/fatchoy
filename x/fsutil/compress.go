// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fsutil

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"io"
)

// 压缩内容
func CompressBytes(data []byte) ([]byte, error) {
	var buf = &bytes.Buffer{}
	w, err := zlib.NewWriterLevel(buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(data); err != nil {
		if er := w.Close(); er != nil {
			return nil, er
		}
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 解压内容
func UncompressBytes(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var buf = &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		if er := r.Close(); er != nil {
			return nil, er
		}
		return nil, err
	}
	if err = r.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
