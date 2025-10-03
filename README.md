# OCR Library

A Go library for extracting information from medical professional credentials (doctors, nurses, pharmacists) using OpenAI's GPT-4o Vision API.

## Features

- 🔍 Extract names from ID cards, licenses, and certificates
- 📋 Extract detailed information based on profession type
- 🏥 Support for multiple medical professions (Doctor, Nurse, Pharmacist)
- 📨 Message queue integration for data pipeline
- 🛡️ Comprehensive error handling
- 🔧 Flexible configuration options

## Installation

```bash
go get github.com/A-pen-app/ocr
```

## Requirements

- Go 1.23.0 or higher
- OpenAI API key
- Message queue (implements `mq.MQ` interface)

## Usage

### Basic Setup

```go
import (
    "context"
    "github.com/A-pen-app/ocr/models"
    "github.com/A-pen-app/ocr/store"
    "github.com/openai/openai-go/v2"
)

// Initialize OpenAI client
client := openai.NewClient(openai.WithAPIKey("your-api-key"))

// Create OCR store with default configuration
ocrStore := store.NewOpenAIStore(mq, client, nil)
```

### Custom Configuration

```go
config := &store.OpenAIConfig{
    MaxToken:    2048,
    Model:       openai.ChatModelGPT4o,
    Topic:       models.OCRTopicProd,
    MessageType: models.OCRMessageTypeIdentifyOCR,
}

ocrStore := store.NewOpenAIStore(mq, client, config)
```

### Scan Name Only

```go
ctx := context.Background()
imageURL := "https://example.com/id-card.jpg"

name, err := ocrStore.ScanName(ctx, imageURL)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Name: %s\n", name)
```

### Scan Complete Information

```go
ctx := context.Background()
userID := "user-123"
imageURL := "https://example.com/doctor-license.jpg"

// For doctor
ocrInfo, err := ocrStore.ScanRawInfo(ctx, userID, imageURL, models.PlatformTypeApen)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Name: %s\n", *ocrInfo.Name)
fmt.Printf("Birthday: %s\n", *ocrInfo.Birthday)
fmt.Printf("Position: %s\n", *ocrInfo.Position)
fmt.Printf("Department: %s\n", *ocrInfo.Department)
fmt.Printf("Facility: %s\n", *ocrInfo.Facility)
```

## Supported Platform Types

| Platform Type | Description | Extracted Fields |
|--------------|-------------|------------------|
| `PlatformTypeApen` | Doctor | Name, Birthday, Position, Department, Facility, Valid Date, Specialty Valid Date |
| `PlatformTypeNurse` | Nurse | Name, Birthday, Department, Facility, Valid Date |
| `PlatformTypePhar` | Pharmacist | Name, Birthday, Facility, Valid Date |

## API Reference

### `NewOpenAIStore`

Creates a new OCR store instance.

```go
func NewOpenAIStore(
    mq mq.MQ,
    client *openai.Client,
    config *OpenAIConfig,
) OCR
```

**Parameters:**
- `mq`: Message queue for publishing OCR results
- `client`: OpenAI client instance
- `config`: Configuration (optional, uses default if nil)

**Returns:** OCR interface implementation

### `ScanName`

Extracts only the name from an image.

```go
func (os *ocrStore) ScanName(
    ctx context.Context,
    link string,
) (string, error)
```

**Parameters:**
- `ctx`: Context for request cancellation
- `link`: URL of the image to scan

**Returns:** Extracted name and error (if any)

### `ScanRawInfo`

Extracts comprehensive information based on profession type.

```go
func (os *ocrStore) ScanRawInfo(
    ctx context.Context,
    userID string,
    link string,
    platformType models.PlatformType,
) (*models.OCRRawInfo, error)
```

**Parameters:**
- `ctx`: Context for request cancellation
- `userID`: User identifier for tracking
- `link`: URL of the image to scan
- `platformType`: Type of profession (Apen/Nurse/Phar)

**Returns:** Extracted OCR information and error (if any)

## Data Models

### `OCRRawInfo`

```go
type OCRRawInfo struct {
    IdentifyURL        *string `json:"identify_url,omitempty"`
    Name               *string `json:"name"`
    Birthday           *string `json:"birthday"`
    Position           *string `json:"position,omitempty"`
    Department         *string `json:"department,omitempty"`
    Facility           *string `json:"facility,omitempty"`
    ValidDate          *string `json:"valid_date,omitempty"`
    SpecialtyValidDate *string `json:"specialty_valid_date,omitempty"`
}
```

### `OpenAIConfig`

```go
type OpenAIConfig struct {
    MaxToken    int64                  // Maximum tokens for response
    Model       openai.ChatModel       // OpenAI model to use
    Topic       models.OCRTopic        // Message queue topic
    MessageType models.OCRMessageType  // Message type identifier
}
```

**Default Values:**
- `MaxToken`: 1024
- `Model`: GPT-4o
- `Topic`: wanderer-dev
- `MessageType`: identify_ocr

## Position Types (for Doctors)

- `PGY` - Post-Graduate Year (不分科醫師)
- `Resident` - Resident Doctor (住院醫師)
- `VS` - Visiting Staff / Attending Physician (主治醫師)

## Error Handling

The library provides detailed error messages for common issues:

- `"openai client is not initialized"` - Client not properly configured
- `"empty response choices from OCR"` - No response from OpenAI API
- `"empty response content from OCR"` - Empty content in API response
- JSON unmarshal errors for invalid response format

## Message Queue Integration

When `ScanRawInfo` is called, the result is automatically published to the configured message queue topic as an `OCREventMessage`:

```go
type OCREventMessage struct {
    UserID    string     `json:"user_id"`
    Payload   OCRRawInfo `json:"payload"`
    CreatedAt time.Time  `json:"created_at"`
    Type      string     `json:"type"`
    Source    string     `json:"source"`
}
```

## Examples

### Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/A-pen-app/ocr/models"
    "github.com/A-pen-app/ocr/store"
    "github.com/openai/openai-go/v2"
)

func main() {
    ctx := context.Background()

    // Initialize OpenAI client
    client := openai.NewClient(
        openai.WithAPIKey("your-api-key"),
    )

    // Initialize your message queue
    // mq := ... (your MQ implementation)

    // Create OCR store
    ocrStore := store.NewOpenAIStore(mq, client, nil)

    // Scan doctor's license
    imageURL := "https://example.com/doctor-license.jpg"
    result, err := ocrStore.ScanRawInfo(
        ctx,
        "user-123",
        imageURL,
        models.PlatformTypeApen,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Print results
    if result.Name != nil {
        fmt.Printf("Name: %s\n", *result.Name)
    }
    if result.Position != nil {
        fmt.Printf("Position: %s\n", *result.Position)
    }
    if result.Department != nil {
        fmt.Printf("Department: %s\n", *result.Department)
    }
}
```

## License

Copyright © 2025 A-pen

## Contributing

This is a private library. For issues or questions, please contact the development team.

## Dependencies

- [openai-go](https://github.com/openai/openai-go) - OpenAI API client
- [A-pen-app/mq](https://github.com/A-pen-app/mq) - Message queue abstraction
- [A-pen-app/logging](https://github.com/A-pen-app/logging) - Logging utilities

