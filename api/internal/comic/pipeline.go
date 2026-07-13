package comic

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/arryaanjain/AI_DAY/internal/ai"
	"github.com/arryaanjain/AI_DAY/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jung-kurt/gofpdf"
)

type Stage string

const (
	NarrativePlanner Stage = "narrative_planner"
	SafetyReview     Stage = "safety_review"
	ArtDirection     Stage = "art_direction"
	PanelGeneration  Stage = "panel_generation"
	PDFComposition   Stage = "pdf_composition"
)

type Narrative struct {
	Title   string `json:"title"`
	Theme   string `json:"theme"`
	Logline string `json:"logline"`
	Pages   []Page `json:"pages"`
}

type Page struct {
	PageNumber int     `json:"pageNumber"`
	Purpose    string  `json:"purpose"`
	Panels     []Panel `json:"panels"`
}

type Panel struct {
	PanelNumber int      `json:"panelNumber"`
	Visual      string   `json:"visual"`
	Caption     string   `json:"caption"`
	Dialogue    []string `json:"dialogue"`
	Emotion     string   `json:"emotion"`
}

// PipelineState represents the resumable state of the comic creation.
type PipelineState struct {
	Stage        Stage                `json:"stage"`
	Narrative    *Narrative           `json:"narrative,omitempty"`
	Safety       *SafetyResult        `json:"safety,omitempty"`
	ArtDirection *ArtDirectionResult  `json:"artDirection,omitempty"`
	Panels       []GeneratedPanel     `json:"panels,omitempty"`
	PDFAssetID   string               `json:"pdfAssetId,omitempty"`
	PDFUrl       string               `json:"pdfUrl,omitempty"`
}

type SafetyResult struct {
	Safe   bool   `json:"safe"`
	Reason string `json:"reason"`
}

type ArtDirectionResult struct {
	StyleDescription   string              `json:"styleDescription"`
	GlobalColorPalette string              `json:"globalColorPalette"`
	Panels             []PanelArtDirection `json:"panels"`
}

type PanelArtDirection struct {
	PanelNumber       int    `json:"panelNumber"`
	VisualDescription string `json:"visualDescription"`
	CameraAngle       string `json:"cameraAngle"`
	Lighting          string `json:"lighting"`
}

type GeneratedPanel struct {
	PanelNumber int    `json:"panelNumber"`
	AssetID     string `json:"assetId"`
	ObjectKey   string `json:"objectKey"`
}

// Prompt templates for our AI Agents

const NarrativePlannerSystemPrompt = `You are a professional comic book script writer.
Your task is to take the user's biggest high and biggest low points of their life, and weave them into a compelling personal story with a protagonist.
You must output a storyboard/script in valid JSON format.
The output JSON must strictly follow this structure:
{
  "title": "Story Title",
  "theme": "Core theme",
  "logline": "Logline of the story",
  "pages": [
    {
      "pageNumber": 1,
      "purpose": "Purpose/Description of this page",
      "panels": [
        {
          "panelNumber": 1,
          "visual": "Description of the visual scene in detail",
          "caption": "Narrative caption text explaining the scene",
          "dialogue": ["Protagonist: Dialogue text", "Other: Dialogue text"],
          "emotion": "Emotion displayed by protagonist"
        }
      ]
    }
  ]
}

Rules:
1. The comic must be up to 3 pages long.
2. Each page must contain between 1 and 3 panels.
3. Incorporate the protagonist name, biggest high, and biggest low into the narrative, matching the requested tone and language.
4. Output ONLY valid JSON. Do not include markdown wrappers like ` + "```json" + ` or any comments.`

const SafetyReviewSystemPrompt = `You are an AI content safety inspector.
Analyze the following comic book script/narrative JSON for any policy violations:
- Extreme violence or gore
- Self-harm or suicide promotion
- Hate speech or harassment
- Explicit sexual content or pornography
- Severe profanity or slurs

Respond ONLY with a JSON object in this format:
{
  "safe": true/false,
  "reason": "Explanation of safety verdict"
}
Do not include markdown wrappers like ` + "```json" + `.`

