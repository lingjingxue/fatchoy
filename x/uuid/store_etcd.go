// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var (
	ErrCannotPutEtcd = errors.New("cannot put counter to etcd")
)

// 使用etcd的key的版本号自增实现
type EtcdStore struct {
	key    string           // 使用的key
	lastId int64            // 保存最近一次生成的ID
	ctx    context.Context  // context对象
	cli    *clientv3.Client //
}

func NewEtcdStore(ctx context.Context, cli *clientv3.Client, key string) Storage {
	return &EtcdStore{
		key: key,
		cli: cli,
		ctx: ctx,
	}
}

func (s *EtcdStore) Close() error {
	return nil
}

func (s *EtcdStore) Incr() (int64, error) {
	rev, err := s.doIncr()
	if err != nil {
		return 0, err
	}
	if s.lastId != 0 && s.lastId >= rev {
		return 0, ErrIDOutOfRange
	}
	s.lastId = rev
	return rev, nil
}

func (s *EtcdStore) doIncr() (int64, error) {
	resp, err := s.putKey()
	if err != nil {
		return 0, err
	}
	// 没有prevKv表明还没有set过这个key，使用初始版本号1
	if resp.PrevKv == nil {
		return 1, nil
	}
	// 否则使用版本号+1
	rev := resp.PrevKv.Version
	return rev + 1, nil
}

func (s *EtcdStore) putKey() (*clientv3.PutResponse, error) {
	// 最多3秒延迟
	ctx, cancel := context.WithTimeout(s.ctx, time.Second*OpTimeout)
	defer cancel()

	// key的value暂时设置为时间戳
	value := strconv.FormatInt(time.Now().Unix(), 10)

	// 内部grpc会重试请求，见 etcd/clientv3/retry_interceptor.go
	resp, err := s.cli.Put(ctx, s.key, value, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}
	return resp, nil
}
