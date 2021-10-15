// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	UUIDCollection = "uuid"
)

// 计数器
type Counter struct {
	Label string `bson:"label"` // 识别符
	Count int64  `bson:"count"` // 计数器
	Step  int32  `bson:"step"`  //
}

type MongoStore struct {
	uri   string          // 连接uri
	db    string          // DB名称
	label string          // 识别符
	step  int32           // ID递增步长
	ctx   context.Context // context对象
	cli   *mongo.Client   //
	last  int64           // 保存最近一次生成的ID
}

func NewMongoDBStore(ctx context.Context, uri, db, label string, step int32) Storage {
	if step <= 0 {
		step = DefaultSeqStep
	}
	store := &MongoStore{
		ctx:   ctx,
		uri:   uri,
		db:    db,
		label: label,
		step:  step,
	}
	if err := store.makeClient(); err != nil {
		log.Panicf("%v", err)
	}
	return store
}

func (s *MongoStore) makeClient() error {
	ctx, cancel := context.WithTimeout(s.ctx, time.Second*5)
	defer cancel()

	clientOpts := options.Client().ApplyURI(s.uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}
	if err = client.Ping(ctx, nil); err != nil {
		return err
	}
	s.cli = client
	return nil
}

func (s *MongoStore) Close() error {
	s.cli = nil
	return nil
}

func (s *MongoStore) Incr() (int64, error) {
	var ctr = &Counter{
		Label: s.label,
		Step:  s.step,
	}

	// 最多3秒延迟
	ctx, cancel := context.WithTimeout(s.ctx, time.Second*300)
	defer cancel()

	// 把counter自增再读取最新的counter
	if err := s.incrementAndLoad(ctx, ctr); err != nil {
		return 0, err
	}
	if s.last != 0 && s.last >= ctr.Count {
		return 0, ErrIDOutOfRange
	}
	s.last = ctr.Count
	return ctr.Count, nil
}

// 把counter自增再读取最新的counter
// see https://docs.mongodb.com/manual/core/write-operations-atomicity/
func (s *MongoStore) incrementAndLoad(ctx context.Context, ctr *Counter) error {
	var filter = bson.M{"label": s.label}
	var update = bson.M{
		"$setOnInsert": bson.M{
			"label": ctr.Label,
			"step":  ctr.Step,
		},
		"$inc": bson.M{"count": 1},
	}
	var opt = options.FindOneAndUpdate()
	opt.SetUpsert(true).SetReturnDocument(options.After)
	result := s.cli.Database(s.db).Collection(UUIDCollection).FindOneAndUpdate(ctx, filter, update, opt)
	if err := result.Err(); err != nil {
		return err
	}
	if err := result.Decode(ctr); err != nil {
		return err
	}
	return nil
}
