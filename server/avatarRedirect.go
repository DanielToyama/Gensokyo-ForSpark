package server

// Modified by DanielToyama on 2026-08-20 (Gensokyo-ForSpark fork)
//
// 野鸡 qlogo 兼容路由:  GET /g?b=qq&nk=<数字ID>&s=<尺寸>
//
// 背景: 部分客户端/前端习惯用  https://q1.qlogo.cn/g?b=qq&nk=<QQ号码>&s=640 取 QQ 头像,
// 但官方 bot 没有 QQ 号, 出站消息的 avatar 是官方格式
// https://q.qlogo.cn/qqapp/<appid>/<用户openid>/640
// (见 Processor.GenerateAvatarURLV2)。
//
// 本路由把应用端视角的"虚构QQ号"(idmap 数字行ID)反查为官方用户 openid,
// 302 重定向到官方头像地址, 尺寸 s 原样透传(100/140/640...)。

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
)

var reDigitsOnly = regexp.MustCompile(`^\d+$`)

// AvatarRedirectHandler 处理 /g?b=qq&nk=<数字ID>&s=<尺寸>
func AvatarRedirectHandler(c *gin.Context) {
	if c.Query("b") != "qq" {
		c.String(http.StatusBadRequest, "invalid param b")
		return
	}
	nk := c.Query("nk")
	if nk == "" || !reDigitsOnly.MatchString(nk) {
		c.String(http.StatusBadRequest, "invalid param nk")
		return
	}
	s := c.Query("s")
	if s == "" {
		s = "640"
	} else if !reDigitsOnly.MatchString(s) {
		c.String(http.StatusBadRequest, "invalid param s")
		return
	}

	// 数字行ID -> 官方用户 openid
	openid, err := idmap.RetrieveRowByIDv2(nk)
	if err != nil || openid == "" {
		mylog.Printf("avatarRedirect: 无法反查用户[%v]的openid: %v", nk, err)
		c.String(http.StatusNotFound, fmt.Sprintf("user %s not found", nk))
		return
	}

	appid := config.GetAppIDStr()
	url := fmt.Sprintf("https://q.qlogo.cn/qqapp/%s/%s/%s", appid, openid, s)
	mylog.Printf("avatarRedirect: nk[%v] -> openid[%v], redirect: %s", nk, openid, url)
	c.Redirect(http.StatusFound, url)
}