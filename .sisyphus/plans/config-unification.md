# 配置文件参数统一计划

## TL;DR

> **快速Summary**: 统一 cloud-clipboard-go 项目的配置参数，解决默认值不一致、命名混乱、文档缺失等问题

> **Deliverables**:
> - 修复 entrypoint.sh 默认值 (history: 10→100, cleanupInterval: 300→86400)
> - 添加缺失的 CLI flags (rate_limit, room_list, cors_allowed_origins 等)
> - 统一参数命名为 UPPER_SNAKE_CASE
> - 完善 config.md 文档
> - 删除根目录重复的 docker-compose.yml

> **Estimated Effort**: Short
> **Parallel Execution**: NO - sequential
> **Critical Path**: 修复默认值 → 添加 CLI flags → 更新文档 → 删除重复文件

---

## Context

### Original Request
用户要求检查项目参数配置明细是否统一，包括说明。分析后发现存在多个不一致问题。

### 用户决策
- **命名规范**: UPPER_SNAKE_CASE (如 MESSAGE_NUM, TEXT_LIMIT)
- **默认值**: 以 config.go 为准 (history=100, cleanupInterval=86400)
- **docker-compose**: 删除根目录重复版本

### 研究发现
1. **默认值差异巨大**:
   - history: config.go=100, entrypoint.sh=10 (10倍)
   - cleanupInterval: config.go=86400, entrypoint.sh=300 (288倍)
   
2. **三套命名风格**:
   - CLI: kebab-case (text_limit, file_expire)
   - Docker: UPPER_SNAKE_CASE (TEXT_LIMIT, FILE_EXPIRE)
   - Config JSON: camelCase/mixed (text.limit, file.expire)

3. **缺失参数**:
   - CLI 缺少: rateLimit, rateLimitBurst, roomList, roomCleanup, corsAllowedOrigins
   - 文档缺少: rateLimit, corsAllowedOrigins, cleanupInterval

---

## Work Objectives

### Core Objective
统一所有配置参数的默认值、命名和文档，消除生产环境与预期不符的行为

### Concrete Deliverables
1. 修改 `cloud-clip/entrypoint.sh` 默认值
2. 修改 `cloud-clip/lib/flags.go` 添加缺失 CLI 参数
3. 修改 `cloud-clip/config.md` 完善文档
4. 删除根目录 `docker-compose.yml`

### Definition of Done
- [ ] entrypoint.sh 中 history 默认值从 10 改为 100
- [ ] entrypoint.sh 中 cleanupInterval 默认值从 300 改为 86400
- [ ] flags.go 添加 rate_limit, rate_limit_burst, room_list, room_cleanup, cors_allowed_origins 参数
- [ ] config.md 添加 rateLimit, corsAllowedOrigins, cleanupInterval 字段说明
- [ ] 删除根目录 docker-compose.yml

### Must Have
- 不破坏现有功能 (向后兼容)
- 修复后运行 `go build` 成功
- 文档与代码行为一致

### Must NOT Have
- 不要修改 config.go 中的默认值 (已正确)
- 不要添加新的配置项 (只修复现有)
- 不要删除 cloud-clip/docker-compose.yml

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: NO (Go 项目无测试框架)
- **Automated tests**: None
- **Agent-Executed QA**: YES - 验证构建成功

### Agent-Executed QA Scenarios

**Scenario: Go build verification**
  Tool: Bash
  Preconditions: Go 1.23+ installed
  Steps:
    1. cd cloud-clip
    2. go build -o /tmp/cloud-clip-test .
    3. Assert: exit code 0
    4. /tmp/cloud-clip-test -h | head -20
    5. Assert: output contains new CLI flags
  Expected Result: Build succeeds, help shows new parameters
  Evidence: Build output captured

---

## Execution Strategy

### Sequential Steps

```
Step 1: Fix entrypoint.sh defaults
Step 2: Add missing CLI flags to flags.go
Step 3: Update config.md documentation
Step 4: Delete duplicate docker-compose.yml
Step 5: Verify build
```

---

## TODOs

### Task 1: 修复 entrypoint.sh 默认值

