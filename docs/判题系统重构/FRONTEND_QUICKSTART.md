# 前端开发快速开始指南

> **文档版本**: v1.0  
> **更新日期**: 2025-10-05  
> **预计阅读时间**: 5 分钟

---

## 🎯 目标

本指南帮助前端开发人员在 **5 分钟内** 完成环境配置，开始开发测试用例管理功能。

---

## 📦 资源清单

以下文件已为您准备好：

| 文件名 | 说明 | 用途 |
|--------|------|------|
| `FRONTEND_API_SPEC.md` | API 接口规范 | 查看接口定义和示例 |
| `MOCK_CONFIG.json` | Mock 数据配置 | 导入到 Mock 工具 |
| `FRONTEND_MOCK_GUIDE.md` | Mock 使用指南 | 配置 Mock 环境 |
| `types.ts` | TypeScript 类型定义 | 获得类型提示 |
| `API_CHANGES_前端开发.md` | API 变更说明 | 了解数据模型变更 |

---

## 🚀 5 分钟快速开始

### 方案 A：使用 Apifox（推荐）

**适合场景**：团队协作、需要接口文档管理

#### 步骤 1：导入 Mock 配置（1 分钟）

1. 打开 Apifox，创建新项目
2. 点击「导入」→「导入数据」→ 选择 `MOCK_CONFIG.json`
3. 点击「确认导入」

#### 步骤 2：启动 Mock 服务（30 秒）

1. 点击「Mock」标签
2. 点击「启动 Mock 服务」
3. 复制 Mock 地址（例如：`http://127.0.0.1:4523/m1/xxxxx`）

#### 步骤 3：配置前端项目（1 分钟）

```javascript
// src/api/index.js
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://127.0.0.1:4523/m1/xxxxx/api/v1', // 替换为你的 Mock 地址
  timeout: 5000,
});

// 自动携带 Token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;
```

#### 步骤 4：开始开发（2 分钟）

```javascript
// src/api/testcase.js
import api from './index';

// 获取测试用例列表
export const getTestCases = (problemId) => {
  return api.get(`/testcase/problem/${problemId}`);
};

// 创建测试用例
export const createTestCase = (data) => {
  return api.post('/testcase/', data);
};

// 更新测试用例
export const updateTestCase = (id, data) => {
  return api.put(`/testcase/${id}`, data);
};

// 删除测试用例
export const deleteTestCase = (id) => {
  return api.delete(`/testcase/${id}`);
};
```

✅ **完成！** 现在你可以开始开发了，所有接口都会返回 Mock 数据。

---

### 方案 B：使用 Mock.js（本地开发）

**适合场景**：个人开发、不依赖外部服务

#### 步骤 1：安装依赖（30 秒）

```bash
npm install mockjs --save-dev
```

#### 步骤 2：创建 Mock 文件（2 分钟）

复制以下代码到 `src/mock/index.js`：

```javascript
import Mock from 'mockjs';

Mock.setup({ timeout: '200-600' });

const baseURL = '/api/v1';

// 获取测试用例列表
Mock.mock(new RegExp(`${baseURL}/testcase/problem/.*`), 'get', {
  code: 200,
  message: 'success',
  data: {
    'test_cases|3-5': [{
      'id': '@id',
      'problem_id': '@id',
      'input': '1 2\n',
      'output': '3\n',
      'is_sample': '@boolean',
      'time_limit|1000-5000': 1000,
      'memory_limit|256-1024': 256,
      'score|0-20': 10,
      'created_at': '@datetime'
    }],
    'total|3-5': 3
  }
});

// 创建测试用例
Mock.mock(`${baseURL}/testcase/`, 'post', (options) => {
  const body = JSON.parse(options.body);
  return {
    code: 200,
    message: 'success',
    data: {
      id: Mock.Random.id(),
      ...body,
      created_at: new Date().toISOString()
    }
  };
});

// 更新测试用例
Mock.mock(new RegExp(`${baseURL}/testcase/[^/]+$`), 'put', (options) => {
  const body = JSON.parse(options.body);
  return {
    code: 200,
    message: 'success',
    data: {
      id: Mock.Random.id(),
      problem_id: Mock.Random.id(),
      ...body,
      created_at: new Date().toISOString()
    }
  };
});

// 删除测试用例
Mock.mock(new RegExp(`${baseURL}/testcase/[^/]+$`), 'delete', {
  code: 200,
  message: 'success',
  data: { message: '测试用例删除成功' }
});

console.log('✅ Mock 数据已加载');
```

