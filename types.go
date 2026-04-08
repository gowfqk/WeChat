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

type MailRequestBody struct {
	Sendkey  string   `json:"sendkey"`
	To       string   `json:"to"`
	Cc       string   `json:"cc,omitempty"`
	Subject  string   `json:"subject"`
	Content  string   `json:"content"`
	ReplyTo  string   `json:"reply_to,omitempty"`
}