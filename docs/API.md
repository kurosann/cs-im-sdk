# cs-im-sdk 接口契约

本文是业务渠道 SDK 的语言无关契约。Go SDK 是当前参考实现；后续 TypeScript、JavaScript、Kotlin 或其他语言 SDK 必须保持同一端点、字段、错误语义和能力边界。

## 定位与能力边界

`cs-im-sdk` 供业务渠道使用。业务渠道是消息的入口：客户消息从渠道进入系统，客服回复从系统返回渠道。

SDK 只提供：

- 创建或复用客户会话
- 提交客户消息
- 查询会话历史消息
- 订阅客服消息

SDK 不提供：

- 发送客服消息
- 接收客户消息回推
- App、员工、租户等管理接口
- 转发渠道私有接口
- 文件上传、媒体存储、转码和内容审核

`sender.role` 由 SDK 固定为 `customer`。即使底层 HTTP 请求体可写，语言 SDK 的公共类型也不应暴露修改角色或发送客服消息的能力。

## 基础约定

### 服务地址

Go SDK 使用两个地址：

- HTTP / SSE：`NewClient(baseURL, appToken)`
- gRPC：`DialChannelGateway(ctx, endpoint, accountToken)`

HTTP 默认超时是 15 秒，只用于普通请求。SSE 使用独立的 client，不设置整体超时，避免长连接被固定超时切断。

### 认证

所有请求使用租户账号 token：

```http
Authorization: Bearer <accountToken>
```

服务端先用 token 找到账号，再校验账号所属租户是否拥有 `appId` 对应的 App。因此渠道请求必须传目标 `appId`，服务端不把 token 本身当作 App 授权码。

### JSON 响应信封

HTTP 响应格式固定为：

```json
{
  "code": "OK",
  "message": "",
  "data": {}
}
```

`data` 才是业务负载。SDK 在 HTTP 状态码不是 2xx，或 `code != "OK"` 时返回错误。

时间字段使用 RFC3339，例如 `2025-01-01T08:00:00Z`。客户端消息必须保留渠道上传的原始发送时间。

## 通用模型

### Conversation

```ts
interface Conversation {
  conversationId: string;
  appId: string;
  customerId: string;
  channel?: string;
  subject?: string;
  status?: string;
}
```

### Message

```ts
interface MessageSender {
  role: "customer" | "agent" | "ai" | "system";
  id: string;
}

interface MessageSource {
  appId?: string;
  externalId?: string;
  sentAt?: string;
}

interface MessageContent {
  url: string;
  mimeType?: string;
  fileName?: string;
  size?: number;
  width?: number;
  height?: number;
  durationSeconds?: number;
  thumbnailUrl?: string;
  checksum?: string;
}

interface Message {
  messageId: string;
  seq: number;
  conversationId: string;
  customerId: string;
  sender: MessageSender;
  source?: MessageSource;
  type: string;
  text?: string;
  content?: MessageContent;
  status: string;
  createdAt: string;
  updatedAt: string;
}
```

字段说明：

- `seq`：会话内递增序号，用作历史分页游标。
- `messageId`：服务端消息 ID，用于客户端本地去重和更新。
- `source.externalId`：渠道侧唯一消息 ID，用于服务端幂等去重。
- `source.sentAt`：客户在渠道实际发送时间。这是业务时间；服务端另有接收时间。
- `createdAt`：优先返回客户实际发送时间，否则返回服务端接收时间。
- `content`：媒体引用。`type` 为 `image`、`file`、`video` 时存在；`url` 由调用方保证可访问。

## HTTP API

### 创建或复用会话

```http
POST /api/v1/channels/{appId}/conversations/ensure
```

请求：

```json
{
  "customerExternalId": "wa-user-123",
  "customerName": "Kuro",
  "channel": "whatsapp",
  "subject": "WhatsApp inquiry"
}
```

响应 `data`：

