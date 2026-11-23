package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	flow "wholesale.local/orbit/generated/CommandFlow"
	d "wholesale.local/orbit/generated/dafny"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/store"
)

func (s *Service) Process(ctx context.Context, id string) error {
	effects := &commandEffects{ctx: ctx, service: s, id: id}
	defer func() {
		if effects.tx != nil {
			_ = effects.tx.Rollback(ctx)
		}
	}()
	flow.Companion_Default___.ExecuteCommand(effects)
	return effects.err
}

type commandEffects struct {
	ctx     context.Context
	service *Service
	id      string

	tx       pgx.Tx
	command  Command
	payload  []byte
	result   any
	err      error
	unit     *store.Unit
	actor    domain.Actor
	now      time.Time
	cart     domain.Cart
	customer string
	stock    *domain.Stock
	order    domain.Order
	loyalty  domain.Loyalty
	pdf      []byte
	address  domain.ShippingAddress
	lines    []domain.Line
	product  domain.Product
}

var _ flow.Effects = (*commandEffects)(nil)

func (*commandEffects) String() string { return "commandEffects" }

func (e *commandEffects) step(err error) flow.Step {
	e.err = err
	var rejection *Rejection
	switch {
	case err == nil:
		return flow.Companion_Step_.Create_Ok_()
	case errors.Is(err, domain.ErrConflict):
		return flow.Companion_Step_.Create_Conflict_()
	case errors.As(err, &rejection):
		return flow.Companion_Step_.Create_Rejected_()
	default:
		return flow.Companion_Step_.Create_Failed_()
	}
}

func (e *commandEffects) Begin() flow.Step {
	*e = commandEffects{ctx: e.ctx, service: e.service, id: e.id}
	tx, err := e.service.DB.Pool.Begin(e.ctx)
	e.tx = tx
	return e.step(err)
}

func (e *commandEffects) LoadCommand() (flow.Step, d.Sequence) {
	c := &e.command
	err := e.tx.QueryRow(e.ctx, "SELECT id,actor_id,type,payload,status FROM commands WHERE id=$1", e.id).
		Scan(&c.ID, &c.ActorID, &c.Type, &e.payload, &c.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return e.step(nil), d.EmptySeq
	}
	return e.step(err), d.UnicodeSeqOfUtf8Bytes(c.Status)
}

func (e *commandEffects) DecodePayload() flow.Step {
	return e.step(json.Unmarshal(e.payload, &e.command.Payload))
}

func (e *commandEffects) LoadActor() flow.Step {
	c := e.command
	var actor domain.Actor
	if c.ActorID == store.ID("system") {
		actor = domain.Actor{ID: c.ActorID, Role: "system", Name: "system"}
	} else {
		var b []byte
		if err := e.tx.QueryRow(e.ctx, store.ActorQuery+" WHERE a.id=$1", c.ActorID).Scan(&b); err != nil {
			return e.step(err)
		}
		if err := json.Unmarshal(b, &actor); err != nil {
			return e.step(err)
		}
	}
	e.actor = actor
	return e.step(nil)
}

func (e *commandEffects) ReadClock() flow.Step {
	e.unit = store.NewUnit(e.ctx, e.tx)
	return e.step(e.tx.QueryRow(e.ctx, "SELECT clock_timestamp()").Scan(&e.now))
}

func (e *commandEffects) Input() flow.CommandInput {
	p, a := e.command.Payload, e.actor
	return flow.Companion_CommandInput_.Create_CommandInput_(
		d.UnicodeSeqOfUtf8Bytes(e.command.Type), d.UnicodeSeqOfUtf8Bytes(a.Role),
		d.UnicodeSeqOfUtf8Bytes(a.ID), d.UnicodeSeqOfUtf8Bytes(p.CustomerID),
		d.IntOf(len(p.Note)), d.IntOf(len(p.Tracking)), d.IntOf(len(p.RequestedDate)),
		p.Coupon != "", d.IntOfInt64(p.ExpectedVersion), validID(p.ProductID), d.IntOfInt64(p.Delta),
	)
}

func (e *commandEffects) SaveCompletion() flow.Step {
	result, err := json.Marshal(e.result)
	if err != nil {
		return e.step(err)
	}
	tag, err := e.tx.Exec(e.ctx, "UPDATE commands SET status='succeeded',result=$2,error='',finished_at=now() WHERE id=$1 AND status='queued'", e.id, result)
	if err == nil && tag.RowsAffected() != 1 {
		err = domain.ErrConflict
	}
	return e.step(err)
}

func (e *commandEffects) Commit() flow.Step {
	err := e.tx.Commit(e.ctx)
	e.tx = nil
	return e.step(err)
}

func (e *commandEffects) Rollback() flow.Step {
	err := e.tx.Rollback(e.ctx)
	e.tx = nil
	if err == nil {
		return flow.Companion_Step_.Create_Ok_()
	}
	var rejection *Rejection
	if errors.As(e.err, &rejection) {
		return e.step(err)
	}
	return flow.Companion_Step_.Create_Failed_()
}

func (e *commandEffects) SaveRejection() flow.Step {
	_, err := e.service.DB.Pool.Exec(e.ctx, "UPDATE commands SET status='rejected',error=$2,finished_at=now() WHERE id=$1 AND status='queued'", e.id, e.err.Error())
	return e.step(err)
}

func (e *commandEffects) RecordFailure(maxAttempts d.Int) {
	_, _ = e.service.DB.Pool.Exec(e.ctx, `UPDATE commands SET attempts=attempts+1,error=$2,status=CASE WHEN attempts+1 >= $3 THEN 'dead' ELSE status END WHERE id=$1 AND status='queued'`, e.id, e.err.Error(), maxAttempts.Int())
}
