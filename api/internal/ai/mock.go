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
