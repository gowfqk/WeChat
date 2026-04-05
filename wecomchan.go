package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
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
	if string(jsonStr) != "" {
		err := json.Unmarshal([]byte(string(jsonStr)), &wecomResponse)
		if err != nil {
			log.Println("生成json字符串错误")
		}
	}
	return wecomResponse
}

// GetRemoteToken 从企业微信服务端API获取access_token，根据配置缓存到redis或内存
func GetRemoteToken(corpId, appSecret string) string {
	getTokenUrl := fmt.Sprintf(GetTokenApi, corpId, appSecret)
	log.Println("getTokenUrl==>", getTokenUrl)
	resp, err := http.Get(getTokenUrl)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	tokenResponse := ParseJson(string(respData))
	log.Println("企业微信获取access_token接口返回==>", tokenResponse)
	accessToken := tokenResponse[RedisTokenKey].(string)

	// 根据缓存类型选择存储方式
	if CacheType == "redis" && RedisStat == "ON" {
		log.Println("prepare to set redis key")
		rdb := RedisClient()
		// access_token有效时间为7200秒(2小时)
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
	log.Println("postJson ", string(postJson))
	log.Println("postUrl ", postUrl)
	msgReq, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(postJson))
	if err != nil {
		log.Println(err)
	}
	msgReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(msgReq)
	if err != nil {
		log.Fatalln("企业微信发送应用消息接口报错==>", err)
	}
	defer msgReq.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	mediaResp := ParseJson(string(body))
	log.Println("企业微信发送应用消息接口返回==>", mediaResp)
	return string(body)
}

// UploadMedia  上传临时素材并返回mediaId
func UploadMedia(msgType string, req *http.Request, accessToken string) (string, float64) {
	// 企业微信图片上传不能大于2M
	_ = req.ParseMultipartForm(2 << 20)
	imgFile, imgHeader, err := req.FormFile("media")
	log.Printf("文件大小==>%d字节", imgHeader.Size)
	if err != nil {
		log.Fatalln("图片文件出错==>", err)
		// 自定义code无效的图片文件
		return "", 400
	}
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	if createFormFile, err := writer.CreateFormFile("media", imgHeader.Filename); err == nil {
		readAll, _ := ioutil.ReadAll(imgFile)
		createFormFile.Write(readAll)
	}
	writer.Close()

	uploadMediaUrl := fmt.Sprintf(UploadMediaApi, accessToken, msgType)
	log.Println("uploadMediaUrl==>", uploadMediaUrl)
	newRequest, _ := http.NewRequest("POST", uploadMediaUrl, buf)
	newRequest.Header.Set("Content-Type", writer.FormDataContentType())
	log.Println("Content-Type ", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(newRequest)
	respData, _ := ioutil.ReadAll(resp.Body)
	mediaResp := ParseJson(string(respData))
	log.Println("企业微信上传临时素材接口返回==>", mediaResp)
	if err != nil {
		log.Fatalln("上传临时素材出错==>", err)
		return "", mediaResp["errcode"].(float64)
	} else {
		return mediaResp["media_id"].(string), float64(0)
	}
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

// 主函数入口
func main() {
	// 设置日志内容显示文件名和行号
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	wecomChan := func(res http.ResponseWriter, req *http.Request) {
		// 获取token
		accessToken := GetAccessToken()
		// 默认token有效
		tokenValid := true

		_ = req.ParseForm()

		// 定义变量用于存储参数
		var sendkey, msgContent, msgType, toUser, agentId string

		// 尝试从请求体（body）中解析JSON参数
		var requestBody RequestBody
		body, err := ioutil.ReadAll(req.Body)
		if err == nil && len(body) > 0 {
			err = json.Unmarshal(body, &requestBody)
			if err == nil {
				// 成功解析JSON请求体
				sendkey = requestBody.Sendkey
				msgType = requestBody.MsgType
				toUser = requestBody.ToUser
				agentId = requestBody.AgentId

				// 优先使用简化格式（向后兼容）
				if requestBody.Msg != "" {
					msgContent = requestBody.Msg
					log.Println("使用body传参（简化格式）")
				} else {
					// 使用官方格式（text.content 或 markdown.content）
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
				// JSON解析失败，回退到URL参数
				log.Printf("JSON解析失败: %v，回退到URL参数\n", err)
			}
		}

		// 如果body中没有获取到参数，则从URL参数中获取（保持向后兼容）
		if sendkey == "" {
			sendkey = req.FormValue("sendkey")
			msgContent = req.FormValue("msg")
			msgType = req.FormValue("msg_type")
			toUser = req.FormValue("touser")
			agentId = req.FormValue("agentid")
			log.Println("使用URL参数传参")
		}

		// 验证sendkey
		if sendkey != Sendkey {
			log.Panicln("sendkey 错误，请检查")
		}

		// 设置默认值
		if msgType == "" {
			msgType = "text"
		}
		if toUser == "" {
			toUser = WecomToUid
		}
		if agentId == "" {
			agentId = WecomAid
		}

		log.Println("mes_type=", msgType)
		// 默认mediaId为空
		mediaId := ""
		if msgType != "image" {
			log.Println("消息类型不是图片")
		} else {
			// token有效则跳出循环继续执行，否则重试3次
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

		// 准备发送应用消息所需参数
		postData := JsonData{
			ToUser:                 toUser,
			AgentId:                agentId,
			MsgType:                msgType,
			DuplicateCheckInterval: 600,
		}
			// 根据消息类型设置对应的内容字段
			if msgType == "markdown" {
				postData.Markdown = Markdown{
					Content: msgContent,
				}
			} else {
				postData.Text = Msg{
					Content: msgContent,
				}
			}

			// 如果是图片消息，设置MediaId
			if msgType == "image" {
				postData.Image = Pic{
					MediaId: mediaId,
				}
			}

		postStatus := ""
		for i := 0; i <= 3; i++ {
			sendMessageUrl := fmt.Sprintf(SendMessageApi, accessToken)
			postStatus = PostMsg(postData, sendMessageUrl)
			postResponse := ParseJson(postStatus)
			errcode := postResponse["errcode"]
			log.Println("发送应用消息接口返回errcode==>", errcode)
			tokenValid = ValidateToken(errcode)
			// token有效则跳出循环继续执行，否则重试3次
			if tokenValid {
				break
			}
			// 刷新token
			accessToken = GetAccessToken()
		}

		res.Header().Set("Content-type", "application/json")
		_, _ = res.Write([]byte(postStatus))
	}
	http.HandleFunc("/wecomchan", wecomChan)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
