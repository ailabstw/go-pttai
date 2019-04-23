# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gptt android ios gptt-cross swarm all test clean
.PHONY: gptt-linux gptt-linux-386 gptt-linux-amd64 gptt-linux-mips64 gptt-linux-mips64le
.PHONY: gptt-linux-arm gptt-linux-arm-5 gptt-linux-arm-6 gptt-linux-arm-7 gptt-linux-arm64
.PHONY: gptt-darwin gptt-darwin-386 gptt-darwin-amd64
.PHONY: gptt-windows gptt-windows-386 gptt-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

gptt:
	build/env.sh go run build/ci.go install ./cmd/gptt
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gptt\" to launch gptt."

bootnode:
	build/env.sh go run build/ci.go install ./cmd/bootnode
	@echo "Done building."

all:
	build/env.sh go run build/ci.go install

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint

clean:
	go clean && rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gptt-cross: gptt-linux gptt-darwin gptt-windows gptt-android gptt-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gptt-*

gptt-linux: gptt-linux-386 gptt-linux-amd64 gptt-linux-arm gptt-linux-mips64 gptt-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-*

gptt-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gptt
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep 386

gptt-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gptt
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep amd64

gptt-linux-arm: gptt-linux-arm-5 gptt-linux-arm-6 gptt-linux-arm-7 gptt-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep arm

gptt-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gptt
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep arm-5

gptt-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gptt
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep arm-6

gptt-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gptt
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep arm-7

gptt-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gptt
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep arm64

gptt-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gptt
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep mips

gptt-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gptt
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep mipsle

gptt-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gptt
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep mips64

gptt-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gptt
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gptt-linux-* | grep mips64le

gptt-darwin: gptt-darwin-386 gptt-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gptt-darwin-*

gptt-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gptt
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-darwin-* | grep 386

gptt-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gptt
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-darwin-* | grep amd64

gptt-windows: gptt-windows-386 gptt-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gptt-windows-*

gptt-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gptt
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-windows-* | grep 386

gptt-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gptt
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gptt-windows-* | grep amd64
