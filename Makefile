MESSAGE?="test"

.PHONY: gen_proto
gen_proto:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
    pkg/echo/echo.proto

.PHONY: run_kxds
run_kxds:
	go run ./cmd/kxds

.PHONY: run_server
run_server:
	go run ./cmd/server

.PHONY: send_echo
send_echo:
	GRPC_GO_LOG_VERBOSITY_LEVEL=99 GRPC_GO_LOG_SEVERITY_LEVEL=info GRPC_XDS_BOOTSTRAP=./cmd/client/xds-bootstrap.json go run ./cmd/client -addr xds:///echo-server $(MESSAGE)
