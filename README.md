# cs-im-sdk

业务渠道接入 SDK。渠道通过此 SDK 将客户消息提交到 cs-im-server，并接收客服回复推送到渠道。

## 能力边界

- **可以**：发送客户消息（`sender.role` 固定为 `customer`）
- **可以**：订阅客服消息推送（只收到 `sender.role == "agent"` 的消息）
- **不能**：发送客服消息
- **不能**：接收客户消息推送（客户消息来自渠道本身，无需回推）
- **不能**：调用管理接口

## 使用

```go
package main

import (
	"context"
	"fmt"
	"time"

	csim "github.com/kurosann/cs-im-sdk"
)

func main() {
	client := csim.NewClient("http://127.0.0.1:18081", "your-app-access-token")

	// 发送客户消息
	msg, err := client.SendCustomerMessage(context.Background(), csim.SendCustomerMessageRequest{
		ConversationID: "conv_xxx",
		CustomerID:     "cus_xxx",
		Text:           "你好，我需要帮助",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("sent: seq=%d id=%s\n", msg.Seq, msg.ID)

	// 订阅客服回复
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, closeFn, err := client.SubscribeAgentMessages(ctx, "conv_xxx")
	if err != nil {
		panic(err)
	}
	defer closeFn()
	for agentMsg := range events {
		fmt.Printf("agent reply: seq=%d text=%s\n", agentMsg.Seq, agentMsg.Text)
		// 将 agentMsg.Text 推送到渠道（WhatsApp/Telegram/...）
		_ = time.Now // placeholder
	}
}
```
