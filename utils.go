package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

func ParseJson(jsonStr string) map[string]interface{} {
	var wecomResponse map[string]interface{}
	if jsonStr != "" {
		err := json.Unmarshal([]byte(jsonStr), &wecomResponse)
		if err != nil {
			log.Println("生成json字符串错误")
		}
	}
	return wecomResponse
}

func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 6 {
		return "***"
	}
	return s[:3] + strings.Repeat("*", len(s)-6) + s[len(s)-3:]
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func recoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				writeJSON(w, http.StatusInternalServerError, `{"errcode":50000,"errmsg":"internal server error"}`)
			}
		}()
		next(w, r)
	}
}

func requirePost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, `{"errcode":405,"errmsg":"method not allowed"}`)
		return false
	}
	return true
}

func getErrorCode(m map[string]interface{}) float64 {
	if m == nil {
		return 0
	}
	v, ok := m["errcode"]
	if !ok || v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func normalizeAppMsgType(msgType string) string {
	msgType = strings.TrimSpace(strings.ToLower(msgType))
	if msgType == "" {
		return "text"
	}
	return msgType
}

func validateExternalRequestBody(requestBody *ExternalRequestBody) (int, string) {
	if len(requestBody.ExternalUserIds) == 0 {
		return http.StatusBadRequest, `{"errcode":40003,"errmsg":"external_userid is required"}`
	}
	if len(requestBody.ExternalUserIds) > 1000 {
		return http.StatusBadRequest, `{"errcode":40005,"errmsg":"external_userid exceeds limit 1000"}`
	}
	seen := make(map[string]struct{}, len(requestBody.ExternalUserIds))
	clean := make([]string, 0, len(requestBody.ExternalUserIds))
	for _, id := range requestBody.ExternalUserIds {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		clean = append(clean, id)
	}
	requestBody.ExternalUserIds = clean
	if len(requestBody.ExternalUserIds) == 0 {
		return http.StatusBadRequest, `{"errcode":40003,"errmsg":"external_userid is required"}`
	}
	requestBody.Sender = strings.TrimSpace(requestBody.Sender)
	if requestBody.Sender == "" {
		return http.StatusBadRequest, `{"errcode":40004,"errmsg":"sender is required"}`
	}
	requestBody.MsgType = strings.TrimSpace(strings.ToLower(requestBody.MsgType))
	if requestBody.MsgType == "" {
		requestBody.MsgType = "text"
	}
	switch requestBody.MsgType {
	case "text":
		if requestBody.Text == nil || strings.TrimSpace(requestBody.Text.Content) == "" {
			return http.StatusBadRequest, `{"errcode":44004,"errmsg":"text content is empty"}`
		}
	case "image":
		if requestBody.Image == nil || strings.TrimSpace(requestBody.Image.MediaId) == "" {
			return http.StatusBadRequest, `{"errcode":40006,"errmsg":"image.media_id is required"}`
		}
	case "markdown":
		if requestBody.Markdown == nil || strings.TrimSpace(requestBody.Markdown.Content) == "" {
			return http.StatusBadRequest, `{"errcode":44004,"errmsg":"markdown content is empty"}`
		}
	case "link":
		if requestBody.Link == nil || strings.TrimSpace(requestBody.Link.Title) == "" || strings.TrimSpace(requestBody.Link.Url) == "" {
			return http.StatusBadRequest, `{"errcode":40007,"errmsg":"link.title and link.url are required"}`
		}
	case "miniprogram":
		if requestBody.MiniProgram == nil || strings.TrimSpace(requestBody.MiniProgram.Title) == "" || strings.TrimSpace(requestBody.MiniProgram.AppId) == "" || strings.TrimSpace(requestBody.MiniProgram.PagePath) == "" || strings.TrimSpace(requestBody.MiniProgram.ThumbMediaId) == "" {
			return http.StatusBadRequest, `{"errcode":40008,"errmsg":"miniprogram title/appid/pagepath/thumb_media_id are required"}`
		}
	default:
		return http.StatusBadRequest, `{"errcode":40009,"errmsg":"unsupported msgtype"}`
	}
	return 0, ""
}
