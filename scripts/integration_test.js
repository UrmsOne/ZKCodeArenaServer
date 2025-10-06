/**
 * 判题系统集成测试脚本
 * 
 * 使用方法:
 * node scripts/integration_test.js
 * 
 * 环境要求:
 * - 服务器已启动 (默认: http://localhost:8080)
 * - MongoDB 已运行并包含测试数据
 * - go-judge 沙箱已启动 (http://localhost:5050)
 */

const axios = require('axios');
const { MongoClient, ObjectId } = require('mongodb');

// 配置
const API_BASE_URL = process.env.API_URL || 'http://localhost:8080';
const MONGO_URI = process.env.MONGO_URI || 'mongodb://42.194.245.236:27017';
const DB_NAME = 'zk_code_arena';

// 测试用户凭证
const TEST_USER = {
    username: 'test_user_1',
    password: 'test123'
};

// 测试代码
const TEST_CODES = {
    // A+B Problem - 正确答案 (Java)
    apb_correct_java: `
import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int a = sc.nextInt();
        int b = sc.nextInt();
        System.out.println(a + b);
        sc.close();
    }
}
`.trim(),

    // A+B Problem - 错误答案 (Java)
    apb_wrong_java: `
import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int a = sc.nextInt();
        int b = sc.nextInt();
        System.out.println(a * b); // 错误: 应该是加法，这里是乘法
        sc.close();
    }
}
`.trim(),

    // A+B Problem - 编译错误 (Java)
    apb_compile_error_java: `
import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int a = sc.nextInt();
        int b = sc.nextInt();
        System.out.println(a + b) // 缺少分号
        sc.close();
    }
}
`.trim(),

    // A+B Problem - 运行时错误 (Java)
    apb_runtime_error_java: `
import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        int[] arr = new int[1];
        System.out.println(arr[10]); // 数组越界
    }
}
`.trim(),

    // A+B Problem - 超时 (Java - 死循环)
    apb_timeout_java: `
import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        while(true) {
            // 死循环
        }
    }
}
`.trim()
};

// 测试结果统计
const testResults = {
    total: 0,
    passed: 0,
    failed: 0,
    errors: []
};

// HTTP 客户端
let authToken = null;
const apiClient = axios.create({
    baseURL: API_BASE_URL,
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json'
    }
});

// 请求拦截器 - 添加认证 token
apiClient.interceptors.request.use(config => {
    if (authToken) {
        config.headers['Authorization'] = `Bearer ${authToken}`;
    }
    return config;
});

// MongoDB 客户端
let mongoClient = null;
let db = null;

/**
 * 工具函数: 延迟
 */
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * 工具函数: 打印测试结果
 */
function printTestResult(testName, passed, message = '') {
    testResults.total++;
    if (passed) {
        testResults.passed++;
        console.log(`✅ ${testName}`);
    } else {
        testResults.failed++;
        console.log(`❌ ${testName}`);
        if (message) {
            console.log(`   错误: ${message}`);
            testResults.errors.push({ test: testName, error: message });
        }
    }
}

/**
 * 测试 1: 服务健康检查
 */
async function testHealthCheck() {
    console.log('\n=== 测试 1: 服务健康检查 ===');
    
    try {
        const response = await axios.get(`${API_BASE_URL}/health`, { timeout: 5000 });
        printTestResult('服务健康检查', response.status === 200, '');
    } catch (error) {
        printTestResult('服务健康检查', false, error.message);
    }
}

/**
 * 测试 2: 用户登录
 */
async function testUserLogin() {
    console.log('\n=== 测试 2: 用户登录 ===');
    
    try {
        const response = await apiClient.post('/api/v1/user/login', TEST_USER);
        
        if (response.data && response.data.data && response.data.data.token) {
            authToken = response.data.data.token;
            printTestResult('用户登录成功', true);
            return true;
        } else {
            printTestResult('用户登录失败', false, '未返回 token');
            return false;
        }
    } catch (error) {
        printTestResult('用户登录失败', false, error.response?.data?.message || error.message);
        return false;
    }
}

/**
 * 测试 3: 获取题目列表
 */
async function testGetProblems() {
    console.log('\n=== 测试 3: 获取题目列表 ===');
    
    try {
        const response = await apiClient.get('/api/v1/problem');
        
        if (response.data && response.data.data && response.data.data.problems && Array.isArray(response.data.data.problems)) {
            const problems = response.data.data.problems;
            printTestResult(`获取题目列表 (共 ${problems.length} 个题目)`, problems.length > 0);
            return problems;
        } else {
            printTestResult('获取题目列表失败', false, '返回数据格式错误');
            console.log('   实际返回:', JSON.stringify(response.data, null, 2));
            return [];
        }
    } catch (error) {
        printTestResult('获取题目列表失败', false, error.response?.data?.message || error.message);
        return [];
    }
}

/**
 * 测试 4: 提交代码 (正确答案)
 */
