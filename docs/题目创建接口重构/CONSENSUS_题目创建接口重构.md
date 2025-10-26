# 题目创建接口重构 - 共识文档

## ✅ 已确认的决策

### 1. 状态流转规则
- **采用自由变更模式**：Draft ↔ Published ↔ Archived
- 状态可以灵活变更，无硬性限制

### 2. IsPublic与Status的关联规则
```
Draft (草稿)     → IsPublic = false（强制私有，不可公开）
Published (已发布) → IsPublic = 可选（创建者可自由设置）
Archived (已归档)  → IsPublic = 可选（保持原有状态）
```

### 3. 默认值设置
```go
Status:       StatusDraft  // "draft" - 默认草稿状态
IsPublic:     false        // 默认私有
ACCount:      0            // 通过次数
SubmitCount:  0            // 提交次数
TimeLimit:    1000         // 1秒（如果未传）
MemoryLimit:  256          // 256MB（如果未传）
Tags:         []           // 空数组
```

### 4. 字段验证规则
```go
Title:        required, min=1, max=200
Description:  required, min=10
Difficulty:   required, oneof=easy medium hard
TimeLimit:    min=100, max=10000    // 0.1s-10s
MemoryLimit:  min=32, max=1024      // 32MB-1GB
Tags:         max=10                // 最多10个标签
Status:       omitempty, oneof=draft published archived
IsPublic:     boolean
```

---

## 🎯 技术实现方案

### 核心变更点

1. **新增Request/Response模型**
   - `CreateProblemRequest` - 创建题目请求
   - `ProblemResponse` - 题目响应（可选，复用Problem）

2. **重构Handler层**
   - 使用专用Request结构接收参数
   - 完整的字段验证
   - 状态规则校验
   - 细化错误处理

3. **优化Service层**
   - Status默认值改为Draft
   - 实现IsPublic与Status关联规则
   - 完善业务逻辑验证
   - 添加详细日志

4. **保持兼容性**
   - 现有数据库数据不受影响
   - 新接口向后兼容

---

## 📋 技术约束

1. **框架约束**
   - 使用Gin的binding验证
   - 保持Swagger注释规范
   - 遵循现有错误处理模式

2. **数据库约束**
   - MongoDB字段名保持不变
   - 支持现有数据格式

3. **业务约束**
   - 权限：Admin和Teacher可创建
   - 草稿状态强制私有
   - 创建者自动设置

---

## ✅ 验收标准

### 功能验收
- [x] 创建题目默认状态为Draft
- [x] 创建题目默认IsPublic为false
- [x] Draft状态强制IsPublic=false
- [x] Published状态允许设置IsPublic
- [x] 所有字段验证正确
- [x] 默认值自动设置

### 代码质量
- [x] 遵循项目代码规范
- [x] Swagger文档完整
- [x] 错误处理规范
- [x] 关键逻辑有日志

### 测试用例
- [x] 正常创建草稿题目
- [x] 字段验证测试
- [x] 状态规则测试
- [x] 权限测试

---

## 🔄 下一步

进入 **DESIGN阶段**，详细设计实现方案。

