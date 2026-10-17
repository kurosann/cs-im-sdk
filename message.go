package csim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kurosann/cs-im-sdk/model"
)

type SendCustomerMessageRequest struct {
	ConversationID string
	CustomerID     string
	Type           string
	Text           string
}

func (c *Client) SendCustomerMessage(ctx context.Context, req SendCustomerMessageRequest) (*model.Message, error) {
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
	var message model.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("parse message response: %w", err)
	}
	return &message, nil
}
