// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

const (
	VersionV1    = 1
	V1HeaderSize = 14 // 包头大小(包含长度）
)

//  协议头，little endian表示，len包含header和body
//       ---------------------------------
// field | len | flag |  seq | cmd | crc |
//       ---------------------------------
// bytes |  3  |   1  |   2  |  4  |  4  |

type V1Header []byte
