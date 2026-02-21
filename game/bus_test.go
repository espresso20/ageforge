package game

import (
	"testing"
)

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewEventBus()

	received := false
	var receivedData EventData

	bus.Subscribe("test_event", func(e EventData) {
		received = true
		receivedData = e
	})

	bus.Publish(EventData{
		Type:    "test_event",
		Payload: map[string]interface{}{"key": "value"},
	})

	if !received {
		t.Error("handler should have been called")
	}
	if receivedData.Type != "test_event" {
		t.Errorf("received type = %v, want test_event", receivedData.Type)
	}
	if receivedData.Payload["key"] != "value" {
		t.Errorf("received payload key = %v, want value", receivedData.Payload["key"])
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus()

	count := 0
	bus.Subscribe("test", func(e EventData) { count++ })
	bus.Subscribe("test", func(e EventData) { count++ })
	bus.Subscribe("test", func(e EventData) { count++ })

	bus.Publish(EventData{Type: "test"})

	if count != 3 {
		t.Errorf("expected 3 handlers called, got %d", count)
	}
}

func TestEventBus_NoSubscribers(t *testing.T) {
	bus := NewEventBus()
	// Should not panic
	bus.Publish(EventData{Type: "nobody_listening"})
}

func TestEventBus_DifferentEvents(t *testing.T) {
	bus := NewEventBus()

	aCalled := false
	bCalled := false
	bus.Subscribe("event_a", func(e EventData) { aCalled = true })
	bus.Subscribe("event_b", func(e EventData) { bCalled = true })

	bus.Publish(EventData{Type: "event_a"})

	if !aCalled {
		t.Error("event_a handler should have been called")
	}
	if bCalled {
		t.Error("event_b handler should not have been called")
	}
}
