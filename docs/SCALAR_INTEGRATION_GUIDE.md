# Scalar API 文档集成指南

## 概述

本项目已集成 Scalar，这是一个现代化的 API 文档工具，提供比传统 Swagger UI 更好的用户体验。

## 功能特性

### 🎨 现代化界面
- **响应式设计**：完美适配桌面和移动设备
- **多主题支持**：支持亮色、暗色和自动主题
- **现代布局**：提供 Modern 和 Classic 两种布局风格

### 🚀 增强功能
- **实时搜索**：支持快捷键 `K` 快速搜索
- **交互式测试**：内置 API 测试功能
- **代码示例**：自动生成多语言代码示例
- **认证支持**：完整的 API 认证流程支持

### ⚡ 性能优化
- **CDN 加载**：使用 CDN 加速资源加载
- **缓存优化**：智能缓存策略
- **懒加载**：按需加载内容

## 访问地址

### 主要访问地址
- **Scalar 文档**：`http://localhost:8080/docs`
- **传统 Swagger**：`http://localhost:8080/swagger/index.html`（保留兼容性）

### 主题变体
- **亮色主题**：`http://localhost:8080/docs/light`
- **暗色主题**：`http://localhost:8080/docs/dark`
- **经典布局**：`http://localhost:8080/docs/classic`

## 配置说明

### 配置文件位置
- 主配置：`conf/config.yaml`
- Scalar 专用配置：`conf/scalar.yaml`

### 主要配置项

```yaml
Scalar:
  Title: "ZK Code Arena API"                    # 文档标题
  Description: "在线编程竞赛平台 API 文档"        # 文档描述
  SwaggerUrl: "/swagger/doc.json"               # Swagger JSON 路径
  Theme: "auto"                                 # 主题：light, dark, auto
  Layout: "modern"                              # 布局：modern, classic
  ShowSidebar: true                             # 显示侧边栏
  HideDownloadButton: false                     # 隐藏下载按钮
  CustomCss: |                                  # 自定义样式
    .scalar-app {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    }
```

## 自定义样式

### 品牌定制
```css
.scalar-app .sidebar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.scalar-app .method-get {
  background: #10b981;
}

.scalar-app .method-post {
  background: #3b82f6;
}

.scalar-app .method-put {
  background: #f59e0b;
}

.scalar-app .method-delete {
  background: #ef4444;
}
```

### 字体定制
```css
.scalar-app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}
```

## 开发指南

### 添加新的主题
1. 在 `scalar_config.go` 中添加新的路由
2. 创建对应的配置
3. 更新模板文件

### 自定义配置加载
```go
// 从配置文件加载
config := &ScalarConfig{
    Title:       conf.Config.Scalar.Title,
    Description: conf.Config.Scalar.Description,
    Theme:       conf.Config.Scalar.Theme,
    // ... 其他配置
}
```

### 添加中间件
```go
// 添加缓存中间件
r.GET("/docs", ScalarMiddleware(), func(c *gin.Context) {
    s.renderScalar(c, config)
})
```

## 与 Swagger 的对比

| 特性 | Scalar | Swagger UI |
|------|--------|------------|
| 界面设计 | 现代化、响应式 | 传统设计 |
| 性能 | 更快的加载速度 | 较慢 |
| 搜索功能 | 实时搜索 | 基础搜索 |
| 主题支持 | 多主题 | 有限 |
| 移动端 | 完美适配 | 基础支持 |
| 代码示例 | 多语言自动生成 | 基础示例 |

## 最佳实践

### 1. 文档编写
- 使用清晰的 API 描述
- 提供完整的参数说明
- 添加示例请求和响应

### 2. 配置优化
- 根据团队偏好选择主题
- 自定义品牌样式
- 配置合适的服务器地址

### 3. 性能优化
- 启用 CDN 加速
- 配置适当的缓存策略
- 使用压缩传输

## 故障排除

### 常见问题

1. **文档无法加载**
   - 检查 Swagger JSON 路径是否正确
   - 确认服务器正常运行

2. **样式不生效**
   - 检查 CustomCSS 配置
   - 确认 CSS 语法正确

3. **主题切换失败**
   - 检查主题配置值
   - 清除浏览器缓存

### 调试模式
```go
// 启用详细日志
s.lg.Debugf("Scalar config: %+v", config)
```

## 升级指南

### 从 Swagger UI 迁移
1. 保留原有 Swagger 路由（兼容性）
2. 添加 Scalar 路由
3. 逐步迁移用户到新界面
4. 收集用户反馈
5. 最终移除旧界面（可选）

### 版本更新
- 定期更新 Scalar CDN 版本
- 测试新功能兼容性
- 更新配置文档

## 技术支持

- **Scalar 官方文档**：https://github.com/scalar/scalar
- **项目 Issues**：提交到项目 GitHub 仓库
- **社区支持**：Scalar 社区论坛
