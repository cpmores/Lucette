set shell := ["bash", "-euo", "pipefail", "-c"]

BINARY_NAME := "node"
CMD_PKG := "./cmd/node"
BIN_DIR := "bin"
COVER_FILE  := "coverage.out"

# 默认：列出所有命令
default:
    @just --list

# 构建 node 可执行文件
build:
    go build -o {{BIN_DIR}}/{{BINARY_NAME}} {{CMD_PKG}}

# 快速运行
run:
    go run {{CMD_PKG}}

# 只编译所有包，检查是否能通过编译
compile:
    go build ./...

# 运行所有测试
test:
    go test ./...

# 运行指定包的测试：just test-pkg ./core/...
test-pkg PKG="./...":
    go test {{PKG}}

# 详细运行单个测试以查看其输出：just trace TestPrintTrace
# 第二个参数可以换包：just trace TestFoo ./core/
trace NAME="TestPrintTrace" PKG="./sim/":
    go test {{PKG}} -v -count=1 -run {{NAME}}

# 带覆盖率的测试
cover:
    go test -coverprofile={{COVER_FILE}} ./...
    go tool cover -html={{COVER_FILE}}

# 格式化代码
fmt:
    go fmt ./...

# 静态检查
vet:
    go vet ./...

# 整理依赖
tidy:
    go mod tidy

# 生成代码
generate:
    go generate ./...

# 清理构建产物
clean:
    rm -rf {{BIN_DIR}} {{COVER_FILE}}

# 开发前检查：格式化 + 静态检查 + 测试
check: fmt vet test

# CI 常用：整理依赖 + 检查 + 构建
ci: tidy check build

# test hello
hello:
    echo "Hello, Just!"