#### 步骤 3：引入 Mock（30 秒）

```javascript
// main.js
import { createApp } from 'vue';
import App from './App.vue';

// 开发环境启用 Mock
if (process.env.NODE_ENV === 'development') {
  require('./mock');
}

createApp(App).mount('#app');
```

#### 步骤 4：配置 API（1 分钟）

```javascript
// src/api/index.js
import axios from 'axios';

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 5000,
});

export default api;
```

✅ **完成！** Mock.js 会自动拦截请求并返回 Mock 数据。

---

## 📘 TypeScript 支持

### 步骤 1：复制类型定义文件

将 `types.ts` 复制到你的项目中（例如：`src/types/api.ts`）

### 步骤 2：使用类型

```typescript
// src/api/testcase.ts
import api from './index';
import type { 
  ApiResponse, 
  TestCaseListResponse, 
  TestCase,
  CreateTestCaseRequest,
  UpdateTestCaseRequest 
} from '@/types/api';

// 获取测试用例列表（带类型提示）
export const getTestCases = (problemId: string): Promise<ApiResponse<TestCaseListResponse>> => {
  return api.get(`/testcase/problem/${problemId}`);
};

// 创建测试用例（带类型检查）
export const createTestCase = (data: CreateTestCaseRequest): Promise<ApiResponse<TestCase>> => {
  return api.post('/testcase/', data);
};

// 更新测试用例
export const updateTestCase = (
  id: string, 
  data: UpdateTestCaseRequest
): Promise<ApiResponse<TestCase>> => {
  return api.put(`/testcase/${id}`, data);
};

// 删除测试用例
export const deleteTestCase = (id: string): Promise<ApiResponse<{ message: string }>> => {
  return api.delete(`/testcase/${id}`);
};
```

### 步骤 3：在组件中使用

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getTestCases, createTestCase } from '@/api/testcase';
import type { TestCase, CreateTestCaseRequest } from '@/types/api';

const testCases = ref<TestCase[]>([]);
const loading = ref(false);

// 获取测试用例列表
const fetchTestCases = async (problemId: string) => {
  loading.value = true;
  try {
    const { data } = await getTestCases(problemId);
    testCases.value = data.test_cases;
  } catch (error) {
    console.error('获取测试用例失败:', error);
  } finally {
    loading.value = false;
  }
};

// 创建测试用例
const handleCreate = async () => {
  const newTestCase: CreateTestCaseRequest = {
    problem_id: '507f1f77bcf86cd799439012',
    input: '1 2\n',
    output: '3\n',
    is_sample: true,
    time_limit: 2000,
    memory_limit: 512,
    score: 10
  };
  
  try {
    const { data } = await createTestCase(newTestCase);
    testCases.value.push(data);
  } catch (error) {
    console.error('创建测试用例失败:', error);
  }
};