async function testSubmitCorrectCode(problemId) {
    console.log('\n=== 测试 4: 提交正确代码 ===');
    
    try {
        const response = await apiClient.post('/api/v1/submit', {
            problem_id: problemId,
            code: TEST_CODES.apb_correct_java,
            language: 'java'
        });
        
        if (response.data && response.data.data && response.data.data.submit_id) {
            const submitId = response.data.data.submit_id;
            printTestResult('提交代码成功', true);
            
            // 等待判题完成
            console.log('   ⏳ 等待判题完成...');
            const result = await waitForJudgeResult(submitId, 30000);
            
            if (result) {
                const isAccepted = result.status === 'accepted';
                printTestResult(
                    `判题结果: ${result.status}`,
                    isAccepted,
                    isAccepted ? '' : `预期 accepted, 实际 ${result.status}`
                );
                
                if (result.result) {
                    console.log(`   时间: ${result.result.time_used}ms, 内存: ${result.result.memory_used}KB`);
                    console.log(`   通过用例: ${result.result.passed_cases}/${result.result.total_cases}`);
                }
            }
            
            return submitId;
        } else {
            printTestResult('提交代码失败', false, '未返回 submit_id');
            return null;
        }
    } catch (error) {
        printTestResult('提交代码失败', false, error.response?.data?.message || error.message);
        return null;
    }
}

/**
 * 测试 5: 提交代码 (错误答案)
 */
async function testSubmitWrongCode(problemId) {
    console.log('\n=== 测试 5: 提交错误代码 (Wrong Answer) ===');
    
    try {
        const response = await apiClient.post('/api/v1/submit', {
            problem_id: problemId,
            code: TEST_CODES.apb_wrong_java,
            language: 'java'
        });
        
        if (response.data && response.data.data && response.data.data.submit_id) {
            const submitId = response.data.data.submit_id;
            printTestResult('提交代码成功', true);
            
            // 等待判题完成
            console.log('   ⏳ 等待判题完成...');
            const result = await waitForJudgeResult(submitId, 30000);
            
            if (result) {
                const isWrongAnswer = result.status === 'wrong_answer';
                printTestResult(
                    `判题结果: ${result.status}`,
                    isWrongAnswer,
                    isWrongAnswer ? '' : `预期 wrong_answer, 实际 ${result.status}`
                );
            }
            
            return submitId;
        } else {
            printTestResult('提交代码失败', false, '未返回 submit_id');
            return null;
        }
    } catch (error) {
        printTestResult('提交代码失败', false, error.response?.data?.message || error.message);
        return null;
    }
}

/**
 * 测试 6: 提交代码 (编译错误)
 */
async function testSubmitCompileError(problemId) {
    console.log('\n=== 测试 6: 提交编译错误代码 ===');
    
    try {
        const response = await apiClient.post('/api/v1/submit', {
            problem_id: problemId,
            code: TEST_CODES.apb_compile_error_java,
            language: 'java'
        });
        
        if (response.data && response.data.data && response.data.data.submit_id) {
            const submitId = response.data.data.submit_id;
            printTestResult('提交代码成功', true);
            
            // 等待判题完成
            console.log('   ⏳ 等待判题完成...');
            const result = await waitForJudgeResult(submitId, 30000);
            
            if (result) {
                const isCompileError = result.status === 'compile_error';
                printTestResult(
                    `判题结果: ${result.status}`,
                    isCompileError,
                    isCompileError ? '' : `预期 compile_error, 实际 ${result.status}`
                );
                
                if (result.result && result.result.compile_error) {
                    console.log(`   编译错误信息: ${result.result.compile_error.substring(0, 100)}...`);
                }
            }
            
            return submitId;
        } else {
            printTestResult('提交代码失败', false, '未返回 submit_id');
            return null;
        }
    } catch (error) {
        printTestResult('提交代码失败', false, error.response?.data?.message || error.message);
        return null;
    }
}

/**
 * 测试 7: 提交代码 (运行时错误)
 */
async function testSubmitRuntimeError(problemId) {
    console.log('\n=== 测试 7: 提交运行时错误代码 ===');
    
    try {
        const response = await apiClient.post('/api/v1/submit', {
            problem_id: problemId,
            code: TEST_CODES.apb_runtime_error_java,
            language: 'java'
        });
        
        if (response.data && response.data.data && response.data.data.submit_id) {
            const submitId = response.data.data.submit_id;
            printTestResult('提交代码成功', true);
            
            // 等待判题完成
            console.log('   ⏳ 等待判题完成...');
            const result = await waitForJudgeResult(submitId, 30000);
            
            if (result) {
                const isRuntimeError = result.status === 'runtime_error';
                printTestResult(
                    `判题结果: ${result.status}`,
                    isRuntimeError,
                    isRuntimeError ? '' : `预期 runtime_error, 实际 ${result.status}`
                );
            }
            
            return submitId;
        } else {
            printTestResult('提交代码失败', false, '未返回 submit_id');
            return null;
        }
    } catch (error) {
        printTestResult('提交代码失败', false, error.response?.data?.message || error.message);
        return null;
    }
}

