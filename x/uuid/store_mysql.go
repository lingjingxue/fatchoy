// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var ErrNoRowsAffected = errors.New("no rows affected")

type MySQLStore struct {
	dsn    string          // 连接DSN
	table  string          // table名
	label  string          // 唯一名称
	step   int32           // 步长
	db     *sql.DB         //
	ctx    context.Context // context对象
	guard  sync.Mutex      //
	lastId int64           // 保存最近一次生成的ID
}

func NewMySQLStore(ctx context.Context, dsn, table, label string, step int) Storage {
	store := &MySQLStore{
		ctx:   ctx,
		dsn:   dsn,
		step:  int32(step),
		table: table,
		label: label,
	}
	if err := store.setupInit(); err != nil {
		log.Panicf("%v", err)
	}
	return store
}

func (s *MySQLStore) setupInit() error {
	s.guard.Lock()
	defer s.guard.Unlock()

	ctx, cancel := context.WithTimeout(s.ctx, time.Second*OpTimeout)
	defer cancel()

	if err := s.createConn(ctx); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Defer a rollback in case anything fails
	defer tx.Rollback()

	if err = s.createTable(ctx, tx); err != nil {
		return err
	}
	var counter int64
	if counter, err = s.loadSeqID(ctx, tx); err != nil {
		if err == sql.ErrNoRows {
			err = s.insertRecord(ctx, tx)
		} else {
			return err
		}
	}
	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}
	s.lastId = counter
	return nil
}

func (s *MySQLStore) createConn(ctx context.Context) error {
	db, err := sql.Open("mysql", s.dsn)
	if err != nil {
		return err
	}
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *MySQLStore) createTable(ctx context.Context, tx *sql.Tx) error {
	var stmt = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n"+
		"\t `id` bigint NOT NULL AUTO_INCREMENT, \n"+
		"\t `label` varchar(50) NOT NULL, \n"+
		"\t `step` int NOT NULL, \n"+
		"\t `seq_id` bigint NOT NULL, \n"+
		"\t PRIMARY KEY (`id`) USING BTREE, \n"+
		"\t UNIQUE INDEX `idx_label` (`label`) USING BTREE \n"+
		") COLLATE='utf8mb4_general_ci' ENGINE=InnoDB", s.table)
	if _, err := tx.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

func (s *MySQLStore) insertRecord(ctx context.Context, tx *sql.Tx) error {
	var stmt = fmt.Sprintf("INSERT INTO `%s`(`label`, `step`, `seq_id`) VALUES(?, ?, 1)", s.table)
	result, err := tx.ExecContext(ctx, stmt, s.label, s.step)
	if err != nil {
		return err
	}
	n, er := result.RowsAffected()
	if er != nil {
		return er
	}
	if n != 1 {
		return ErrNoRowsAffected
	}
	return nil
}

func (s *MySQLStore) Incr() (int64, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	ctx, cancel := context.WithTimeout(s.ctx, time.Second*OpTimeout)
	defer cancel()

	counter, err := s.incrSeqID(ctx)
	if err != nil {
		return 0, err
	}

	if s.lastId != 0 && s.lastId >= counter {
		return 0, ErrIDOutOfRange
	}
	s.lastId = counter
	return counter, nil
}

// see https://dev.mysql.com/doc/refman/8.0/en/information-functions.html#function_last-insert-id
func (s *MySQLStore) incrSeqID(ctx context.Context) (int64, error) {
	var stmt = fmt.Sprintf("UPDATE `%s` SET `seq_id` = LAST_INSERT_ID(`seq_id`) + 1 WHERE `label`=? LIMIT 1", s.table)
	result, err := s.db.ExecContext(ctx, stmt, s.label)
	if err != nil {
		return 0, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, ErrNoRowsAffected
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastId + 1, nil
}

func (s *MySQLStore) loadSeqID(ctx context.Context, tx *sql.Tx) (int64, error) {
	var seqId int64
	var stmt = fmt.Sprintf("SELECT `seq_id` FROM `%s` WHERE `label`=?", s.table)
	if err := tx.QueryRowContext(ctx, stmt, s.label).Scan(&seqId); err != nil {
		return 0, err
	}
	return seqId, nil
}

func (s *MySQLStore) Close() error {
	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}
	return nil
}
