# CORS 任意来源支持计划

## TL;DR

> **快速Summary**: 修改 CORS 检查逻辑，支持通配符 "*" 允许任意来源

> **Deliverables**:
> - 修改 `checkOriginAllowed` 函数支持通配符
> - 修改 `isAllowedOrigin` 函数支持通配符
> - 修改 `authMiddleware` 处理通配符时的响应头
> - 更新 `config.md` 文档

> **Estimated Effort**: Short
> **Parallel Execution**: NO - sequential

---

## Context

### 问题描述
当前 CORS 只支持精确匹配，不支持通配符。用户需要支持任意来源。

### 当前行为
- `corsAllowedOrigins: []` - 拒绝所有跨域请求
- `corsAllowedOrigins: ["http://example.com"]` - 只允许指定域名

### 期望行为
- `corsAllowedOrigins: ["*"]` - 允许任意来源
- `corsAllowedOrigins: ["null"]` - 允许 null 来源 (如 file://)
- 仍支持精确匹配

### CORS 规范说明
- 使用 `*` 时，不能同时使用 `Access-Control-Allow-Credentials: true`
- 如果需要 credentials，只能列出具体域名

---

## Work Objectives

### Core Objective
支持通过设置 `corsAllowedOrigins: ["*"]` 允许任意来源访问

### Concrete Deliverables
1. 修改 `cloud-clip/lib/main.go` 中的 `checkOriginAllowed` 函数
2. 修改 `cloud-clip/lib/main.go` 中的 `isAllowedOrigin` 函数
3. 修改 `cloud-clip/lib/main.go` 中的 `authMiddleware` 函数
4. 更新 `cloud-clip/config.md` 文档

### Definition of Done
- [ ] `checkOriginAllowed` 支持 "*" 通配符
- [ ] `isAllowedOrigin` 支持 "*" 通配符
- [ ] authMiddleware 正确设置响应头
- [ ] config.md 文档更新

### Must Have
- 支持 "*" 允许所有来源
- 支持 "null" 允许 null 来源
- 保持向后兼容 (精确匹配仍然有效)

### Must NOT Have
- 不破坏现有精确匹配逻辑

---

## Verification Strategy

### Agent-Executed QA Scenarios

**Scenario: Build verification**
  Tool: Bash
  Preconditions: Go 1.23+ installed
  Steps:
    1. cd cloud-clip
    2. go build -buildvcs=false -o /tmp/cors-test .
    3. Assert: exit code 0
  Expected Result: Build succeeds

---

## TODOs

### Task 1: 修改 checkOriginAllowed 函数

**What to do**:
修改 `cloud-clip/lib/main.go` 中的 `checkOriginAllowed` 函数，支持：
- `"*"` - 允许任意来源
- `"null"` - 允许 null 来源

**New Logic**:
```go
func checkOriginAllowed(origin, host string, allowedOrigins []string) bool {
	// 支持通配符 "*"
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		// 支持 null (file:// 协议)
		if allowed == "null" && origin == "null" {
			return true
		}
		if origin == allowed {
			return true
		}
	}
	
	// 默认情况下，只允许同源请求
	if origin == "" {
		return true
	}
	
	return strings.Contains(origin, "://"+host)
}
```

**References**:
- `cloud-clip/lib/main.go:42-59` - 当前实现

**Acceptance Criteria**:
- [ ] 函数支持 "*" 返回 true
- [ ] 函数支持 "null" 返回 true
- [ ] 精确匹配仍然有效

---

### Task 2: 修改 isAllowedOrigin 函数

**What to do**:
修改 `cloud-clip/lib/main.go` 中的 `isAllowedOrigin` 函数，保持与 `checkOriginAllowed` 一致

**New Logic**:
```go
func (s *ClipboardServer) isAllowedOrigin(origin string, host string) bool {
	allowedOrigins := s.config.Server.CORSAllowedOrigins

	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowed := range allowedOrigins {
		// 支持通配符 "*"
		if allowed == "*" {
			return true
		}
		// 支持 null
		if allowed == "null" && origin == "null" {
			return true
		}
		if origin == allowed {
			return true
		}
	}
	return false
}
```

**References**:
- `cloud-clip/lib/main.go:706-719` - 当前实现

**Acceptance Criteria**:
- [ ] 函数支持 "*" 返回 true
- [ ] 精确匹配仍然有效

---

### Task 3: 修改 authMiddleware 处理通配符响应头

**What to do**:
修改 `cloud-clip/lib/main.go` 中的 `authMiddleware` 函数

当使用 "*" 时：
- `Access-Control-Allow-Origin` 设置为 "*"
- 不能设置 `Access-Control-Allow-Credentials: true`

**New Logic**:
```go
func (s *ClipboardServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" {
			// 检查是否使用通配符
			hasWildcard := false
			for _, allowed := range s.config.Server.CORSAllowedOrigins {
				if allowed == "*" {
					hasWildcard = true
					break
				}
			}

			if s.isAllowedOrigin(origin, r.Host) {
				if hasWildcard {
					// 通配符模式：不能使用 credentials
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					// 精确匹配模式
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")
				if !hasWildcard {
					w.Header().Set("Vary", "Origin")
				}
			} else {
				http.Error(w, "Origin not allowed", http.StatusForbidden)
				return
			}
		}
		// ... rest of function
	}
}
```

**References**:
- `cloud-clip/lib/main.go:721-737` - 当前实现

**Acceptance Criteria**:
- [ ] 使用 "*" 时响应头正确
- [ ] 精确匹配时 credentials 仍为 true

---

### Task 4: 更新 config.md 文档

**What to do**:
更新 `cloud-clip/config.md` 文档，添加通配符说明

**Add to server section**:
```json
"corsAllowedOrigins": ["*"], // CORS允许的来源，支持:
//   - 具体域名: "http://example.com"
//   - 通配符: "*" (允许所有来源，但不支持credentials)
//   - null: "null" (允许file://协议)
```

**References**:
- `cloud-clip/config.md` - 配置文件说明

**Acceptance Criteria**:
- [ ] 文档包含通配符使用说明

---

### Task 5: 验证构建

**What to do**:
运行 `go build` 验证代码无错误

**Acceptance Criteria**:
- [ ] `cd cloud-clip && go build -buildvcs=false .` 成功

---

## Success Criteria

### Verification Commands
```bash
cd cloud-clip && go build -buildvcs=false -o /tmp/cors-test .
```

### Final Checklist
- [ ] checkOriginAllowed 支持 "*"
- [ ] isAllowedOrigin 支持 "*"
- [ ] authMiddleware 正确处理通配符
- [ ] 文档已更新
- [ ] 构建成功
