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
// 本路由把应用端视角的"虚构QQ号"(idmap 数字行ID)反查为官方用户 openid, 然后:
//   - 默认(代理模式): 由 Gensokyo 后台下载官方头像, 直接返回图片字节流,
//     客户端无需理解任何重定向 —— 覆盖 axios/浏览器/官方服务器/不跟随302的客户端等所有场景;
//   - 显式加 &redirect=1 时: 302 重定向到官方头像地址(给想省流量的客户端保留)。
//
// 尺寸 s 原样透传(100/140/640...), 缺省 640。

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
)

var reDigitsOnly = regexp.MustCompile(`^\d+$`)

// AvatarProxyHandler 处理 /g?b=qq&nk=<数字ID>&s=<尺寸>(默认代理下载头像, redirect=1 走302)
func AvatarProxyHandler(c *gin.Context) {
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
		mylog.Printf("avatarProxy: 无法反查用户[%v]的openid: %v", nk, err)
		c.String(http.StatusNotFound, fmt.Sprintf("user %s not found", nk))
		return
	}

	appid := config.GetAppIDStr()
	target := fmt.Sprintf("https://q.qlogo.cn/qqapp/%s/%s/%s", appid, openid, s)

	// 显式 redirect=1: 保留 302 模式
	if c.Query("redirect") == "1" {
		mylog.Printf("avatarProxy: nk[%v] -> 302 %s", nk, target)
		c.Redirect(http.StatusFound, target)
		return
	}

	// 默认代理模式: 后台下载官方头像, 直接回图片字节流
	client := &http.Client{Timeout: 10 * time.Second} // Go 默认跟随重定向
	resp, err := client.Get(target)
	if err != nil {
		mylog.Printf("avatarProxy: 下载头像失败 %s: %v", target, err)
		c.String(http.StatusBadGateway, "failed to fetch avatar")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		mylog.Printf("avatarProxy: 头像源返回 %d (%s)", resp.StatusCode, target)
		c.Data(resp.StatusCode, "text/plain; charset=utf-8", []byte("avatar source error"))
		return
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	c.Header("Content-Type", ct)
	c.Header("Cache-Control", "public, max-age=86400") // 头像源本身可缓存, 客户端缓存一天
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		mylog.Printf("avatarProxy: 透传头像失败: %v", err)
		return
	}
	mylog.Printf("avatarProxy: nk[%v] -> openid[%v], 代理返回 %s", nk, openid, target)
}