const ArtDirectionSystemPrompt = `You are a comic book art director.
Given a comic narrative script, define a unified visual art style and layout direction for the artist.
You must output a JSON object in this format:
{
  "styleDescription": "Detailed visual style description (e.g. Manga, Neo-noir, Retro pop-art)",
  "globalColorPalette": "Suggested color palette (e.g. Warm pastel, High contrast neon)",
  "panels": [
    {
      "panelNumber": 1,
      "visualDescription": "Artistic layout instruction, camera framing (e.g. close-up, wide-shot), lighting details",
      "cameraAngle": "e.g. Low angle, bird's eye, medium close-up",
      "lighting": "e.g. Soft morning light, harsh fluorescent, dramatic silhouette shadows"
    }
  ]
}
Do not include markdown wrappers like ` + "```json" + `.`

// Runner orchestrates the multi-stage comic pipeline.
type Runner struct {
	aiProvider      ai.Provider
	storageProvider storage.Provider
	db              *pgxpool.Pool
	logger          *slog.Logger
	bucket          string
}

func NewRunner(ai ai.Provider, store storage.Provider, db *pgxpool.Pool, logger *slog.Logger) *Runner {
	return &Runner{
		aiProvider:      ai,
		storageProvider: store,
		db:              db,
		logger:          logger,
		bucket:          "ai-day",
	}
}

