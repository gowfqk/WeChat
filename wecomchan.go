package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

/*-------------------------------  环境变量配置 begin  -------------------------------*/

var Sendkey = GetEnvDefault("SENDKEY", "set_a_sendkey")
var WecomCid = GetEnvDefault("WECOM_CID", "企业微信公司ID")
var WecomSecret = GetEnvDefault("WECOM_SECRET", "企业微信应用Secret")
var WecomAid = GetEnvDefault("WECOM_AID", "企业微信应用ID")
var WecomToUid = GetEnvDefault("WECOM_TOUID", "@all")
var CacheType = GetEnvDefault("CACHE_TYPE", "none") // 可选值: none, memory, redis
var RedisStat = GetEnvDefault("REDIS_STAT", "OFF")
var RedisAddr = GetEnvDefault("REDIS_ADDR", "localhost:6379")
var RedisPassword = GetEnvDefault("REDIS_PASSWORD", "")
var ctx = context.Background()
var httpClient = &http.Client{Timeout: 10 * time.Second}

/*-------------------------------  环境变量配置 end  -------------------------------*/

/*-------------------------------  内存缓存配置 begin  -------------------------------*/

type MemoryCache struct {
	token      string
	expireTime time.Time
}

var memoryCache MemoryCache
var cacheMutex sync.RWMutex

/*-------------------------------  内存缓存配置 end  -------------------------------*/

/*-------------------------------  企业微信服务端API begin  -------------------------------*/

var GetTokenApi = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s"
var SendMessageApi = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s"
var ExternalSendMessageApi = "https://qyapi.weixin.qq.com/cgi-bin/externalcontact/message/send?access_token=%s"
var UploadMediaApi = "https://qyapi.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s"

/*-------------------------------  企业微信服务端API end  -------------------------------*/

const RedisTokenKey = "access_token"

// RequestBody 请求体结构体（支持JSON传参）
// 支持两种格式：
// 1. 简化格式（向后兼容）：{"msg_type": "text", "msg": "内容"}
// 2. 官方格式（新增）：{"msg_type": "text", "text": {"content": "内容"}}
type RequestBody struct {
	Sendkey  string          `json:"sendkey"`
	Msg      string          `json:"msg"`               // 简化格式：文本/Markdown内容
	MsgType  string          `json:"msg_type"`
	ToUser   string          `json:"touser,omitempty"`  // 可选：覆盖默认的接收人
	AgentId  string          `json:"agentid,omitempty"` // 可选：覆盖默认的应用ID
	Text     *Msg            `json:"text,omitempty"`    // 官方格式：文本消息
	Markdown *Markdown       `json:"markdown,omitempty"` // 官方格式：Markdown消息
}

type Msg struct {
	Content string `json:"content"`
}
type Pic struct {
	MediaId string `json:"media_id"`
}
type Markdown struct {
	Content string `json:"content"`
}
type JsonData struct {
	ToUser                 string `json:"touser"`
	AgentId                string `json:"agentid"`
	MsgType                string `json:"msgtype"`
	DuplicateCheckInterval int    `json:"duplicate_check_interval"`
	Text                   Msg      `json:"text"`
	Image                  Pic      `json:"image"`
	Markdown               Markdown `json:"markdown"`
}

/*-------------------------------  外部联系人消息结构体 begin  -------------------------------*/

// ExternalRequestBody 外部联系人消息请求体
type ExternalRequestBody struct {
	Sendkey         string      `json:"sendkey"`
	ExternalUserIds []string    `json:"external_userid"` // 外部联系人userid列表
	Sender          string      `json:"sender"`          // 发送企业成员的userid
	MsgType         string      `json:"msgtype"`         // 消息类型
	Text            *Msg        `json:"text,omitempty"`   // 文本消息
	Image           *Pic        `json:"image,omitempty"`  // 图片消息
	Markdown        *Markdown   `json:"markdown,omitempty"` // Markdown消息
	Link            *LinkMsg    `json:"link,omitempty"`   // 链接消息
	MiniProgram     *MiniProgramMsg `json:"miniprogram,omitempty"` // 小程序消息
}

