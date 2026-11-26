package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPeekSingleProcess(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.yml")

	cases := []struct {
		name string
		data string
		want bool
	}{
		{"enabled", "version: 1\nsettings:\n  single_process: true\n", true},
		{"disabled", "version: 1\nsettings:\n  single_process: false\n", false},
		{"comment", "version: 1\nsettings:\n  single_process: true # 面板模式\n", true},
		// 历史配置可能存在重复键，yaml.Unmarshal 会整体失败，需要按行兜底。
		{"duplicate-keys-fallback", "version: 1\nsettings:\n  single_process: true\n  allow_proactive_msg: true\n  allow_proactive_msg: true\n", true},
		{"missing", "version: 1\nsettings:\n  port: \"15630\"\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(file, []byte(tc.data), 0644); err != nil {
				t.Fatal(err)
			}
			if got := PeekSingleProcess(file); got != tc.want {
				t.Fatalf("PeekSingleProcess() = %v, want %v", got, tc.want)
			}
		})
	}
}
