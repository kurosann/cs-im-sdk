package csim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SendCustomerMessageRequest struct {
	ConversationID string
	CustomerID     string
	Type           string
	Text           string
}

type Message struct {
	ID             string        `json:"messageId"`
	Seq            int64         `json:"seq"`
	ConversationID string        `json:"conversationId"`
	CustomerID     string        `json:"customerId"`
	Sender         MessageSender `json:"sender"`
	Source         MessageSource `json:"source,omitempty"`
	Type           string        `json:"type"`
	Text           string        `json:"text,omitempty"`
	Status         string        `json:"status"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type MessageSender struct {
	Role string `json:"role"`
	ID   string `json:"id"`
}

type MessageSource struct {
	AppID      string `json:"appId,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
}

func (c *Client) SendCustomerMessage(ctx context.Context, req SendCustomerMessageRequest) (*Message, error) {
	if req.ConversationID == "" {
		return nil, fmt.Errorf("conversationId is required")
	}
	if req.CustomerID == "" {
		return nil, fmt.Errorf("customerId is required")
	}
	if req.Text == "" {
		return nil, fmt.Errorf("text is required")
	}
	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}

	payload := map[string]any{
		"conversationId": req.ConversationID,
		"customerId":     req.CustomerID,
		"sender": map[string]string{
			"role": "customer",
			"id":   req.CustomerID,
		},
		"type": msgType,
		"text": req.Text,
	}

	data, err := c.do(ctx, http.MethodPost, "/api/v1/messages", payload)
	if err != nil {
		return nil, err
	}
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("parse message response: %w", err)
	}
	return &message, nil
}
