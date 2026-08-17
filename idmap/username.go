package idmap

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"encoding/json"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

// 用户昵称缓存: 官方群事件里携带的 username (如入群申请的申请人昵称、群消息发送者昵称)
// 持久化到 idmap.db 的 usernames 桶, 缓存 7 天, 重启不丢失。
// 用途: get_stranger_info 反查 / 出站 at 转 @昵称 文本。
// 说明: QQ官方机器人API没有"按openid查用户资料"的接口, 昵称只能靠事件顺带缓存;
// 直接拉进群(无申请事件)的成员可能没有昵称, 会走 @Openid 保底。
const (
	usernameBucket = "usernames"
	usernameTTL    = 7 * 24 * time.Hour
)

type usernameEntry struct {
	Username string `json:"username"`
	StoredAt int64  `json:"stored_at"` // Unix 秒
}

// StoreUsernameV2 缓存 openid 对应的用户昵称(覆盖旧值并刷新存储时间)
func StoreUsernameV2(openid, username string) {
	if openid == "" || username == "" {
		return
	}
	entry, _ := json.Marshal(usernameEntry{Username: username, StoredAt: time.Now().Unix()})
	_ = db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(usernameBucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(openid), entry)
	})
}

// RetrieveUsernameByOpenID 读取缓存的用户昵称; 超过 7 天自动失效并删除, 无则返回空串
func RetrieveUsernameByOpenID(openid string) string {
	if openid == "" {
		return ""
	}
	var username string
	now := time.Now().Unix()
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(usernameBucket))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(openid))
		if v == nil {
			return nil
		}
		var e usernameEntry
		if err := json.Unmarshal(v, &e); err != nil {
			_ = b.Delete([]byte(openid))
			return nil
		}
		if now-e.StoredAt > int64(usernameTTL/time.Second) {
			_ = b.Delete([]byte(openid))
			return nil
		}
		username = e.Username
		return nil
	})
	return username
}

// 入群申请缓存: join_request_id -> {group_openid, member_openid}
// 原因: SparkBridge groupRequest 插件审批时只传 flag(join_request_id)/approve/reason,
// 不带 group_id/user_id, 而官方审批接口需要 group_openid + member_openid 两个路径参数,
// 因此事件到达时先缓存映射, 审批时按 flag 反查。
var joinRequestCache sync.Map

// StoreJoinRequestV2 缓存入群申请 id 对应的群/成员 openid
func StoreJoinRequestV2(joinRequestID, groupOpenID, memberOpenID string) {
	if joinRequestID == "" || groupOpenID == "" || memberOpenID == "" {
		return
	}
	joinRequestCache.Store(joinRequestID, [2]string{groupOpenID, memberOpenID})
}

// RetrieveJoinRequestV2 按入群申请 id 反查群/成员 openid, 无则返回空串
func RetrieveJoinRequestV2(joinRequestID string) (string, string) {
	if joinRequestID == "" {
		return "", ""
	}
	v, ok := joinRequestCache.Load(joinRequestID)
	if !ok {
		return "", ""
	}
	pair, _ := v.([2]string)
	return pair[0], pair[1]
}
