package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func externalContactHandler(res http.ResponseWriter, req *http.Request) {
	if !requirePost(res, req) {
		return
	}
	res.Header().Set("Content-Type", "application/json")
	req.Body = http.MaxBytesReader(res, req.Body, 1<<20)
	accessToken := GetAccessToken()
	if accessToken == "" {
		writeJSON(res, http.StatusBadGateway, `{"errcode":50001,"errmsg":"failed to get access token"}`)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeJSON(res, http.StatusBadRequest, `{"errcode":40001,"errmsg":"invalid request body"}`)
		return
	}
	var requestBody ExternalRequestBody
	if err = json.Unmarshal(body, &requestBody); err != nil {
		writeJSON(res, http.StatusBadRequest, `{"errcode":40002,"errmsg":"invalid json format"}`)
		return
	}
	if requestBody.Sendkey != Sendkey {
		writeJSON(res, http.StatusUnauthorized, `{"errcode":40001,"errmsg":"invalid sendkey"}`)
		return
	}
	if status, body := validateExternalRequestBody(&requestBody); status != 0 {
		writeJSON(res, status, body)
		return
	}
	baseData := map[string]interface{}{"external_userid": requestBody.ExternalUserIds, "sender": requestBody.Sender, "msgtype": requestBody.MsgType}
	var postData interface{}
	switch requestBody.MsgType {
	case "text":
		postData = map[string]interface{}{"external_userid": requestBody.ExternalUserIds, "sender": requestBody.Sender, "msgtype": "text", "text": map[string]interface{}{"content": requestBody.Text.Content}}
	case "image":
		postData = map[string]interface{}{"external_userid": requestBody.ExternalUserIds, "sender": requestBody.Sender, "msgtype": "image", "image": map[string]interface{}{"media_id": requestBody.Image.MediaId}}
	case "markdown":
		postData = map[string]interface{}{"external_userid": requestBody.ExternalUserIds, "sender": requestBody.Sender, "msgtype": "markdown", "markdown": map[string]interface{}{"content": requestBody.Markdown.Content}}
	case "link":
		postData = map[string]interface{}{"external_userid": requestBody.ExternalUserIds, "sender": requestBody.Sender, "msgtype": "link", "link": map[string]interface{}{"title": requestBody.Link.Title, "desc": requestBody.Link.Description, "url": requestBody.Link.Url, "thumb_media_id": requestBody.Link.ThumbMediaId}}
	case "miniprogram":
		postData = map[string]interface{}{"external_userid": requestBody.ExternalUserIds, "sender": requestBody.Sender, "msgtype": "miniprogram", "miniprogram": map[string]interface{}{"title": requestBody.MiniProgram.Title, "appid": requestBody.MiniProgram.AppId, "pagepath": requestBody.MiniProgram.PagePath, "thumb_media_id": requestBody.MiniProgram.ThumbMediaId}}
	default:
		postData = baseData
	}
	writeJSON(res, http.StatusOK, SendExternalMessage(accessToken, postData))
}

func wecomChan(res http.ResponseWriter, req *http.Request) {
	if !requirePost(res, req) {
		return
	}
	res.Header().Set("Content-Type", "application/json")
	req.Body = http.MaxBytesReader(res, req.Body, 1<<20)
	accessToken := GetAccessToken()
	if accessToken == "" {
		writeJSON(res, http.StatusBadGateway, `{"errcode":50001,"errmsg":"failed to get access token"}`)
		return
	}
	var sendkey, msgContent, msgType, toUser, agentId string
	var requestBody RequestBody
	body, err := io.ReadAll(req.Body)
	if err == nil && len(body) > 0 && json.Unmarshal(body, &requestBody) == nil {
		sendkey = requestBody.Sendkey
		msgType = requestBody.MsgType
		toUser = requestBody.ToUser
		agentId = requestBody.AgentId
		if requestBody.Msg != "" {
			msgContent = requestBody.Msg
		} else if requestBody.Text != nil && requestBody.Text.Content != "" {
			msgContent = requestBody.Text.Content
		} else if requestBody.Markdown != nil && requestBody.Markdown.Content != "" {
			msgContent = requestBody.Markdown.Content
		}
	}
	if sendkey == "" {
		sendkey = req.FormValue("sendkey")
		msgContent = req.FormValue("msg")
		msgType = req.FormValue("msg_type")
		if msgType == "" { msgType = req.FormValue("msgtype") }
		toUser = req.FormValue("touser")
		agentId = req.FormValue("agentid")
	}
	msgType = normalizeAppMsgType(msgType)
	msgContent = strings.TrimSpace(msgContent)
	toUser = strings.TrimSpace(toUser)
	agentId = strings.TrimSpace(agentId)
	if sendkey != Sendkey {
		writeJSON(res, http.StatusUnauthorized, `{"errcode":40001,"errmsg":"invalid sendkey"}`)
		return
	}
	if msgContent == "" {
		writeJSON(res, http.StatusBadRequest, `{"errcode":44004,"errmsg":"text content is empty"}`)
		return
	}
	if msgType != "text" && msgType != "markdown" && msgType != "image" {
		writeJSON(res, http.StatusBadRequest, `{"errcode":40009,"errmsg":"unsupported msgtype"}`)
		return
	}
	if toUser == "" { toUser = WecomToUid }
	if agentId == "" { agentId = WecomAid }
	mediaId := ""
	tokenValid := true
	if msgType == "image" {
		for i := 0; i <= 3; i++ {
			var errcode float64
			mediaId, errcode = UploadMedia(msgType, req, accessToken)
			tokenValid = ValidateToken(errcode)
			if tokenValid { break }
			accessToken = GetAccessToken()
		}
		if mediaId == "" {
			writeJSON(res, http.StatusBadRequest, `{"errcode":40006,"errmsg":"image upload failed or media missing"}`)
			return
		}
	}
	postData := JsonData{ToUser: toUser, AgentId: agentId, MsgType: msgType, DuplicateCheckInterval: 600}
	if msgType == "markdown" { postData.Markdown = Markdown{Content: msgContent} } else { postData.Text = Msg{Content: msgContent} }
	if msgType == "image" { postData.Image = Pic{MediaId: mediaId} }
	postStatus := ""
	for i := 0; i <= 3; i++ {
		postStatus = PostMsg(postData, fmt.Sprintf(SendMessageApi, accessToken))
		postResponse := ParseJson(postStatus)
		tokenValid = ValidateToken(postResponse["errcode"])
		if tokenValid { break }
		accessToken = GetAccessToken()
	}
	writeJSON(res, http.StatusOK, postStatus)
}

func healthz(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet { writeJSON(res, http.StatusMethodNotAllowed, `{"errcode":405,"errmsg":"method not allowed"}`); return }
	writeJSON(res, http.StatusOK, `{"status":"ok"}`)
}

func readyz(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet { writeJSON(res, http.StatusMethodNotAllowed, `{"errcode":405,"errmsg":"method not allowed"}`); return }
	if GetAccessToken() == "" { writeJSON(res, http.StatusServiceUnavailable, `{"status":"degraded","errmsg":"failed to get access token"}`); return }
	writeJSON(res, http.StatusOK, `{"status":"ready"}`)
}