**What to do**:
- 修改 `cloud-clip/entrypoint.sh` 中:
  - `MESSAGE_NUM:-10` → `MESSAGE_NUM:-100`
  - `FILE_CLEANUP_INTERVAL:-300` → `FILE_CLEANUP_INTERVAL:-86400`

**Must NOT do**:
- 不要修改其他逻辑

**References**:
- `cloud-clip/entrypoint.sh:116` - history 默认值位置
- `cloud-clip/entrypoint.sh:131` - cleanupInterval 默认值位置
- `cloud-clip/lib/config.go:109,135` - 标准默认值参考

**Acceptance Criteria**:
- [ ] entrypoint.sh 第116行: `${MESSAGE_NUM:-100}`
- [ ] entrypoint.sh 第131行: `${FILE_CLEANUP_INTERVAL:-86400}`

---

### Task 2: 添加缺失的 CLI flags

**What to do**:
- 在 `cloud-clip/lib/flags.go` 添加以下 CLI 参数:
  - `-rate_limit` (对应 rateLimit)
  - `-rate_limit_burst` (对应 rateLimitBurst)
  - `-room_list` (对应 roomList)
  - `-room_cleanup` (对应 roomCleanup)
  - `-cors_allowed_origins` (对应 corsAllowedOrigins)

- 在 `applyCommandLineArgs()` 函数中添加处理逻辑

**Must NOT do**:
- 不要删除现有参数
- 不要修改现有参数的名称

**References**:
- `cloud-clip/lib/flags.go:22-28` - 现有参数定义模式
- `cloud-clip/lib/flags.go:98-182` - applyCommandLineArgs 实现模式
- `cloud-clip/lib/config.go:26-30` - Config 结构体定义

**Acceptance Criteria**:
- [ ] flags.go 添加 5 个新 flag 变量
- [ ] applyCommandLineArgs 处理这 5 个参数
- [ ] go build 成功

---

### Task 3: 更新 config.md 文档

**What to do**:
- 在 `cloud-clip/config.md` 添加以下字段说明:
  - `rateLimit` - 每秒最大请求数
  - `rateLimitBurst` - 突发请求数
  - `corsAllowedOrigins` - CORS 允许来源
  - `file.cleanupInterval` - 文件清理间隔

**Must NOT do**:
- 不要删除现有文档
- 不要修改现有字段说明 (除非错误)

**References**:
- `cloud-clip/lib/config.go:29-30,39` - 字段定义和注释

**Acceptance Criteria**:
- [ ] config.md 包含 rateLimit 说明
- [ ] config.md 包含 rateLimitBurst 说明
- [ ] config.md 包含 corsAllowedOrigins 说明
- [ ] config.md 包含 cleanupInterval 说明

---

### Task 4: 删除重复的 docker-compose.yml

**What to do**:
- 删除根目录的 `docker-compose.yml`
- 保留 `cloud-clip/docker-compose.yml`

**Must NOT do**:
- 不要删除 cloud-clip/docker-compose.yml
- 不要修改 cloud-clip/docker-compose.yml 内容

**References**:
- `cloud-clip/docker-compose.yml` - 保留版本 (更完整)

**Acceptance Criteria**:
- [ ] 根目录 docker-compose.yml 已删除
- [ ] cloud-clip/docker-compose.yml 存在

---

### Task 5: 验证构建

**What to do**:
- 运行 `go build` 验证代码无语法错误
- 运行 `./cloud-clipboard-go -h` 验证帮助信息

**Must NOT do**:
- 不要修改任何文件

**References**:
- `cloud-clip/Makefile` - 构建命令

**Acceptance Criteria**:
- [ ] `cd cloud-clip && go build -o /tmp/test-build .` 成功
- [ ] `/tmp/test-build -h` 显示帮助信息

---

## Success Criteria

### Verification Commands
```bash
cd cloud-clip && go build -o /tmp/test-build .
/tmp/test-build -h
```

### Final Checklist
- [ ] entrypoint.sh 历史记录默认值为 100
- [ ] entrypoint.sh 清理间隔默认值为 86400
- [ ] flags.go 包含 5 个新参数
- [ ] config.md 文档完整
- [ ] 根目录 docker-compose.yml 已删除
- [ ] go build 成功