```json
{
  "conversation": {
    "conversationId": "conv_123",
    "appId": "app_123",
    "customerId": "cus_123",
    "channel": "whatsapp",
    "status": "open"
  },
  "created": true
}
```

规则：

- `customerExternalId` 必填。
- 同一 `appId` 下，同一个外部客户如果已有 `open` 会话，则复用该会话并返回 `created=false`。
- 否则创建会话并返回 `created=true`。

### 发送客户消息

```http
POST /api/v1/channels/{appId}/messages
```

请求：

```json
{
  "conversationId": "conv_123",
  "customerId": "cus_123",
  "sender": {
    "role": "customer",
    "id": "cus_123"
  },
  "type": "text",
  "text": "hello",
  "source": {
    "appId": "app_123",
    "externalId": "wa-message-id-001",
    "sentAt": "2025-01-01T08:00:00.123456789Z"
  }
}
```

媒体消息示例：

```json
{
  "conversationId": "conv_123",
  "customerId": "cus_123",
  "sender": {
    "role": "customer",
    "id": "cus_123"
  },
  "type": "image",
  "text": "截图说明",
  "content": {
    "url": "https://cdn.example.com/photo.png",
    "mimeType": "image/png",
    "fileName": "photo.png",
    "size": 102400,
    "width": 1280,
    "height": 720
  },
  "source": {
    "appId": "app_123",
    "externalId": "wa-message-id-002",
    "sentAt": "2025-01-01T08:01:00.123456789Z"
  }
}
```

响应 `data` 是单个 `Message`。

文本消息必填项：

- `conversationId`
- `customerId`
- `text`
- `source.externalId`
- `source.sentAt`

规则：

- `type` 支持 `text`、`image`、`file`、`video`；默认是 `text`。
- 媒体消息必须传 `content.url`，协议必须是 `http` 或 `https`；`text` 作为可选说明文字。
- 服务端不负责文件上传和下载，也不保证外部 URL 长期可访问。
- `sender.role` 固定为 `customer`。
- `sender.id` 必须等于 `conversation.customerId`。
- `conversationId` 必须属于路径中的 `appId`。
- `source.externalId` 与 `appId` 组成幂等键。重试同一条消息时返回已有消息，不新增 `seq`。
- 如果渠道没有提供唯一消息 ID，SDK 层应该为每次渠道事件生成稳定 ID，避免网络重试产生重复消息。

### 查询历史消息

```http
GET /api/v1/channels/{appId}/conversations/{conversationId}/messages
```

查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `afterSeq` | int64 | 大于该序号。0 表示不限制。 |
| `beforeSeq` | int64 | 小于该序号。0 表示不限制。 |
| `limit` | int | 默认 50，最大 100。 |
| `order` | string | `asc` 或 `desc`，默认 `desc`。 |

响应 `data`：

```ts
interface MessagePage {
  items: Message[];
  nextAfterSeq: number;
  nextBeforeSeq: number;
  hasMore: boolean;
}
```

分页规则：

- `afterSeq` 和 `beforeSeq` 都是排除边界。
- 拉新消息：`order=asc&afterSeq=<localCursor>`。
- 向上翻页：`order=desc&beforeSeq=<oldestLocalSeq>`。
- `asc` 页使用 `nextAfterSeq` 继续拉取。
- `desc` 页使用 `nextBeforeSeq` 继续向前翻页。
- 建议客户端把本会话的最大 `seq` 和最小 `seq` 持久化，作为本地游标。

## 客服消息订阅

业务渠道通过 SSE 订阅客服回复：

```http
GET /api/v1/channels/{appId}/realtime
Accept: text/event-stream
Authorization: Bearer <accountToken>
```

事件格式：

```
data: {"type":"message.created","id":"msg_123","data":{},"at":"2025-01-01T08:00:00Z"}
```

SDK 行为：

