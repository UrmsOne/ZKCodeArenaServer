# TODO - 题目列表用户状态功能部署清单

## 🚀 待办事项

### 1. 数据库索引创建 (必须)
**优先级**: 🔴 高  
**操作**: 在生产环境创建查询优化索引

#### Docker环境部署
```bash
# 执行索引创建脚本
docker exec <mongo-container-name> mongo zk_code_arena /scripts/add_user_status_index.js
```

#### 本地环境部署  
```bash
# 进入项目根目录执行
mongo zk_code_arena scripts/add_user_status_index.js
```

#### 验证索引创建成功
```bash
# 连接到MongoDB
mongo zk_code_arena

# 查看submits集合的索引
db.submits.getIndexes()

# 应该能看到名为 "idx_user_problem_status" 的索引
```

---

### 2. 应用重新部署 (必须)
**优先级**: 🔴 高  
**操作**: 使用最新代码重新构建和部署应用

#### Docker部署流程
```bash
# 1. 拉取最新代码
git pull origin <branch-name>

# 2. 停止现有服务
docker-compose down

# 3. 重新构建镜像（包含最新代码）
docker-compose -f docker-compose.dev.yml build --no-cache

# 4. 启动服务
docker-compose -f docker-compose.dev.yml up -d
```

---

### 3. 功能验证测试 (必须)
**优先级**: 🔴 高  
**操作**: 验证新功能正常工作

#### 测试清单
- [ ] **登录用户测试**: 访问题目列表，确认返回数据包含`user_status`字段
- [ ] **匿名用户测试**: 未登录访问题目列表，确认不受影响
- [ ] **搜索功能测试**: 搜索题目时也应包含用户状态
- [ ] **性能测试**: 对比部署前后的响应时间

#### 测试API
```bash
# 登录用户访问题目列表 (需要先登录获取token)
curl -H "Authorization: Bearer <token>" \
     "http://your-domain/api/problems?page=1&page_size=10"

# 匿名用户访问  
curl "http://your-domain/api/problems?page=1&page_size=10"

# 搜索功能
curl -H "Authorization: Bearer <token>" \
     "http://your-domain/api/problems/search?keyword=algorithm&page=1&page_size=10"
```

---

### 4. 监控配置 (推荐)
**优先级**: 🟡 中  
**操作**: 设置性能监控和告警

#### 监控指标
- 题目列表API响应时间
- 数据库查询执行时间  
- 内存使用情况
- 用户状态查询成功率

#### 告警设置
- API响应时间 > 500ms
- 数据库查询超时
- 内存使用率 > 80%

---

### 5. 备份恢复测试 (推荐)
**优先级**: 🟡 中  
**操作**: 确认数据备份包含新索引

#### 操作步骤
```bash
# 1. 创建数据库备份
mongodump --db zk_code_arena --out /backup/$(date +%Y%m%d)

# 2. 验证备份包含索引信息
mongorestore --db test_db /backup/$(date +%Y%m%d)/zk_code_arena --dryRun

# 3. 恢复测试（建议在测试环境）
mongorestore --db test_db /backup/$(date +%Y%m%d)/zk_code_arena
```

---

## 🔧 故障排除

### 常见问题及解决方案

#### 1. 索引创建失败
**症状**: 执行索引脚本报错  
**解决**: 
- 检查MongoDB连接
- 确认数据库名称正确
- 检查是否有足够权限

#### 2. API返回格式异常
**症状**: `user_status`字段缺失或格式错误  
**解决**:
- 检查用户是否已登录
- 确认JWT token有效
- 查看应用日志排查错误

#### 3. 性能下降明显
**症状**: 题目列表加载变慢  
**解决**:
- 确认索引创建成功：`db.submits.getIndexes()`
- 检查查询执行计划：`db.submits.find().explain()`
- 监控数据库CPU和内存使用

#### 4. Docker构建失败
**症状**: 镜像构建过程中出错  
**解决**:
- 清理Docker缓存：`docker system prune`
- 检查Dockerfile语法
- 确认网络连接正常

---

## 📞 技术支持

### 紧急联系
- **功能异常**: 立即回滚到上一版本
- **数据异常**: 停止服务，检查数据一致性
- **性能问题**: 监控系统资源，必要时扩容

### 回滚方案
```bash
# 1. 快速回滚代码
git checkout <previous-commit-hash>

# 2. 重新构建部署
docker-compose -f docker-compose.dev.yml build --no-cache
docker-compose -f docker-compose.dev.yml up -d

# 3. 删除新创建的索引（可选）
mongo zk_code_arena --eval 'db.submits.dropIndex("idx_user_problem_status")'
```

---

## ✅ 完成确认

部署完成后，请确认以下事项：

- [ ] 数据库索引创建成功
- [ ] 应用重新部署完成
- [ ] 登录用户能看到题目状态
- [ ] 匿名用户访问正常
- [ ] API响应时间在可接受范围
- [ ] 无错误日志出现
- [ ] Swagger文档更新正确

**部署负责人**: _______________  
**完成时间**: _______________  
**确认签名**: _______________  

---

**文档版本**: v1.0  
**创建时间**: 2025-10-25  
**适用环境**: 生产环境、测试环境

