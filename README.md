# fake-api

一个模拟 OpenAI Chat Completions API 的假服务，任何请求都返回 `Hello, World!`。

## 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions` | 聊天补全，支持流式 |
| GET  | `/v1/models` | 模型列表 |
| GET  | `/v1/models/{id}` | 模型详情 |

## 启动

```bash
go build -o fake-api .
./fake-api
```

或

```bash
go run main.go
```

默认监听 `:8080`。

## 请求示例

```bash
# 普通请求
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}'

# 流式请求
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stream":true}'

# 模型列表
curl http://localhost:8080/v1/models
```