package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// MockProvider implements the ai.Provider interface using deterministic mock responses.
type MockProvider struct{}

// NewMockProvider creates a new mock AI provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// GenerateImage returns a mock-generated image result. In production, this would call OpenAI or similar.
func (m *MockProvider) GenerateImage(ctx context.Context, req ImageRequest) (ImageResult, error) {
	// Generate a unique request ID
	requestID := generateRequestID()

	// Simulate a simple deterministic image generation based on the prompt
	// In a real implementation, this would call OpenAI's image API
	imageURL := fmt.Sprintf("https://placeholder-images.local/mock/%s.png", requestID)

	return ImageResult{
		URL:               imageURL,
		ProviderRequestID: requestID,
	}, nil
}

// ModerateText returns whether text violates content policies. For mock, we return false (no violations).
func (m *MockProvider) ModerateText(ctx context.Context, text string) (bool, error) {
	// Mock: check if text contains obviously harmful keywords (simplistic demo)
	harmfulKeywords := []string{"violence", "hate", "illegal"}
	lowerText := strings.ToLower(text)

	for _, keyword := range harmfulKeywords {
		if strings.Contains(lowerText, keyword) {
			return true, nil // flagged as harmful
		}
	}

	return false, nil // safe
}

// generateRequestID creates a unique request ID.
func generateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// fallback: use timestamp
		return fmt.Sprintf("mock_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("mock_%s", hex.EncodeToString(bytes))
}

// GenerateText returns a mock response for LLM calls based on keywords in the prompts.
func (m *MockProvider) GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	lowerSys := strings.ToLower(systemPrompt)
	if strings.Contains(lowerSys, "safety") || strings.Contains(lowerSys, "moderate") || strings.Contains(lowerSys, "violation") {
		return `{"safe": true, "reason": "content is safe"}`, nil
	}
	if strings.Contains(lowerSys, "narrative") || strings.Contains(lowerSys, "comic writer") {
		// Return a 2 page, 2 panels per page sample narrative JSON
		return `{
			"title": "A Day of Learning",
			"theme": "Overcoming obstacles and succeeding",
			"logline": "A developer faces a tough bug, but triumphs in the end.",
			"pages": [
				{
					"pageNumber": 1,
					"purpose": "Introduce the challenge",
					"panels": [
						{
							"panelNumber": 1,
							"visual": "A developer staring intently at a glowing monitor in a dark room.",
							"caption": "It was a dark and stormy night when the server crashed.",
							"dialogue": ["Why won't this compile?!"],
							"emotion": "frustrated"
						},
						{
							"panelNumber": 2,
							"visual": "A close up of lines of red error logs on screen.",
							"caption": "The bugs were piling up, and time was running out.",
							"dialogue": ["I need to find the root cause."],
							"emotion": "focused"
						}
					]
				},
				{
					"pageNumber": 2,
					"purpose": "Resolve the challenge",
					"panels": [
						{
							"panelNumber": 3,
							"visual": "The developer smiles as they write a brilliant fix.",
							"caption": "Suddenly, the solution became crystal clear.",
							"dialogue": ["Aha! That's the bug!"],
							"emotion": "happy"
						},
						{
							"panelNumber": 4,
							"visual": "The developer cheering in front of the working system.",
							"caption": "With the fix deployed, peace returned to the codebase.",
							"dialogue": ["It works! All tests green!"],
							"emotion": "triumphant"
						}
					]
				}
			]
		}`, nil
	}
	if strings.Contains(lowerSys, "art direction") || strings.Contains(lowerSys, "visual style") {
		return `{
			"styleDescription": "Modern vibrant comic book style",
			"globalColorPalette": "High contrast cyan and warm amber",
			"panels": [
				{"panelNumber": 1, "visualDescription": "Heroic portrait framing", "cameraAngle": "Medium shot", "lighting": "Dramatic neon rim lighting"},
				{"panelNumber": 2, "visualDescription": "Detailed terminal screen close-up", "cameraAngle": "Close-up", "lighting": "Soft blue screen glow"},
				{"panelNumber": 3, "visualDescription": "Triumphant eureka moment", "cameraAngle": "Low angle", "lighting": "Warm golden sunlight"},
				{"panelNumber": 4, "visualDescription": "Panoramic success view", "cameraAngle": "Wide shot", "lighting": "Bright natural lighting"}
			]
		}`, nil
	}
	return "Mock text response", nil
}
