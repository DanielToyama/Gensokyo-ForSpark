// onclient 是一个极简 OneBot v11 测试客户端，用于验证 Gensokyo-ForSpark 的 API。
//
// 它直连 Gensokyo 的正向 WebSocket（配置项 enable_ws_server + port + ws_server_path），
// 发送 {"action","params","echo"} 并打印完整响应，不依赖 SparkBridge / Aemeath 等任何框架。
//
// 用法：
//
//	# 1) 单次调用
//	go run ./tools/onclient -url ws://127.0.0.1:15630 -token 你的token \
//	    -action get_group_info -params '{"group_id":187781763}'
//
//	# 2) 运行测试套件（只读项默认执行，危险项需 -allow-danger）
//	go run ./tools/onclient -url ws://127.0.0.1:15630 -token 你的token \
//	    -group 187781763 -user 800512121 \
//	    -script tools/onclient/suite.example.json
//
//	# 3) 交互模式（不带 -action/-script）
//	go run ./tools/onclient -url ws://127.0.0.1:15630 -token 你的token
//	> get_group_member_list {"group_id":187781763}
//
// 说明：
//   - 群成员列表返回的 user_id 就是后续禁言/踢人等接口要传的 ID（idmap 虚拟号），
//     可直接复制到 -user 或套件变量里。
//   - 套件中的 ${group_id} / ${user_id} 会被 -group / -user 替换。
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// kvFlag 支持重复的 key=value 参数
type kvFlag []string

func (k *kvFlag) String() string { return strings.Join(*k, ",") }

func (k *kvFlag) Set(v string) error {
	*k = append(*k, v)
	return nil
}

// Request 测试套件中的一项
type Request struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Danger bool                   `json:"danger,omitempty"` // 危险操作，需 -allow-danger
}

// Suite 测试套件文件
type Suite struct {
	Name     string    `json:"name"`
	Requests []Request `json:"requests"`
}

// Client 极简 OneBot v11 WebSocket 客户端
type Client struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
	seq     uint64
	respMu  sync.Mutex
	respCh  map[string]chan map[string]interface{}
}

func dial(url, token string, timeout time.Duration) (*Client, error) {
	header := map[string][]string{}
	if token != "" {
		// Gensokyo 支持 Authorization: Bearer <token> 或 Token <token>
		header["Authorization"] = []string{"Bearer " + token}
	}
	d := websocket.Dialer{HandshakeTimeout: timeout}
	conn, _, err := d.Dial(url, header)
	if err != nil {
		return nil, err
	}
	c := &Client{conn: conn, respCh: make(map[string]chan map[string]interface{})}
	go c.readLoop()
	return c, nil
}

// readLoop 读取服务端消息：带 echo 的当作响应分发，其余（事件/心跳）忽略
func (c *Client) readLoop() {
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.respMu.Lock()
			for k, ch := range c.respCh {
				close(ch)
				delete(c.respCh, k)
			}
			c.respMu.Unlock()
			return
		}
		var pack map[string]interface{}
		if err := json.Unmarshal(data, &pack); err != nil {
			continue
		}
		echo, ok := pack["echo"].(string)
		if !ok || echo == "" {
			continue
		}
		c.respMu.Lock()
		ch, exists := c.respCh[echo]
		if exists {
			delete(c.respCh, echo)
		}
		c.respMu.Unlock()
		if exists {
			ch <- pack
			close(ch)
		}
	}
}

// Call 发送 action 并等待对应 echo 的响应
func (c *Client) Call(action string, params map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	echo := fmt.Sprintf("onclient-%d-%d", time.Now().UnixNano(), atomic.AddUint64(&c.seq, 1))

	ch := make(chan map[string]interface{}, 1)
	c.respMu.Lock()
	c.respCh[echo] = ch
	c.respMu.Unlock()

	payload := map[string]interface{}{"action": action, "echo": echo}
	if params != nil {
		payload["params"] = params
	} else {
		payload["params"] = map[string]interface{}{}
	}
	raw, _ := json.Marshal(payload)

	c.writeMu.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, raw)
	c.writeMu.Unlock()
	if err != nil {
		c.respMu.Lock()
		delete(c.respCh, echo)
		c.respMu.Unlock()
		return nil, err
	}

	select {
	case pack, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("连接已关闭")
		}
		return pack, nil
	case <-time.After(timeout):
		c.respMu.Lock()
		delete(c.respCh, echo)
		c.respMu.Unlock()
		return nil, fmt.Errorf("请求超时(%s)", timeout)
	}
}

func (c *Client) Close() { _ = c.conn.Close() }

// isOK 判断 OneBot 成功响应
func isOK(pack map[string]interface{}) bool {
	status, _ := pack["status"].(string)
	code, hasCode := pack["retcode"].(float64)
	if !hasCode {
		return status == "ok"
	}
	return status == "ok" && int(code) == 0
}

