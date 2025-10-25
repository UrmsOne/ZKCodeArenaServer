# 前端 Mock 数据使用指南

> **文档版本**: v1.0  
> **更新日期**: 2025-10-05  
> **目标读者**: 前端开发人员

---

## 📋 目录

- [快速开始](#快速开始)
- [Mock 工具配置](#mock-工具配置)
  - [Apifox 配置](#apifox-配置)
  - [Mock.js 配置](#mockjs-配置)
  - [MSW 配置](#msw-配置)
  - [Axios Mock Adapter 配置](#axios-mock-adapter-配置)
- [TypeScript 类型定义](#typescript-类型定义)
- [常见问题](#常见问题)

---

## 快速开始

### 1. 文件清单

本次提供的前端开发资源包括：

```
docs/判题系统重构/
├── FRONTEND_API_SPEC.md      # API 接口规范文档
├── MOCK_CONFIG.json           # Mock 数据配置（通用）
├── FRONTEND_MOCK_GUIDE.md     # Mock 使用指南（本文档）
└── types.ts                   # TypeScript 类型定义
```

### 2. 推荐工作流

1. **阅读 API 规范**：查看 `FRONTEND_API_SPEC.md` 了解接口定义
2. **选择 Mock 工具**：根据项目技术栈选择合适的 Mock 方案
3. **导入 Mock 配置**：使用 `MOCK_CONFIG.json` 快速配置
4. **引入类型定义**：使用 `types.ts` 获得 TypeScript 支持
5. **开始开发**：无需等待后端接口，直接开发前端功能

---

## Mock 工具配置

### Apifox 配置

**适用场景**：团队协作、接口文档管理、自动化测试

#### 步骤 1：导入 Mock 配置

1. 打开 Apifox，创建新项目或选择现有项目
2. 点击「导入」→「导入数据」
3. 选择「导入格式」为「JSON」
4. 上传 `MOCK_CONFIG.json` 文件
5. 点击「确认导入」

#### 步骤 2：启用 Mock 服务

1. 在 Apifox 项目中，点击「Mock」标签
2. 点击「启动 Mock 服务」
3. 复制 Mock 服务地址（例如：`http://127.0.0.1:4523/m1/xxxxx`）
4. 在前端项目中配置 `baseURL` 为 Mock 服务地址

#### 步骤 3：前端配置

```javascript
// axios 配置示例
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://127.0.0.1:4523/m1/xxxxx/api/v1', // Apifox Mock 地址
  timeout: 5000,
});

// 添加请求拦截器（自动携带 Token）
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;
```

---

### Mock.js 配置

**适用场景**：本地开发、快速原型、不依赖外部服务

#### 步骤 1：安装依赖

```bash
npm install mockjs --save-dev
# 或
yarn add mockjs -D
```

#### 步骤 2：创建 Mock 文件

创建 `src/mock/index.js`：

```javascript
import Mock from 'mockjs';

// 设置延迟时间（模拟网络请求）
Mock.setup({
  timeout: '200-600'
});

// 基础 URL
const baseURL = '/api/v1';

// ========== 测试用例管理 API ==========

// 获取题目的测试用例列表
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

// 获取单个测试用例
Mock.mock(new RegExp(`${baseURL}/testcase/[^/]+$`), 'get', {
  code: 200,
  message: 'success',
  data: {
    'id': '@id',
    'problem_id': '@id',
    'input': '1 2\n',
    'output': '3\n',
    'is_sample': true,
    'time_limit': 2000,
    'memory_limit': 512,
    'score': 10,
    'created_at': '@datetime'
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
  data: {
    message: '测试用例删除成功'
  }
});

// ========== 题目管理 API ==========

// 获取题目列表
Mock.mock(new RegExp(`${baseURL}/problem/`), 'get', {
  code: 200,
  message: 'success',
  data: {
    'problems|10': [{
      'id': '@id',
      'title': '@ctitle(5, 10)',
      'description': '@cparagraph(3, 5)',
      'difficulty|1': ['easy', 'medium', 'hard'],
      'time_limit': 1000,
      'memory_limit': 256,
      'tags|1-3': ['@cword(2, 4)'],
      'is_public': true,
      'created_at': '@datetime'
    }],
    'total': 100,
    'page': 1,
    'page_size': 10
  }
});

// 获取题目详情
Mock.mock(new RegExp(`${baseURL}/problem/[^/]+$`), 'get', {
  code: 200,
  message: 'success',
  data: {
    'id': '@id',
    'title': '两数之和',
    'description': '给定一个整数数组 nums 和一个目标值 target，请你在该数组中找出和为目标值的那两个整数。',
    'difficulty': 'easy',
    'time_limit': 1000,
    'memory_limit': 256,
    'tags': ['数组', '哈希表'],
    'is_public': true,
    'sample_input': '1 2\n',
    'sample_output': '3\n',
    'created_at': '@datetime',
    'updated_at': '@datetime'
  }
});

// ========== 提交管理 API ==========

// 提交代码
Mock.mock(`${baseURL}/submit/`, 'post', {
  code: 200,
  message: 'success',
  data: {
    'id': '@id',
    'problem_id': '@id',
    'user_id': '@id',
    'code': 'public class Main { ... }',
    'language': 'java',
    'status': 'pending',
    'created_at': '@datetime'
  }
});

// 获取提交详情
Mock.mock(new RegExp(`${baseURL}/submit/[^/]+$`), 'get', {
  code: 200,
  message: 'success',
  data: {
    'id': '@id',
    'problem_id': '@id',
    'user_id': '@id',
    'code': 'public class Main { ... }',
    'language': 'java',
    'status|1': ['accepted', 'wrong_answer', 'time_limit_exceeded'],
    'result': {
      'status|1': ['accepted', 'wrong_answer'],
      'time_used|100-200': 150,
      'memory_used|1024-4096': 2048,
      'test_results|2-5': [{
        'test_case_id': '@id',
        'status|1': ['accepted', 'wrong_answer'],
        'time_used|100-200': 120,
        'memory_used|1024-4096': 1536,
        'is_sample': '@boolean',
        'output': '3\n',
        'expected': '3\n',
        'error': ''
      }]
    },
    'created_at': '@datetime',
    'updated_at': '@datetime'
  }
});

// ========== 用户管理 API ==========

// 用户注册
Mock.mock(`${baseURL}/user/register`, 'post', {
  code: 200,
  message: 'success',
  data: {
    'id': '@id',
    'username': '@name',
    'email': '@email',
    'created_at': '@datetime'
  }
});

// 用户登录
Mock.mock(`${baseURL}/user/login`, 'post', {
  code: 200,
  message: 'success',
  data: {
    'token': '@guid',
    'user': {
      'id': '@id',
      'username': '@name',
      'email': '@email'
    }
  }
});

console.log('✅ Mock 数据已加载');
```

#### 步骤 3：在项目中引入

在 `main.js` 或 `App.vue` 中引入：

```javascript
// main.js
import { createApp } from 'vue';
import App from './App.vue';

// 仅在开发环境启用 Mock
if (process.env.NODE_ENV === 'development') {
  require('./mock');
}

createApp(App).mount('#app');
```

---

### MSW (Mock Service Worker) 配置

**适用场景**：现代前端项目、Service Worker 支持、真实网络请求模拟

#### 步骤 1：安装依赖

```bash
npm install msw --save-dev
# 或
yarn add msw -D
```

#### 步骤 2：初始化 MSW

```bash
npx msw init public/ --save
```

#### 步骤 3：创建 Mock 处理器

创建 `src/mocks/handlers.js`：

```javascript
import { rest } from 'msw';

const baseURL = 'http://localhost:8080/api/v1';

export const handlers = [
  // ========== 测试用例管理 API ==========
  
  // 获取题目的测试用例列表
  rest.get(`${baseURL}/testcase/problem/:problemId`, (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        code: 200,
        message: 'success',
        data: {
          test_cases: [
            {
              id: '507f1f77bcf86cd799439011',
              problem_id: req.params.problemId,
              input: '1 2\n',
              output: '3\n',
              is_sample: true,
              time_limit: 2000,
              memory_limit: 512,
              score: 10,
              created_at: new Date().toISOString()
            }
          ],
          total: 1
        }
      })
    );
  }),

  // 创建测试用例
  rest.post(`${baseURL}/testcase/`, async (req, res, ctx) => {
    const body = await req.json();
    return res(
      ctx.status(200),
      ctx.json({
        code: 200,
        message: 'success',
        data: {
          id: '507f1f77bcf86cd799439015',
          ...body,
          created_at: new Date().toISOString()
        }
      })
    );
  }),

  // 更新测试用例
  rest.put(`${baseURL}/testcase/:id`, async (req, res, ctx) => {
    const body = await req.json();
    return res(
      ctx.status(200),
      ctx.json({
        code: 200,
        message: 'success',
        data: {
          id: req.params.id,
          problem_id: '507f1f77bcf86cd799439012',
          ...body,
          created_at: new Date().toISOString()
        }
      })
    );
  }),

  // 删除测试用例
  rest.delete(`${baseURL}/testcase/:id`, (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        code: 200,
        message: 'success',
        data: {
          message: '测试用例删除成功'
        }
      })
    );
  }),

  // ========== 题目管理 API ==========
  
  // 获取题目列表
  rest.get(`${baseURL}/problem/`, (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        code: 200,
        message: 'success',
        data: {
          problems: [
            {
              id: '507f1f77bcf86cd799439012',
              title: '两数之和',
              description: '给定一个整数数组...',
              difficulty: 'easy',
              time_limit: 1000,
              memory_limit: 256,
              tags: ['数组', '哈希表'],
              is_public: true,
              created_at: new Date().toISOString()
            }
          ],
          total: 100,
          page: 1,
          page_size: 10
        }
      })
    );
  }),

  // 提交代码
  rest.post(`${baseURL}/submit/`, async (req, res, ctx) => {
    const body = await req.json();
    return res(
      ctx.status(200),
      ctx.json({
        code: 200,
        message: 'success',
        data: {
          id: '507f1f77bcf86cd799439020',
          ...body,
          user_id: '507f1f77bcf86cd799439001',
          status: 'pending',
          created_at: new Date().toISOString()
        }
      })
    );
  }),

  // 用户登录
  rest.post(`${baseURL}/user/login`, async (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        code: 200,
        message: 'success',
        data: {
          token: 'mock_jwt_token_' + Date.now(),
          user: {
            id: '507f1f77bcf86cd799439001',
            username: 'testuser',
            email: 'test@example.com'
          }
        }
      })
    );
  }),
];
```

#### 步骤 4：启动 MSW

创建 `src/mocks/browser.js`：

```javascript
import { setupWorker } from 'msw';
import { handlers } from './handlers';

export const worker = setupWorker(...handlers);
```

在 `main.js` 中启动：

```javascript
// main.js
import { createApp } from 'vue';
import App from './App.vue';

async function enableMocking() {
  if (process.env.NODE_ENV !== 'development') {
    return;
  }

  const { worker } = await import('./mocks/browser');
  return worker.start();
}

enableMocking().then(() => {
  createApp(App).mount('#app');
});
```

---

### Axios Mock Adapter 配置

**适用场景**：使用 Axios 的项目、简单快速的 Mock 方案

#### 步骤 1：安装依赖

```bash
npm install axios-mock-adapter --save-dev
# 或
yarn add axios-mock-adapter -D
```

#### 步骤 2：创建 Mock 配置

创建 `src/api/mock.js`：

```javascript
import MockAdapter from 'axios-mock-adapter';
import api from './index'; // 你的 axios 实例

// 创建 Mock 实例
const mock = new MockAdapter(api, { delayResponse: 500 });

// ========== 测试用例管理 API ==========

// 获取题目的测试用例列表
mock.onGet(/\/testcase\/problem\/.*/).reply(200, {
  code: 200,
  message: 'success',
  data: {
    test_cases: [
      {
        id: '507f1f77bcf86cd799439011',
        problem_id: '507f1f77bcf86cd799439012',
        input: '1 2\n',
        output: '3\n',
        is_sample: true,
        time_limit: 2000,
        memory_limit: 512,
        score: 10,
        created_at: new Date().toISOString()
      }
    ],
    total: 1
  }
});

// 创建测试用例
mock.onPost('/testcase/').reply((config) => {
  const data = JSON.parse(config.data);
  return [200, {
    code: 200,
    message: 'success',
    data: {
      id: '507f1f77bcf86cd799439015',
      ...data,
      created_at: new Date().toISOString()
    }
  }];
});

// 更新测试用例
mock.onPut(/\/testcase\/.*/).reply((config) => {
  const data = JSON.parse(config.data);
  return [200, {
    code: 200,
    message: 'success',
    data: {
      id: '507f1f77bcf86cd799439011',
      problem_id: '507f1f77bcf86cd799439012',
      ...data,
      created_at: new Date().toISOString()
    }
  }];
});

// 删除测试用例
mock.onDelete(/\/testcase\/.*/).reply(200, {
  code: 200,
  message: 'success',
  data: {
    message: '测试用例删除成功'
  }
});

// ========== 用户登录 ==========
mock.onPost('/user/login').reply(200, {
  code: 200,
  message: 'success',
  data: {
    token: 'mock_jwt_token_' + Date.now(),
    user: {
      id: '507f1f77bcf86cd799439001',
      username: 'testuser',
      email: 'test@example.com'
    }
  }
});

console.log('✅ Axios Mock Adapter 已启用');

export default mock;
```

#### 步骤 3：在项目中引入

```javascript
// main.js
import { createApp } from 'vue';
import App from './App.vue';

// 仅在开发环境启用 Mock
if (process.env.NODE_ENV === 'development') {
  require('./api/mock');
}

createApp(App).mount('#app');
```

---

## TypeScript 类型定义

创建 `src/types/api.ts`：

```typescript
// ========== 基础类型 ==========

export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

// ========== 测试用例相关 ==========

export interface TestCase {
  id: string;
  problem_id: string;
  input: string;
  output: string;
  is_sample: boolean;
  time_limit?: number;  // ms
  memory_limit?: number;  // MB
  score?: number;
  created_at: string;
}

export interface TestCaseListResponse {
  test_cases: TestCase[];
  total: number;
}

export interface CreateTestCaseRequest {
  problem_id: string;
  input: string;
  output: string;
  is_sample?: boolean;
  time_limit?: number;
  memory_limit?: number;
  score?: number;
}

export interface UpdateTestCaseRequest {
  input?: string;
  output?: string;
  is_sample?: boolean;
  time_limit?: number;
  memory_limit?: number;
  score?: number;
}

// ========== 题目相关 ==========

export type Difficulty = 'easy' | 'medium' | 'hard';

export interface Problem {
  id: string;
  title: string;
  description: string;
  difficulty: Difficulty;
  time_limit: number;  // ms
  memory_limit: number;  // MB
  tags: string[];
  is_public: boolean;
  sample_input?: string;
  sample_output?: string;
  created_at: string;
  updated_at?: string;
}

export interface ProblemListResponse {
  problems: Problem[];
  total: number;
  page: number;
  page_size: number;
}

// ========== 提交相关 ==========

export type SubmitStatus =
  | 'pending'
  | 'running'
  | 'accepted'
  | 'wrong_answer'
  | 'time_limit_exceeded'
  | 'memory_limit_exceeded'
  | 'runtime_error'
  | 'compile_error'
  | 'system_error';

export type Language = 'java' | 'cpp' | 'python' | 'go' | 'c';

export interface TestResult {
  test_case_id: string;
  status: SubmitStatus;
  time_used: number;  // ms
  memory_used: number;  // KB
  is_sample: boolean;
  output?: string;  // 仅示例用例返回
  expected?: string;  // 仅示例用例返回
  error?: string;  // 仅示例用例返回
}

export interface JudgeResult {
  status: SubmitStatus;
  time_used: number;  // ms
  memory_used: number;  // KB
  test_results: TestResult[];
}

export interface Submit {
  id: string;
  problem_id: string;
  user_id: string;
  code: string;
  language: Language;
  status: SubmitStatus;
  result?: JudgeResult;
  created_at: string;
  updated_at?: string;
}

export interface CreateSubmitRequest {
  problem_id: string;
  code: string;
  language: Language;
}

// ========== 用户相关 ==========

export interface User {
  id: string;
  username: string;
  email: string;
  created_at?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
}
```

---

## 常见问题

### Q1: Mock 数据不生效？

**A**: 检查以下几点：
1. 确认 Mock 文件已正确引入
2. 检查 URL 匹配规则是否正确
3. 确认开发环境判断逻辑（`process.env.NODE_ENV === 'development'`）
4. 查看浏览器控制台是否有错误信息

### Q2: 如何切换真实接口和 Mock 接口？

**A**: 推荐使用环境变量：

```javascript
// .env.development
VITE_USE_MOCK=true
VITE_API_BASE_URL=http://localhost:8080/api/v1

// .env.production
VITE_USE_MOCK=false
VITE_API_BASE_URL=https://api.example.com/api/v1
```

```javascript
// api/index.js
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 5000,
});

// 仅在需要时启用 Mock
if (import.meta.env.VITE_USE_MOCK === 'true') {
  require('./mock');
}

export default api;
```

### Q3: 如何模拟网络延迟？

**A**: 各工具都支持延迟配置：

- **Mock.js**: `Mock.setup({ timeout: '200-600' })`
- **MSW**: `ctx.delay(500)`
- **Axios Mock Adapter**: `new MockAdapter(api, { delayResponse: 500 })`

### Q4: 如何模拟错误响应？

**A**: 在 Mock 配置中返回错误状态码：

```javascript
// Mock.js 示例
Mock.mock('/api/v1/testcase/error', 'get', {
  code: 404,
  message: '资源不存在',
  data: null
});

// MSW 示例
rest.get('/api/v1/testcase/error', (req, res, ctx) => {
  return res(
    ctx.status(404),
    ctx.json({
      code: 404,
      message: '资源不存在',
      data: null
    })
  );
});
```

---

## 📞 技术支持

如有疑问，请联系后端开发团队：
- **项目负责人**: omenkk7
- **更新日期**: 2025-10-05
