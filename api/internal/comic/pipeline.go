package comic

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log/slog"
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
	PageNumber        int    `json:"pageNumber"`
	PanelNumber       int    `json:"panelNumber"`
	VisualDescription string `json:"visualDescription"`
	CameraAngle       string `json:"cameraAngle"`
	Lighting          string `json:"lighting"`
}

type GeneratedPanel struct {
	PageNumber  int    `json:"pageNumber"`
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
      "pageNumber": 1,
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
	err := r.db.QueryRow(ctx, `SELECT output FROM generation_jobs WHERE id = $1::uuid`, jobID).Scan(&outputJSON)
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
			_ = r.db.QueryRow(ctx, `SELECT object_key FROM assets WHERE id = $1::uuid`, input.SourceAssetID).Scan(&sourceKey)

			// Generate each panel sequentially
			for _, page := range state.Narrative.Pages {
				for _, panel := range page.Panels {
					// Check if this panel was already generated previously
					alreadyGenerated := false
					for _, gp := range state.Panels {
						if gp.PageNumber == page.PageNumber && gp.PanelNumber == panel.PanelNumber {
							alreadyGenerated = true
							break
						}
					}
					if alreadyGenerated {
						continue
					}

					r.logger.Info("generating panel image", "jobId", jobID, "pageNumber", page.PageNumber, "panelNumber", panel.PanelNumber)

					// Get art direction for this panel
					var panelArt PanelArtDirection
					for _, pa := range state.ArtDirection.Panels {
						if (pa.PageNumber == page.PageNumber || pa.PageNumber == 0) && pa.PanelNumber == panel.PanelNumber {
							panelArt = pa
							if pa.PageNumber == page.PageNumber {
								break
							}
						}
					}

					// Build DALL-E prompt with family-friendly PG-rated styling hint
					dallePrompt := fmt.Sprintf(
						"A family-friendly, PG-rated comic book panel in vibrant art style. Visual: %s. Caption: %s. Style: %s. Color palette: %s. Layout: %s. Camera angle: %s. Lighting: %s. Emotion: %s.",
						panel.Visual, panel.Caption, state.ArtDirection.StyleDescription, state.ArtDirection.GlobalColorPalette,
						panelArt.VisualDescription, panelArt.CameraAngle, panelArt.Lighting, panel.Emotion,
					)

					imgRes, err := r.aiProvider.GenerateImage(ctx, ai.ImageRequest{
						Prompt:         dallePrompt,
						SourceAssetURL: sourceKey,
					})
					if err != nil {
						if strings.Contains(err.Error(), "moderation_blocked") || strings.Contains(err.Error(), "safety system") {
							r.logger.Warn("panel prompt flagged by DALL-E safety filter, retrying with safe fallback prompt", "panelNumber", panel.PanelNumber, "error", err)
							safePrompt := fmt.Sprintf(
								"A wholesome, family-friendly comic book scene showing character perseverance and hope. Style: %s. Color palette: %s.",
								state.ArtDirection.StyleDescription, state.ArtDirection.GlobalColorPalette,
							)
							imgRes, err = r.aiProvider.GenerateImage(ctx, ai.ImageRequest{
								Prompt: safePrompt,
							})
						}
					}
					if err != nil {
						return state, fmt.Errorf("failed to generate image for panel %d: %w", panel.PanelNumber, err)
					}

					// Use raw bytes from b64_json response (populated by OpenAI provider)
					var imgBytes []byte
					if len(imgRes.Bytes) > 0 {
						imgBytes = imgRes.Bytes
					}
					if len(imgBytes) == 0 {
						imgBytes = createDummyPNG()
					}

					// Save to storage provider
					objectKey := fmt.Sprintf("users/%s/comic/%s/page_%d_panel_%d.png", userID, jobID, page.PageNumber, panel.PanelNumber)
					err = r.storageProvider.Put(ctx, objectKey, imgBytes, "image/png")
					if err != nil {
						return state, fmt.Errorf("failed to write page %d panel %d image to storage: %w", page.PageNumber, panel.PanelNumber, err)
					}

					// Register asset in db
					assetID, err := uuid()
					if err != nil {
						return state, err
					}

					panelFilename := fmt.Sprintf("page_%d_panel_%d.png", page.PageNumber, panel.PanelNumber)
					_, err = r.db.Exec(ctx, `
						INSERT INTO assets(id, user_id, generation_job_id, asset_type, bucket, object_key, original_filename, mime_type, size_bytes)
						VALUES($1::uuid, $2::uuid, $3::uuid, 'comic_panel', $4, $5, $6, 'image/png', $7)
					`, assetID, userID, jobID, r.bucket, objectKey, panelFilename, int64(len(imgBytes)))
					if err != nil {
						return state, fmt.Errorf("failed to register page %d panel %d asset in db: %w", page.PageNumber, panel.PanelNumber, err)
					}

					state.Panels = append(state.Panels, GeneratedPanel{
						PageNumber:  page.PageNumber,
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

				pdf.SetTextColor(15, 23, 42)
				pdf.SetFont("Helvetica", "B", 12)
				pdf.SetXY(10, 10)
				headerText := fmt.Sprintf("%s — Page %d", cleanPDFText(pdf, state.Narrative.Title), page.PageNumber)
				pdf.Cell(190, 6, headerText)

				if page.Purpose != "" {
					pdf.SetTextColor(71, 85, 105)
					pdf.SetFont("Helvetica", "I", 9)
					pdf.SetXY(10, 16)
					pdf.Cell(190, 4, cleanPDFText(pdf, page.Purpose))
				}

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
						if gp.PageNumber == page.PageNumber && gp.PanelNumber == p.PanelNumber {
							pAssets = append(pAssets, panelWithAsset{panel: p, asset: gp})
							found = true
							break
						}
					}
					if !found {
						return state, fmt.Errorf("missing generated image asset for page %d panel %d", page.PageNumber, p.PanelNumber)
					}
				}

				// Render panels using our dynamic layouts
				for i, pa := range pAssets {
					// Load image bytes directly from storage — avoids defer-in-loop and
					// file:// URL parsing bugs that caused wrong images to be reused.
					imgBytes, err := r.storageProvider.Get(ctx, pa.asset.ObjectKey)
					if err != nil || len(imgBytes) == 0 {
						r.logger.Warn("could not load panel image from storage, using placeholder",
							"objectKey", pa.asset.ObjectKey, "error", err)
						imgBytes = createDummyPNG()
					}

					// Ensure we have valid, standard PNG bytes for gofpdf
					imgBytes = ensurePNG(imgBytes)

					// Include pageNumber in the key so gofpdf doesn't reuse a cached image
					// from a different page that happened to have the same panelNumber.
					imgName := fmt.Sprintf("job_%s_page_%d_panel_%d", jobID, page.PageNumber, pa.panel.PanelNumber)
					reader := bytes.NewReader(imgBytes)
					pdf.RegisterImageReader(imgName, "PNG", reader)

					// Coordinate calculation
					var x, y, w, h float64
					if numPanels == 1 {
						// Single panel fills the page
						x, y, w, h = 10, 25, 190, 230
					} else if numPanels == 2 {
						// Two panels stacked vertically
						h = 110
						w = 190
						x = 10
						if i == 0 {
							y = 25
						} else {
							y = 145
						}
					} else if numPanels == 3 {
						// Three panels: 1 top wide panel, 2 side-by-side bottom panels
						if i == 0 {
							x, y, w, h = 10, 25, 190, 110
						} else if i == 1 {
							x, y, w, h = 10, 145, 92, 110
						} else {
							x, y, w, h = 108, 145, 92, 110
						}
					} else if numPanels == 4 {
						// Four panels in a 2x2 grid
						w = 92
						h = 110
						if i == 0 {
							x, y = 10, 25
						} else if i == 1 {
							x, y = 108, 25
						} else if i == 2 {
							x, y = 10, 145
						} else {
							x, y = 108, 145
						}
					} else {
						// Generic grid layout for any higher number of panels
						cols := 2
						rows := (numPanels + 1) / 2
						colW := 92.0
						rowH := 240.0 / float64(rows)
						if rowH > 110 {
							rowH = 110
						}
						colIndex := i % cols
						rowIndex := i / cols
						x = 10.0 + float64(colIndex)*98.0
						y = 25.0 + float64(rowIndex)*(rowH+10.0)
						w = colW
						h = rowH
					}

					// Draw panel image
					pdf.Image(imgName, x, y, w, h, false, "", 0, "")

					// Draw a neat comic border around the panel
					pdf.SetLineWidth(0.8)
					pdf.SetDrawColor(15, 23, 42)
					pdf.Rect(x, y, w, h, "D")

					// Calculate text wrapping and required height
					pdf.SetFont("Helvetica", "B", 8)
					captionLines := pdf.SplitText(cleanPDFText(pdf, pa.panel.Caption), w-6)
					
					var dialogueLines []string
					if len(pa.panel.Dialogue) > 0 {
						pdf.SetFont("Helvetica", "I", 7)
						for _, d := range pa.panel.Dialogue {
							dLines := pdf.SplitText(cleanPDFText(pdf, d), w-6)
							dialogueLines = append(dialogueLines, dLines...)
						}
					}

					captionHeight := float64(len(captionLines)) * 4.0
					dialogueHeight := float64(len(dialogueLines)) * 3.5
					
					boxPadding := 4.0
					spacing := 0.0
					if len(dialogueLines) > 0 && len(captionLines) > 0 {
						spacing = 2.0
					}
					
					boxH := captionHeight + dialogueHeight + spacing + boxPadding
					maxBoxH := h * 0.45 // Limit to 45% of panel height
					if boxH > maxBoxH {
						boxH = maxBoxH
					}

					boxY := y + h - boxH

					// Draw a styled semi-transparent caption box at the bottom of the panel
					pdf.SetFillColor(255, 255, 255)
					pdf.SetDrawColor(15, 23, 42)
					pdf.SetLineWidth(0.3)
					pdf.SetAlpha(0.92, "Normal")
					pdf.Rect(x, boxY, w, boxH, "FD")
					pdf.SetAlpha(1.0, "Normal")

					// Render Caption
					pdf.SetTextColor(15, 23, 42)
					pdf.SetFont("Helvetica", "B", 8)
					currentY := boxY + 2.5
					for _, line := range captionLines {
						if currentY+4.0 > boxY+boxH {
							break
						}
						pdf.SetXY(x+3, currentY)
						pdf.Cell(w-6, 4, line)
						currentY += 4.0
					}

					// Render Dialogue
					if len(dialogueLines) > 0 {
						pdf.SetTextColor(30, 41, 59)
						pdf.SetFont("Helvetica", "I", 7)
						currentY += spacing
						for _, line := range dialogueLines {
							if currentY+3.5 > boxY+boxH {
								break
							}
							pdf.SetXY(x+3, currentY)
							pdf.Cell(w-6, 3.5, line)
							currentY += 3.5
						}
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
				INSERT INTO assets(id, user_id, generation_job_id, asset_type, bucket, object_key, original_filename, mime_type, size_bytes)
				VALUES($1::uuid, $2::uuid, $3::uuid, 'comic_pdf', $4, $5, 'story.pdf', 'application/pdf', $6)
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
		WHERE id = $2::uuid
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

func ensurePNG(data []byte) []byte {
	if len(data) == 0 {
		return createDummyPNG()
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return createDummyPNG()
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return createDummyPNG()
	}
	return buf.Bytes()
}

func createDummyPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	// Draw a vibrant gradient background with dark comic border
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			if x < 8 || x > 391 || y < 8 || y > 291 {
				img.Set(x, y, color.RGBA{R: 15, G: 23, B: 42, A: 255}) // Dark slate border
			} else {
				// Cyan to indigo gradient
				r := uint8(14 + (x * 30 / 400))
				g := uint8(116 + (y * 50 / 300))
				b := uint8(144 + ((x + y) * 40 / 700))
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
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

func cleanPDFText(pdf *gofpdf.Fpdf, s string) string {
	s = strings.ReplaceAll(s, "’", "'")
	s = strings.ReplaceAll(s, "‘", "'")
	s = strings.ReplaceAll(s, "“", "\"")
	s = strings.ReplaceAll(s, "”", "\"")
	s = strings.ReplaceAll(s, "—", "-")
	s = strings.ReplaceAll(s, "–", "-")
	s = strings.ReplaceAll(s, "…", "...")

	tr := pdf.UnicodeTranslatorFromDescriptor("")
	s = tr(s)

	var buf bytes.Buffer
	for _, r := range s {
		if r < 256 {
			buf.WriteRune(r)
		} else {
			buf.WriteRune('?')
		}
	}
	return buf.String()
}
