package types

import (
	"encoding/json"
	"testing"
)

// TestRunAgentInputMarshalOmitsUnsetOptionals guards the request body the SSE
// client sends. A RunAgentInput with only the required fields set used to be
// written with "messages", "tools", "context", "forwardedProps" and "state"
// all null. The TypeScript RunAgentInputSchema rejects the first four, so a
// server validating with it refused the run.
func TestRunAgentInputMarshalOmitsUnsetOptionals(t *testing.T) {
	data, err := json.Marshal(RunAgentInput{ThreadID: "t1", RunID: "r1"})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got, want any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	_ = json.Unmarshal([]byte(`{"threadId":"t1","runId":"r1","messages":[]}`), &want)
	if !jsonEqual(got, want) {
		t.Errorf("Marshal = %s, want {\"threadId\":\"t1\",\"runId\":\"r1\",\"messages\":[]}", data)
	}
}

// TestRunAgentInputMarshalKeepsSetValues checks set fields are still written,
// including an empty state object, which is a value and not an absence.
func TestRunAgentInputMarshalKeepsSetValues(t *testing.T) {
	data, err := json.Marshal(RunAgentInput{
		ThreadID:       "t1",
		RunID:          "r1",
		State:          map[string]any{},
		Messages:       []Message{{ID: "m1", Role: RoleUser, Content: "hi"}},
		Tools:          []Tool{{Name: "search", Description: "d", Parameters: map[string]any{}}},
		Context:        []Context{{Description: "c", Value: "v"}},
		ForwardedProps: map[string]any{"k": 1},
	})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	for _, key := range []string{"state", "messages", "tools", "context", "forwardedProps"} {
		if got[key] == nil {
			t.Errorf("%s missing or null in %s", key, data)
		}
	}
}

func jsonEqual(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}
