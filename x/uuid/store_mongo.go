// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
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
	cli       *mongo.Client
	parentCtx context.Context
	db        string
	label     string
	step      int32
}

func NewMongoDBStore(cli *mongo.Client, ctx context.Context, db, label string, step int32) Storage {
	if ctx == nil {
		ctx = context.TODO()
	}
	if step <= 0 {
		step = DefaultSeqStep
	}
	return &MongoStore{
		cli:       cli,
		parentCtx: ctx,
		db:        db,
		label:     label,
		step:      step,
	}
}

func (s *MongoStore) Close() error {
	return nil
}

func (s *MongoStore) Incr() (int64, error) {
	var counter = &Counter{
		Label: s.label,
		Step:  s.step,
	}
	ctx, cancel := context.WithTimeout(s.parentCtx, time.Second*2)
	defer cancel()
	// 把counter自增再读取最新的counter
	if err := s.incrementAndLoad(ctx, counter); err != nil {
		return 0, err
	}
	return counter.Count, nil
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
