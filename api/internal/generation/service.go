package generation

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientCredits = errors.New("insufficient credits")
var ErrJobNotFound = errors.New("generation job not found")

type JobDetails struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"userId"`
	Module        string                 `json:"module"`
	Status        string                 `json:"status"`
	SourceAssetID *string                `json:"sourceAssetId"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output"`
	ErrorCode     *string                `json:"errorCode"`
	ErrorMessage  *string                `json:"errorMessage"`
	StartedAt     *time.Time             `json:"startedAt"`
	CompletedAt   *time.Time             `json:"completedAt"`
	CreatedAt     time.Time              `json:"createdAt"`
}

type Service struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) CreateWithCredit(ctx context.Context, userID, module, assetID, key string, input []byte) (string, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id string
	err = tx.QueryRow(ctx, `SELECT id::text FROM generation_jobs WHERE user_id=$1::uuid AND idempotency_key=$2`, userID, key).Scan(&id)
	if err == nil {
		return id, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	var dummy string
	_ = tx.QueryRow(ctx, `SELECT id::text FROM users WHERE id=$1::uuid FOR UPDATE`, userID).Scan(&dummy)
	var balance int
	if err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(quantity),0) FROM credit_ledger WHERE user_id=$1::uuid`, userID).Scan(&balance); err != nil {
		return "", err
	}
	if balance < 1 {
		return "", ErrInsufficientCredits
	}
	if err = tx.QueryRow(ctx, `INSERT INTO generation_jobs(user_id,module,source_asset_id,idempotency_key,input,status) VALUES($1::uuid,$2::module_type,$3::uuid,$4,$5,'queued') RETURNING id::text`, userID, module, assetID, key, input).Scan(&id); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO credit_ledger(user_id,entry_type,quantity,generation_job_id,idempotency_key,description) VALUES($1::uuid,'reservation',-1,$2::uuid,$3,'Generation credit reserved')`, userID, id, "reservation:"+key); err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}

func (s *Service) Get(ctx context.Context, jobID, userID string) (*JobDetails, error) {
	var job JobDetails
	var sourceAssetID *string
	var inputBytes, outputBytes []byte
	err := s.db.QueryRow(ctx, `
		SELECT id::text, user_id::text, module, status, source_asset_id::text, input, output, error_code, error_message, started_at, completed_at, created_at
		FROM generation_jobs
		WHERE id = $1 AND user_id = $2
	`, jobID, userID).Scan(
		&job.ID, &job.UserID, &job.Module, &job.Status, &sourceAssetID, &inputBytes, &outputBytes, &job.ErrorCode, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	job.SourceAssetID = sourceAssetID
	if len(inputBytes) > 0 {
		_ = json.Unmarshal(inputBytes, &job.Input)
	}
	if len(outputBytes) > 0 {
		_ = json.Unmarshal(outputBytes, &job.Output)
	}
	return &job, nil
}

func (s *Service) ListUserJobs(ctx context.Context, userID string) ([]JobDetails, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id::text, user_id::text, module, status, source_asset_id::text, input, output, error_code, error_message, started_at, completed_at, created_at
		FROM generation_jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]JobDetails, 0)
	for rows.Next() {
		var job JobDetails
		var sourceAssetID *string
		var inputBytes, outputBytes []byte
		if err := rows.Scan(&job.ID, &job.UserID, &job.Module, &job.Status, &sourceAssetID, &inputBytes, &outputBytes, &job.ErrorCode, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt); err != nil {
			return nil, err
		}
		job.SourceAssetID = sourceAssetID
		if len(inputBytes) > 0 {
			_ = json.Unmarshal(inputBytes, &job.Input)
		}
		if len(outputBytes) > 0 {
			_ = json.Unmarshal(outputBytes, &job.Output)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