onMounted(() => {
  fetchTestCases('507f1f77bcf86cd799439012');
});
</script>
```

---

## 🎨 UI 组件示例

### 测试用例列表组件

```vue
<template>
  <div class="testcase-list">
    <div class="header">
      <h2>测试用例管理</h2>
      <button @click="showCreateDialog = true">新增测试用例</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else class="testcase-items">
      <div 
        v-for="testCase in testCases" 
        :key="testCase.id" 
        class="testcase-item"
      >
        <div class="testcase-header">
          <span class="badge" :class="{ 'sample': testCase.is_sample }">
            {{ testCase.is_sample ? '示例' : '隐藏' }}
          </span>
          <div class="actions">
            <button @click="handleEdit(testCase)">编辑</button>
            <button @click="handleDelete(testCase.id)" class="danger">删除</button>
          </div>
        </div>

        <div class="testcase-content">
          <div class="field">
            <label>输入:</label>
            <pre>{{ testCase.input }}</pre>
          </div>
          <div class="field">
            <label>输出:</label>
            <pre>{{ testCase.output }}</pre>
          </div>
        </div>

        <div class="testcase-meta">
          <span v-if="testCase.time_limit">
            时间限制: {{ testCase.time_limit }}ms
          </span>
          <span v-if="testCase.memory_limit">
            内存限制: {{ testCase.memory_limit }}MB
          </span>
          <span v-if="testCase.score">
            分数: {{ testCase.score }}
          </span>
        </div>
      </div>
    </div>

    <!-- 创建/编辑对话框 -->
    <TestCaseDialog
      v-model="showCreateDialog"
      :testCase="editingTestCase"
      @submit="handleSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getTestCases, deleteTestCase } from '@/api/testcase';
import type { TestCase } from '@/types/api';
import TestCaseDialog from './TestCaseDialog.vue';

const props = defineProps<{
  problemId: string;
}>();

const testCases = ref<TestCase[]>([]);
const loading = ref(false);
const showCreateDialog = ref(false);
const editingTestCase = ref<TestCase | null>(null);

const fetchTestCases = async () => {
  loading.value = true;
  try {
    const { data } = await getTestCases(props.problemId);
    testCases.value = data.test_cases;
  } catch (error) {
    console.error('获取测试用例失败:', error);
  } finally {
    loading.value = false;
  }
};

const handleEdit = (testCase: TestCase) => {
  editingTestCase.value = testCase;
  showCreateDialog.value = true;
};

const handleDelete = async (id: string) => {
  if (!confirm('确定要删除这个测试用例吗？')) return;
  
  try {
    await deleteTestCase(id);
    testCases.value = testCases.value.filter(tc => tc.id !== id);
  } catch (error) {
    console.error('删除测试用例失败:', error);
  }
};

const handleSubmit = () => {
  showCreateDialog.value = false;
  editingTestCase.value = null;
  fetchTestCases();
};

onMounted(() => {
  fetchTestCases();
});
</script>

<style scoped>
.testcase-list {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.testcase-item {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.badge.sample {
  background: #e6f7ff;
  color: #1890ff;
}

.testcase-content pre {
  background: #f5f5f5;
  padding: 8px;
  border-radius: 4px;
  margin-top: 4px;
}

.testcase-meta {
  display: flex;
  gap: 16px;
  margin-top: 12px;
  font-size: 12px;
  color: #666;
}
</style>
```

---

## 📚 下一步

1. **阅读完整文档**：
   - `FRONTEND_API_SPEC.md` - 查看所有接口定义
   - `API_CHANGES_前端开发.md` - 了解数据模型变更

2. **选择 Mock 方案**：
   - 查看 `FRONTEND_MOCK_GUIDE.md` 了解更多 Mock 工具配置

3. **开始开发**：
   - 测试用例管理页面
   - 判题结果展示页面
   - 题目管理页面

---

## ❓ 常见问题

### Q: Mock 数据不生效？

A: 检查以下几点：
1. 确认 Mock 文件已正确引入
2. 检查 URL 是否匹配
3. 查看浏览器控制台是否有错误

### Q: 如何切换真实接口？

A: 修改 `baseURL` 配置：
```javascript
// 开发环境使用 Mock
const baseURL = process.env.NODE_ENV === 'development' 
  ? 'http://127.0.0.1:4523/m1/xxxxx/api/v1'  // Mock 地址
  : 'http://localhost:8080/api/v1';          // 真实接口
```

### Q: TypeScript 类型报错？

A: 确保：
1. `types.ts` 文件已正确导入
2. `tsconfig.json` 配置正确
3. IDE 已重启（刷新类型缓存）

---

## 📞 技术支持

如有疑问，请联系：
- **项目负责人**: omenkk7
- **更新日期**: 2025-10-05

---

**祝开发顺利！🎉**
