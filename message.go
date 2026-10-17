package csim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SendCustomerMessageRequest struct {
	AppID          string
	ConversationID string
	CustomerID     string
	ExternalID     string
	SentAt         time.Time
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
	SentAt     string `json:"sentAt,omitempty"`
}

type Conversation struct {
	ID         string `json:"conversationId"`
	AppID      string `json:"appId"`
	CustomerID string `json:"customerId"`
	Channel    string `json:"channel,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Status     string `json:"status,omitempty"`
}

type EnsureConversationRequest struct {
	AppID              string
	CustomerExternalID string
	CustomerName       string
	Channel            string
	Subject            string
}

type MessageFilter struct {
	AfterSeq  int64
	BeforeSeq int64
	Limit     int32
	Order     string
}

type MessagePage struct {
	Items         []Message `json:"items"`
	NextAfterSeq  int64     `json:"nextAfterSeq"`
	NextBeforeSeq int64     `json:"nextBeforeSeq"`
	HasMore       bool      `json:"hasMore"`
}

func (c *Client) SendCustomerMessage(ctx context.Context, req SendCustomerMessageRequest) (*Message, error) {
	if req.AppID == "" {
		return nil, fmt.Errorf("appId is required")
	}
	if req.ConversationID == "" {
		return nil, fmt.Errorf("conversationId is required")
	}
	if req.CustomerID == "" {
		return nil, fmt.Errorf("customerId is required")
	}
	if req.Text == "" {
		return nil, fmt.Errorf("text is required")
	}
	if req.ExternalID == "" {
		return nil, fmt.Errorf("externalId is required")
	}
	if req.SentAt.IsZero() {
		return nil, fmt.Errorf("sentAt is required")
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
		"source": map[string]string{
			"appId":      req.AppID,
			"externalId": req.ExternalID,
			"sentAt":     req.SentAt.UTC().Format(time.RFC3339Nano),
		},
	}

	data, err := c.do(ctx, http.MethodPost, "/api/v1/channels/"+req.AppID+"/messages", payload)
	if err != nil {
		return nil, err
	}
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("parse message response: %w", err)
	}
	return &message, nil
}

func (c *Client) EnsureConversation(ctx context.Context, req EnsureConversationRequest) (*Conversation, bool, error) {
	if req.AppID == "" {
		return nil, false, fmt.Errorf("appId is required")
	}
	if req.CustomerExternalID == "" {
		return nil, false, fmt.Errorf("customerExternalId is required")
	}
	payload := map[string]any{
		"customerExternalId": req.CustomerExternalID,
		"customerName":       req.CustomerName,
		"channel":            req.Channel,
		"subject":            req.Subject,
	}
	data, err := c.do(ctx, http.MethodPost, "/api/v1/channels/"+req.AppID+"/conversations/ensure", payload)
	if err != nil {
		return nil, false, err
	}
	var response struct {
		Conversation Conversation `json:"conversation"`
		Created      bool         `json:"created"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, false, fmt.Errorf("parse conversation response: %w", err)
	}
	return &response.Conversation, response.Created, nil
}

func (c *Client) ListMessages(ctx context.Context, appID, conversationID string, filter MessageFilter) (*MessagePage, error) {
	if appID == "" {
		return nil, fmt.Errorf("appId is required")
	}
	if conversationID == "" {
		return nil, fmt.Errorf("conversationId is required")
	}
	path := fmt.Sprintf("/api/v1/channels/%s/conversations/%s/messages", appID, conversationID)
	path += fmt.Sprintf("?afterSeq=%d&beforeSeq=%d&limit=%d&order=%s", filter.AfterSeq, filter.BeforeSeq, filter.Limit, filter.Order)
	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var page MessagePage
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("parse message page response: %w", err)
	}
	return &page, nil
}
