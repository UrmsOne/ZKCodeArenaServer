# 实时判题状态功能 - 快速开始指南

## 🚀 5分钟快速集成

### 第一步：了解接口

```bash
# 轻量级状态查询（轮询）
GET /api/v1/submit/{id}/status

# WebSocket实时推送
WebSocket: ws://localhost:8080/api/v1/ws/submit/{id}?token={jwt_token}
```

### 第二步：选择实现方式

#### 方案A：简单轮询（推荐新手）

```javascript
// 最简单的实现
async function monitorSubmitStatus(submitId, token) {
  const poll = async () => {
    const response = await fetch(`/api/v1/submit/${submitId}/status`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    const result = await response.json();
    const status = result.data;
    
    console.log('状态:', status.status, status.message);
    
    // 显示进度
    if (status.progress) {
      console.log(`进度: ${status.progress.percentage}%`);
    }
    
    // 判断是否完成
    const finalStatuses = ['accepted', 'wrong_answer', 'compile_error', 'time_limit', 'memory_limit', 'runtime_error', 'system_error'];
    if (!finalStatuses.includes(status.status)) {
      setTimeout(poll, 2000); // 2秒后再次检查
    } else {
      console.log('判题完成!', status);
    }
  };
  
  poll();
}

// 使用
monitorSubmitStatus('67124f5a8b8a4c123d456789', 'your-jwt-token');
```

#### 方案B：WebSocket实时推送（推荐进阶）

```javascript
// WebSocket实时监控
function connectSubmitWebSocket(submitId, token) {
  const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/submit/${submitId}?token=${token}`);
  
  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === 'status_update') {
      const status = message.data;
      console.log('实时状态:', status.status, status.message);
      
      // 更新UI
      updateStatusUI(status);
    }
  };
  
  ws.onerror = (error) => {
    console.error('WebSocket错误:', error);
    // 降级到轮询
    monitorSubmitStatus(submitId, token);
  };
  
  return ws;
}

// 使用
const ws = connectSubmitWebSocket('67124f5a8b8a4c123d456789', 'your-jwt-token');
```

### 第三步：更新UI

```javascript
function updateStatusUI(status) {
  // 状态显示
  document.getElementById('status-text').textContent = status.message;
  document.getElementById('status-badge').className = `badge ${status.status}`;
  
  // 进度条
  if (status.progress) {
    const progressBar = document.getElementById('progress-bar');
    progressBar.style.width = `${status.progress.percentage}%`;
    document.getElementById('progress-text').textContent = 
      `${status.progress.current_test_case}/${status.progress.total_test_cases}`;
  }
  
  // 结果显示
  if (status.time_used || status.memory_used) {
    document.getElementById('time-used').textContent = `${status.time_used}ms`;
    document.getElementById('memory-used').textContent = `${status.memory_used}KB`;
  }
}
```

## 📱 前端框架集成示例

### React 示例

```jsx
import React, { useState, useEffect } from 'react';

function SubmitMonitor({ submitId, token }) {
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 使用轮询方式
    const poll = async () => {
      try {
        const response = await fetch(`/api/v1/submit/${submitId}/status`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        const result = await response.json();
        
        setStatus(result.data);
        setLoading(false);
        
        // 继续轮询直到完成
        const finalStatuses = ['accepted', 'wrong_answer', 'compile_error', 'time_limit', 'memory_limit', 'runtime_error', 'system_error'];
        if (!finalStatuses.includes(result.data.status)) {
          setTimeout(poll, 2000);
        }
      } catch (error) {
        console.error('获取状态失败:', error);
        setLoading(false);
      }
    };

    poll();
  }, [submitId, token]);

  if (loading) return <div>加载中...</div>;

  return (
    <div className="submit-monitor">
      <div className={`status-badge ${status.status}`}>
        {status.message}
      </div>
      
      {status.progress && (
        <div className="progress">
          <div 
            className="progress-bar"
            style={{ width: `${status.progress.percentage}%` }}
          />
          <span>{status.progress.current_test_case}/{status.progress.total_test_cases}</span>
        </div>
      )}
      
      {(status.time_used || status.memory_used) && (
        <div className="results">
          {status.time_used && <span>时间: {status.time_used}ms</span>}
          {status.memory_used && <span>内存: {status.memory_used}KB</span>}
        </div>
      )}
    </div>
  );
}
```

### Vue 示例

```vue
<template>
  <div class="submit-monitor">
    <div v-if="loading">加载中...</div>
    <div v-else>
      <div :class="['status-badge', status.status]">
        {{ status.message }}
      </div>
      
      <div v-if="status.progress" class="progress">
        <div 
          class="progress-bar"
          :style="{ width: status.progress.percentage + '%' }"
        />
        <span>{{ status.progress.current_test_case }}/{{ status.progress.total_test_cases }}</span>
      </div>
      
      <div v-if="status.time_used || status.memory_used" class="results">
        <span v-if="status.time_used">时间: {{ status.time_used }}ms</span>
        <span v-if="status.memory_used">内存: {{ status.memory_used }}KB</span>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'SubmitMonitor',
  props: ['submitId', 'token'],
  data() {
    return {
      status: null,
      loading: true,
      pollTimer: null
    };
  },
  
  async mounted() {
    this.startPolling();
  },
  
  beforeDestroy() {
    if (this.pollTimer) {
      clearTimeout(this.pollTimer);
    }
  },
  
  methods: {
    async startPolling() {
      try {
        const response = await fetch(`/api/v1/submit/${this.submitId}/status`, {
          headers: { 'Authorization': `Bearer ${this.token}` }
        });
        const result = await response.json();
        
        this.status = result.data;
        this.loading = false;
        
        // 检查是否需要继续轮询
        const finalStatuses = ['accepted', 'wrong_answer', 'compile_error', 'time_limit', 'memory_limit', 'runtime_error', 'system_error'];
        if (!finalStatuses.includes(this.status.status)) {
          this.pollTimer = setTimeout(() => this.startPolling(), 2000);
        }
      } catch (error) {
        console.error('获取状态失败:', error);
        this.loading = false;
      }
    }
  }
};
</script>
```

## 🎨 基础样式

```css
/* 状态徽章 */
.status-badge {
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 500;
  display: inline-block;
}

