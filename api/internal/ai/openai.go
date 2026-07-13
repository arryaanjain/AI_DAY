package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// OpenAIProvider implements the ai.Provider interface using the official OpenAI Go SDK.
type OpenAIProvider struct {
	client     openai.Client
	imageModel string
	chatModel  string
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(apiKey, imageModel, chatModel string) *OpenAIProvider {
	if chatModel == "" {
		chatModel = string(openai.ChatModelGPT4o)
	}
	if imageModel == "" {
		imageModel = string(openai.ImageModelDallE3)
	}
	return &OpenAIProvider{
		client:     openai.NewClient(option.WithAPIKey(apiKey)),
		imageModel: imageModel,
		chatModel:  chatModel,
	}
}

// GenerateImage calls DALL-E to generate an image from a prompt.
func (o *OpenAIProvider) GenerateImage(ctx context.Context, req ImageRequest) (ImageResult, error) {
	// Construct the prompt. If a SourceAssetURL is provided, we can mention it in the prompt.
	prompt := req.Prompt
	if req.SourceAssetURL != "" {
		// Just a hint for DALL-E to maintain context, though DALL-E 3 doesn't do true image-to-image.
		prompt = fmt.Sprintf("%s (inspired by reference image: %s)", req.Prompt, req.SourceAssetURL)
	}

	result, err := o.client.Images.Generate(ctx, openai.ImageGenerateParams{
		Prompt:         prompt,
		Model:          openai.ImageModel(o.imageModel),
		N:              openai.Int(1),
		Size:           openai.ImageGenerateParamsSize1024x1024,
		ResponseFormat: openai.ImageGenerateParamsResponseFormatURL,
	})
	if err != nil {
		return ImageResult{}, fmt.Errorf("openai image generation failed: %w", err)
	}

	if len(result.Data) == 0 {
		return ImageResult{}, fmt.Errorf("openai returned no image data")
	}

	// We can use a generated or timestamped provider request ID since OpenAI doesn't return a direct API request ID in the model response
	providerReqID := fmt.Sprintf("oai_%d", result.Created)

	return ImageResult{
		URL:               result.Data[0].URL,
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
