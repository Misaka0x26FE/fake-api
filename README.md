# 许家印 API

一个把许家印“空城计”做成 API 的 OpenAI Chat Completions 兼容恶搞服务。

它不调用真实模型，也不理解 prompt。对于合法的聊天补全请求，它会：

1. 输出 `<think>`。
2. 在流式模式下，每秒输出一个空格，持续 60 秒。
3. 结束响应。

非流式请求会直接返回 `<think>` 加 60 个空格；流式请求则把这 60 个空格拆成 60 个 SSE chunk。

> 这是一个技术玩具和网络文化二创项目，与许家印、恒大集团或 OpenAI 没有任何官方关联。项目中的名称、梗和描述仅用于讽刺性演示，不构成新闻、投资或法律意见。

## API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/v1/chat/completions` | OpenAI 风格聊天补全，支持 `stream` |
| `GET` | `/v1/models` | 返回 `xujiayin` 模型 |
| `GET` | `/v1/models/xujiayin` | 返回模型详情 |
| `GET` | `/health` | 健康检查 |

服务默认监听 `0.0.0.0:8080`。

## 快速开始

需要 Go 1.22 或更高版本。

```bash
go run .
```

或者编译后运行：

```bash
go build -o xujiayin-api .
./xujiayin-api
```

使用 tmux 长时间运行：

```bash
tmux new -s xujiayin-api
go run .
```

按 `Ctrl-b`、`d` 脱离会话，重新查看：

```bash
tmux attach -t xujiayin-api
```

## 使用示例

### 非流式

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"xujiayin","messages":[{"role":"user","content":"请开始思考"}]}'
```

响应中的 `content` 是 `<think>` 后跟 60 个空格。尾部空格在终端中不明显，可以用 `jq` 查看：

```bash
curl -s http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"xujiayin","messages":[{"role":"user","content":"hi"}]}' \
  | jq -r '.choices[0].message.content | @json'
```

### 流式

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"xujiayin","messages":[{"role":"user","content":"请开始思考"}],"stream":true}'
```

流式响应顺序为：assistant role chunk、`<think>` chunk、60 个间隔约 1 秒的空格 chunk、结束 chunk 和 `data: [DONE]`。

## 梗背景：许家印“空城计”

本项目映射的核心 meme 不是“许皮带”，而是近期网络二创里把诸葛亮的“空城计”套到许家印和恒大语境的说法。它抓住的是一种夸张的反差：门面、招牌和流程看起来都在，真正应该出现的内容却迟迟不来，最后只剩一座“空城”。

网络二创中可以看到“诸葛亮一觉醒来排第二了”“古有诸葛空城计，今有许家印空城计”等戏仿表达。本项目把这个结构直接翻译成程序行为：

- `<think>` 是看起来马上要开始工作的门面。
- 60 个逐秒发送的空格是“城里什么也没有”。
- 最后的结束 chunk 是等待 60 秒后才姗姗来迟的“结局”。

因此，这个 API 的笑点不是模型真的在思考，而是它非常认真地模拟了一个看似有进展、实际只输出空气的系统。本项目不打包、不托管相关图片，也不把网络段子当作具体案件事实。

## 参考链接

- [小梗人物志之许家印](https://m.ixigua.com/dx/7668981479787089161)：可见“许家印”“空城计”组合用法的网络二创示例。
- [许家印人物纪录片解说](https://www.sina.cn/news/detail/5334508560716523.html)：包含“诸葛亮一觉醒来排第二了”等戏仿表达的传播示例。
- [“许家印空城计”表情包页面](https://www.qiubiaoqing.com/index.php/img_detail/931884459811147452.html)：仅作为表情包线索，本仓库不重新分发页面图片。

## License

[MIT License](LICENSE)
