// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"log"
	"sync"
	"time"
)

// 一个64位SnowflakeID由以下组成
//     1  bit sign
//     37 bits time units in centi-seconds
//     16 bits machine id
//     10 bits sequence number
//
//  37位时间戳加上2020-01-01的纪元可以表示到2063-06-21
// 		twepoch + (1 << 37) ==> 2063-07-21

const (
	SequenceBits   = 10
	MachineIDBits  = 16
	TimestampShift = SequenceBits + MachineIDBits
	TimeUnitBits   = 63 - TimestampShift

	Twepoch = 1577836800_000_000_000 // custom epoch in nanosecond, (2020-01-01 00:00:00 UTC)
)

func currentTimeUnit() int64 {
	return (time.Now().UTC().UnixNano() - Twepoch) / 10_000_000 // to centi-seconds
}

// 雪花算法生成uuid
type SnowFlake struct {
	sync.Mutex
	seq          int64 // last sequence ID
	lastTimeUnit int64 // last time unit
	lastID       int64 // last generated id
	machineID    int64 // id of this machine(process)
}

func NewSnowFlake(machineId uint16) *SnowFlake {
	var sf = &SnowFlake{
		machineID:    int64(machineId),
		lastTimeUnit: currentTimeUnit(),
	}
	return sf
}

func (sf *SnowFlake) Next() int64 {
	sf.Lock()
	defer sf.Unlock() // open-coded defer has less overhead
	var ts = currentTimeUnit()
	if ts < sf.lastTimeUnit {
		log.Panicf("SnowFlake: time has gone backwards, %v -> %v", ts, sf.lastTimeUnit)
	}
	if ts == sf.lastTimeUnit {
		sf.seq++
		if sf.seq >= (1 << SequenceBits) { // sequence expired, tick to next time unit
			sf.seq = 0
			ts = sf.tilNextTimeUnit(ts)
		}
	} else {
		sf.seq = 0
	}
	sf.lastTimeUnit = ts

	var uuid = (ts << TimestampShift) | (sf.machineID << SequenceBits) | sf.seq
	if uuid <= sf.lastID {
		log.Panicf("SnowFlake: integer overflow, %x -> %x, %x", uuid, sf.lastID, ts)
	}
	sf.lastID = uuid
	return uuid
}

func (sf *SnowFlake) tilNextTimeUnit(ts int64) int64 {
	for {
		time.Sleep(time.Millisecond)
		var now = currentTimeUnit()
		if now > ts {
			return now
		}
	}
}
