#!/bin/bash
echo "测试1: 发送文本消息（不包含 miniprogram 字段）"
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wmbaMZEAAAM3CVZx_rFIXyO39ZLIpY0w"],
    "sender": "GuoWenQing",
    "msgtype": "text",
    "text": {
      "content": "测试消息 - 验证修复"
    }
  }' 2>&1

echo -e "\n\n测试2: 发送 Markdown 消息（不包含 miniprogram 字段）"
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wmbaMZEAAAM3CVZx_rFIXyO39ZLIpY0w"],
    "sender": "GuoWenQing",
    "msgtype": "markdown",
    "markdown": {
      "content": "**测试Markdown消息**"
    }
  }' 2>&1

echo -e "\n\n测试3: 发送小程序消息（包含必需字段）"
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wmbaMZEAAAM3CVZx_rFIXyO39ZLIpY0w"],
    "sender": "GuoWenQing",
    "msgtype": "miniprogram",
    "miniprogram": {
      "title": "小程序测试",
      "appid": "wx1234567890abcdef",
      "pagepath": "pages/index",
      "thumb_media_id": "MEDIA_ID"
    }
  }' 2>&1

echo -e "\n\n测试完成"