// LinkMsg 链接消息
type LinkMsg struct {
	Title       string `json:"title"`
	Description string `json:"desc"`
	Url         string `json:"url"`
	ThumbMediaId string `json:"thumb_media_id"`
}

// MiniProgramMsg 小程序消息
type MiniProgramMsg struct {
	Title        string `json:"title"`
	AppId        string `json:"appid"`
	PagePath     string `json:"pagepath"`
	ThumbMediaId string `json:"thumb_media_id"`
}

// ExternalMessageData 外部联系人消息数据
type ExternalMessageData struct {
	ExternalUserIds []string          `json:"external_userid"`
	Sender          string            `json:"sender"`
	MsgType         string            `json:"msgtype"`
	Text            Msg               `json:"text,omitempty"`
	Image           Pic               `json:"image,omitempty"`
	Markdown        Markdown          `json:"markdown,omitempty"`
	Link            LinkMsg           `json:"link,omitempty"`
	MiniProgram     MiniProgramMsg    `json:"miniprogram,omitempty"`
}

/*-------------------------------  外部联系人消息结构体 end  -------------------------------*/

// GetEnvDefault 获取配置信息，未获取到则取默认值
func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

// ParseJson 将json字符串解析为map
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

// GetRemoteToken 从企业微信服务端API获取access_token，根据配置缓存到redis或内存
func GetRemoteToken(corpId, appSecret string) string {
	getTokenUrl := fmt.Sprintf(GetTokenApi, corpId, appSecret)
	log.Printf("getTokenUrl ==> %s", strings.Replace(getTokenUrl, appSecret, "***", 1))
	resp, err := httpClient.Get(getTokenUrl)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return ""
	}
	tokenResponse := ParseJson(string(respData))
	log.Println("企业微信获取access_token接口返回==>", tokenResponse)
	accessToken, ok := tokenResponse[RedisTokenKey].(string)
	if !ok || accessToken == "" {
		log.Println("企业微信获取access_token失败: missing access_token")
		return ""
	}

	// 根据缓存类型选择存储方式
	if CacheType == "redis" && RedisStat == "ON" {
		log.Println("prepare to set redis key")
		rdb := RedisClient()
		set, err := rdb.SetNX(ctx, RedisTokenKey, accessToken, 7000*time.Second).Result()
		log.Println(set)
		if err != nil {
			log.Println(err)
		}
	} else if CacheType == "memory" {
		log.Println("prepare to set memory cache")
		cacheMutex.Lock()
		memoryCache = MemoryCache{
			token:      accessToken,
			expireTime: time.Now().Add(7000 * time.Second),
		}
		cacheMutex.Unlock()
		log.Println("memory cache set successfully")
	}
	return accessToken
}

// RedisClient redis客户端
func RedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: RedisPassword, // no password set
		DB:       0,             // use default DB
	})
	return rdb
}

// PostMsg 推送消息
func PostMsg(postData JsonData, postUrl string) string {
	postJson, _ := json.Marshal(postData)
	log.Println("postUrl ", postUrl)
	msgReq, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(postJson))
	if err != nil {
		log.Println(err)
		return `{"errcode":500,"errmsg":"create request failed"}`
	}
	msgReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(msgReq)
	if err != nil {
		log.Println("企业微信发送应用消息接口报错==>", err)
		return `{"errcode":500,"errmsg":"upstream request failed"}`
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取企业微信响应失败==>", err)
		return `{"errcode":500,"errmsg":"read upstream response failed"}`
	}
	mediaResp := ParseJson(string(body))
	log.Println("企业微信发送应用消息接口返回==>", mediaResp)
	return string(body)
}

