/**
 * ZK Code Arena Server - TypeScript 类型定义
 * @version 1.0
 * @date 2025-10-05
 * @description 判题系统 API 的 TypeScript 类型定义
 */

// ========== 基础类型 ==========

/**
 * API 统一响应格式
 */
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

/**
 * 分页参数
 */
export interface PaginationParams {
  page?: number;
  page_size?: number;
}

/**
 * 分页响应
 */
export interface PaginationResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

// ========== 测试用例相关 ==========

/**
 * 测试用例
 */
export interface TestCase {
  /** 测试用例 ID */
  id: string;
  /** 题目 ID */
  problem_id: string;
  /** 输入数据 */
  input: string;
  /** 期望输出 */
  output: string;
  /** 是否为示例用例 */
  is_sample: boolean;
  /** 时间限制（ms），可选，不填则使用题目默认值 */
  time_limit?: number;
  /** 内存限制（MB），可选，不填则使用题目默认值 */
  memory_limit?: number;
  /** 用例分数，用于部分分支持 */
  score?: number;
  /** 创建时间 */
  created_at: string;
}

/**
 * 测试用例列表响应
 */
export interface TestCaseListResponse {
  test_cases: TestCase[];
  total: number;
}

/**
 * 创建测试用例请求
 */
export interface CreateTestCaseRequest {
  /** 题目 ID */
  problem_id: string;
  /** 输入数据 */
  input: string;
  /** 期望输出 */
  output: string;
  /** 是否为示例用例，默认 false */
  is_sample?: boolean;
  /** 时间限制（ms） */
  time_limit?: number;
  /** 内存限制（MB） */
  memory_limit?: number;
  /** 用例分数 */
  score?: number;
}

/**
 * 更新测试用例请求
 */
export interface UpdateTestCaseRequest {
  /** 输入数据 */
  input?: string;
  /** 期望输出 */
  output?: string;
  /** 是否为示例用例 */
  is_sample?: boolean;
  /** 时间限制（ms） */
  time_limit?: number;
  /** 内存限制（MB） */
  memory_limit?: number;
  /** 用例分数 */
  score?: number;
}

// ========== 题目相关 ==========

/**
 * 题目难度
 */
export type Difficulty = 'easy' | 'medium' | 'hard';

/**
 * 题目
 */
export interface Problem {
  /** 题目 ID */
  id: string;
  /** 题目标题 */
  title: string;
  /** 题目描述 */
  description: string;
  /** 难度 */
  difficulty: Difficulty;
  /** 时间限制（ms） */
  time_limit: number;
  /** 内存限制（MB） */
  memory_limit: number;
  /** 标签 */
  tags: string[];
  /** 是否公开 */
  is_public: boolean;
  /** 示例输入 */
  sample_input?: string;
  /** 示例输出 */
  sample_output?: string;
  /** 创建时间 */
  created_at: string;
  /** 更新时间 */
  updated_at?: string;
}

/**
 * 题目列表响应
 */
export interface ProblemListResponse {
  problems: Problem[];
  total: number;
  page: number;
  page_size: number;
}

/**
 * 题目列表查询参数
 */
export interface ProblemListParams extends PaginationParams {
  /** 难度筛选 */
  difficulty?: Difficulty;
  /** 标签筛选 */
  tags?: string[];
  /** 是否公开 */
  is_public?: boolean;
}

// ========== 提交相关 ==========

/**
 * 提交状态
 */
export type SubmitStatus =
  | 'pending'                  // 等待判题
  | 'running'                  // 判题中
  | 'accepted'                 // 通过
  | 'wrong_answer'             // 答案错误
  | 'time_limit_exceeded'      // 超时
  | 'memory_limit_exceeded'    // 内存超限
  | 'runtime_error'            // 运行时错误
  | 'compile_error'            // 编译错误
  | 'system_error';            // 系统错误

/**
 * 编程语言
 */
export type Language = 'java' | 'cpp' | 'python' | 'go' | 'c';

/**
 * 测试用例结果
 */
export interface TestResult {
  /** 测试用例 ID */
  test_case_id: string;
  /** 判题状态 */
  status: SubmitStatus;
  /** 时间消耗（ms） */
  time_used: number;
  /** 内存消耗（KB） */
  memory_used: number;
  /** 是否为示例用例 */
  is_sample: boolean;
  /** 实际输出（仅示例用例返回） */
  output?: string;
  /** 期望输出（仅示例用例返回） */
  expected?: string;
  /** 错误信息（仅示例用例返回） */
  error?: string;
}

/**
 * 判题结果
 */
export interface JudgeResult {
  /** 最终状态 */
  status: SubmitStatus;
  /** 总时间消耗（ms） */
  time_used: number;
  /** 总内存消耗（KB） */
  memory_used: number;
  /** 测试用例结果列表 */
  test_results: TestResult[];
}

/**
 * 提交记录
 */
export interface Submit {
  /** 提交 ID */
  id: string;
  /** 题目 ID */
  problem_id: string;
  /** 用户 ID */
  user_id: string;
  /** 代码内容 */
  code: string;
  /** 编程语言 */
  language: Language;
  /** 提交状态 */
  status: SubmitStatus;
  /** 判题结果 */
  result?: JudgeResult;
  /** 创建时间 */
  created_at: string;
  /** 更新时间 */
  updated_at?: string;
}

