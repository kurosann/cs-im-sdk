package csim

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnsureConversation(t *testing.T) {
	var gotPath string
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"OK","data":{"conversation":{"conversationId":"conv_1"},"created":true}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token")
	conversation, created, err := client.EnsureConversation(context.Background(), EnsureConversationRequest{
		AppID: "app_1", CustomerExternalID: "customer-1", Channel: "web",
	})
	if err != nil {
		t.Fatalf("EnsureConversation() error = %v", err)
	}
	if gotPath != "/api/v1/channels/app_1/conversations/ensure" || !created || conversation.ID != "conv_1" {
		t.Fatalf("EnsureConversation() path=%s created=%t conversation=%+v", gotPath, created, conversation)
	}
	if gotPayload["customerExternalId"] != "customer-1" {
		t.Fatalf("EnsureConversation() payload = %+v", gotPayload)
	}
}

func TestSendCustomerMessage(t *testing.T) {
	var gotPath string
	var gotPayload struct {
		Sender struct {
			Role string `json:"role"`
		} `json:"sender"`
		Source struct {
			AppID      string `json:"appId"`
			ExternalID string `json:"externalId"`
			SentAt     string `json:"sentAt"`
		} `json:"source"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"OK","data":{"messageId":"msg_1","sender":{"role":"customer"}}}`))
	}))
	defer server.Close()

	sentAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	message, err := NewClient(server.URL, "token").SendCustomerMessage(context.Background(), SendCustomerMessageRequest{
		AppID: "app_1", ConversationID: "conv_1", CustomerID: "cus_1",
		ExternalID: "external-1", SentAt: sentAt, Text: "hello",
	})
	if err != nil {
		t.Fatalf("SendCustomerMessage() error = %v", err)
	}
	if gotPath != "/api/v1/channels/app_1/messages" {
		t.Fatalf("SendCustomerMessage() path = %s", gotPath)
	}
	if message.ID != "msg_1" || gotPayload.Sender.Role != "customer" {
		t.Fatalf("SendCustomerMessage() response = %+v payload = %+v", message, gotPayload)
	}
	if gotPayload.Source.ExternalID != "external-1" || gotPayload.Source.SentAt == "" {
		t.Fatalf("SendCustomerMessage() source = %+v", gotPayload.Source)
	}
}

func TestSendCustomerMessageValidations(t *testing.T) {
	if _, err := NewClient("http://localhost", "token").SendCustomerMessage(context.Background(), SendCustomerMessageRequest{
		AppID: "app_1", ConversationID: "conv_1", CustomerID: "cus_1",
		SentAt: time.Now(), Text: "hello",
	}); err == nil {
		t.Fatal("SendCustomerMessage() expected missing externalId error")
	}
	if _, err := NewClient("http://localhost", "token").SendCustomerMessage(context.Background(), SendCustomerMessageRequest{
		AppID: "app_1", ConversationID: "conv_1", CustomerID: "cus_1",
		ExternalID: "external-1", Text: "hello",
	}); err == nil {
		t.Fatal("SendCustomerMessage() expected missing sentAt error")
	}
}

func TestListMessages(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"OK","data":{"items":[{"messageId":"msg_1"}],"hasMore":true}}`))
	}))
	defer server.Close()

	page, err := NewClient(server.URL, "token").ListMessages(context.Background(), "app_1", "conv_1", MessageFilter{AfterSeq: 2, Limit: 10})
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	want := "/api/v1/channels/app_1/conversations/conv_1/messages?afterSeq=2&beforeSeq=0&limit=10&order="
	if gotPath != want {
		t.Fatalf("ListMessages() path = %s, want %s", gotPath, want)
	}
	if !page.HasMore || len(page.Items) != 1 || page.Items[0].ID != "msg_1" {
		t.Fatalf("ListMessages() page = %+v", page)
	}
}