// UploadMedia  上传临时素材并返回mediaId
func UploadMedia(msgType string, req *http.Request, accessToken string) (string, float64) {
	_ = req.ParseMultipartForm(2 << 20)
	imgFile, imgHeader, err := req.FormFile("media")
	if err != nil {
		log.Println("图片文件出错==>", err)
		return "", 400
	}
	defer imgFile.Close()
	log.Printf("文件大小==>%d字节", imgHeader.Size)
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	createFormFile, err := writer.CreateFormFile("media", imgHeader.Filename)
	if err != nil {
		log.Println("创建 multipart 文件失败==>", err)
		return "", 500
	}
	readAll, err := io.ReadAll(imgFile)
	if err != nil {
		log.Println("读取图片文件失败==>", err)
		return "", 500
	}
	_, _ = createFormFile.Write(readAll)
	_ = writer.Close()

	uploadMediaUrl := fmt.Sprintf(UploadMediaApi, accessToken, msgType)
	log.Println("uploadMediaUrl==>", uploadMediaUrl)
	newRequest, err := http.NewRequest("POST", uploadMediaUrl, buf)
	if err != nil {
		log.Println("创建上传请求失败==>", err)
		return "", 500
	}
	newRequest.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := httpClient.Do(newRequest)
	if err != nil {
		log.Println("上传临时素材出错==>", err)
		return "", 500
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取上传响应失败==>", err)
		return "", 500
	}
	mediaResp := ParseJson(string(respData))
	log.Println("企业微信上传临时素材接口返回==>", mediaResp)
	if mediaID, ok := mediaResp["media_id"].(string); ok && mediaID != "" {
		return mediaID, 0
	}
	return "", getErrorCode(mediaResp)
}

// ValidateToken 判断accessToken是否失效
// true-未失效, false-失效需重新获取
func ValidateToken(errcode interface{}) bool {
	codeTyp := reflect.TypeOf(errcode)
	log.Println("errcode的数据类型==>", codeTyp)
	if !codeTyp.Comparable() {
		log.Printf("type is not comparable: %v", codeTyp)
		return true
	}

	// 如果errcode为42001表明token已失效，则清空缓存中的token
	// 已知codeType为float64
	if math.Abs(errcode.(float64)-float64(42001)) < 1e-3 {
		if CacheType == "redis" && RedisStat == "ON" {
			log.Printf("token已失效，开始删除redis中的key==>%s", RedisTokenKey)
			rdb := RedisClient()
			rdb.Del(ctx, RedisTokenKey)
			log.Printf("删除redis中的key==>%s完毕", RedisTokenKey)
		} else if CacheType == "memory" {
			log.Printf("token已失效，开始清除内存缓存")
			cacheMutex.Lock()
			memoryCache = MemoryCache{
				token:      "",
				expireTime: time.Time{},
			}
			cacheMutex.Unlock()
			log.Printf("清除内存缓存完毕")
		}
		log.Println("现需重新获取token")
		return false
	}
	return true
}

// GetAccessToken 获取企业微信的access_token
func GetAccessToken() string {
	accessToken := ""

	if CacheType == "redis" && RedisStat == "ON" {
		log.Println("尝试从redis获取token")
		rdb := RedisClient()
		value, err := rdb.Get(ctx, RedisTokenKey).Result()
		if err == redis.Nil {
			log.Println("access_token does not exist, need get it from remote API")
		} else if err != nil {
			log.Println("从redis获取token失败==>", err)
		}
		accessToken = value
	} else if CacheType == "memory" {
		log.Println("尝试从内存缓存获取token")
		cacheMutex.RLock()
		if !memoryCache.expireTime.IsZero() && time.Now().Before(memoryCache.expireTime) {
			accessToken = memoryCache.token
			log.Println("get access_token from memory cache")
		} else {
			log.Println("memory cache expired or empty")
		}
		cacheMutex.RUnlock()
	}

	if accessToken == "" {
		log.Println("get access_token from remote API")
		accessToken = GetRemoteToken(WecomCid, WecomSecret)
	}
	return accessToken
}

// InitJsonData 初始化Json公共部分数据
func InitJsonData(msgType string) JsonData {
	return JsonData{
		ToUser:                 WecomToUid,
		AgentId:                WecomAid,
		MsgType:                msgType,
		DuplicateCheckInterval: 600,
	}
}