/**
 * 创建提交请求
 */
export interface CreateSubmitRequest {
  /** 题目 ID */
  problem_id: string;
  /** 代码内容 */
  code: string;
  /** 编程语言 */
  language: Language;
}

// ========== 用户相关 ==========

/**
 * 用户
 */
export interface User {
  /** 用户 ID */
  id: string;
  /** 用户名 */
  username: string;
  /** 邮箱 */
  email: string;
  /** 创建时间 */
  created_at?: string;
}

/**
 * 登录请求
 */
export interface LoginRequest {
  /** 用户名 */
  username: string;
  /** 密码 */
  password: string;
}

/**
 * 登录响应
 */
export interface LoginResponse {
  /** JWT Token */
  token: string;
  /** 用户信息 */
  user: User;
}

/**
 * 注册请求
 */
export interface RegisterRequest {
  /** 用户名 */
  username: string;
  /** 密码 */
  password: string;
  /** 邮箱 */
  email: string;
}

// ========== API 服务类型 ==========

/**
 * 测试用例 API 服务
 */
export interface TestCaseAPI {
  /** 获取题目的测试用例列表 */
  getTestCasesByProblemId(problemId: string): Promise<ApiResponse<TestCaseListResponse>>;
  /** 获取单个测试用例 */
  getTestCaseById(id: string): Promise<ApiResponse<TestCase>>;
  /** 创建测试用例 */
  createTestCase(data: CreateTestCaseRequest): Promise<ApiResponse<TestCase>>;
  /** 更新测试用例 */
  updateTestCase(id: string, data: UpdateTestCaseRequest): Promise<ApiResponse<TestCase>>;
  /** 删除测试用例 */
  deleteTestCase(id: string): Promise<ApiResponse<{ message: string }>>;
}

/**
 * 题目 API 服务
 */
export interface ProblemAPI {
  /** 获取题目列表 */
  getProblemList(params?: ProblemListParams): Promise<ApiResponse<ProblemListResponse>>;
  /** 获取题目详情 */
  getProblemById(id: string): Promise<ApiResponse<Problem>>;
  /** 创建题目 */
  createProblem(data: Partial<Problem>): Promise<ApiResponse<Problem>>;
  /** 更新题目 */
  updateProblem(id: string, data: Partial<Problem>): Promise<ApiResponse<Problem>>;
  /** 删除题目 */
  deleteProblem(id: string): Promise<ApiResponse<{ message: string }>>;
}

/**
 * 提交 API 服务
 */
export interface SubmitAPI {
  /** 提交代码 */
  createSubmit(data: CreateSubmitRequest): Promise<ApiResponse<Submit>>;
  /** 获取提交详情 */
  getSubmitById(id: string): Promise<ApiResponse<Submit>>;
  /** 获取用户提交列表 */
  getUserSubmits(params?: PaginationParams): Promise<ApiResponse<PaginationResponse<Submit>>>;
}

/**
 * 用户 API 服务
 */
export interface UserAPI {
  /** 用户注册 */
  register(data: RegisterRequest): Promise<ApiResponse<User>>;
  /** 用户登录 */
  login(data: LoginRequest): Promise<ApiResponse<LoginResponse>>;
  /** 获取当前用户信息 */
  getCurrentUser(): Promise<ApiResponse<User>>;
}

// ========== 工具类型 ==========

/**
 * 状态码映射
 */
export const StatusCodeMap: Record<SubmitStatus, string> = {
  pending: '等待判题',
  running: '判题中',
  accepted: '通过',
  wrong_answer: '答案错误',
  time_limit_exceeded: '超时',
  memory_limit_exceeded: '内存超限',
  runtime_error: '运行时错误',
  compile_error: '编译错误',
  system_error: '系统错误',
};

/**
 * 难度映射
 */
export const DifficultyMap: Record<Difficulty, string> = {
  easy: '简单',
  medium: '中等',
  hard: '困难',
};

/**
 * 语言映射
 */
export const LanguageMap: Record<Language, string> = {
  java: 'Java',
  cpp: 'C++',
  python: 'Python',
  go: 'Go',
  c: 'C',
};

/**
 * 状态颜色映射（用于 UI 展示）
 */
export const StatusColorMap: Record<SubmitStatus, string> = {
  pending: '#faad14',      // 橙色
  running: '#1890ff',      // 蓝色
  accepted: '#52c41a',     // 绿色
  wrong_answer: '#f5222d', // 红色
  time_limit_exceeded: '#fa8c16',
  memory_limit_exceeded: '#fa8c16',
  runtime_error: '#ff4d4f',
  compile_error: '#ff7a45',
  system_error: '#8c8c8c', // 灰色
};

/**
 * 难度颜色映射（用于 UI 展示）
 */
export const DifficultyColorMap: Record<Difficulty, string> = {
  easy: '#52c41a',    // 绿色
  medium: '#faad14',  // 橙色
  hard: '#f5222d',    // 红色
};
