package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ==================== Bark ====================

func SendBark(barkUrl, title, content string) string {
	if barkUrl == "" {
		barkUrl = BarkDefaultUrl
	}
	if barkUrl == "" {
		return `{"errcode":40020,"errmsg":"bark_url is not configured"}`
	}
	// Bark 支持 GET: /key/title/body
	u := fmt.Sprintf("%s/%s/%s", trimSuffix(barkUrl, "/"),
		url.PathEscape(title), url.PathEscape(content))
	resp, err := httpClient.Get(u)
	if err != nil {
		log.Println("Bark推送失败==>", err)
		return `{"errcode":500,"errmsg":"bark request failed"}`
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

// ==================== 钉钉机器人 ====================

func SendDingTalk(token, secret, title, content string) string {
	if token == "" {
		token = DingDefaultToken
	}
	if token == "" {
		return `{"errcode":40021,"errmsg":"dingtalk token is not configured"}`
	}
	if secret == "" {
		secret = DingDefaultSecret
	}

	postUrl := fmt.Sprintf(DingTalkWebhookApi, token)
	if secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		sign := dingTalkSign(timestamp, secret)
		postUrl = fmt.Sprintf("%s&timestamp=%s&sign=%s", postUrl, timestamp, sign)
	}

	postData := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  fmt.Sprintf("### %s\n\n%s", title, content),
		},
	}
	return doHttpPost(postUrl, postData)
}

func dingTalkSign(timestamp, secret string) string {
	stringToSign := timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ==================== 飞书机器人 ====================

func SendFeishu(webhook, title, content string) string {
	if webhook == "" {
		webhook = FeishuDefaultWebhook
	}
	if webhook == "" {
		return `{"errcode":40022,"errmsg":"feishu webhook is not configured"}`
	}

	postData := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]string{
					"content": title,
					"tag":     "plain_text",
				},
			},
			"elements": []map[string]interface{}{
				{
					"tag": "markdown",
					"content": content,
				},
			},
		},
	}
	return doHttpPost(webhook, postData)
}

// ==================== Telegram Bot ====================

func SendTelegram(botToken, chatId, title, content string) string {
	if botToken == "" {
		botToken = TelegramBotToken
	}
	if botToken == "" {
		return `{"errcode":40023,"errmsg":"telegram bot token is not configured"}`
	}
	if chatId == "" {
		chatId = TelegramDefaultChatId
	}
	if chatId == "" {
		return `{"errcode":40024,"errmsg":"telegram chat_id is not configured"}`
	}

	postUrl := fmt.Sprintf(TelegramSendMessageApi, botToken)
	text := title + "\n\n" + content
	postData := map[string]string{
		"chat_id":    chatId,
		"text":       text,
		"parse_mode": "Markdown",
	}
	return doHttpPost(postUrl, postData)
}

// ==================== Server酱 ====================

func SendServerChan(key, title, content string) string {
	if key == "" {
		key = ServerChanKey
	}
	if key == "" {
		return `{"errcode":40025,"errmsg":"serverchan key is not configured"}`
	}

	postUrl := fmt.Sprintf(ServerChanApi, key)
	postData := map[string]string{
		"title": title,
		"desp":  content,
	}
	return doHttpPost(postUrl, postData)
}

// ==================== PushPlus ====================

func SendPushPlus(token, title, content string) string {
	if token == "" {
		token = PushPlusToken
	}
	if token == "" {
		return `{"errcode":40026,"errmsg":"pushplus token is not configured"}`
	}

	postData := map[string]string{
		"token":   token,
		"title":   title,
		"content": content,
	}
	return doHttpPost(PushPlusApi, postData)
}

// ==================== 通用工具 ====================

func doHttpPost(postUrl string, data interface{}) string {
	postJson, _ := json.Marshal(data)
	msgReq, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(postJson))
	if err != nil {
		log.Println("创建请求失败==>", err)
		return `{"errcode":500,"errmsg":"create request failed"}`
	}
	msgReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(msgReq)
	if err != nil {
		log.Println("请求失败==>", err)
		return `{"errcode":500,"errmsg":"upstream request failed"}`
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func trimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}

// SendPush 统一推送入口，根据 channel 分发
func SendPush(req *PushRequestBody) string {
	switch req.Channel {
	case "bark":
		return SendBark(req.BarkUrl, req.Title, req.Content)
	case "dingtalk":
		return SendDingTalk("", req.DingSecret, req.Title, req.Content)
	case "feishu":
		return SendFeishu("", req.Title, req.Content)
	case "telegram":
		return SendTelegram("", req.ChatId, req.Title, req.Content)
	case "serverchan":
		return SendServerChan("", req.Title, req.Content)
	case "pushplus":
		return SendPushPlus("", req.Title, req.Content)
	case "wecom", "":
		return sendWecomFromPush(req)
	default:
		return fmt.Sprintf(`{"errcode":40030,"errmsg":"unsupported channel: %s"}`, req.Channel)
	}
}

func sendWecomFromPush(req *PushRequestBody) string {
	accessToken := GetAccessToken()
	if accessToken == "" {
		return `{"errcode":50001,"errmsg":"failed to get access token"}`
	}
	toUser := req.ToUser
	if toUser == "" {
		toUser = WecomToUid
	}
	agentId := req.AgentId
	if agentId == "" {
		agentId = WecomAid
	}
	msgType := req.MsgType
	if msgType == "" {
		msgType = "text"
	}

	postData := JsonData{
		ToUser:                 toUser,
		AgentId:                agentId,
		MsgType:                msgType,
		DuplicateCheckInterval: 600,
	}
	if msgType == "markdown" {
		postData.Markdown = Markdown{Content: req.Content}
	} else {
		postData.Text = Msg{Content: req.Title + "\n" + req.Content}
	}

	var postStatus string
	for i := 0; i <= 3; i++ {
		postStatus = PostMsg(postData, fmt.Sprintf(SendMessageApi, accessToken))
		postResponse := ParseJson(postStatus)
		if ValidateToken(postResponse["errcode"]) {
			break
		}
		accessToken = GetAccessToken()
	}
	return postStatus
}
