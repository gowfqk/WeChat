package main

type RequestBody struct {
	Sendkey  string    `json:"sendkey"`
	Msg      string    `json:"msg"`
	MsgType  string    `json:"msg_type"`
	ToUser   string    `json:"touser,omitempty"`
	AgentId  string    `json:"agentid,omitempty"`
	Text     *Msg      `json:"text,omitempty"`
	Markdown *Markdown `json:"markdown,omitempty"`
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
	ToUser                 string   `json:"touser"`
	AgentId                string   `json:"agentid"`
	MsgType                string   `json:"msgtype"`
	DuplicateCheckInterval int      `json:"duplicate_check_interval"`
	Text                   Msg      `json:"text"`
	Image                  Pic      `json:"image"`
	Markdown               Markdown `json:"markdown"`
}

type ExternalRequestBody struct {
	Sendkey         string          `json:"sendkey"`
	ExternalUserIds []string        `json:"external_userid"`
	Sender          string          `json:"sender"`
	MsgType         string          `json:"msgtype"`
	Text            *Msg            `json:"text,omitempty"`
	Image           *Pic            `json:"image,omitempty"`
	Markdown        *Markdown       `json:"markdown,omitempty"`
	Link            *LinkMsg        `json:"link,omitempty"`
	MiniProgram     *MiniProgramMsg `json:"miniprogram,omitempty"`
}

type LinkMsg struct {
	Title        string `json:"title"`
	Description  string `json:"desc"`
	Url          string `json:"url"`
	ThumbMediaId string `json:"thumb_media_id"`
}

type MiniProgramMsg struct {
	Title        string `json:"title"`
	AppId        string `json:"appid"`
	PagePath     string `json:"pagepath"`
	ThumbMediaId string `json:"thumb_media_id"`
}
