# wechat-alert

Grafana 告警 → 企业微信应用 转发服务

## 项目简介

这是一个轻量级的 Go 语言服务，用于接收 Grafana 的 Webhook 告警，并转发到企业微信应用（非群机器人）。

**特点：**
- 使用 **text** 格式发送消息（不是 markdown），确保个人微信也能正常查看
- 纯 Go 标准库实现，无第三方依赖
- 自动缓存 access_token，避免频繁请求
- 支持 Docker 一键部署

## 架构

```
Grafana → wechat-alert → 企业微信应用 → 用户（企业微信/个人微信）
```

## 快速开始

### 1. 配置环境变量

编辑 `docker-compose.yml`，填入企业微信配置：

```yaml
environment:
  - WECHAT_CORP_ID=your_corp_id        # 企业ID
  - WECHAT_AGENT_ID=your_agent_id      # 应用ID
  - WECHAT_AGENT_SECRET=your_secret    # 应用Secret
  - WECHAT_TO_USER=@all                # 接收用户
```

### 2. 启动服务

```bash
docker compose up -d
```

### 3. 配置 Grafana

在 Grafana 中配置 Contact Point：
- **Type**: Webhook
- **URL**: `http://your-server-ip:18091/webhook`
- **HTTP Method**: POST

## 环境变量配置

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `WECHAT_CORP_ID` | ✅ | - | 企业微信企业ID |
| `WECHAT_AGENT_ID` | ✅ | - | 企业微信应用ID |
| `WECHAT_AGENT_SECRET` | ✅ | - | 企业微信应用Secret |
| `WECHAT_TO_USER` | ❌ | `@all` | 接收用户，多个用 `\|` 分隔 |
| `WECHAT_TO_PARTY` | ❌ | 空 | 接收部门ID，多个用 `\|` 分隔 |
| `WECHAT_TO_TAG` | ❌ | 空 | 接收标签 |
| `SERVER_PORT` | ❌ | `8080` | 服务监听端口 |

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查，返回 `{"status":"ok"}` |
| `/webhook` | POST | Grafana Webhook 接收端点 |

## 测试验证

### 健康检查

```bash
curl http://localhost:18091/health
```

### 模拟 Grafana 告警（curl 测试）

```bash
curl -X POST http://localhost:18091/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "wechat",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "TestAlert",
          "instance": "server1:9090",
          "severity": "critical"
        },
        "annotations": {
          "summary": "This is a test alert",
          "description": "Testing wechat-alert webhook forwarding"
        },
        "startsAt": "2026-03-13T10:00:00Z",
        "endsAt": "0001-01-01T00:00:00Z",
        "generatorURL": "http://grafana:3000/alerting/test"
      }
    ],
    "groupLabels": {"alertname": "TestAlert"},
    "commonLabels": {"alertname": "TestAlert"},
    "commonAnnotations": {},
    "externalURL": "http://grafana:3000/"
  }'
```

如果配置正确，你的企业微信（和个人微信）会收到一条测试告警消息。

### 模拟告警恢复

```bash
curl -X POST http://localhost:18091/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "wechat",
    "status": "resolved",
    "alerts": [
      {
        "status": "resolved",
        "labels": {
          "alertname": "TestAlert",
          "instance": "server1:9090",
          "severity": "critical"
        },
        "annotations": {
          "summary": "This is a test alert",
          "description": "Testing wechat-alert webhook forwarding"
        },
        "startsAt": "2026-03-13T10:00:00Z",
        "endsAt": "2026-03-13T10:05:00Z",
        "generatorURL": "http://grafana:3000/alerting/test"
      }
    ],
    "groupLabels": {"alertname": "TestAlert"},
    "commonLabels": {"alertname": "TestAlert"},
    "commonAnnotations": {},
    "externalURL": "http://grafana:3000/"
  }'
```

## 消息格式示例

```
[Grafana告警]
状态: firing
告警名称: HighCPU
级别: critical
实例: server1:9090
摘要: CPU usage is high
详情: CPU usage above 90%
开始时间: 2026-03-12 10:00:00
```

多条告警之间用分隔线隔开。

## 本地开发

```bash
# 设置环境变量
export WECHAT_CORP_ID=your_corp_id
export WECHAT_AGENT_ID=your_agent_id
export WECHAT_AGENT_SECRET=your_secret

# 运行服务
go run .

# 测试健康检查
curl http://localhost:8080/health
```

## License

MIT