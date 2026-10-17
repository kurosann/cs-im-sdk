package model

import "time"

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
