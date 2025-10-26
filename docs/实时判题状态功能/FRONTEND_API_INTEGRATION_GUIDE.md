# 实时判题状态功能 - 前端对接指南

## 📋 功能概述

本功能提供了两种方式来获取提交判题的实时状态：

1. **轻量级状态查询 API** - 适合轮询场景
2. **WebSocket 实时推送** - 适合实时性要求高的场景

---

## 🔗 1. 轻量级状态查询 API

### 接口信息

```http
GET /api/v1/submit/{id}/status
```

### 请求参数

| 参数 | 类型   | 必填 | 说明   |
|------|--------|------|--------|
| id   | string | 是   | 提交ID |

### 请求头

```http
Authorization: Bearer {token}
```

### 响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "67124f5a8b8a4c123d456789",
    "status": "running",
    "progress": {
      "current_test_case": 3,
      "total_test_cases": 10,
      "percentage": 30
    },
    "message": "正在判题...",
    "updated_at": "2024-10-26T10:30:45Z",
    "time_used": null,
    "memory_used": null
  }
}
```

### 响应字段说明

| 字段          | 类型     | 说明                    |
|---------------|----------|-------------------------|
| id            | string   | 提交ID                  |
| status        | string   | 当前状态（见状态列表）    |
| progress      | object   | 判题进度（仅running状态） |
| message       | string   | 状态描述信息            |
| updated_at    | string   | 最后更新时间            |
| time_used     | int      | 时间消耗(ms)，完成后才有 |
| memory_used   | int      | 内存消耗(KB)，完成后才有 |

### 状态列表

| 状态值          | 说明       | 是否最终状态 |
|----------------|------------|-------------|
| pending        | 等待判题    | 否          |
| running        | 正在判题    | 否          |
| accepted       | 通过       | 是          |
| wrong_answer   | 答案错误    | 是          |
| time_limit     | 时间超限    | 是          |
| memory_limit   | 内存超限    | 是          |
| runtime_error  | 运行时错误  | 是          |
| compile_error  | 编译错误    | 是          |
| system_error   | 系统错误    | 是          |

---

## 📡 2. WebSocket 实时推送

### 连接信息

```http
WebSocket: ws://localhost:8080/api/v1/ws/submit/{id}?token={jwt_token}
```

### 连接参数

| 参数  | 类型   | 必填 | 说明              |
|-------|--------|------|-------------------|
| id    | string | 是   | 提交ID            |
| token | string | 是   | JWT访问令牌       |

### 消息格式

WebSocket 接收到的消息格式：

```json
{
  "type": "status_update",
  "submit_id": "67124f5a8b8a4c123d456789",
  "timestamp": "2024-10-26T10:30:45Z",
  "data": {
    "status": "running",
    "progress": {
      "current_test_case": 5,
      "total_test_cases": 10,
      "percentage": 50
    },
    "message": "正在判题...",
    "result": null
  }
}
```

### 最终结果消息

```json
{
  "type": "status_update",
  "submit_id": "67124f5a8b8a4c123d456789",
  "timestamp": "2024-10-26T10:31:20Z",
  "data": {
    "status": "accepted",
    "progress": null,
    "message": "通过",
    "result": {
      "status": "accepted",
      "time_used": 150,
      "memory_used": 1024,
      "passed_cases": 10,
      "total_cases": 10,
      "compile_error": ""
    }
  }
}
```

---

## 💻 前端实现示例

### 1. 轻量级 API 轮询实现

```javascript
class SubmitStatusPoller {
  constructor(submitId, token) {
    this.submitId = submitId;
    this.token = token;
    this.isPolling = false;
    this.pollInterval = null;
  }

  async startPolling(onUpdate, interval = 2000) {
    this.isPolling = true;
    
    const poll = async () => {
      if (!this.isPolling) return;
      
      try {
        const response = await fetch(`/api/v1/submit/${this.submitId}/status`, {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        });
        
        const result = await response.json();
        
        if (result.code === 200) {
          onUpdate(result.data);
          
          // 如果是最终状态，停止轮询
          if (this.isFinalStatus(result.data.status)) {
            this.stopPolling();
          }
        }
      } catch (error) {
        console.error('获取状态失败:', error);
        onUpdate({ error: '获取状态失败' });
      }
    };

    // 立即执行一次
    await poll();
    
    // 设置定时器
    if (this.isPolling) {
      this.pollInterval = setInterval(poll, interval);
    }
  }