// SendExternalMessage 发送外部联系人消息
func SendExternalMessage(accessToken string, postData interface{}) string {
	postJson, _ := json.Marshal(postData)
	log.Println("发送外部联系人消息 postJson prepared")

	sendMessageUrl := fmt.Sprintf(ExternalSendMessageApi, accessToken)
	log.Println("发送外部联系人消息 URL ", sendMessageUrl)

	msgReq, err := http.NewRequest("POST", sendMessageUrl, bytes.NewBuffer(postJson))
	if err != nil {
		log.Println("创建请求失败:", err)
		return `{"errcode":500,"errmsg":"create request failed"}`
	}
	msgReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(msgReq)
	if err != nil {
		log.Println("发送外部联系人消息失败==>", err)
		return `{"errcode":500,"errmsg":"upstream request failed"}`
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取外部联系人响应失败==>", err)
		return `{"errcode":500,"errmsg":"read upstream response failed"}`
	}
	respData := ParseJson(string(body))
	log.Println("发送外部联系人消息接口返回==>", respData)

	return string(body)
}

// externalContactHandler 外部联系人消息处理器
func externalContactHandler(res http.ResponseWriter, req *http.Request) {
	if !requirePost(res, req) {
		return
	}
	log.Println("========== 收到外部联系人消息请求 ==========")
	log.Printf("请求方法: %s\n", req.Method)
	log.Printf("请求URL: %s\n", req.URL.String())
	log.Printf("Content-Type: %s\n", req.Header.Get("Content-Type"))

	res.Header().Set("Content-Type", "application/json")
	req.Body = http.MaxBytesReader(res, req.Body, 1<<20)

	accessToken := GetAccessToken()
	if accessToken == "" {
		writeJSON(res, http.StatusBadGateway, `{"errcode":50001,"errmsg":"failed to get access token"}`)
		return
	}

	_ = req.ParseForm()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("读取请求体失败:", err)
		writeJSON(res, http.StatusBadRequest, `{"errcode":40001,"errmsg":"invalid request body"}`)
		return
	}

	log.Printf("请求体长度: %d\n", len(body))

	var requestBody ExternalRequestBody
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		log.Printf("JSON解析失败: %v\n", err)
		writeJSON(res, http.StatusBadRequest, `{"errcode":40002,"errmsg":"invalid json format"}`)
		return
	}

	log.Printf("解析结果 - sendkey: '%s', sender: '%s', msgType: '%s'\n", maskSecret(requestBody.Sendkey), requestBody.Sender, requestBody.MsgType)
	log.Printf("外部联系人数量: %d\n", len(requestBody.ExternalUserIds))

	if requestBody.Sendkey != Sendkey {
		log.Printf("sendkey验证失败 - 期望: '%s', 实际: '%s'\n", maskSecret(Sendkey), maskSecret(requestBody.Sendkey))
		writeJSON(res, http.StatusUnauthorized, `{"errcode":40001,"errmsg":"invalid sendkey"}`)
		return
	}

	if len(requestBody.ExternalUserIds) == 0 {
		log.Println("错误：external_userid为空")
		writeJSON(res, http.StatusBadRequest, `{"errcode":40003,"errmsg":"external_userid is required"}`)
		return
	}
	if len(requestBody.ExternalUserIds) > 1000 {
		writeJSON(res, http.StatusBadRequest, `{"errcode":40005,"errmsg":"external_userid exceeds limit 1000"}`)
		return
	}
	if requestBody.Sender == "" {
		log.Println("错误：sender为空")
		writeJSON(res, http.StatusBadRequest, `{"errcode":40004,"errmsg":"sender is required"}`)
		return
	}
	if requestBody.MsgType == "" {
		requestBody.MsgType = "text"
	}

	var postData interface{}
	baseData := map[string]interface{}{
		"external_userid": requestBody.ExternalUserIds,
		"sender":          requestBody.Sender,
		"msgtype":         requestBody.MsgType,
	}

	switch requestBody.MsgType {
	case "text":
		if requestBody.Text != nil {
			postData = map[string]interface{}{
				"external_userid": requestBody.ExternalUserIds,
				"sender":          requestBody.Sender,
				"msgtype":         "text",
				"text": map[string]interface{}{
					"content": requestBody.Text.Content,
				},
			}
			log.Printf("文本消息内容长度: %d\n", len(requestBody.Text.Content))
		} else {
			// 兜底：发送空文本消息
			postData = baseData
		}
	case "image":
		if requestBody.Image != nil {
			postData = map[string]interface{}{
				"external_userid": requestBody.ExternalUserIds,
				"sender":          requestBody.Sender,
				"msgtype":         "image",
				"image": map[string]interface{}{
					"media_id": requestBody.Image.MediaId,
				},
			}
			log.Printf("图片消息media_id: %s\n", requestBody.Image.MediaId)
		} else {
			postData = baseData
		}
	case "markdown":
		if requestBody.Markdown != nil {
			postData = map[string]interface{}{
				"external_userid": requestBody.ExternalUserIds,
				"sender":          requestBody.Sender,
				"msgtype":         "markdown",
				"markdown": map[string]interface{}{
					"content": requestBody.Markdown.Content,
				},
			}
			log.Printf("Markdown消息内容长度: %d\n", len(requestBody.Markdown.Content))
		} else {
			postData = baseData
		}
	case "link":
		if requestBody.Link != nil {
			postData = map[string]interface{}{
				"external_userid": requestBody.ExternalUserIds,
				"sender":          requestBody.Sender,
				"msgtype":         "link",
				"link": map[string]interface{}{
					"title":          requestBody.Link.Title,
					"desc":           requestBody.Link.Description,
					"url":            requestBody.Link.Url,
					"thumb_media_id": requestBody.Link.ThumbMediaId,
				},
			}
			log.Printf("链接消息标题: %s\n", requestBody.Link.Title)
		} else {
			postData = baseData
		}
	case "miniprogram":
		if requestBody.MiniProgram != nil {
			postData = map[string]interface{}{
				"external_userid": requestBody.ExternalUserIds,
				"sender":          requestBody.Sender,
				"msgtype":         "miniprogram",
				"miniprogram": map[string]interface{}{
					"title":           requestBody.MiniProgram.Title,
					"appid":           requestBody.MiniProgram.AppId,
					"pagepath":        requestBody.MiniProgram.PagePath,
					"thumb_media_id":  requestBody.MiniProgram.ThumbMediaId,
				},
			}
			log.Printf("小程序消息标题: %s\n", requestBody.MiniProgram.Title)
		} else {
			postData = baseData
		}
	default:
		log.Printf("未知消息类型: %s，使用text类型\n", requestBody.MsgType)
		postData = baseData
	}

	log.Printf("准备发送的外部联系人数据: %+v\n", postData)

	// 发送消息
	response := SendExternalMessage(accessToken, postData)

	writeJSON(res, http.StatusOK, response)
	log.Println("========== 外部联系人消息请求处理完成 ==========")
}

