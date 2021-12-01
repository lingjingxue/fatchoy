// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

var (
	registry = make(map[string]Encoder)
)

func Register(v Encoder) {
	registry[v.Name()] = v
}

func GetEncoder(name string) Encoder {
	return registry[name]
}
