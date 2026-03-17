# HarborLink 编译与运行指南

## 环境要求

- **Go**: 1.25+
- **操作系统**: macOS / Linux / Windows

## 快速开始（无需外部依赖）

### 1. 下载依赖

```bash
cd harborlink
GOPROXY=https://goproxy.cn,direct go mod download
```

### 2. 编译

```bash
GOPROXY=https://goproxy.cn,direct go build -o bin/harborlink ./cmd/api
```

### 3. 运行

```bash
./bin/harborlink
```

**默认配置使用：**
- **数据库**: SQLite 内存数据库（无需安装 PostgreSQL）
- **缓存**: miniredis 内存 Redis（无需安装 Redis）

### 4. 测试

```bash
# 健康检查
curl http://localhost:8080/v2/health

# 服务信息
curl http://localhost:8080/

# 列出预订
curl http://localhost:8080/v2/bookings
```

---

## 生产环境部署

### 使用 PostgreSQL 和 Redis

1. **修改配置文件 `config.yaml`**:

```yaml
database:
  host: localhost      # PostgreSQL 主机
  port: 5432
  name: harborlink
  user: postgres
  password: "your_password"
  ssl_mode: disable

redis:
  host: localhost      # Redis 主机
  port: 6379
  password: ""
  db: 0
```

2. **设置环境变量（可选）**:

```bash
export HARBORLINK_DATABASE_HOST=localhost
export HARBORLINK_DATABASE_PASSWORD=your_password
export HARBORLINK_REDIS_HOST=localhost
```

3. **运行**:

```bash
./bin/harborlink
```

---

## 配置说明

### 服务器配置

```yaml
server:
  port: 8080           # 服务端口
  host: "0.0.0.0"      # 监听地址
  mode: debug          # debug / release / test
```

### 数据库配置

```yaml
database:
  host: sqlite         # 使用 "sqlite" 启用内存数据库，或设置 PostgreSQL 主机
  port: 5432
  name: harborlink
  user: postgres
  password: ""
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
```

### Redis 配置

```yaml
redis:
  host: miniredis      # 使用 "miniredis" 启用内存 Redis，或设置 Redis 主机
  port: 6379
  password: ""
  db: 0
```

### 航运公司配置

```yaml
carriers:
  - name: maersk
    code: MAEU
    adapter: mock
    enabled: true
    base_url: https://api.maersk.com
    api_key: ""
    rate_limit: 100
    poll_enabled: true
    poll_interval: 5s
```

---

## 可用的 API 端点

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/v2/health` | 健康检查 |
| GET | `/` | 服务信息 |
| GET | `/v2/bookings` | 列出预订 |
| POST | `/v2/bookings` | 创建预订 |
| GET | `/v2/bookings/:reference` | 获取预订 |
| PUT | `/v2/bookings/:reference` | 更新预订 |
| DELETE | `/v2/bookings/:reference` | 取消预订 |
| GET | `/v2/bookings/:reference/status` | 获取状态 |
| GET | `/v2/slot-watches` | 列出监控 |
| POST | `/v2/slot-watches` | 创建监控 |
| GET | `/v2/slot-watches/:reference` | 获取监控 |
| DELETE | `/v2/slot-watches/:reference` | 取消监控 |
| POST | `/v2/slot-watches/:reference/confirm` | 确认锁定 |
| GET | `/v2/ws` | WebSocket 连接 |

---

## 开发模式特性

### SQLite 内存数据库
- 无需安装 PostgreSQL
- 数据存储在内存中，重启后清空
- 适合开发和测试

### miniredis 内存 Redis
- 无需安装 Redis
- 完全兼容 Redis 协议
- 适合开发和测试

---

## 常见问题

### 1. 端口被占用

```bash
# 修改端口
export HARBORLINK_SERVER_PORT=8081
./bin/harborlink
```

### 2. 数据库连接失败

- 检查 PostgreSQL 是否运行
- 或使用 `database.host: sqlite` 启用内存数据库

### 3. Redis 连接失败

- 检查 Redis 是否运行
- 或使用 `redis.host: miniredis` 启用内存 Redis

---

## 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行特定包的测试
go test ./internal/repository/... -v
```

---

## 构建

```bash
# 本地构建
go build -o bin/harborlink ./cmd/api

# 跨平台构建
GOOS=linux GOARCH=amd64 go build -o bin/harborlink-linux ./cmd/api
GOOS=darwin GOARCH=arm64 go build -o bin/harborlink-darwin ./cmd/api
GOOS=windows GOARCH=amd64 go build -o bin/harborlink.exe ./cmd/api
```

---

## 版本信息

当前版本: **dev**

---

## 更新日志

### 2026-03-17
- ✅ 添加 SQLite 内存数据库支持
- ✅ 添加 miniredis 内存 Redis 支持
- ✅ 支持无外部依赖运行
- ✅ 添加编译和运行文档
