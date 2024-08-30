package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/liushuangls/go-anthropic/v2"
)

var sysPromptPath = flag.String("s", "", "file containing system prompt")
var model = flag.String("m", "claude-3-5-sonnet", "anthropic model name")
var maxTokens = flag.Int("max", 1000, "max number of tokens")

func main() {
	flag.Parse()

	token := os.Getenv("ANTHROPIC_API_KEY")
	var c *anthropic.Client

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c = anthropic.NewClient(token)
	}()

	var sysPrompt string

	wg.Add(1)
	go func() {
		defer wg.Done()

		if *sysPromptPath == "" {
			return
		}

		sysPromptFile, err := os.Open(*sysPromptPath)

		if err != nil {
			sysPrompt = "You are a helpful assistant"
			return
		}

		content, err := io.ReadAll(sysPromptFile)

		if err != nil {
			panic(err)
		}

		sysPrompt = string(content)
	}()

	var inputSysPrompt string
	var inputMessages []anthropic.Message
	wg.Add(1)
	go func() {
		defer wg.Done()
		input, err := io.ReadAll(os.Stdin)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %s\n", err)
			os.Exit(1)
		}

		inputSysPrompt, inputMessages = parseInput(string(input))
	}()

	wg.Wait()

	ctx := context.Background()

	// Use the system prompt from file if provided, otherwise use the one from input
	if sysPrompt == "" {
		sysPrompt = inputSysPrompt
	}

	req := anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model:     *model,
			MaxTokens: *maxTokens,
			Messages:  inputMessages,
			System:    sysPrompt,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			fmt.Print(data.Delta.GetText())
		},
	}

	_, err := c.CreateMessagesStream(ctx, req)
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Fprintf(os.Stderr, "Messages stream error, type: %s, message: %s\n", e.Type, e.Message)
		} else {
			fmt.Fprintf(os.Stderr, "Messages stream error: %v\n", err)
		}
		return
	}

	fmt.Println() // Add a newline at the end
}

func parseInput(input string) (string, []anthropic.Message) {
	parts := strings.Split(input, "%%")

	var sysPrompt string
	messages := []anthropic.Message{}

	if len(parts) > 0 && strings.TrimSpace(parts[0]) == "" {
		// If input starts with a line containing only %%, we assume that's an explicit system prompt
		if len(parts) > 1 {
			sysPrompt = strings.TrimSpace(parts[1])
			parts = parts[2:] // Skip the system prompt for message parsing
		} else {
			return "", messages // Empty input after %%
		}
	}

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i%2 == 0 {
			messages = append(messages, anthropic.NewUserTextMessage(part))
		} else {
			messages = append(messages, anthropic.NewAssistantTextMessage(part))
		}
	}

	return sysPrompt, messages
}
