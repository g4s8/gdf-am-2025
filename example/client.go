package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type OpenAIClient struct {
	cli    *http.Client
	apiKey string
}

func NewOpenAIClient(cli *http.Client, apiKey string) *OpenAIClient {
	return &OpenAIClient{
		cli:    cli,
		apiKey: apiKey,
	}
}

func (c *OpenAIClient) Stream(ctx context.Context, input string, model string, errCh chan<- error) <-chan string {
	upPayload := openAIRequest{
		Model:  model,
		Input:  input,
		Stream: true,
	}
	bodyBytes, err := json.Marshal(upPayload)
	if err != nil {
		errCh <- fmt.Errorf("failed to marshal request: %w", err)
		return nil
	}

	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(bodyBytes))
	if err != nil {
		errCh <- fmt.Errorf("failed to create upstream request: %w", err)
		return nil
	}
	upReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Accept", "text/event-stream")

	resCh := make(chan string)
	go func() {
		defer close(resCh)

		upResp, err := c.cli.Do(upReq)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Printf("upstream request canceled")
				return
			}
			errCh <- fmt.Errorf("upstream request failed: %w", err)
		}

		defer upResp.Body.Close()

		if upResp.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("upstream returned non-200 status: %d", upResp.StatusCode)
			return
		}

		buf := make([]byte, 8*1024)
		for {
			n, err := upResp.Body.Read(buf)
			if n > 0 {
				resCh <- string(buf[:n])
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Printf("upstream response read completed")
					return
				}

				if errors.Is(err, context.Canceled) {
					log.Printf("upstream response read canceled")
					return
				}
				errCh <- fmt.Errorf("error reading upstream response: %w", err)
				return
			}
		}
	}()

	return resCh
}

type openAIRequest struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Stream bool   `json:"stream"`
}
