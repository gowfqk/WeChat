package main

import (
	"context"
	"net/http"
	"sync"
	"time"
)

var Sendkey = GetEnvDefault("SENDKEY", "set_a_sendkey")
var WecomCid = GetEnvDefault("WECOM_CID", "企业微信公司ID")
var WecomSecret = GetEnvDefault("WECOM_SECRET", "企业微信应用Secret")
var WecomAid = GetEnvDefault("WECOM_AID", "企业微信应用ID")
var WecomToUid = GetEnvDefault("WECOM_TOUID", "@all")

// Bark
var BarkDefaultUrl = GetEnvDefault("BARK_URL", "")

// 钉钉机器人
var DingDefaultToken = GetEnvDefault("DINGTALK_TOKEN", "")
var DingDefaultSecret = GetEnvDefault("DINGTALK_SECRET", "")

// 飞书机器人
var FeishuDefaultWebhook = GetEnvDefault("FEISHU_WEBHOOK", "")

// Telegram Bot
var TelegramBotToken = GetEnvDefault("TELEGRAM_BOT_TOKEN", "")
var TelegramDefaultChatId = GetEnvDefault("TELEGRAM_CHAT_ID", "")

// Server酱
var ServerChanKey = GetEnvDefault("SERVERCHAN_KEY", "")

// PushPlus
var PushPlusToken = GetEnvDefault("PUSHPLUS_TOKEN", "")

var CacheType = GetEnvDefault("CACHE_TYPE", "none")
var RedisStat = GetEnvDefault("REDIS_STAT", "OFF")
var RedisAddr = GetEnvDefault("REDIS_ADDR", "localhost:6379")
var RedisPassword = GetEnvDefault("REDIS_PASSWORD", "")

var ctx = context.Background()
var httpClient = &http.Client{Timeout: 10 * time.Second}
var serverReadTimeout = 15 * time.Second
var serverWriteTimeout = 15 * time.Second
var serverIdleTimeout = 60 * time.Second

type MemoryCache struct {
	token      string
	expireTime time.Time
}

var memoryCache MemoryCache
var cacheMutex sync.RWMutex

var GetTokenApi = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s"
var SendMessageApi = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s"
var UploadMediaApi = "https://qyapi.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s"
var MailComposeSendApi = "https://qyapi.weixin.qq.com/cgi-bin/exmail/app/compose_send?access_token=%s"

// 钉钉机器人 Webhook
var DingTalkWebhookApi = "https://oapi.dingtalk.com/robot/send?access_token=%s"

// 飞书机器人 Webhook
// 直接使用完整 webhook URL，无需模板拼接

// Telegram Bot API
var TelegramSendMessageApi = "https://api.telegram.org/bot%s/sendMessage"

// Server酱
var ServerChanApi = "https://sctapi.ftqq.com/%s.send"

// PushPlus
var PushPlusApi = "https://www.pushplus.plus/send"

const RedisTokenKey = "access_token"
