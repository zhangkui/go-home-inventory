# go-home-inventory candidate 3

## 项目说明
家庭物品和保修管理 Web API，使用 golang:1.22。

## 标准命令
cd '/app' && GOTOOLCHAIN=local go build ./...
cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
分别构建 linux/amd64 与 linux/arm64。

## 题目验证命令
go test -buildvcs=false -count=1 ./...
go build -buildvcs=false ./...

## Bug 复现
见 BUG_REPRO.md。
