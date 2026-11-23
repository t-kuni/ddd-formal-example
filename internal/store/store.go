package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"wholesale.local/orbit/internal/domain"
)

//go:embed schema.sql
var schema string

type Store struct{ Pool *pgxpool.Pool }

func Open(ctx context.Context, url string) (*Store, error) {
	p, e := pgxpool.New(ctx, url)
	if e != nil {
		return nil, e
	}
	return &Store{p}, p.Ping(ctx)
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
	e := u.Tx.QueryRow(u.Ctx, "SELECT version,body FROM aggregates WHERE kind=$1 AND id=$2", kind, id).Scan(&v, &b)
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
	b, e := json.Marshal(value)
	if e != nil {
		return e
	}
	v, known := u.Versions[kind+id]
	if !known {
		return errors.New("aggregate must be loaded before save")
	}
	if v == 0 {
		tag, e := u.Tx.Exec(u.Ctx, "INSERT INTO aggregates(kind,id,version,body) VALUES($1,$2,1,$3) ON CONFLICT DO NOTHING", kind, id, b)
		if e != nil {
			return e
		}
		if tag.RowsAffected() != 1 {
			return domain.ErrConflict
		}
	} else {
		tag, e := u.Tx.Exec(u.Ctx, "UPDATE aggregates SET version=version+1,body=$4 WHERE kind=$1 AND id=$2 AND version=$3", kind, id, v, b)
		if e != nil {
			return e
		}
		if tag.RowsAffected() != 1 {
			return domain.ErrConflict
		}
	}
	u.Versions[kind+id] = v + 1
	_, e = u.Tx.Exec(u.Ctx, "INSERT INTO projection_events(id,kind,aggregate_id,version,body) VALUES($1,$2,$3,$4,$5)", uuid.NewString(), kind, id, v+1, b)
	return e
}
func (s *Store) Project(ctx context.Context) error {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, "SELECT id,kind,aggregate_id,version,body FROM projection_events ORDER BY created_at,id LIMIT 200")
	if e != nil {
		return e
	}
	type event struct {
		id, kind, aid string
		v             int64
		b             []byte
	}
	var events []event
	for rows.Next() {
		var x event
		if e = rows.Scan(&x.id, &x.kind, &x.aid, &x.v, &x.b); e != nil {
			rows.Close()
			return e
		}
		events = append(events, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, x := range events {
		if _, e = tx.Exec(ctx, `INSERT INTO projections(kind,id,version,body) VALUES($1,$2,$3,$4) ON CONFLICT(kind,id) DO UPDATE SET version=EXCLUDED.version,body=EXCLUDED.body WHERE projections.version<EXCLUDED.version`, x.kind, x.aid, x.v, x.b); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "DELETE FROM projection_events WHERE id=$1", x.id); e != nil {
			return e
		}
	}
	return tx.Commit(ctx)
}
