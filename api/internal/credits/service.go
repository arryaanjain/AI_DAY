package credits

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientCredits = errors.New("insufficient credits")

type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }
func (s *Service) Balance(ctx context.Context, userID string) (int, error) {
	var balance int
	err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(quantity),0) FROM credit_ledger WHERE user_id=$1`, userID).Scan(&balance)
	return balance, err
}
func (s *Service) Reserve(ctx context.Context, userID, jobID, key string) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var balance int
	if err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(quantity),0) FROM credit_ledger WHERE user_id=$1 FOR UPDATE`, userID).Scan(&balance); err != nil {
		return err
	}
	if balance < 1 {
		return ErrInsufficientCredits
	}
	if _, err = tx.Exec(ctx, `INSERT INTO credit_ledger(user_id,entry_type,quantity,generation_job_id,idempotency_key,description) VALUES($1,'reservation',-1,$2,$3,'Generation credit reserved')`, userID, jobID, key); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
