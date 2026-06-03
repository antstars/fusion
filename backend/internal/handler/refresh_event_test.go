package handler

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestRefreshEventBrokerPublishesAndUnsubscribes(t *testing.T) {
	broker := newRefreshEventBroker()
	id, events := broker.subscribe()

	feedID := int64(42)
	want := refreshEvent{
		Type:       "refresh_completed",
		Scope:      refreshEventScopeFeed,
		FeedID:     &feedID,
		FinishedAt: time.Now().Unix(),
	}
	broker.publish(want)

	got := waitForRefreshEvent(t, events)
	if got.Type != want.Type || got.Scope != want.Scope || got.FeedID == nil || *got.FeedID != feedID {
		t.Fatalf("unexpected event: %#v", got)
	}

	broker.unsubscribe(id)
	broker.mu.Lock()
	subscriberCount := len(broker.subscribers)
	broker.mu.Unlock()
	if subscriberCount != 0 {
		t.Fatalf("expected subscribers to be cleaned up, got %d", subscriberCount)
	}

	broker.publish(refreshEvent{Type: "refresh_completed", Scope: refreshEventScopeAll})

	select {
	case event := <-events:
		t.Fatalf("received event after unsubscribe: %#v", event)
	case <-time.After(25 * time.Millisecond):
	}
}

func TestWriteRefreshSSE(t *testing.T) {
	var buf bytes.Buffer
	event := refreshEvent{
		Type:       "refresh_completed",
		Scope:      refreshEventScopeAll,
		FinishedAt: 123,
	}

	if err := writeRefreshSSE(&buf, event); err != nil {
		t.Fatalf("write SSE: %v", err)
	}

	output := buf.String()
	if !bytes.HasPrefix(buf.Bytes(), []byte("event: refresh-completed\ndata: ")) {
		t.Fatalf("unexpected SSE prefix: %q", output)
	}

	dataStart := len("event: refresh-completed\ndata: ")
	dataEnd := len(output) - len("\n\n")
	var got refreshEvent
	if err := json.Unmarshal([]byte(output[dataStart:dataEnd]), &got); err != nil {
		t.Fatalf("decode SSE payload: %v\n%s", err, output)
	}
	if got.Type != event.Type || got.Scope != event.Scope || got.FinishedAt != event.FinishedAt {
		t.Fatalf("unexpected SSE payload: %#v", got)
	}
}

func waitForRefreshEvent(t *testing.T, events <-chan refreshEvent) refreshEvent {
	t.Helper()

	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatal("refresh event was not published")
		return refreshEvent{}
	}
}
