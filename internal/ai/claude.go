package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const defaultModel = "claude-opus-4-8"

// Client wraps the Anthropic Messages API for generating growth summaries.
// It is a no-op (Enabled() == false) when ANTHROPIC_API_KEY is unset, so the
// rest of the app never has to special-case a missing key.
type Client struct {
	model  string
	client anthropic.Client
}

func New() *Client {
	model := os.Getenv("GTZY_AI_MODEL")
	if model == "" {
		model = defaultModel
	}
	var opts []option.RequestOption
	return &Client{
		model:  model,
		client: anthropic.NewClient(opts...),
	}
}

func (c *Client) Enabled() bool {
	return os.Getenv("ANTHROPIC_API_KEY") != ""
}

func (c *Client) Model() string {
	return c.model
}

// Generate sends prompt to Claude and returns the plain-text response.
// periodType/periodKey are accepted for interface symmetry with the summary
// store but aren't used in the request itself.
func (c *Client) Generate(periodType, periodKey, prompt string) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("AI summaries are disabled: ANTHROPIC_API_KEY is not set")
	}

	msg, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 4096,
		Thinking:  anthropic.ThinkingConfigParamUnion{OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("claude request failed: %w", err)
	}

	var out string
	for _, block := range msg.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			out += tb.Text
		}
	}
	if out == "" {
		return "", fmt.Errorf("claude returned no text content (stop_reason=%s)", msg.StopReason)
	}
	return out, nil
}
