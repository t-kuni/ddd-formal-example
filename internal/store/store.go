package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"wholesale.local/orbit/internal/domain"
)

//go:embed schema.sql
var schema string

type Store struct {
	Pool *pgxpool.Pool
	orm  *gorm.DB
}

func Open(ctx context.Context, url string) (*Store, error) {
	p, e := pgxpool.New(ctx, url)
	if e != nil {
		return nil, e
	}
	if e = p.Ping(ctx); e != nil {
		p.Close()
		return nil, e
	}
	sqlDB := stdlib.OpenDBFromPool(p)
	orm, e := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true})
	if e != nil {
		sqlDB.Close()
		p.Close()
		return nil, e
	}
	return &Store{Pool: p, orm: orm}, nil
}
func (s *Store) Migrate(ctx context.Context) error { _, e := s.Pool.Exec(ctx, schema); return e }

type Unit struct {
	Tx       pgx.Tx
	Ctx      context.Context
	Versions map[string]int64
}

func NewUnit(ctx context.Context, tx pgx.Tx) *Unit { return &Unit{tx, ctx, map[string]int64{}} }
func (u *Unit) Load(kind, id string, dst domain.RestorableAggregate) error {
	if string(dst.Kind()) != kind {
		return errors.New("aggregate kind mismatch")
	}
	var b []byte
	var v int64
	e := u.Tx.QueryRow(u.Ctx, aggregateQueries[dst.Kind()], id).Scan(&v, &b)
	if errors.Is(e, pgx.ErrNoRows) {
		u.Versions[kind+id] = 0
		return nil
	}
	if e != nil {
		return e
	}
	u.Versions[kind+id] = v
	return json.Unmarshal(b, dst)
}
func (u *Unit) Save(kind, id string, value domain.Aggregate) error {
	if string(value.Kind()) != kind {
		return errors.New("aggregate kind mismatch")
	}
	key := kind + id
	version, known := u.Versions[key]
	if !known {
		return errors.New("aggregate must be loaded before save")
	}
	if err := u.save(id, version, value); err != nil {
		return err
	}
	u.Versions[key] = version + 1
	return nil
}
