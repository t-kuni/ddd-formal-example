package app

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	plan "wholesale.local/orbit/generated/CommandFlow"
	d "wholesale.local/orbit/generated/dafny"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/store"
)

func (s *Service) Process(ctx context.Context, id string) error {
	effects := &commandEffects{ctx: ctx, service: s, id: id}
	defer effects.Rollback()
	plan.Companion_Default___.ExecuteCommand(effects)
	return effects.err
}

type commandEffects struct {
	ctx     context.Context
	service *Service
	id      string

	tx      pgx.Tx
	work    pgx.Tx
	attempt commandAttempt
	raw     []byte
	result  map[string]any
	err     error
}

var _ plan.Effects = (*commandEffects)(nil)

func (*commandEffects) String() string { return "commandEffects" }

func (e *commandEffects) status(err error) plan.Status {
	e.err = err
	var denied *Rejection
	switch {
	case err == nil:
		return plan.Companion_Status_.Create_Ok_()
	case errors.Is(err, domain.ErrConflict):
		return plan.Companion_Status_.Create_Conflict_()
	case errors.As(err, &denied):
		return plan.Companion_Status_.Create_Rejected_()
	default:
		return plan.Companion_Status_.Create_Failed_()
	}
}

func (e *commandEffects) Begin() plan.Status {
	*e = commandEffects{ctx: e.ctx, service: e.service, id: e.id}
	var err error
	e.tx, err = e.service.DB.Pool.Begin(e.ctx)
	if err == nil {
		e.attempt.unit = store.NewUnit(e.ctx, e.tx)
	}
	return e.status(err)
}

func (e *commandEffects) LockCommand() (plan.Status, d.Sequence) {
	c := &e.attempt.command
	err := e.tx.QueryRow(e.ctx, "SELECT id,actor_id,type,payload,status FROM commands WHERE id=$1 FOR UPDATE", e.id).
		Scan(&c.ID, &c.ActorID, &c.Type, &e.raw, &c.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return e.status(nil), d.EmptySeq
	}
	return e.status(err), textValue(c.Status)
}

func (e *commandEffects) ReadInput() (plan.Status, plan.CommandInput) {
	a := &e.attempt
	c := &a.command
	if err := json.Unmarshal(e.raw, &c.Payload); err != nil {
		return e.status(err), plan.CommandInput{}
	}
	if c.ActorID == store.ID("system") {
		a.actor = domain.Actor{ID: c.ActorID, Role: "system", Name: "system"}
	} else {
		var raw []byte
		if err := e.tx.QueryRow(e.ctx, store.ActorQuery+" WHERE a.id=$1", c.ActorID).Scan(&raw); err != nil {
			return e.status(err), plan.CommandInput{}
		}
		if err := json.Unmarshal(raw, &a.actor); err != nil {
			return e.status(err), plan.CommandInput{}
		}
	}
	if err := e.tx.QueryRow(e.ctx, "SELECT clock_timestamp()").Scan(&a.now); err != nil {
		return e.status(err), plan.CommandInput{}
	}
	return e.status(nil), a.input()
}

func (e *commandEffects) Read(request plan.ReadRequest) (plan.Status, plan.Snapshot) {
	snapshot, err := e.attempt.read(request)
	return e.status(err), snapshot
}

func (e *commandEffects) ReadCoupon(code d.Sequence) (plan.Status, plan.Coupon) {
	var percent int64
	err := e.tx.QueryRow(e.ctx, "SELECT percent FROM coupons WHERE code=$1", code.VerbatimString(false)).Scan(&percent)
	if errors.Is(err, pgx.ErrNoRows) {
		return e.status(nil), plan.Companion_Coupon_.Create_UnknownCoupon_()
	}
	if err != nil {
		return e.status(err), plan.Coupon{}
	}
	return e.status(nil), plan.Companion_Coupon_.Create_CouponRate_(d.IntOfInt64(percent))
}

func (e *commandEffects) ReadDependencies(reads plan.Dependencies) plan.Status {
	return e.status(e.attempt.readDependencies(reads))
}

func (e *commandEffects) Apply(decision plan.CommandPlan) plan.Status {
	// Keep the Command row lock if applying the plan is rejected.
	var err error
	e.work, err = e.tx.Begin(e.ctx)
	if err != nil {
		return e.status(err)
	}
	e.attempt.unit.Tx = e.work
	e.result, err = e.attempt.apply(decision)
	return e.status(err)
}

func (e *commandEffects) SetRejection(reason d.Sequence) {
	e.err = reject(reason.VerbatimString(false))
}

func (e *commandEffects) RecordRejection() plan.Status {
	if e.work != nil {
		if err := e.work.Rollback(e.ctx); err != nil {
			return e.status(err)
		}
		e.work = nil
	}
	_, err := e.tx.Exec(e.ctx, "UPDATE commands SET status='rejected',error=$2,finished_at=now() WHERE id=$1", e.id, e.err.Error())
	return e.status(err)
}

func (e *commandEffects) RecordSuccess() plan.Status {
	if err := e.work.Commit(e.ctx); err != nil {
		return e.status(err)
	}
	e.work = nil
	body, err := json.Marshal(e.result)
	if err != nil {
		return e.status(err)
	}
	tag, err := e.tx.Exec(e.ctx, "UPDATE commands SET status='succeeded',result=$2,error='',finished_at=now() WHERE id=$1 AND status='queued'", e.id, body)
	if err == nil && tag.RowsAffected() != 1 {
		err = domain.ErrConflict
	}
	return e.status(err)
}

func (e *commandEffects) Commit() plan.Status {
	err := e.tx.Commit(e.ctx)
	e.tx = nil
	return e.status(err)
}

func (e *commandEffects) Rollback() {
	if e.tx != nil {
		_ = e.tx.Rollback(e.ctx)
		e.tx = nil
	}
}

func (e *commandEffects) RecordFailure(maxAttempts d.Int) {
	_, _ = e.service.DB.Pool.Exec(e.ctx, `UPDATE commands SET attempts=attempts+1,error=$2,
		status=CASE WHEN attempts+1 >= $3 THEN 'dead' ELSE status END WHERE id=$1 AND status='queued'`,
		e.id, e.err.Error(), maxAttempts.Int())
}