- 只处理 `type == "message.created"` 的事件。
- 只输出 `sender.role == "agent"` 的消息。
- `conversationId` 不为空时，只输出该会话的消息。
- 服务端也会按业务渠道边界过滤，客户消息不会回推给业务渠道。

断线重连建议：

1. 使用指数退避重试。
2. 重连后先用 `afterSeq` 拉取本会话缺失消息。
3. 再恢复 SSE。
4. 不要假设 SSE 能覆盖断线期间全部事件。

## gRPC 契约

proto 源文件是 `proto/gateway/v1/channel_gateway.proto`，服务名：

```text
cs.im.gateway.v1.ChannelGatewayService
```

认证元数据：

```text
authorization: Bearer <accountToken>
```

### 方法

#### EnsureConversation

```proto
rpc EnsureConversation(EnsureConversationRequest) returns (ConversationResponse);
```

语义与 HTTP 的 ensure conversation 相同。`created` 表示是否新创建。

#### SendMessage

```proto
rpc SendMessage(SendMessageRequest) returns (MessageResponse);
```

语义与 HTTP 的 send customer message 相同。`sender.role` 必须是 `customer`，`source.external_id` 和 `source.sent_at` 必填。

#### ListMessages

```proto
rpc ListMessages(ListMessagesRequest) returns (ListMessagesResponse);
```

`after_seq`、`before_seq`、`limit`、`order` 的语义与 HTTP 历史查询一致。

#### StreamMessages

```proto
rpc StreamMessages(StreamMessagesRequest) returns (stream MessageEventResponse);
```

服务端只推送允许业务渠道接收的客服消息。实现语言 SDK 时建议把 gRPC stream 包装成 `AsyncIterable`、`Flowable` 或其他平台惯用的流类型。

## 错误处理

HTTP 错误由 Go SDK 映射为：

```ts
interface APIError {
  statusCode: number;
  code: string;
  message: string;
}
```

常见 gRPC code：

| gRPC code | 语义 |
| --- | --- |
| `InvalidArgument` | 参数缺失或格式错误。 |
| `Unauthenticated` | token 缺失或无效。 |
| `PermissionDenied` | token 有效但不允许访问 App，或业务渠道越界。 |
| `NotFound` | App 或会话不存在。 |
| `Internal` | 服务端内部错误。 |

SDK 应区分可重试错误和业务错误：

- 网络、超时、5xx：可退避重试。
- `Unauthenticated` / `PermissionDenied`：不要简单重试，先修复凭据或授权。
- `InvalidArgument`：不要重试，修正调用参数。
- `NotFound`：确认 `appId`、`conversationId` 是否正确。

## 多语言 SDK 实现要求

新增 TS / JS SDK 时，建议目录和能力如下：

```text
cs-im-sdk-ts/
├── src/
│   ├── client.ts
│   ├── message.ts
│   ├── realtime.ts
│   ├── errors.ts
│   └── types.ts
├── tests/
└── package.json
```

最低公共接口：

```ts
class ChannelClient {
  constructor(baseUrl: string, accountToken: string, options?: ClientOptions);

  ensureConversation(input: EnsureConversationInput): Promise<EnsureConversationResult>;
  sendCustomerMessage(input: SendCustomerMessageInput): Promise<Message>;
  listMessages(appId: string, conversationId: string, filter?: MessageFilter): Promise<MessagePage>;
  subscribeAgentMessages(appId: string, conversationId: string): AgentMessageStream;
}
```

实现要求：

- 类型中的发送角色固定为 `customer`，不要暴露 `sender` 字段。
- `source.externalId` 和 `source.sentAt` 必填。
- SSE 使用独立 fetch client，不要复用带全局超时的 axios/fetch wrapper。
- 关闭 SSE 时要同时取消请求、退出读取循环并关闭事件流。
- 单元测试必须覆盖：请求路径、请求体、角色固定、幂等字段、分页边界、只接收客服消息。
