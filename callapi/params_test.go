package callapi

import (
	"encoding/json"
	"testing"
)

func TestParamsPresenceForOneBotDefaults(t *testing.T) {
	var withValues ActionMessage
	if err := json.Unmarshal([]byte(`{"action":"set_group_ban","params":{"group_id":1,"user_id":2,"duration":0,"enable":false,"approve":false}}`), &withValues); err != nil {
		t.Fatal(err)
	}
	if !withValues.Params.DurationSet || withValues.Params.Duration != 0 {
		t.Fatalf("duration presence not tracked: %+v", withValues.Params)
	}
	if !withValues.Params.EnableSet || withValues.Params.Enable {
		t.Fatalf("enable presence not tracked: %+v", withValues.Params)
	}
	if !withValues.Params.ApproveSet || withValues.Params.Approve {
		t.Fatalf("approve presence not tracked: %+v", withValues.Params)
	}

	var omitted ActionMessage
	if err := json.Unmarshal([]byte(`{"action":"set_group_ban","params":{"group_id":1,"user_id":2}}`), &omitted); err != nil {
		t.Fatal(err)
	}
	if omitted.Params.DurationSet || omitted.Params.EnableSet || omitted.Params.ApproveSet {
		t.Fatalf("omitted fields should not be marked set: %+v", omitted.Params)
	}
}
