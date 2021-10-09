// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"log"
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	TimeoutSec = 3
	MaxRetry   = 2
)

// 使用redis INCR命令实现
type RedisStore struct {
	addr string     // redis服务器地址
	key  string     //
	conn redis.Conn //
}

func NewRedisStore(addr, key string) Storage {
	store := &RedisStore{
		addr: addr,
		key:  key,
	}
	if err := store.createConn(TimeoutSec); err != nil {
		log.Panicf("%v", err)
	}
	return store
}

func (s *RedisStore) Close() error {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	return nil
}

func (s *RedisStore) createConn(timeout int32) error {
	conn, err := redis.Dial("tcp", s.addr,
		redis.DialConnectTimeout(time.Second*time.Duration(timeout)),
		redis.DialReadTimeout(time.Second*TimeoutSec),
		redis.DialWriteTimeout(time.Second*TimeoutSec),
	)
	if err != nil {
		return err
	}
	pong, err := redis.String(conn.Do("PING"))
	if err != nil || pong != "PONG" {
		return err
	}
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.conn = conn
	return nil
}

func (s *RedisStore) Incr() (int64, error) {
	counter, err := s.doIncr(MaxRetry)
	if err != nil {
		return 0, err
	}
	return counter, nil
}

func (s *RedisStore) doIncr(retry int) (int64, error) {
	counter, err := redis.Int64(s.conn.Do("INCR", s.key))
	if err == nil {
		return counter, nil
	}
	if retry == 0 {
		return 0, err
	}
	if er, ok := err.(*net.OpError); ok {
		if e := s.tryReconnect(er); e == nil {
			return s.doIncr(retry - 1)
		}
	}
	return 0, err
}

func (s *RedisStore) tryReconnect(err *net.OpError) error {
	if err.Op == "write" || err.Op == "read" {
		if er := s.createConn(TimeoutSec); er != nil {
			return er
		}
		return nil
	}
	return err
}
