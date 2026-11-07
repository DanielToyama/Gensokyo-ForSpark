package msgmap

import (
	"os"
	"testing"
)

func TestOutboundRoundTrip(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	InitializeDB()
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	// 出站记录: 官方 id + 压缩行 id 双键, 群聊
	RecordOutbound([]string{"ROBOT1.0_abc", "1387423"}, MsgInfo{
		GroupID: "187781763",
		MsgType: "group",
	})
	for _, k := range []string{"ROBOT1.0_abc", "1387423"} {
		info, ok := LookupMessage(k)
		if !ok {
			t.Fatalf("key %s not found", k)
		}
		if !info.Outbound {
			t.Fatalf("key %s outbound flag lost", k)
		}
		if info.MsgType != "group" || info.GroupID != "187781763" {
			t.Fatalf("key %s wrong info: %+v", k, info)
		}
	}

	// 入站记录不受影响(Outbound=false)
	RecordMessage([]string{"999"}, MsgInfo{GroupID: "1", MsgType: "group"})
	if info, ok := LookupMessage("999"); !ok || info.Outbound {
		t.Fatalf("inbound record broken: %+v ok=%v", info, ok)
	}

	// C2C 出站
	RecordOutbound([]string{"555"}, MsgInfo{UserID: "u1", GroupID: "u1", MsgType: "group_private"})
	if info, ok := LookupMessage("555"); !ok || info.MsgType != "group_private" || info.UserID != "u1" {
		t.Fatalf("c2c outbound broken: %+v ok=%v", info, ok)
	}
}
