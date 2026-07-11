package generation

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientCredits = errors.New("insufficient credits")

type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }
func (s *Service) CreateWithCredit(ctx context.Context, userID, module, assetID, key string, input []byte) (string, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id string
	err = tx.QueryRow(ctx, `SELECT id::text FROM generation_jobs WHERE user_id=$1 AND idempotency_key=$2`, userID, key).Scan(&id)
	if err == nil {
		return id, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	var balance int
	if err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(quantity),0) FROM credit_ledger WHERE user_id=$1 FOR UPDATE`, userID).Scan(&balance); err != nil {
		return "", err
	}
	if balance < 1 {
		return "", ErrInsufficientCredits
	}
	if err = tx.QueryRow(ctx, `INSERT INTO generation_jobs(user_id,module,source_asset_id,idempotency_key,input,status) VALUES($1,$2,$3,$4,$5,'queued') RETURNING id::text`, userID, module, assetID, key, input).Scan(&id); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO credit_ledger(user_id,entry_type,quantity,generation_job_id,idempotency_key,description) VALUES($1,'reservation',-1,$2,$3,'Generation credit reserved')`, userID, id, "reservation:"+key); err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}
