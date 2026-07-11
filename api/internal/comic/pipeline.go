package comic

// Pipeline is a deterministic, resumable workflow. Each stage persists its
// validated JSON output in generation_jobs.output before the next stage starts.
// Workers can therefore safely resume after a restart without regenerating
// completed panels.
type Stage string
const (NarrativePlanner Stage = "narrative_planner"; SafetyReview Stage = "safety_review"; ArtDirection Stage = "art_direction"; PanelGeneration Stage = "panel_generation"; PDFComposition Stage = "pdf_composition")
type Narrative struct { Title string `json:"title"`; Theme string `json:"theme"`; Logline string `json:"logline"`; Pages []Page `json:"pages"` }
type Page struct { PageNumber int `json:"pageNumber"`; Purpose string `json:"purpose"`; Panels []Panel `json:"panels"` }
type Panel struct { PanelNumber int `json:"panelNumber"`; Visual string `json:"visual"`; Caption string `json:"caption"`; Dialogue []string `json:"dialogue"`; Emotion string `json:"emotion"` }
