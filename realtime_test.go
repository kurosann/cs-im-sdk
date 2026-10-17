package csim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSubscribeAgentMessagesFiltersCustomerMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/channels/app_1/realtime" {
			t.Fatalf("realtime path = %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("Accept header = %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		_, _ = w.Write([]byte("data: {\"type\":\"message.created\",\"data\":{\"conversationId\":\"conv_1\",\"sender\":{\"role\":\"customer\"}}}\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("data: {\"type\":\"message.created\",\"data\":{\"conversationId\":\"conv_1\",\"sender\":{\"role\":\"agent\",\"id\":\"agent_1\"},\"messageId\":\"msg_1\"}}\n\n"))
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	events, closeFn, err := NewClient(server.URL, "token").SubscribeAgentMessages(ctx, "app_1", "conv_1")
	if err != nil {
		t.Fatalf("SubscribeAgentMessages() error = %v", err)
	}
	defer closeFn()

	event, ok := <-events
	if !ok || event.ID != "msg_1" {
		t.Fatalf("agent event = %+v ok=%t", event, ok)
	}
	select {
	case event := <-events:
		t.Fatalf("customer message leaked: %+v", event)
	case <-ctx.Done():
	}
}