/**
 * 测试 8: 数据库状态验证
 */
async function testDatabaseStatus() {
    console.log('\n=== 测试 8: 数据库状态验证 ===');
    
    try {
        // 连接 MongoDB
        if (!mongoClient) {
            mongoClient = new MongoClient(MONGO_URI);
            await mongoClient.connect();
            db = mongoClient.db(DB_NAME);
        }
        
        // 检查提交记录数量
        const submitCount = await db.collection('submits').countDocuments();
        printTestResult(`数据库中有 ${submitCount} 条提交记录`, submitCount > 0);
        
        // 检查是否有 pending 状态的任务 (不应该有)
        const pendingCount = await db.collection('submits').countDocuments({ status: 'pending' });
        printTestResult(
            '没有遗留的 pending 任务',
            pendingCount === 0,
            pendingCount > 0 ? `发现 ${pendingCount} 个 pending 任务` : ''
        );
        
        // 检查最近的提交状态分布
        const statusStats = await db.collection('submits').aggregate([
            { $group: { _id: '$status', count: { $sum: 1 } } }
        ]).toArray();
        
        console.log('   提交状态分布:');
        statusStats.forEach(stat => {
            console.log(`     ${stat._id}: ${stat.count}`);
        });
        
    } catch (error) {
        printTestResult('数据库状态验证失败', false, error.message);
    }
}

/**
 * 辅助函数: 等待判题结果
 */
async function waitForJudgeResult(submitId, timeout = 30000) {
    const startTime = Date.now();
    const pollInterval = 1000; // 1秒轮询一次
    
    while (Date.now() - startTime < timeout) {
        try {
            const response = await apiClient.get(`/api/v1/submit/${submitId}`);
            
            if (response.data && response.data.data) {
                const submit = response.data.data;
                
                // 如果状态不是 pending 或 running，说明判题完成
                if (submit.status !== 'pending' && submit.status !== 'running') {
                    return submit;
                }
            }
        } catch (error) {
            console.log(`   ⚠️  查询提交状态失败: ${error.message}`);
        }
        
        await sleep(pollInterval);
    }
    
    printTestResult('等待判题结果超时', false, `超过 ${timeout}ms`);
    return null;
}

/**
 * 主测试流程
 */
async function runTests() {
    console.log('========================================');
    console.log('判题系统集成测试');
    console.log('========================================');
    console.log(`API 地址: ${API_BASE_URL}`);
    console.log(`MongoDB: ${MONGO_URI}`);
    console.log('========================================\n');
    
    try {
        // 1. 健康检查
        await testHealthCheck();
        
        // 2. 用户登录
        const loginSuccess = await testUserLogin();
        if (!loginSuccess) {
            console.log('\n❌ 用户登录失败，无法继续测试');
            return;
        }
        
        // 3. 获取题目列表
        const problems = await testGetProblems();
        if (problems.length === 0) {
            console.log('\n❌ 没有可用的题目，无法继续测试');
            return;
        }
        
        // 找到 A+B Problem
        const apbProblem = problems.find(p => p.title === 'A+B Problem');
        if (!apbProblem) {
            console.log('\n❌ 未找到 "A+B Problem" 题目，无法继续测试');
            return;
        }
        
        console.log(`\n使用题目: ${apbProblem.title} (ID: ${apbProblem.id})`);
        
        // 4-7. 提交不同类型的代码
        await testSubmitCorrectCode(apbProblem.id);
        await sleep(2000); // 等待2秒，避免请求过快
        
        await testSubmitWrongCode(apbProblem.id);
        await sleep(2000);
        
        await testSubmitCompileError(apbProblem.id);
        await sleep(2000);
        
        await testSubmitRuntimeError(apbProblem.id);
        await sleep(2000);
        
        // 8. 数据库状态验证
        await testDatabaseStatus();
        
    } catch (error) {
        console.error('\n❌ 测试执行出错:', error.message);
        console.error(error.stack);
    } finally {
        // 关闭 MongoDB 连接
        if (mongoClient) {
            await mongoClient.close();
        }
    }
    
    // 打印测试总结
    console.log('\n========================================');
    console.log('测试总结');
    console.log('========================================');
    console.log(`总测试数: ${testResults.total}`);
    console.log(`✅ 通过: ${testResults.passed}`);
    console.log(`❌ 失败: ${testResults.failed}`);
    console.log(`通过率: ${((testResults.passed / testResults.total) * 100).toFixed(2)}%`);
    
    if (testResults.errors.length > 0) {
        console.log('\n失败的测试:');
        testResults.errors.forEach((err, index) => {
            console.log(`${index + 1}. ${err.test}`);
            console.log(`   ${err.error}`);
        });
    }
    
    console.log('\n========================================');
    
    // 返回退出码
    process.exit(testResults.failed > 0 ? 1 : 0);
}

// 运行测试
runTests().catch(error => {
    console.error('测试执行失败:', error);
    process.exit(1);
});
