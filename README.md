# GrabSeat

CCNU图书馆预约座位抢座工具

## 简介

GrabSeat 是一个用 Go 编写的自动化工具，用于帮助中南民族大学 (CCNU) 图书馆进行座位预约和抢座。它支持自动登录、查询可用座位、提交预约请求等功能，旨在提高预约成功率并减少人工操作。

## 特性

- 使用 Go 编写，轻量且高性能
- 支持自动登录与会话管理
- 支持按条件搜索可用座位并自动提交预约
- 可通过 Docker 打包部署（仓库包含 Dockerfile）

## 依赖

- Go 1.18+

## 快速开始

1. 克隆仓库：

   git clone https://github.com/Serendipity565/GrabSeat.git
   cd GrabSeat

2. 配置

   在项目根目录创建一个配置文件（例如 config.json 或 .env），配置登录信息与预约偏好。具体字段请参考代码中 `config` 或相关 README 示例（若仓库中已有示例文件，请以示例为准）。

3. 编译并运行：

   go build -o grabseat ./...
   ./grabseat

4. 使用 Docker（可选）：

   docker build -t grabseat:latest .
   docker run -v $(pwd)/config.json:/app/config.json grabseat:latest

## 配置示例（示意）

```json
{
  "username": "your_student_id",
  "password": "your_password",
  "library": "xxx",
  "seat_preferences": ["A1","B2"],
  "notify": {
    "enabled": false
  }
}
```

请根据项目实际代码调整字段名与结构。

## 贡献

欢迎提交 issue 或 PR。提交 PR 前请先说明变更目的，并确保测试通过。

## 注意事项

- 请确保遵守学校与图书馆的相关使用政策与法律法规。不要滥用自动化抢座导致资源滥占。
- 仓库仅供学习与研究使用。

## 许可证

MIT