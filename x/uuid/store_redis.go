// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	TimeoutSec = 3
	MaxRetry   = 2
)

// 使用redis INCR命令实现
type RedisStore struct {
	addr   string          // redis服务器地址
	key    string          // 使用的key
	ctx    context.Context // context对象
	client *redis.Client   //
	last   int64           // 保存最近一次生成的ID
}

func NewRedisStore(ctx context.Context, addr, key string) Storage {
	var client = redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  7 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolTimeout:  10 * time.Second,
		PoolSize:     3,
		MaxRetries:   3,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		log.Panicf("%v", err)
	}
	return &RedisStore{
		ctx:    ctx,
		addr:   addr,
		key:    key,
		client: client,
	}
}

func (s *RedisStore) Close() error {
	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
	return nil
}

func (s *RedisStore) Incr() (int64, error) {
	cnt, err := s.doIncr()
	if err != nil {
		return 0, err
	}
	if s.last != 0 && s.last >= cnt {
		return 0, ErrIDOutOfRange
	}
	s.last = cnt
	return cnt, nil
}

func (s *RedisStore) doIncr() (int64, error) {
	counter, err := s.client.Do(s.ctx, "INCR", s.key).Int64()
	if err == nil {
		return counter, nil
	}
	return 0, err
}
