# Implementation Plan - HTML Comic Generation Agent Stage

Add an AI Agent stage to the comic generation pipeline (`api/internal/comic/pipeline.go`) that creates a rich, self-contained HTML/CSS web comic page based on the narrative and generated panel images. Persist the generated HTML asset in object storage and PostgreSQL database, updating the frontend pipeline progress and result views.

## User Review Required

> [!IMPORTANT]
> The HTML generator agent will run after `PanelGeneration` (or alongside/after `PDFComposition`). It will synthesize the comic script, panel layout, visual art style, and generated panel image URLs into a single responsive, styled HTML/CSS file.
> 
> The HTML asset will be saved to storage at `users/<userID>/comic/<jobID>/story.html`, registered in the `assets` database table (`asset_type = 'comic_html'`), and linked in `PipelineState` with `HTMLAssetID` and `HTMLUrl`.

## Proposed Changes

---

### Backend Components (`api`)

#### [MODIFY] [pipeline.go](file:///c:/Users/Admin/Desktop/Everything/AI_DAY/api/internal/comic/pipeline.go)
- **New Stage**: Define `HTMLComposition Stage = "html_composition"`.
- **System Prompt**: Add `HTMLCompositionSystemPrompt` instructing the AI provider to generate responsive HTML5 & CSS (modern web comic template with clean styling, speech bubbles, fonts, and dark theme).
- **Pipeline Execution**:
  1. Add `HTMLComposition` step after `PanelGeneration` (before or after `PDFComposition`).
  2. Substitute image src paths with `/api/v1/assets/<assetId>/download` or storage URLs for the panel images.
  3. Call `aiProvider.GenerateText` to produce the styled HTML structure.
  4. Write the HTML file to object storage (`storageProvider.Put`) as `application/html`.
  5. Register the new asset record in `assets` table with `asset_type = 'comic_html'`.
  6. Update `PipelineState` struct to include `HTMLAssetID` and `HTMLUrl`.

---

### Frontend Components (`ui`)

#### [MODIFY] [types.ts](file:///c:/Users/Admin/Desktop/Everything/AI_DAY/ui/src/api/types.ts)
- Update `ComicPipelineState` type definition to include `htmlAssetId?: string` and `htmlUrl?: string`.
- Update stage union to include `'html_composition'`.

#### [MODIFY] [JobProgress.tsx](file:///c:/Users/Admin/Desktop/Everything/AI_DAY/ui/src/components/generation/JobProgress.tsx)
- Update `STAGES` array to include the new 6th stage: `{ key: 'html_composition', label: '6. Web Comic (HTML)', icon: Code }`.

#### [MODIFY] [JobResultCard.tsx](file:///c:/Users/Admin/Desktop/Everything/AI_DAY/ui/src/components/generation/JobResultCard.tsx)
- Add a **View / Download Web Comic (HTML)** button alongside the PDF download button, pointing to the newly generated `comic_html` asset endpoint (`/api/v1/assets/${comicOutput.htmlAssetId}/download`).

---

## Verification Plan

### Automated / Backend Verification
- Execute tests or trigger comic pipeline job to verify pipeline advances through `HTMLComposition` stage without errors.
- Confirm state persistence in PostgreSQL `generation_jobs` output column.
- Verify asset record creation in `assets` table with `asset_type = 'comic_html'`.

### Manual & UI Verification
- Run local frontend (`npm run dev`) and test triggering a comic generation job.
- Verify 6-stage pipeline progress in `JobProgress`.
- Verify presence and clickability of the "Download Web Comic (HTML)" button in `JobResultCard`.
