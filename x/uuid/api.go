// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"log"

	gouuid "github.com/google/uuid"
)

// 分布式UUID
var (
	seqGen  *SeqIDGen  // 发号器算法
	uuidGen *Snowflake // 雪花算法
)

func Init(workerId uint16, store Storage) error {
	var seq = NewSeqIDGen(store, DefaultSeqStep)
	if err := seq.Init(); err != nil {
		return err
	}
	seqGen = seq
	uuidGen = NewSnowflake(workerId)
	return nil
}

// 生成依赖存储，可用于角色ID
func NextID() int64 {
	return seqGen.MustNext()
}

// 生成的值与时钟有关，通常值比较大，可用于日志ID
func NextUUID() int64 {
	return uuidGen.MustNext()
}

// 生成GUID
func NextGUID() string {
	return gouuid.NewString()
}
