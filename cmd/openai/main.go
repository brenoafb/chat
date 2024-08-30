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

	openai "github.com/sashabaranov/go-openai"
)

var sysPromptPath = flag.String("s", "", "file containing system prompt")
var model = flag.String("m", "gpt-4o", "openai model name")
var maxTokens = flag.Int("max", 1000, "max number of tokens")

func main() {
	flag.Parse()

	token := os.Getenv("OPENAI_TOKEN")
	var c *openai.Client

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c = openai.NewClient(token)
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

	var inputMessages []openai.ChatCompletionMessage
	wg.Add(1)
	go func() {
		defer wg.Done()
		input, err := io.ReadAll(os.Stdin)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %s\n", err)
			os.Exit(1)
		}

		inputMessages = parseInput(string(input))
	}()

	wg.Wait()

	ctx := context.Background()

	if sysPrompt != "" {
		inputMessages = append([]openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleSystem,
			Content: sysPrompt,
		}}, inputMessages...)
	}

	req := openai.ChatCompletionRequest{
		Model:     *model,
		MaxTokens: *maxTokens,
		Messages:  inputMessages,
		Stream:    true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ChatCompletionStream error: %v\n", err)
		return
	}

	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Printf("\n")
			return
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			return
		}

		fmt.Print(response.Choices[0].Delta.Content)
	}
}

func parseInput(input string) []openai.ChatCompletionMessage {
	parts := strings.Split(input, "%%")

	role := openai.ChatMessageRoleUser
	// if input starts with a line containing only %%,
	// we assume that there's an explicit system prompt
	if parts[0] == "%%" {
		role = openai.ChatMessageRoleSystem
	}

	messages := []openai.ChatCompletionMessage{}

	for _, part := range parts {
		part := strings.TrimSpace(part)
		message := openai.ChatCompletionMessage{
			Role:    role,
			Content: part,
		}
		messages = append(messages, message)
	}

	return messages
}

func switchRole(role string) string {
	if role == openai.ChatMessageRoleSystem {
		return openai.ChatMessageRoleUser
	}
	if role == openai.ChatMessageRoleUser {
		return openai.ChatMessageRoleAssistant
	}
	if role == openai.ChatMessageRoleAssistant {
		return openai.ChatMessageRoleUser
	}

	return ""
}
