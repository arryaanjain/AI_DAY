package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/arryaanjain/AI_DAY/internal/ai"
	"github.com/arryaanjain/AI_DAY/internal/assets"
	"github.com/arryaanjain/AI_DAY/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Job payloads always carry jobId, attempt and traceId.
type Payload struct {
	JobID   string `json:"jobId"`
	Attempt int    `json:"attempt"`
	TraceID string `json:"traceId"`
}

// Worker processes generation jobs from the database.
type Worker struct {
	db                *pgxpool.Pool
	aiProvider        ai.Provider
	storageProvider   storage.Provider
	assetService      *assets.Service
	logger            *slog.Logger
	retryLimit        int
	processingTimeout time.Duration
}

// NewWorker creates a new worker instance.
func NewWorker(
	db *pgxpool.Pool,
	aiProv ai.Provider,
	storageProv storage.Provider,
	assetSvc *assets.Service,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		db:                db,
		aiProvider:        aiProv,
		storageProvider:   storageProv,
		assetService:      assetSvc,
		logger:            logger,
		retryLimit:        3,
		processingTimeout: 5 * time.Minute,
	}
}

// Start begins processing jobs from the database.
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("worker started, polling for jobs")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker shutting down")
			return
		case <-ticker.C:
			w.processNextJob(ctx)
		}
	}
}

// processNextJob retrieves and processes the next job from the database.
func (w *Worker) processNextJob(ctx context.Context) {
	// Query for the next queued job
	var jobID string
	err := w.db.QueryRow(ctx, `
		SELECT id FROM generation_jobs
		WHERE status = 'queued'
		ORDER BY created_at ASC
		LIMIT 1
	`).Scan(&jobID)

	if err != nil {
		// No jobs available
		return
	}

	w.processJob(ctx, jobID)
}

// processJob handles a single job.
func (w *Worker) processJob(ctx context.Context, jobID string) {
	jobCtx, cancel := context.WithTimeout(ctx, w.processingTimeout)
	defer cancel()

	// Mark job as processing
	_, err := w.db.Exec(jobCtx, `
		UPDATE generation_jobs
		SET status = 'processing'
		WHERE id = $1
	`, jobID)

	if err != nil {
		w.logger.Error("failed to mark job as processing", "jobId", jobID, "error", err)
		return
	}

	// Fetch the job from the database
	job, err := w.fetchJob(jobCtx, jobID)
	if err != nil {
		w.logger.Error("failed to fetch job", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "FETCH_ERROR", err.Error())
		return
	}

	w.logger.Info("processing job", "jobId", jobID, "module", job.Module)

	// Parse the input
	var input map[string]interface{}
	if err := json.Unmarshal(job.Input, &input); err != nil {
		w.logger.Error("failed to parse job input", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "PARSE_ERROR", err.Error())
		return
	}

	// Get the source asset for context
	sourceAsset, _ := w.assetService.Get(jobCtx, job.SourceAssetID)

	// Generate the image
	genResult, err := w.aiProvider.GenerateImage(jobCtx, ai.ImageRequest{
		Prompt:         fmt.Sprintf("%v", input),
		SourceAssetURL: sourceAsset.ObjectKey,
	})

	if err != nil {
		w.logger.Error("job processing failed", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "GENERATION_ERROR", err.Error())
		return
	}

	// Update job with output
	output := map[string]interface{}{
		"imageUrl":          genResult.URL,
		"providerRequestId": genResult.ProviderRequestID,
	}

	outputJSON, _ := json.Marshal(output)
	w.updateJobSuccess(jobCtx, jobID, outputJSON)

	w.logger.Info("job completed", "jobId", jobID)
}

// fetchJob retrieves a job from the database.
func (w *Worker) fetchJob(ctx context.Context, jobID string) (*GenerationJob, error) {
	var job GenerationJob
	err := w.db.QueryRow(ctx, `
		SELECT id, user_id, module, status, source_asset_id, input
		FROM generation_jobs
		WHERE id = $1
	`, jobID).Scan(&job.ID, &job.UserID, &job.Module, &job.Status, &job.SourceAssetID, &job.Input)

	if err != nil {
		return nil, err
	}

	return &job, nil
}

// updateJobSuccess marks a job as completed with output.
func (w *Worker) updateJobSuccess(ctx context.Context, jobID string, output []byte) error {
	_, err := w.db.Exec(ctx, `
		UPDATE generation_jobs
		SET status = 'completed', output = $1, completed_at = NOW()
		WHERE id = $2
	`, output, jobID)

	return err
}

// updateJobError marks a job as failed with an error message.
func (w *Worker) updateJobError(ctx context.Context, jobID, errorCode, errorMessage string) error {
	_, err := w.db.Exec(ctx, `
		UPDATE generation_jobs
		SET status = 'failed', error_code = $1, error_message = $2, completed_at = NOW()
		WHERE id = $3
	`, errorCode, errorMessage, jobID)

	return err
}

// GenerationJob represents a generation job.
type GenerationJob struct {
	ID            string
	UserID        string
	Module        string
	Status        string
	SourceAssetID string
	Input         []byte
}