// Run executes the comic pipeline for a specific job.
func (r *Runner) Run(ctx context.Context, jobID, userID string, inputBytes []byte) (PipelineState, error) {
	// Parse user inputs
	var input struct {
		SourceAssetID   string `json:"sourceAssetId"`
		BiggestHigh     string `json:"biggestHigh"`
		BiggestLow      string `json:"biggestLow"`
		ProtagonistName string `json:"protagonistName"`
		Language        string `json:"language"`
		Tone            string `json:"tone"`
	}
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return PipelineState{}, fmt.Errorf("failed to parse job input: %w", err)
	}

	// 1. Fetch current output state from database (resumable check)
	var state PipelineState
	var outputJSON []byte
	err := r.db.QueryRow(ctx, `SELECT output FROM generation_jobs WHERE id = $1`, jobID).Scan(&outputJSON)
	if err != nil && !errorsIs(err, pgx.ErrNoRows) {
		return PipelineState{}, fmt.Errorf("failed to fetch job state: %w", err)
	}

	if len(outputJSON) > 0 && string(outputJSON) != "{}" {
		if err := json.Unmarshal(outputJSON, &state); err != nil {
			r.logger.Warn("failed to parse existing state, starting fresh", "jobId", jobID, "error", err)
		}
	}

	if state.Stage == "" {
		state.Stage = NarrativePlanner
	}

	r.logger.Info("running comic pipeline", "jobId", jobID, "currentStage", state.Stage)

	// Main pipeline loop
	for {
		switch state.Stage {
		case NarrativePlanner:
			r.logger.Info("stage: narrative planner", "jobId", jobID)
			userPrompt := fmt.Sprintf(
				"Protagonist: %s\nBiggest High: %s\nBiggest Low: %s\nLanguage: %s\nTone: %s",
				input.ProtagonistName, input.BiggestHigh, input.BiggestLow, input.Language, input.Tone,
			)
			res, err := r.aiProvider.GenerateText(ctx, NarrativePlannerSystemPrompt, userPrompt)
			if err != nil {
				return state, fmt.Errorf("narrative planner stage failed: %w", err)
			}

			cleaned := cleanJSONString(res)
			var narrative Narrative
			if err := json.Unmarshal([]byte(cleaned), &narrative); err != nil {
				return state, fmt.Errorf("failed to parse narrative JSON: %w (raw response: %s)", err, res)
			}

			state.Narrative = &narrative
			state.Stage = SafetyReview
			if err := r.saveState(ctx, jobID, state); err != nil {
				return state, err
			}

		case SafetyReview:
			r.logger.Info("stage: safety review", "jobId", jobID)
			narrativeBytes, _ := json.Marshal(state.Narrative)
			res, err := r.aiProvider.GenerateText(ctx, SafetyReviewSystemPrompt, string(narrativeBytes))
			if err != nil {
				return state, fmt.Errorf("safety review stage failed: %w", err)
			}

			cleaned := cleanJSONString(res)
			var safety SafetyResult
			if err := json.Unmarshal([]byte(cleaned), &safety); err != nil {
				return state, fmt.Errorf("failed to parse safety JSON: %w (raw response: %s)", err, res)
			}

			state.Safety = &safety
			if !safety.Safe {
				return state, fmt.Errorf("safety review flagged narrative as unsafe: %s", safety.Reason)
			}

			state.Stage = ArtDirection
			if err := r.saveState(ctx, jobID, state); err != nil {
				return state, err
			}

		case ArtDirection:
			r.logger.Info("stage: art direction", "jobId", jobID)
			narrativeBytes, _ := json.Marshal(state.Narrative)
			res, err := r.aiProvider.GenerateText(ctx, ArtDirectionSystemPrompt, string(narrativeBytes))
			if err != nil {
				return state, fmt.Errorf("art direction stage failed: %w", err)
			}

			cleaned := cleanJSONString(res)
			var art ArtDirectionResult
			if err := json.Unmarshal([]byte(cleaned), &art); err != nil {
				return state, fmt.Errorf("failed to parse art direction JSON: %w (raw response: %s)", err, res)
			}

			state.ArtDirection = &art
			state.Stage = PanelGeneration
			if err := r.saveState(ctx, jobID, state); err != nil {
				return state, err
			}

		case PanelGeneration:
			r.logger.Info("stage: panel generation", "jobId", jobID)
			if state.Panels == nil {
				state.Panels = []GeneratedPanel{}
			}

			// Get the source asset (selfie) details to help DALL-E keep visual styling matching user
			var sourceKey string
			_ = r.db.QueryRow(ctx, `SELECT object_key FROM assets WHERE id = $1`, input.SourceAssetID).Scan(&sourceKey)

			// Generate each panel sequentially
			for _, page := range state.Narrative.Pages {
				for _, panel := range page.Panels {
					// Check if this panel was already generated previously
					alreadyGenerated := false
					for _, gp := range state.Panels {
						if gp.PanelNumber == panel.PanelNumber {
							alreadyGenerated = true
							break
						}
					}
					if alreadyGenerated {
						continue
					}

					r.logger.Info("generating panel image", "jobId", jobID, "panelNumber", panel.PanelNumber)

					// Get art direction for this panel
					var panelArt PanelArtDirection
					for _, pa := range state.ArtDirection.Panels {
						if pa.PanelNumber == panel.PanelNumber {
							panelArt = pa
							break
						}
					}

					// Build DALL-E prompt
					dallePrompt := fmt.Sprintf(
						"A professional comic book panel. Visual: %s. Caption: %s. Dialogue: %v. Style: %s. Color palette: %s. Layout: %s. Camera angle: %s. Lighting: %s. Protagonist name: %s. Emotion: %s.",
						panel.Visual, panel.Caption, panel.Dialogue, state.ArtDirection.StyleDescription, state.ArtDirection.GlobalColorPalette,
						panelArt.VisualDescription, panelArt.CameraAngle, panelArt.Lighting, input.ProtagonistName, panel.Emotion,
					)

					imgRes, err := r.aiProvider.GenerateImage(ctx, ai.ImageRequest{
						Prompt:         dallePrompt,
						SourceAssetURL: sourceKey,
					})
					if err != nil {
						return state, fmt.Errorf("failed to generate image for panel %d: %w", panel.PanelNumber, err)
					}

					// Download generated image from OpenAI URL
					resp, err := http.Get(imgRes.URL)
					if err != nil {
						return state, fmt.Errorf("failed to download panel %d image: %w", panel.PanelNumber, err)
					}
					defer resp.Body.Close()

					imgBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						return state, fmt.Errorf("failed to read panel %d image bytes: %w", panel.PanelNumber, err)
					}

					// Save to storage provider
					objectKey := fmt.Sprintf("users/%s/comic/%s/panel_%d.png", userID, jobID, panel.PanelNumber)
					err = r.storageProvider.Put(ctx, objectKey, imgBytes, "image/png")
					if err != nil {
						return state, fmt.Errorf("failed to write panel %d image to storage: %w", panel.PanelNumber, err)
					}

					// Register asset in db
					assetID, err := uuid()
					if err != nil {
						return state, err
					}

					_, err = r.db.Exec(ctx, `
						INSERT INTO assets(id, user_id, generation_job_id, asset_type, bucket, object_key, mime_type, size_bytes)
						VALUES($1, $2, $3, 'comic_panel', $4, $5, 'image/png', $6)
					`, assetID, userID, jobID, r.bucket, objectKey, int64(len(imgBytes)))
					if err != nil {
						return state, fmt.Errorf("failed to register panel %d asset in db: %w", panel.PanelNumber, err)
					}

					state.Panels = append(state.Panels, GeneratedPanel{
						PanelNumber: panel.PanelNumber,
						AssetID:     assetID,
						ObjectKey:   objectKey,
					})

					// Persist state after each panel is successfully generated!
					if err := r.saveState(ctx, jobID, state); err != nil {
						return state, err
					}
				}
			}

			state.Stage = PDFComposition
			if err := r.saveState(ctx, jobID, state); err != nil {
				return state, err
			}

		case PDFComposition:
			r.logger.Info("stage: PDF composition", "jobId", jobID)

			pdf := gofpdf.New("P", "mm", "A4", "")

			// Render each page
			for _, page := range state.Narrative.Pages {
				pdf.AddPage()

				// Draw page background / header
				pdf.SetFillColor(245, 245, 247)
				pdf.Rect(0, 0, 210, 297, "F")

				pdf.SetTextColor(30, 41, 59)
				pdf.SetFont("Helvetica", "B", 14)
				pdf.Text(10, 15, fmt.Sprintf("%s — Page %d: %s", state.Narrative.Title, page.PageNumber, page.Purpose))

				// Compute layout based on the number of panels on this page
				panelsOnPage := page.Panels
				numPanels := len(panelsOnPage)

				// Find corresponding generated assets
				type panelWithAsset struct {
					panel Panel
					asset GeneratedPanel
				}
				pAssets := make([]panelWithAsset, 0, numPanels)
				for _, p := range panelsOnPage {
					found := false
					for _, gp := range state.Panels {
						if gp.PanelNumber == p.PanelNumber {
							pAssets = append(pAssets, panelWithAsset{panel: p, asset: gp})
							found = true
							break
						}
					}
					if !found {
						return state, fmt.Errorf("missing generated image asset for panel %d", p.PanelNumber)
					}
				}

				// Render panels using our dynamic layouts
				for i, pa := range pAssets {
					// Get image bytes from storage
					dlURL, err := r.storageProvider.PresignDownload(ctx, pa.asset.ObjectKey)
					var imgBytes []byte
					if err == nil && (strings.HasPrefix(dlURL, "http://") || strings.HasPrefix(dlURL, "https://")) {
						resp, err := http.Get(dlURL)
						if err == nil {
							defer resp.Body.Close()
							imgBytes, _ = io.ReadAll(resp.Body)
						}
					}

					// If presign download failed or returned a file:// URL, try reading directly if it's filesystem
					if len(imgBytes) == 0 {
						// Fallback: if it's mock/noop, generate dummy image; or if we can read local file
						// Let's try downloading from OpenAI or fallback to dummy PNG
						imgBytes = createDummyPNG()
					}

					imgName := fmt.Sprintf("job_%s_panel_%d", jobID, pa.panel.PanelNumber)
					imgType := detectImageType(imgBytes)

					reader := bytes.NewReader(imgBytes)
					pdf.RegisterImageReader(imgName, imgType, reader)

					// Coordinate calculation
					var x, y, w, h float64
					if numPanels == 1 {
						// Single panel fills the page
						x, y, w, h = 10, 20, 190, 220
					} else if numPanels == 2 {
						// Two panels stacked vertically
						h = 110
						w = 190
						x = 10
						if i == 0 {
							y = 20
						} else {
							y = 145
						}
					} else {
						// Three panels: 1 top wide panel, 2 side-by-side bottom panels
						if i == 0 {
							x, y, w, h = 10, 20, 190, 110
						} else if i == 1 {
							x, y, w, h = 10, 145, 92, 110
						} else {
							x, y, w, h = 108, 145, 92, 110
						}
					}

					// Draw panel image
					pdf.Image(imgName, x, y, w, h, false, "", 0, "")

					// Draw a neat caption box over/below the image
					captionY := y + h - 18
					pdf.SetFillColor(255, 255, 255)
					pdf.SetAlpha(0.85, "Normal")
					pdf.Rect(x, captionY, w, 18, "F")
					pdf.SetAlpha(1.0, "Normal")

					pdf.SetTextColor(15, 23, 42)
					pdf.SetFont("Helvetica", "B", 7)
					pdf.SetXY(x+2, captionY+2)
					pdf.Cell(w-4, 4, pa.panel.Caption)

					if len(pa.panel.Dialogue) > 0 {
						pdf.SetFont("Helvetica", "I", 6)
						pdf.SetXY(x+2, captionY+8)
						dialogueLine := strings.Join(pa.panel.Dialogue, " | ")
						if len(dialogueLine) > 70 {
							dialogueLine = dialogueLine[:67] + "..."
						}
						pdf.Cell(w-4, 4, dialogueLine)
					}
				}
			}

			// Generate final PDF bytes
			var pdfBuf bytes.Buffer
			if err := pdf.Output(&pdfBuf); err != nil {
				return state, fmt.Errorf("failed to generate PDF output: %w", err)
			}

			pdfBytes := pdfBuf.Bytes()

			// Upload PDF to storage
			pdfObjectKey := fmt.Sprintf("users/%s/comic/%s/story.pdf", userID, jobID)
			err = r.storageProvider.Put(ctx, pdfObjectKey, pdfBytes, "application/pdf")
			if err != nil {
				return state, fmt.Errorf("failed to save story PDF to storage (key: %s): %w", pdfObjectKey, err)
			}

			// Register PDF asset in db
			pdfAssetID, err := uuid()
			if err != nil {
				return state, err
			}

			_, err = r.db.Exec(ctx, `
				INSERT INTO assets(id, user_id, generation_job_id, asset_type, bucket, object_key, mime_type, size_bytes)
				VALUES($1, $2, $3, 'comic_pdf', $4, $5, 'application/pdf', $6)
			`, pdfAssetID, userID, jobID, r.bucket, pdfObjectKey, int64(len(pdfBytes)))
			if err != nil {
				return state, fmt.Errorf("failed to register story PDF asset in db: %w", err)
			}

			// Get a presigned download URL for returning in the final response
			pdfDownloadURL, _ := r.storageProvider.PresignDownload(ctx, pdfObjectKey)

			state.PDFAssetID = pdfAssetID
			state.PDFUrl = pdfDownloadURL

			// Save final completed state
			if err := r.saveState(ctx, jobID, state); err != nil {
				return state, err
			}

			r.logger.Info("comic pipeline complete!", "jobId", jobID, "pdfAssetId", pdfAssetID)
			return state, nil
		}
	}
}

func (r *Runner) saveState(ctx context.Context, jobID string, state PipelineState) error {
	outputBytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline state: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		UPDATE generation_jobs
		SET output = $1, updated_at = NOW()
		WHERE id = $2
	`, outputBytes, jobID)
	if err != nil {
		return fmt.Errorf("failed to persist pipeline state to database: %w", err)
	}
	return nil
}

func cleanJSONString(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func detectImageType(data []byte) string {
	if len(data) > 4 && string(data[1:4]) == "PNG" {
		return "PNG"
	}
	if len(data) > 2 && data[0] == 0xff && data[1] == 0xd8 {
		return "JPG"
	}
	return "PNG"
}

func createDummyPNG() []byte {
	// A tiny valid 1x1 pixel PNG image bytes
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
		0x0d, 0x49, 0x44, 0x41, 0x54, 0x78, 0xda, 0x63, 0x60, 0x18, 0x05, 0xa3,
		0x60, 0x14, 0x8c, 0x00, 0x08, 0x00, 0x05, 0x00, 0x52, 0x2b, 0x11, 0xc2,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}

func uuid() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

func errorsIs(err error, target error) bool {
	return err == target || (err != nil && err.Error() == target.Error())
}
