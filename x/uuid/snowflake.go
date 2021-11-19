// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

// 一个64位SnowflakeID由以下组成
//      1 bit sign
//     37 bits time units in centi-seconds(10 msec)
//      1 bit time backwards mask
//     15 bits machine id
//     10 bits sequence number

const (
	SequenceBits   = 10
	MachineIDBits  = 15
	TimeUnitBits   = 37
	MaxSeqID       = (1 << SequenceBits) - 1
	MachineIDMask  = 0x7fff
	TimestampShift = MachineIDBits + SequenceBits
	MaxTimeUnits   = (1 << TimeUnitBits) - 1
	BackwardsMask  = 1 << (MachineIDBits + SequenceBits)

	TimeUnit    = int64(time.Second / 100)        // 厘秒（10毫秒）
	CustomEpoch = int64(1577836800 * time.Second) // 起始纪元 2020-01-01 00:00:00 UTC
)

var (
	ErrClockGoneBackwards = errors.New("clock gone backwards")
	ErrTimeUnitOverflow   = errors.New("clock time unit overflow")
	ErrUUIDIntOverflow    = errors.New("uuid integer overflow")
)

// UUID的生成依赖系统时钟，如果系统时钟被回拨，会有潜在的生成重复ID的情况
// 	1，系统时钟被人为回拨
// 	2，开启了NTP同步且当前系统时钟快了，则NTP会回拨时钟
//  3，UTC闰秒, https://en.wikipedia.org/wiki/Leap_second
//
// 设计中增加了时钟回拨标记位，可以让系统在时钟被回拨时仍正确工作
// 如果需要支持更多的回拨次数，增加此位长度
func currentTimeUnit() int64 {
	return (time.Now().UTC().UnixNano() - CustomEpoch) / TimeUnit // to centi-seconds
}

// sequence expired, tick to next time unit
func waitUntilNextTimeUnit(ts int64) int64 {
	for {
		time.Sleep(time.Millisecond)
		var now = currentTimeUnit()
		if now > ts {
			return now
		}
	}
}

// 雪花ID生成器
type Snowflake struct {
	machineID     int64      // id of this machine(process)
	guard         sync.Mutex //
	seq           int64      // last sequence ID
	lastTimeUnit  int64      // last time unit
	lastID        int64      // last generated id
	backwardsMask int64      // 允许时钟被回拨一次
}

func NewSnowflake(machineId uint16) *Snowflake {
	if machineId == 0 {
		machineId = privateIP4()
		log.Printf("snowflake auto set machine id to %d", machineId)
	}
	var sf = &Snowflake{
		machineID:    int64(machineId) & MachineIDMask,
		lastTimeUnit: currentTimeUnit(),
	}
	return sf
}

// Next generate an ID, panic if time gone backwards or integer overflow
func (sf *Snowflake) Next() (int64, error) {
	sf.guard.Lock()
	defer sf.guard.Unlock()

	var ts = currentTimeUnit()
	if ts > MaxTimeUnits {
		log.Printf("Snowflake: time unit overflow")
		return 0, ErrTimeUnitOverflow
	}
	if ts < sf.lastTimeUnit {
		log.Printf("Snowflake: time has gone backwards")
		if sf.backwardsMask > 0 {
			return 0, ErrClockGoneBackwards
		}
		sf.backwardsMask = BackwardsMask
	}
	if ts == sf.lastTimeUnit {
		sf.seq++
		if sf.seq > MaxSeqID {
			sf.seq = 0
			ts = waitUntilNextTimeUnit(ts)
		}
	} else {
		sf.seq = 0
	}
	sf.lastTimeUnit = ts

	var uuid = (ts << TimestampShift) | (sf.machineID << SequenceBits) | sf.seq
	if uuid <= sf.lastID {
		log.Printf("Snowflake: uuid int64 overflow")
		return 0, ErrUUIDIntOverflow
	}
	sf.lastID = uuid
	return uuid, nil
}

func (sf *Snowflake) MustNext() int64 {
	if n, err := sf.Next(); err != nil {
		panic(err)
	} else {
		return n
	}
}

// lower 16-bits of IPv4
func privateIP4() uint16 {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return 0
	}
	var isPrivateIPv4 = func(ip net.IP) bool {
		return ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 &&
			ip[1] < 32) || ip[0] == 192 && ip[1] == 168
	}
	for _, a := range addr {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}
		var ip = ipnet.IP.To4()
		if ip != nil && isPrivateIPv4(ip) {
			return uint16(ip[2])<<8 + uint16(ip[3])
		}
	}
	return 0
}
