# 许家印 API

一个故意“想很久但不回答”的 OpenAI Chat Completions 兼容恶搞服务。

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

## 梗背景

这个项目借用了中文互联网里与许家印和恒大相关的几个公开标签，并把它们转成了一个“没有实质内容但很有耐心”的 API：

- **“许皮带”**：公开报道和资料中常见的网络称呼，源于 2012 年公开活动中被注意到的带有 H 标志的皮带。项目使用 `xujiayin` 作为模型 ID，即取这个标签的辨识度，不代表任何官方命名。
- **“空城计”式表情包**：网络表情包中曾出现以许家印和恒大冰泉形象为素材、标题为“许家印空城计”的二创内容。本项目不打包、不托管这些图片，只借用“看起来马上要发生什么，实际上只剩等待”的喜剧结构。
- **起高楼与楼塌**：恒大相关公共讨论经常用“起高楼、宴宾客、楼塌了”概括从高速扩张到危机暴露的戏剧性反差。本项目把“楼塌”改写成服务端持续输出空格，纯粹作为程序行为的夸张隐喻。

这些梗来自不同质量的网络内容，不能互相替代，也不应把表情包文案当作事实依据。项目只保留可识别的文化语境，不复述未经核实的传闻。

## 参考链接

- [许家印 - 维基百科](https://zh.wikipedia.org/zh-hans/%E8%AE%B8%E5%AE%B6%E5%8D%B0)：包含“许皮带”称呼及其公开活动背景的资料汇总。
- [“许家印空城计”表情包页面](https://www.qiubiaoqing.com/index.php/img_detail/931884459811147452.html)：仅作为梗的出处线索，本仓库不重新分发页面图片。
- [BBC：恒大许家印被判处无期徒刑](https://www.bbc.com/zhongwen/articles/cvgjdjnynvlo/trad)：用于了解相关公共事件的新闻背景，不代表本项目对新闻之外网络说法的背书。

## License

[MIT License](LICENSE)
