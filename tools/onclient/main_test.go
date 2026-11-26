package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestCall 用本地假 OneBot WS 服务验证请求/响应往返
func TestCall(t *testing.T) {
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var req map[string]interface{}
			if err := json.Unmarshal(data, &req); err != nil {
				continue
			}
			resp := map[string]interface{}{
				"status":  "ok",
				"retcode": 0,
				"echo":    req["echo"],
				"data":    map[string]interface{}{"action": req["action"]},
			}
			raw, _ := json.Marshal(resp)
			_ = conn.WriteMessage(websocket.TextMessage, raw)
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	c, err := dial(wsURL, "", 5*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()

	pack, err := c.Call("get_status", map[string]interface{}{}, 5*time.Second)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !isOK(pack) {
		t.Fatalf("expected ok, got %+v", pack)
	}
	data, _ := pack["data"].(map[string]interface{})
	if data["action"] != "get_status" {
		t.Fatalf("unexpected data: %+v", data)
	}
}

func TestSubst(t *testing.T) {
	vars := map[string]string{"group_id": "187781763", "user_id": "800512121", "victim_id": "42"}
	got, _ := subst(map[string]interface{}{
		"group_id": "${group_id}",
		"user_id":  "${user_id}",
		"text":     "g=${group_id} u=${user_id}",
		"list":     []interface{}{"${victim_id}"},
	}, vars).(map[string]interface{})

	if got["group_id"] != int64(187781763) {
		t.Fatalf("group_id should be int64, got %T %v", got["group_id"], got["group_id"])
	}
	if got["user_id"] != int64(800512121) {
		t.Fatalf("user_id should be int64, got %T %v", got["user_id"], got["user_id"])
	}
	if got["text"] != "g=187781763 u=800512121" {
		t.Fatalf("inline substitution failed: %v", got["text"])
	}
	list := got["list"].([]interface{})
	if list[0] != int64(42) {
		t.Fatalf("list substitution failed: %v", list[0])
	}
}

// TestSubstMissingVar 未提供的变量保持原样，便于在结果里直接看出漏填
func TestSubstMissingVar(t *testing.T) {
	got := subst("${message_id}", map[string]string{})
	if got != "${message_id}" {
		t.Fatalf("missing var should be kept, got %v", got)
	}
}
