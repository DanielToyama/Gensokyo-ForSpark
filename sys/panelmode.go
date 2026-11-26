package sys

import "sync/atomic"

// panelMode 由 single_process 配置、-panel 参数或 GENSOKYO_PANEL 环境变量开启。
// 一旦开启就保持到进程结束：配置热重载把 single_process 改回 false 时，
// 当前进程仍按面板模式运行，避免运行中突然恢复“双击保护/新开窗口”行为。
var panelMode atomic.Bool

// SetPanelMode 开启面板模式；传 false 不会关闭已经开启的模式。
func SetPanelMode(enabled bool) {
	if enabled {
		panelMode.Store(true)
	}
}

// PanelMode 返回当前是否为面板/单进程模式。
func PanelMode() bool {
	return panelMode.Load()
}
