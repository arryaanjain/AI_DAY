package worker

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/arryaanjain/AI_DAY/internal/ai"
	"github.com/arryaanjain/AI_DAY/internal/assets"
	"github.com/arryaanjain/AI_DAY/internal/comic"
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
		SET status = 'processing', started_at = NOW()
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

	if job.Module == "comic" {
		runner := comic.NewRunner(w.aiProvider, w.storageProvider, w.db, w.logger)
		state, err := runner.Run(jobCtx, job.ID, job.UserID, job.Input)
		if err != nil {
			w.logger.Error("comic pipeline execution failed", "jobId", jobID, "error", err)
			errorCode := "GENERATION_ERROR"
			if strings.Contains(err.Error(), "safety review flagged") || strings.Contains(err.Error(), "content is safe") || strings.Contains(err.Error(), "policy") {
				errorCode = "CONTENT_POLICY_VIOLATION"
			}
			w.updateJobError(jobCtx, jobID, errorCode, err.Error())
			return
		}

		outputBytes, _ := json.Marshal(state)
		w.updateJobSuccess(jobCtx, jobID, outputBytes)
		w.logger.Info("comic job completed successfully", "jobId", jobID)
		return
	}

	// Default/PixArt ("pixel_portrait") flow
	// Parse the input
	var input map[string]interface{}
	if err := json.Unmarshal(job.Input, &input); err != nil {
		w.logger.Error("failed to parse job input", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "PARSE_ERROR", err.Error())
		return
	}

	// Get the source asset for context
	var sourceKey string
	if job.SourceAssetID != "" {
		sourceAsset, _ := w.assetService.Get(jobCtx, job.SourceAssetID)
		if sourceAsset != nil {
			sourceKey = sourceAsset.ObjectKey
		}
	}

	// Generate PixArt image using a premium Pixar style prompt
	dallePrompt := "A premium, highly detailed 3D Pixar-style digital art portrait of the person. Cute animated character style, vibrant colors, soft lighting, professional character design."
	genResult, err := w.aiProvider.GenerateImage(jobCtx, ai.ImageRequest{
		Prompt:         dallePrompt,
		SourceAssetURL: sourceKey,
	})

	if err != nil {
		w.logger.Error("job processing failed", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "GENERATION_ERROR", err.Error())
		return
	}

	// Download generated image immediately so DALL-E 3 URL doesn't expire
	resp, err := http.Get(genResult.URL)
	if err != nil {
		w.logger.Error("failed to download generated image", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "DOWNLOAD_ERROR", err.Error())
		return
	}
	defer resp.Body.Close()

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		w.logger.Error("failed to read downloaded image bytes", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "DOWNLOAD_ERROR", err.Error())
		return
	}

	// Save to storage
	objectKey := fmt.Sprintf("users/%s/pixart/%s/image.png", job.UserID, job.ID)
	err = w.storageProvider.Put(jobCtx, objectKey, imgBytes, "image/png")
	if err != nil {
		w.logger.Error("failed to save generated image to storage", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "STORAGE_ERROR", err.Error())
		return
	}

	// Register generated image asset in db
	assetID, err := generateUUID()
	if err != nil {
		w.logger.Error("failed to generate UUID for asset", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "UUID_ERROR", err.Error())
		return
	}

	_, err = w.db.Exec(jobCtx, `
		INSERT INTO assets(id, user_id, generation_job_id, asset_type, bucket, object_key, mime_type, size_bytes)
		VALUES($1, $2, $3, 'generated_image', $4, $5, 'image/png', $6)
	`, assetID, job.UserID, job.ID, "ai-day", objectKey, int64(len(imgBytes)))
	if err != nil {
		w.logger.Error("failed to register asset in db", "jobId", jobID, "error", err)
		w.updateJobError(jobCtx, jobID, "DATABASE_ERROR", err.Error())
		return
	}

	// Generate presigned download URL for the final output
	downloadURL, _ := w.storageProvider.PresignDownload(jobCtx, objectKey)

	// Update job with output containing local downloadURL
	output := map[string]interface{}{
		"imageUrl":          downloadURL,
		"providerRequestId": genResult.ProviderRequestID,
		"assetId":           assetID,
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

func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