// 主函数入口
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	wecomChan := func(res http.ResponseWriter, req *http.Request) {
		if !requirePost(res, req) {
			return
		}
		log.Println("========== 收到新请求 ==========")
		log.Printf("请求方法: %s\n", req.Method)
		log.Printf("请求URL: %s\n", req.URL.String())
		log.Printf("Content-Type: %s\n", req.Header.Get("Content-Type"))
		res.Header().Set("Content-Type", "application/json")
		req.Body = http.MaxBytesReader(res, req.Body, 1<<20)

		accessToken := GetAccessToken()
		if accessToken == "" {
			writeJSON(res, http.StatusBadGateway, `{"errcode":50001,"errmsg":"failed to get access token"}`)
			return
		}
		tokenValid := true

		_ = req.ParseForm()
		var sendkey, msgContent, msgType, toUser, agentId string

		var requestBody RequestBody
		body, err := io.ReadAll(req.Body)
		if err == nil && len(body) > 0 {
			log.Printf("请求体长度: %d\n", len(body))
			err = json.Unmarshal(body, &requestBody)
			if err == nil {
				sendkey = requestBody.Sendkey
				msgType = requestBody.MsgType
				toUser = requestBody.ToUser
				agentId = requestBody.AgentId

				log.Printf("解析结果 - sendkey: '%s', msgType: '%s'\n", maskSecret(sendkey), msgType)
				if requestBody.Msg != "" {
					msgContent = requestBody.Msg
					log.Println("使用body传参（简化格式）")
				} else {
					if requestBody.Text != nil && requestBody.Text.Content != "" {
						msgContent = requestBody.Text.Content
						log.Println("使用body传参（官方格式 - text）")
					} else if requestBody.Markdown != nil && requestBody.Markdown.Content != "" {
						msgContent = requestBody.Markdown.Content
						log.Println("使用body传参（官方格式 - markdown）")
					} else {
						log.Println("警告：未找到消息内容")
					}
				}
			} else {
				log.Printf("JSON解析失败: %v，回退到URL参数\n", err)
			}
		} else {
			log.Println("请求体为空或读取失败")
		}

		if sendkey == "" {
			sendkey = req.FormValue("sendkey")
			msgContent = req.FormValue("msg")
			msgType = req.FormValue("msg_type")
			if msgType == "" {
				msgType = req.FormValue("msgtype")
			}
			toUser = req.FormValue("touser")
			agentId = req.FormValue("agentid")
			log.Println("使用URL参数传参")
		}

		log.Printf("最终参数 - sendkey: '%s', msgType: '%s', msgContent长度: %d\n", maskSecret(sendkey), msgType, len(msgContent))

		if sendkey != Sendkey {
			log.Printf("sendkey验证失败 - 期望: '%s', 实际: '%s'\n", maskSecret(Sendkey), maskSecret(sendkey))
			writeJSON(res, http.StatusUnauthorized, `{"errcode":40001,"errmsg":"invalid sendkey"}`)
			return
		}

		if msgContent == "" {
			log.Println("错误：msgContent为空")
			writeJSON(res, http.StatusBadRequest, `{"errcode":44004,"errmsg":"text content is empty"}`)
			return
		}

		if msgType == "" {
			msgType = "text"
		}
		if toUser == "" {
			toUser = WecomToUid
		}
		if agentId == "" {
			agentId = WecomAid
		}

		mediaId := ""
		if msgType == "image" {
			for i := 0; i <= 3; i++ {
				var errcode float64
				mediaId, errcode = UploadMedia(msgType, req, accessToken)
				log.Printf("企业微信上传临时素材接口返回的media_id==>[%s], errcode==>[%f]\n", mediaId, errcode)
				tokenValid = ValidateToken(errcode)
				if tokenValid {
					break
				}
				accessToken = GetAccessToken()
			}
		}

		postData := JsonData{
			ToUser:                 toUser,
			AgentId:                agentId,
			MsgType:                msgType,
			DuplicateCheckInterval: 600,
		}
		if msgType == "markdown" {
			postData.Markdown = Markdown{Content: msgContent}
		} else {
			postData.Text = Msg{Content: msgContent}
		}
		if msgType == "image" {
			postData.Image = Pic{MediaId: mediaId}
		}

		log.Printf("准备发送的数据: %+v\n", postData)
		log.Printf("消息内容长度: %d\n", len(msgContent))

		postStatus := ""
		for i := 0; i <= 3; i++ {
			sendMessageUrl := fmt.Sprintf(SendMessageApi, accessToken)
			postStatus = PostMsg(postData, sendMessageUrl)
			postResponse := ParseJson(postStatus)
			errcode := postResponse["errcode"]
			log.Println("发送应用消息接口返回errcode==>", errcode)
			tokenValid = ValidateToken(errcode)
			if tokenValid {
				break
			}
			accessToken = GetAccessToken()
		}

		writeJSON(res, http.StatusOK, postStatus)
		log.Println("========== 请求处理完成 ==========")
	}
	healthz := func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			writeJSON(res, http.StatusMethodNotAllowed, `{"errcode":405,"errmsg":"method not allowed"}`)
			return
		}
		writeJSON(res, http.StatusOK, `{"status":"ok"}`)
	}
	http.HandleFunc("/wecomchan", wecomChan)
	http.HandleFunc("/external", externalContactHandler)
	http.HandleFunc("/healthz", healthz)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
