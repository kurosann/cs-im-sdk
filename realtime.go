package csim

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type AgentMessage struct {
	Message
}

type realtimeEvent struct {
	Type string          `json:"type"`
	ID   string          `json:"id,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
	At   string          `json:"at"`
}

func (c *Client) SubscribeAgentMessages(ctx context.Context, conversationID string) (<-chan AgentMessage, func(), error) {
	if conversationID == "" {
		return nil, nil, fmt.Errorf("conversationId is required")
	}

	if conversationID == "" {
		return nil, nil, fmt.Errorf("conversationId is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/ws/v1/realtime", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create realtime request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.appToken)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("connect realtime: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("realtime connect failed: status %d", resp.StatusCode)
	}

	events := make(chan AgentMessage, 64)
	var closeOnce sync.Once
	done := make(chan struct{})

	go func() {
		defer close(events)
		defer closeOnce.Do(func() { resp.Body.Close() })
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			default:
			}
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			var event realtimeEvent
			if err := json.Unmarshal([]byte(line[6:]), &event); err != nil {
				continue
			}
			if event.Type != "message.created" {
				continue
			}
			var message Message
			if err := json.Unmarshal(event.Data, &message); err != nil {
				continue
			}
			if message.Sender.Role != "agent" {
				continue
			}
			if conversationID != "" && message.ConversationID != conversationID {
				continue
			}
			select {
			case events <- AgentMessage{Message: message}:
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		}
	}()

	closeFn := func() {
		closeOnce.Do(func() { close(done); resp.Body.Close() })
	}
	return events, closeFn, nil
}