func shortJSON(v interface{}, n int) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	s := string(raw)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// subst 递归替换字符串里的 ${var}
func subst(v interface{}, vars map[string]string) interface{} {
	switch x := v.(type) {
	case string:
		// 整串就是 ${var} 且变量是数字时，保持数字类型（group_id/user_id 需要）
		if strings.HasPrefix(x, "${") && strings.HasSuffix(x, "}") && strings.Count(x, "${") == 1 {
			key := x[2 : len(x)-1]
			if val, ok := vars[key]; ok {
				if n, err := strconv.ParseInt(val, 10, 64); err == nil {
					return n
				}
				return val
			}
			return x
		}
		return os.Expand(x, func(k string) string { return vars[k] })
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k, val := range x {
			out[k] = subst(val, vars)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, val := range x {
			out[i] = subst(val, vars)
		}
		return out
	}
	return v
}

// runSuite 运行测试套件
func runSuite(c *Client, suite *Suite, vars map[string]string, allowDanger bool, timeout time.Duration) int {
	failed, skipped, passed := 0, 0, 0
	fmt.Printf("套件: %s (%d 项)\n", suite.Name, len(suite.Requests))
	fmt.Println(strings.Repeat("-", 72))
	for _, req := range suite.Requests {
		label := req.ID
		if req.Name != "" {
			label = req.ID + " " + req.Name
		}
		if req.Danger && !allowDanger {
			fmt.Printf("%-46s SKIP (危险操作, 加 -allow-danger 执行)\n", label)
			skipped++
			continue
		}
		params, _ := subst(req.Params, vars).(map[string]interface{})
		pack, err := c.Call(req.Action, params, timeout)
		if err != nil {
			fmt.Printf("%-46s FAIL %v\n", label, err)
			failed++
			continue
		}
		if isOK(pack) {
			passed++
			fmt.Printf("%-46s PASS retcode=%v data=%s\n", label, pack["retcode"], shortJSON(pack["data"], 120))
		} else {
			failed++
			fmt.Printf("%-46s FAIL status=%v retcode=%v msg=%v data=%s\n",
				label, pack["status"], pack["retcode"], pack["message"], shortJSON(pack["data"], 120))
		}
	}
	fmt.Println(strings.Repeat("-", 72))
	fmt.Printf("结果: PASS=%d FAIL=%d SKIP=%d\n", passed, failed, skipped)
	if failed > 0 {
		return 1
	}
	return 0
}

// repl 交互模式：每行 "action {json}"
func repl(c *Client, timeout time.Duration) {
	fmt.Println("交互模式：输入 <action> [json参数]，例如 get_group_info {\"group_id\":187781763}")
	fmt.Println("输入 exit 退出。")
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			return
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			return
		}
		fields := strings.SplitN(line, " ", 2)
		action := fields[0]
		params := map[string]interface{}{}
		if len(fields) == 2 && strings.TrimSpace(fields[1]) != "" {
			if err := json.Unmarshal([]byte(fields[1]), &params); err != nil {
				fmt.Println("参数 JSON 解析失败:", err)
				continue
			}
		}
		pack, err := c.Call(action, params, timeout)
		if err != nil {
			fmt.Println("失败:", err)
			continue
		}
		raw, _ := json.MarshalIndent(pack, "", "  ")
		fmt.Println(string(raw))
	}
}

func main() {
	url := flag.String("url", "ws://127.0.0.1:15630", "Gensokyo 正向 WS 地址(含 path)")
	token := flag.String("token", "", "ws_server_token")
	action := flag.String("action", "", "单次调用的 action")
	paramsJSON := flag.String("params", "{}", "单次调用的 params JSON")
	script := flag.String("script", "", "测试套件 JSON 文件")
	groupID := flag.String("group", "", "套件变量 group_id")
	userID := flag.String("user", "", "套件变量 user_id")
	allowDanger := flag.Bool("allow-danger", false, "执行套件中的危险操作(禁言/踢人/黑名单等)")
	timeoutSec := flag.Int("timeout", 10, "单次请求超时(秒)")
	var sets kvFlag
	flag.Var(&sets, "set", "自定义套件变量 key=value，可重复，如 -set victim_id=123")
	flag.Parse()

	timeout := time.Duration(*timeoutSec) * time.Second
	c, err := dial(*url, *token, timeout)
	if err != nil {
		fmt.Printf("连接失败 %s: %v\n", *url, err)
		os.Exit(2)
	}
	defer c.Close()
	fmt.Printf("已连接 %s\n", *url)

	vars := map[string]string{}
	if *groupID != "" {
		vars["group_id"] = *groupID
		// 常见别名，套件里可写 ${group}
		vars["group"] = *groupID
	}
	if *userID != "" {
		vars["user_id"] = *userID
		vars["user"] = *userID
	}
	for _, kv := range sets {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}

	switch {
	case *action != "":
		params := map[string]interface{}{}
		if err := json.Unmarshal([]byte(*paramsJSON), &params); err != nil {
			fmt.Println("params JSON 解析失败:", err)
			os.Exit(2)
		}
		pack, err := c.Call(*action, subst(params, vars).(map[string]interface{}), timeout)
		if err != nil {
			fmt.Println("失败:", err)
			os.Exit(1)
		}
		raw, _ := json.MarshalIndent(pack, "", "  ")
		fmt.Println(string(raw))
		if !isOK(pack) {
			os.Exit(1)
		}
	case *script != "":
		raw, err := os.ReadFile(*script)
		if err != nil {
			fmt.Println("读取套件失败:", err)
			os.Exit(2)
		}
		var suite Suite
		if err := json.Unmarshal(raw, &suite); err != nil {
			fmt.Println("套件 JSON 解析失败:", err)
			os.Exit(2)
		}
		os.Exit(runSuite(c, &suite, vars, *allowDanger, timeout))
	default:
		repl(c, timeout)
	}
}
