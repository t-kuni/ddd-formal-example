package app

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	plan "wholesale.local/orbit/generated/CommandFlow"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/store"
)

const commandAttemptLimit = 5
const commandFailureLimit = 8

func (s *Service) Process(ctx context.Context, id string) error {
	for range commandAttemptLimit {
		err := s.processCommand(ctx, id)
		if errors.Is(err, domain.ErrConflict) {
			continue
		}
		if err != nil {
			_, _ = s.DB.Pool.Exec(ctx, `UPDATE commands SET attempts=attempts+1,error=$2,
				status=CASE WHEN attempts+1 >= $3 THEN 'dead' ELSE status END WHERE id=$1 AND status='queued'`,
				id, err.Error(), commandFailureLimit)
		}
		return err
	}
	return domain.ErrConflict
}

func (s *Service) processCommand(ctx context.Context, id string) error {
	tx, err := s.DB.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var c Command
	var raw []byte
	err = tx.QueryRow(ctx, "SELECT id,actor_id,type,payload,status FROM commands WHERE id=$1 FOR UPDATE", id).
		Scan(&c.ID, &c.ActorID, &c.Type, &raw, &c.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil || c.Status != "queued" {
		return err
	}
	if err := json.Unmarshal(raw, &c.Payload); err != nil {
		return err
	}
	attempt, err := readCommand(ctx, tx, c)
	if err != nil {
		return err
	}
	snapshot, err := attempt.snapshot()
	if err != nil {
		return err
	}
	decision := plan.Companion_Default___.DecideCommand(attempt.input(), snapshot)
	// Keep the Command row lock when rolling back a partial plan on rejection.
	work, err := tx.Begin(ctx)
	if err != nil {
		return err
	}
	attempt.unit.Tx = work
	result, err := attempt.apply(decision)

	var denied *Rejection
	if errors.As(err, &denied) {
		if err := work.Rollback(ctx); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "UPDATE commands SET status='rejected',error=$2,finished_at=now() WHERE id=$1", id, denied.Message)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	if err != nil {
		return err
	}
	if err := work.Commit(ctx); err != nil {
		return err
	}
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, "UPDATE commands SET status='succeeded',result=$2,error='',finished_at=now() WHERE id=$1 AND status='queued'", id, body)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return tx.Commit(ctx)
}

func readCommand(ctx context.Context, tx pgx.Tx, c Command) (*commandAttempt, error) {
	a := &commandAttempt{unit: store.NewUnit(ctx, tx), command: c}
	if c.ActorID == store.ID("system") {
		a.actor = domain.Actor{ID: c.ActorID, Role: "system", Name: "system"}
	} else {
		var body []byte
		if err := tx.QueryRow(ctx, store.ActorQuery+" WHERE a.id=$1", c.ActorID).Scan(&body); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &a.actor); err != nil {
			return nil, err
		}
	}
	if err := tx.QueryRow(ctx, "SELECT clock_timestamp()").Scan(&a.now); err != nil {
		return nil, err
	}
	return a, nil
}
