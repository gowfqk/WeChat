#!/bin/bash

echo "========== 测试1：简化格式 =========="
echo "请求：{\"msg_type\": \"text\", \"msg\": \"测试消息\"}"
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "msg_type": "text",
    "msg": "测试消息"
  }' \
  2>/dev/null | jq .

echo ""
echo "========== 测试2：官方格式 =========="
echo "请求：{\"msg_type\": \"text\", \"text\": {\"content\": \"测试消息\"}}"
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "msg_type": "text",
    "text": {
      "content": "测试消息"
    }
  }' \
  2>/dev/null | jq .

echo ""
echo "========== 测试3：带sendkey的简化格式 =========="
echo "请求：{\"sendkey\": \"set_a_sendkey\", \"msg_type\": \"text\", \"msg\": \"测试消息\"}"
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "set_a_sendkey",
    "msg_type": "text",
    "msg": "测试消息"
  }' \
  2>/dev/null | jq .

echo ""
echo "========== 测试4：URL参数格式 =========="
echo "请求：/?sendkey=set_a_sendkey&msg_type=text&msg=测试消息"
curl "http://localhost:8080/wecomchan?sendkey=set_a_sendkey&msg_type=text&msg=测试消息" \
  2>/dev/null | jq .
