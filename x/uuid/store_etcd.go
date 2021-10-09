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
	key string //
	cli *clientv3.Client
}

func NewEtcdStore(cli *clientv3.Client, key string) Storage {
	return &EtcdStore{
		key: key,
		cli: cli,
	}
}

func (s *EtcdStore) Close() error {
	return nil
}

func (s *EtcdStore) putKey() (*clientv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	value := strconv.FormatInt(time.Now().Unix(), 10)
	resp, err := s.cli.Put(ctx, s.key, value, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *EtcdStore) Incr() (int64, error) {
	resp, err := s.putKey()
	if err != nil {
		return 0, err
	}
	if resp.PrevKv == nil {
		return 1, nil
	}
	rev := resp.PrevKv.Version
	return rev + 1, nil
}