.status-badge.pending { background: #f0f0f0; color: #666; }
.status-badge.running { background: #e6f3ff; color: #1890ff; }
.status-badge.accepted { background: #e6f7ff; color: #52c41a; }
.status-badge.wrong_answer { background: #fff2e8; color: #fa541c; }
.status-badge.compile_error { background: #fff1f0; color: #f5222d; }
.status-badge.time_limit { background: #fff7e6; color: #fa8c16; }
.status-badge.memory_limit { background: #fff7e6; color: #fa8c16; }
.status-badge.runtime_error { background: #fff1f0; color: #f5222d; }
.status-badge.system_error { background: #f6f6f6; color: #8c8c8c; }

/* 进度条 */
.progress {
  width: 100%;
  height: 8px;
  background: #f0f0f0;
  border-radius: 4px;
  margin: 8px 0;
  position: relative;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #1890ff, #52c41a);
  border-radius: 4px;
  transition: width 0.3s ease;
}

/* 结果信息 */
.results {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #666;
  margin-top: 8px;
}
```

## ⚡ 性能优化技巧

### 1. 智能轮询间隔

```javascript
function getPollingInterval(status) {
  switch (status) {
    case 'pending': return 3000;     // 3秒
    case 'running': return 1000;     // 1秒
    default: return 0;               // 停止轮询
  }
}
```

### 2. 网络错误处理

```javascript
async function fetchWithRetry(url, options, maxRetries = 3) {
  for (let i = 0; i <= maxRetries; i++) {
    try {
      return await fetch(url, options);
    } catch (error) {
      if (i === maxRetries) throw error;
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
    }
  }
}
```

### 3. WebSocket 自动重连

```javascript
class AutoReconnectWebSocket {
  constructor(url) {
    this.url = url;
    this.reconnectDelay = 1000;
    this.maxReconnectDelay = 30000;
    this.connect();
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
      this.reconnectDelay = 1000; // 重置延迟
    };
    
    this.ws.onclose = () => {
      setTimeout(() => {
        this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
        this.connect();
      }, this.reconnectDelay);
    };
  }
}
```

## 🔍 调试工具

### 控制台调试

```javascript
// 在浏览器控制台运行
window.debugSubmit = {
  // 测试轮询
  testPolling: (submitId, token) => {
    console.log('开始测试轮询...');
    monitorSubmitStatus(submitId, token);
  },
  
  // 测试WebSocket
  testWebSocket: (submitId, token) => {
    console.log('开始测试WebSocket...');
    const ws = connectSubmitWebSocket(submitId, token);
    return ws;
  },
  
  // 模拟状态变化
  mockStatusUpdates: () => {
    const statuses = ['pending', 'running', 'accepted'];
    let i = 0;
    
    const interval = setInterval(() => {
      if (i < statuses.length) {
        updateStatusUI({
          status: statuses[i],
          message: `状态 ${i + 1}`,
          progress: i === 1 ? { percentage: 50, current_test_case: 5, total_test_cases: 10 } : null
        });
        i++;
      } else {
        clearInterval(interval);
      }
    }, 2000);
  }
};
```

## 📋 检查清单

### 集成前检查
- [ ] 确认服务器已启动判题功能
- [ ] 获得有效的 JWT token
- [ ] 有一个有效的提交ID用于测试

### 功能测试
- [ ] 轻量级API返回正确的状态格式
- [ ] WebSocket可以正常建立连接
- [ ] 状态更新能够正确触发UI更新
- [ ] 最终状态能够正确显示结果信息

### 错误处理测试
- [ ] 网络断开时的降级处理
- [ ] 无效token的错误提示
- [ ] WebSocket连接失败的处理

### 性能测试
- [ ] 长时间运行无内存泄漏
- [ ] 多个提交同时监控的性能
- [ ] 页面切换时正确清理资源

---

## 🎯 下一步

1. **详细文档**: 阅读完整的 [前端对接指南](./FRONTEND_API_INTEGRATION_GUIDE.md)
2. **API测试**: 使用 Swagger 文档测试接口: `/swagger/index.html`
3. **高级功能**: 了解批量监控、状态持久化等高级特性

---

## ❓ 遇到问题？

- **API问题**: 检查 `/swagger/index.html` 文档
- **WebSocket问题**: 查看浏览器网络控制台
- **权限问题**: 确认token有效性和提交归属

**技术支持**: 如需帮助，请提供错误信息和复现步骤。
