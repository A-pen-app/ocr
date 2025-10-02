package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/A-pen-app/logging"
	"github.com/A-pen-app/mq/v2"
	"github.com/A-pen-app/ocr/models"
	"github.com/openai/openai-go/v2"
)

type OpenAIConfig struct {
	MaxToken    int64
	Model       openai.ChatModel
	Topic       models.OCRTopic
	MessageType models.OCRMessageType
}

type ocrStore struct {
	mq  mq.MQ
	c   *openai.Client
	cfg *OpenAIConfig
}

// NewOpenAIStore creates a new OpenAI store with dependency injection
func NewOpenAIStore(mq mq.MQ, client *openai.Client, config *OpenAIConfig) OCR {

	if config == nil {
		config = &OpenAIConfig{
			MaxToken:    1024,
			Model:       openai.ChatModelGPT4o,
			Topic:       models.OCRTopicDev,
			MessageType: models.OCRMessageTypeIdentifyOCR,
		}
	}

	return &ocrStore{
		mq:  mq,
		c:   client,
		cfg: config,
	}
}

func (os *ocrStore) ScanName(ctx context.Context, link string) (string, error) {

	if os.c == nil {
		return "", fmt.Errorf("openai client is not initialized")
	}

	resp, err := os.c.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:     os.cfg.Model,
		MaxTokens: openai.Int(os.cfg.MaxToken),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(models.SystemContent),
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				openai.TextContentPart(models.NamePrompt),
				openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL: link,
				}),
			}),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		},
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response choices from OCR")
	}

	choice := resp.Choices[0]

	ocr := struct {
		Name string `json:"name"`
	}{}

	if err := json.Unmarshal([]byte(choice.Message.Content), &ocr); err != nil {
		return "", err
	}

	return ocr.Name, nil
}

// ScanRawInfo scans the image and returns raw OCR information based on profession type
func (os *ocrStore) ScanRawInfo(ctx context.Context, userID string, link string, platformType models.PlatformType) (*models.OCRRawInfo, error) {

	if os.c == nil {
		return nil, fmt.Errorf("openai client is not initialized")
	}

	prompt := models.GetInfoPrompt(platformType)

	resp, err := os.c.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:     os.cfg.Model,
		MaxTokens: openai.Int(os.cfg.MaxToken),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(models.SystemContent),
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				openai.TextContentPart(prompt),
				openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL: link,
				}),
			}),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response choices from OCR")
	}

	choice := resp.Choices[0]

	if choice.Message.Content == "" {
		return nil, fmt.Errorf("empty response content from OCR")
	}

	// Parse OCR response
	ocr := models.OCRRawInfo{}
	if err := json.Unmarshal([]byte(choice.Message.Content), &ocr); err != nil {
		return nil, err
	}

	// Set identify_url after parsing
	ocr.IdentifyURL = &link

	// Send OCR raw data to Pub/Sub for BigQuery
	if err := os.mq.Send(string(os.cfg.Topic), models.OCREventMessage{
		UserID:    userID,
		Payload:   ocr,
		CreatedAt: time.Now(),
		Type:      string(os.cfg.MessageType),
		Source:    string(platformType),
	}); err != nil {
		logging.Errorw(ctx, "Failed to send ocr result", "error", err)
	}

	return &ocr, nil
}
