.PHONY: proto test

proto:
	protoc \
		--go_out=. \
		--go_opt=module=github.com/kurosann/cs-im-sdk \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/kurosann/cs-im-sdk \
		proto/gateway/v1/channel_gateway.proto

test:
	go test ./...