  stopPolling() {
    this.isPolling = false;
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
      this.pollInterval = null;
    }
  }

  isFinalStatus(status) {
    const finalStatuses = [
      'accepted', 'wrong_answer', 'time_limit', 
      'memory_limit', 'runtime_error', 'compile_error', 'system_error'
    ];
    return finalStatuses.includes(status);
  }
}

// 使用示例
const poller = new SubmitStatusPoller(submitId, token);
poller.startPolling((status) => {
  console.log('状态更新:', status);
  updateUI(status);
});
```

### 2. WebSocket 实时推送实现

```javascript
class SubmitStatusWebSocket {
  constructor(submitId, token) {
    this.submitId = submitId;
    this.token = token;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
  }

  connect(onUpdate, onError) {
    const wsUrl = `ws://localhost:8080/api/v1/ws/submit/${this.submitId}?token=${this.token}`;
    
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket连接已建立');
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.type === 'status_update') {
          onUpdate(message.data);
        }
      } catch (error) {
        console.error('解析WebSocket消息失败:', error);
      }
    };

    this.ws.onclose = (event) => {
      console.log('WebSocket连接已关闭:', event.code, event.reason);
      this.attemptReconnect(onUpdate, onError);
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket错误:', error);
      if (onError) onError(error);
    };
  }

  attemptReconnect(onUpdate, onError) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      
      setTimeout(() => {
        this.connect(onUpdate, onError);
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      console.error('WebSocket重连失败，超出最大重试次数');
      if (onError) onError(new Error('连接失败'));
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// 使用示例
const wsClient = new SubmitStatusWebSocket(submitId, token);
wsClient.connect(
  (status) => {
    console.log('状态更新:', status);
    updateUI(status);
  },
  (error) => {
    console.error('WebSocket错误:', error);
    // 可以回退到轮询方式
    fallbackToPolling();
  }
);
```

### 3. React Hook 示例

```jsx
import { useState, useEffect, useRef } from 'react';

// 轻量级轮询 Hook
function useSubmitStatus(submitId, token, method = 'websocket') {
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  const wsRef = useRef(null);
  const pollerRef = useRef(null);

  useEffect(() => {
    if (!submitId || !token) return;

    const handleUpdate = (statusData) => {
      setStatus(statusData);
      setLoading(false);
      setError(null);
    };

    const handleError = (err) => {
      setError(err);
      setLoading(false);
    };

    if (method === 'websocket') {
      // 使用WebSocket
      wsRef.current = new SubmitStatusWebSocket(submitId, token);
      wsRef.current.connect(handleUpdate, handleError);
    } else {
      // 使用轮询
      pollerRef.current = new SubmitStatusPoller(submitId, token);
      pollerRef.current.startPolling(handleUpdate);
    }

    return () => {
      if (wsRef.current) {
        wsRef.current.disconnect();
      }
      if (pollerRef.current) {
        pollerRef.current.stopPolling();
      }
    };
  }, [submitId, token, method]);

  return { status, loading, error };
}

// 使用示例组件
function SubmitStatusDisplay({ submitId, token }) {
  const { status, loading, error } = useSubmitStatus(submitId, token, 'websocket');

  if (loading) return <div>加载中...</div>;
  if (error) return <div>错误: {error.message}</div>;

  return (
    <div className="submit-status">
      <div className="status-info">
        <span className={`status-badge ${status.status}`}>
          {status.message}
        </span>
        
        {status.progress && (
          <div className="progress-info">
            <div className="progress-bar">
              <div 
                className="progress-fill" 
                style={{ width: `${status.progress.percentage}%` }}
              />
            </div>
            <span>
              {status.progress.current_test_case}/{status.progress.total_test_cases}
            </span>
          </div>
        )}
        
        {(status.time_used || status.memory_used) && (
          <div className="result-info">
            {status.time_used && <span>时间: {status.time_used}ms</span>}
            {status.memory_used && <span>内存: {status.memory_used}KB</span>}
          </div>
        )}
      </div>
    </div>
  );
}
```

---

## 🎯 最佳实践

### 1. 选择合适的方法

- **WebSocket**: 适合实时性要求高、用户活跃的场景
- **轻量级 API**: 适合简单场景、网络不稳定的环境

### 2. 错误处理

```javascript
// WebSocket 降级策略
function createSubmitMonitor(submitId, token) {
  let currentMethod = 'websocket';
  let monitor = null;

  const startMonitoring = (onUpdate) => {
    if (currentMethod === 'websocket') {
      monitor = new SubmitStatusWebSocket(submitId, token);
      monitor.connect(onUpdate, (error) => {
        console.log('WebSocket失败，降级到轮询');
        currentMethod = 'polling';
        startMonitoring(onUpdate);
      });
    } else {
      monitor = new SubmitStatusPoller(submitId, token);
      monitor.startPolling(onUpdate);
    }
  };

  return { startMonitoring };
}
```

### 3. 性能优化

- **节流更新**: 避免过于频繁的 UI 更新
- **内存管理**: 及时清理监听器和定时器
- **网络优化**: WebSocket 断线重连策略

### 4. 用户体验

- **状态可视化**: 使用进度条、动画效果
- **错误提示**: 友好的错误信息显示
- **加载状态**: 明确的加载指示器

---

## 🔧 调试工具

### 1. 状态监控

```javascript
// 调试工具
const debugSubmitStatus = {
  logAllUpdates: (submitId, token) => {
    const ws = new SubmitStatusWebSocket(submitId, token);
    ws.connect((status) => {
      console.log(`[${new Date().toISOString()}] 状态更新:`, status);
    });
  },
  
  compareWebSocketVsPolling: (submitId, token) => {
    const wsUpdates = [];
    const pollUpdates = [];
    
    // WebSocket 监控
    const ws = new SubmitStatusWebSocket(submitId, token);
    ws.connect((status) => {
      wsUpdates.push({ time: Date.now(), status });
    });
    
    // 轮询监控
    const poller = new SubmitStatusPoller(submitId, token);
    poller.startPolling((status) => {
      pollUpdates.push({ time: Date.now(), status });
    }, 1000);
    
    // 5分钟后输出对比结果
    setTimeout(() => {
      console.log('WebSocket 更新次数:', wsUpdates.length);
      console.log('轮询更新次数:', pollUpdates.length);
      console.log('WebSocket 更新:', wsUpdates);
      console.log('轮询更新:', pollUpdates);
    }, 5 * 60 * 1000);
  }
};
```

### 2. 网络状态检测

```javascript
// 网络状态监控
function createNetworkAwareMonitor(submitId, token) {
  let isOnline = navigator.onLine;
  let currentMonitor = null;

  const updateMethod = () => {
    if (currentMonitor) {
      currentMonitor.disconnect?.() || currentMonitor.stopPolling?.();
    }

    if (isOnline) {
      currentMonitor = new SubmitStatusWebSocket(submitId, token);
    } else {
      currentMonitor = new SubmitStatusPoller(submitId, token);
    }
  };

  window.addEventListener('online', () => {
    isOnline = true;
    updateMethod();
  });

  window.addEventListener('offline', () => {
    isOnline = false;
    updateMethod();
  });

  return { updateMethod };
}
```

---

## ❓ 常见问题

### 1. WebSocket 连接失败

**问题**: WebSocket 连接建立失败
**解决**: 
- 检查 token 是否有效
- 确认服务器支持 WebSocket
- 检查网络代理设置

### 2. 轮询频率设置

**问题**: 轮询太频繁或太慢
**建议**:
- pending/running 状态：1-2秒
- 其他状态：可以停止轮询
- 根据网络情况调整

### 3. 内存泄漏

**问题**: 长时间运行后内存增长
**预防**:
- 及时清理定时器
- 关闭 WebSocket 连接
- 移除事件监听器

### 4. 跨域问题

**问题**: WebSocket 跨域连接失败
**解决**:
- 配置服务器 CORS
- 使用正确的协议 (ws/wss)
- 检查防火墙设置

---

## 📞 技术支持

如有问题，请联系开发团队或查看：
- API 文档: `/swagger/index.html`
- 错误日志: 浏览器开发者工具
- 服务器日志: 查看服务端错误信息
