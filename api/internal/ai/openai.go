package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

type OpenAIConfig struct {
	APIKey             string
	ImageModel         string
	ChatModel          string
	ImageSize          string
	ImageQuality       string
	OutputRequirements string
	OrgID              string
	ProjectID          string
	BaseURL            string
}

// OpenAIProvider implements the ai.Provider interface using the official OpenAI Go SDK.
type OpenAIProvider struct {
	client             openai.Client
	imageModel         string
	chatModel          string
	imageSize          string
	imageQuality       string
	outputRequirements string
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(cfg OpenAIConfig) *OpenAIProvider {
	if cfg.ChatModel == "" {
		cfg.ChatModel = string(openai.ChatModelGPT4o)
	}
	if cfg.ImageModel == "" {
		cfg.ImageModel = string(openai.ImageModelDallE3)
	}

	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	if cfg.OrgID != "" {
		opts = append(opts, option.WithHeader("OpenAI-Organization", cfg.OrgID))
	}
	if cfg.ProjectID != "" {
		opts = append(opts, option.WithHeader("OpenAI-Project", cfg.ProjectID))
	}

	return &OpenAIProvider{
		client:             openai.NewClient(opts...),
		imageModel:         cfg.ImageModel,
		chatModel:          cfg.ChatModel,
		imageSize:          cfg.ImageSize,
		imageQuality:       cfg.ImageQuality,
		outputRequirements: cfg.OutputRequirements,
	}
}

// GenerateImage calls DALL-E to generate an image from a prompt.
func (o *OpenAIProvider) GenerateImage(ctx context.Context, req ImageRequest) (ImageResult, error) {
	// Construct the prompt. If OutputRequirements or SourceAssetURL are provided, append them.
	prompt := req.Prompt
	if o.outputRequirements != "" {
		prompt = fmt.Sprintf("%s\n\n%s", req.Prompt, o.outputRequirements)
	}
	if req.SourceAssetURL != "" {
		prompt = fmt.Sprintf("%s (inspired by reference image: %s)", prompt, req.SourceAssetURL)
	}

	size := openai.ImageGenerateParamsSize1024x1024
	switch o.imageSize {
	case "1024x1792":
		size = openai.ImageGenerateParamsSize1024x1792
	case "1792x1024":
		size = openai.ImageGenerateParamsSize1792x1024
	case "1024x1536":
		size = openai.ImageGenerateParamsSize1024x1792
	}

	quality := openai.ImageGenerateParamsQuality(strings.ToLower(o.imageQuality))
	if quality == "" {
		quality = openai.ImageGenerateParamsQualityStandard
	}
	// For backward compatibility, map "high" to "hd" when using DALL-E models.
	if quality == "high" && strings.Contains(strings.ToLower(o.imageModel), "dall-e") {
		quality = openai.ImageGenerateParamsQualityHD
	}

	result, err := o.client.Images.Generate(ctx, openai.ImageGenerateParams{
		Prompt:  prompt,
		Model:   openai.ImageModel(o.imageModel),
		N:       openai.Int(1),
		Size:    size,
		Quality: quality,
	})
	if err != nil {
		return ImageResult{}, fmt.Errorf("openai image generation failed: %w", err)
	}

	if len(result.Data) == 0 {
		return ImageResult{}, fmt.Errorf("openai returned no image data")
	}

	// We can use a generated or timestamped provider request ID since OpenAI doesn't return a direct API request ID in the model response
	providerReqID := fmt.Sprintf("oai_%d", result.Created)

	var b64Data []byte
	if result.Data[0].B64JSON != "" {
		var err error
		b64Data, err = base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
		if err != nil {
			return ImageResult{}, fmt.Errorf("failed to decode base64 image data: %w", err)
		}
	}

	return ImageResult{
		URL:               result.Data[0].URL,
		Bytes:             b64Data,
		ProviderRequestID: providerReqID,
	}, nil
}

// ModerateText checks for profanity and policy violations.
func (o *OpenAIProvider) ModerateText(ctx context.Context, text string) (bool, error) {
	// 1. Local profanity check to deny common profanities and slurs instantly
	profanities := []string{
		"fuck", "shit", "asshole", "bitch", "cunt", "bastard", "dick", "pussy",
		"nigger", "faggot", "chink", "kike", "retard", "porn", "naked", "sex",
	}
	lower := strings.ToLower(text)
	for _, word := range profanities {
		if strings.Contains(lower, word) {
			return true, nil // Flagged locally
		}
	}

	// 2. Call OpenAI's Moderation API
	res, err := o.client.Moderations.New(ctx, openai.ModerationNewParams{
		Input: openai.ModerationNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model: openai.ModerationModelOmniModerationLatest,
	})
	if err != nil {
		return false, fmt.Errorf("openai moderation failed: %w", err)
	}

	for _, result := range res.Results {
		if result.Flagged {
			return true, nil // Flagged by OpenAI
		}
	}

	return false, nil
}

// GenerateText calls GPT-4o chat completions to generate narrative text.
func (o *OpenAIProvider) GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model: shared.ChatModel(o.chatModel),
	})
	if err != nil {
		return "", fmt.Errorf("openai chat completion failed: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("openai chat completion returned no choices")
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
