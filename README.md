# go-home-inventory

## 项目说明

go-home-inventory 是一个使用内存存储的家庭物品和保修管理 Web API。它支持登记家电和贵重物品、维护房间与具体位置、计算保修状态、记录维修历史，以及查询一段日期范围内即将过保的物品。

所有日期字段使用 RFC3339 格式。服务启动时通过 `APP_TIMEZONE` 指定业务时区，默认使用 `Asia/Shanghai`。

## 标准命令

```bash
go build ./...
go test ./...
go vet ./...
go run ./cmd/server
```

## 使用方式

默认监听 `:8080`，可通过 `HTTP_ADDR` 修改。

- `POST /items`：新增物品
- `GET /items`：查询物品，可用 `room` 和 `position` 筛选
- `GET /items/{id}`：查询单个物品
- `PUT /items/{id}`：修改物品
- `DELETE /items/{id}`：删除物品
- `PUT /items/{id}/location`：设置房间和具体位置
- `GET /items/{id}/warranty`：查询保修状态
- `POST /items/{id}/repairs`：添加维修记录
- `GET /items/{id}/repairs`：查询维修历史
- `GET /reminders/upcoming?from=...&to=...`：查询即将过保物品
- `POST /reminders/summary`：批量生成提醒摘要

内存数据会在进程退出后清空。
