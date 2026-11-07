package msgmap

// Modified by DanielToyama on 2026-08-18 (Gensokyo-ForSpark fork)
//
// 消息ID → 发送者信息 持久化 (fakeReply 假回复用, GPLv3 修改记录见 readme)
//
// 背景: OneBot v11 的 reply 引用段(参考 https://api.luckylillia.com/api-226300082.md,
// send_group_msg 的 message 数组里 {"type":"reply","data":{"id":...}}), 在官方 QQ 协议
// 中无法发送真实引用。但引用能让人辨别"这条消息在回复谁", 因此本包把收到的每条消息
// 的 message_id(应用端可见)映射到发送者昵称/原文内容, 发送侧收到 reply 段时查表,
// 把回复文本伪造为 "回复 @昵称\n————\n原内容"。
//
// 所有方法在 db 未初始化时静默失败(记录日志), 与 botstats 同款容错。
//
// 2026-09-02 新增用途(delete_msg 撤自己的消息): 出站发送成功后也入表(RecordOutbound),
// 见 RecordOutbound 注释。

import (
	"encoding/json"
	"log"
	"time"

	"github.com/hoshinonyaruko/gensokyo/mylog"
	"go.etcd.io/bbolt"
)

var db *bbolt.DB

const bucketName = "msginfo"

// entryTTL 条目保留时长, 过期条目在启动时清理, 防止无限增长
const entryTTL = 7 * 24 * time.Hour

// MsgInfo 一条消息的发送者信息(以应用端可见的 message_id 为键)
type MsgInfo struct {
	UserID   string `json:"user_id"`            // 应用端可见的发送者ID
	Nickname string `json:"nickname"`           // 发送者QQ昵称(官方 username / 频道 member.nick)
	GroupID  string `json:"group_id"`           // 群/频道/私聊对象ID(应用端可见)
	MsgType  string `json:"msg_type"`           // group / group_private / guild / guild_private
	Content  string `json:"content"`            // 消息文本(经@转换后的可读文本)
	Time     int64  `json:"time"`               // 记录时间(Unix秒)
	Outbound bool   `json:"outbound,omitempty"` // true=机器人自己发出的消息(入站记录为 false)
}

// InitializeDB 打开数据库并清理过期条目
func InitializeDB() {
	var err error
	db, err = bbolt.Open("msgmap.db", 0600, nil)
	if err != nil {
		log.Fatalf("Failed to open msgmap database: %v", err)
	}

	db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})

	trimExpired()
}

// RecordMessage 记录一条消息的发送者信息, keys 为该消息在应用端可见的所有 message_id
func RecordMessage(keys []string, info MsgInfo) {
	if db == nil {
		mylog.Printf("msgmap RecordMessage db is nil")
		return
	}
	if len(keys) == 0 {
		return
	}
	info.Time = time.Now().Unix()
	data, err := json.Marshal(info)
	if err != nil {
		return
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		for _, k := range keys {
			if k == "" {
				continue
			}
			b.Put([]byte(k), data)
		}
		return nil
	})
}

// RecordOutbound 记录一条「机器人自己发出的」消息(出站标记), keys 为该消息在应用端可见的所有 message_id
//
// Modified by DanielToyama on 2026-09-02 (Gensokyo-ForSpark fork)
// 背景: 标准 onebot v11 的 delete_msg 只传 message_id, 而 RecordMessage 记录的都是「收到的」
// 消息, 机器人自己发出去的消息从未入表, 导致 delete_msg 反查不到会话对象 ->
// 四个撤回分支全部跳过、响应却依然 ok(应用端以为撤回成功)。
// 现在发送成功后把 message_id -> 会话对象(群/私聊/频道)也写进来, 只凭 message_id 即可撤自己的消息。
func RecordOutbound(keys []string, info MsgInfo) {
	info.Outbound = true
	RecordMessage(keys, info)
}

// LookupMessage 按应用端 message_id 查询发送者信息
func LookupMessage(key string) (MsgInfo, bool) {
	var info MsgInfo
	var found bool
	if db == nil || key == "" {
		return info, false
	}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return nil
		}
		raw := b.Get([]byte(key))
		if raw == nil {
			return nil
		}
		found = true
		return json.Unmarshal(raw, &info)
	})
	if err != nil {
		return MsgInfo{}, false
	}
	return info, found
}

// trimExpired 清理超过 entryTTL 的记录
func trimExpired() {
	if db == nil {
		return
	}
	cutoff := time.Now().Add(-entryTTL)
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var info MsgInfo
			if err := json.Unmarshal(v, &info); err != nil {
				continue
			}
			if info.Time > 0 && time.Unix(info.Time, 0).Before(cutoff) {
				if err := b.Delete(k); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// CloseDB 关闭数据库
func CloseDB() {
	if db != nil {
		db.Close()
	}
}
