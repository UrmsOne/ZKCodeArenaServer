/**
 * 测试数据初始化脚本
 * 
 * 功能:
 * 1. 创建测试用户 (管理员 + 普通用户)
 * 2. 创建测试题目 (3个不同难度的题目)
 * 3. 创建测试用例 (每个题目5个用例: 2个示例 + 3个隐藏)
 * 
 * 使用方法:
 * mongosh mongodb://42.194.245.236:27017/zk_code_arena < scripts/init_test_data.js
 */

// 连接数据库
db = db.getSiblingDB('zk_code_arena');

print("========================================");
print("开始初始化测试数据...");
print("========================================\n");

// ============================================
// 1. 清理旧测试数据 (可选)
// ============================================
print("1. 清理旧测试数据...");

// 删除测试用户 (保留 admin)
db.users.deleteMany({
    username: { $in: ["test_user_1", "test_user_2"] }
});

// 删除测试题目和相关测试用例
const testProblemTitles = ["A+B Problem", "数组排序", "字符串反转"];
const testProblems = db.problems.find({
    title: { $in: testProblemTitles }
}).toArray();

const testProblemIds = testProblems.map(p => p._id);
if (testProblemIds.length > 0) {
    db.test_cases.deleteMany({ problem_id: { $in: testProblemIds } });
    db.submits.deleteMany({ problem_id: { $in: testProblemIds } });
    db.problems.deleteMany({ _id: { $in: testProblemIds } });
}

print("✓ 旧测试数据清理完成\n");

// ============================================
// 2. 创建测试用户
// ============================================
print("2. 创建测试用户...");

// 注意: 实际使用时需要使用 bcrypt 加密密码
// 这里使用明文密码仅用于测试，生产环境必须加密
const testUsers = [
    {
        username: "test_user_1",
        password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // test123
        email: "user1@test.com",
        real_name: "测试用户1",
        role: "student",
        is_active: true,
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        username: "test_user_2",
        password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // test123
        email: "user2@test.com",
        real_name: "测试用户2",
        role: "student",
        is_active: true,
        created_at: new Date(),
        updated_at: new Date()
    }
];

const userInsertResult = db.users.insertMany(testUsers);
print(`✓ 创建了 ${Object.keys(userInsertResult.insertedIds).length} 个测试用户`);
print(`  - test_user_1 (密码: test123)`);
print(`  - test_user_2 (密码: test123)\n`);

// 获取管理员用户ID (用于创建题目)
const adminUser = db.users.findOne({ username: "admin" });
if (!adminUser) {
    print("⚠️  警告: 未找到管理员用户，请先运行初始化脚本");
    quit(1);
}

// ============================================
// 3. 创建测试题目
// ============================================
print("3. 创建测试题目...");

const testProblemsData = [
    {
        title: "A+B Problem",
        description: `# 题目描述

计算两个整数的和。

## 输入格式

一行两个整数 a 和 b，用空格分隔。

## 输出格式

一行一个整数，表示 a + b 的值。

## 数据范围

- -10^9 ≤ a, b ≤ 10^9

## 样例输入 1
\`\`\`
1 2
\`\`\`

## 样例输出 1
\`\`\`
3
\`\`\`

## 样例输入 2
\`\`\`
100 200
\`\`\`

## 样例输出 2
\`\`\`
300
\`\`\``,
        difficulty: "easy",
        time_limit: 1000,      // 1秒
        memory_limit: 128,     // 128MB
        tags: ["数学", "入门", "基础"],
        is_public: true,
        created_by: adminUser._id,
        submit_count: 0,
        accept_count: 0,
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        title: "数组排序",
        description: `# 题目描述

给定一个整数数组，请将其按升序排序。

## 输入格式

第一行一个整数 n，表示数组长度。
第二行 n 个整数，表示数组元素。

## 输出格式

一行 n 个整数，表示排序后的数组，用空格分隔。

## 数据范围

- 1 ≤ n ≤ 10^5
- -10^9 ≤ 数组元素 ≤ 10^9

## 样例输入
\`\`\`
5
3 1 4 2 5
\`\`\`

## 样例输出
\`\`\`
1 2 3 4 5
\`\`\``,
        difficulty: "medium",
        time_limit: 2000,      // 2秒
        memory_limit: 256,     // 256MB
        tags: ["排序", "数组", "算法"],
        is_public: true,
        created_by: adminUser._id,
        submit_count: 0,
        accept_count: 0,
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        title: "字符串反转",
        description: `# 题目描述

给定一个字符串，请将其反转输出。

## 输入格式

一行一个字符串（不含空格）。

## 输出格式

一行一个字符串，表示反转后的结果。

## 数据范围

- 1 ≤ 字符串长度 ≤ 10^5
- 字符串仅包含小写字母和数字

## 样例输入 1
\`\`\`
hello
\`\`\`

## 样例输出 1
\`\`\`
olleh
\`\`\`

## 样例输入 2
\`\`\`
world
\`\`\`

## 样例输出 2
\`\`\`
dlrow
\`\`\``,
        difficulty: "easy",
        time_limit: 1000,      // 1秒
        memory_limit: 128,     // 128MB
        tags: ["字符串", "基础"],
        is_public: true,
        created_by: adminUser._id,
        submit_count: 0,
        accept_count: 0,
        created_at: new Date(),
        updated_at: new Date()
    }
];

const problemInsertResult = db.problems.insertMany(testProblemsData);
const insertedProblemIds = Object.values(problemInsertResult.insertedIds);
print(`✓ 创建了 ${insertedProblemIds.length} 个测试题目`);
print(`  - A+B Problem (简单)`);
print(`  - 数组排序 (中等)`);
print(`  - 字符串反转 (简单)\n`);

// ============================================
// 4. 创建测试用例
// ============================================
print("4. 创建测试用例...");

// A+B Problem 测试用例
const apbTestCases = [
    { input: "1 2", output: "3", is_sample: true },
    { input: "100 200", output: "300", is_sample: true },
    { input: "0 0", output: "0", is_sample: false },
    { input: "-1 1", output: "0", is_sample: false },
    { input: "999999 1", output: "1000000", is_sample: false }
].map(tc => ({
    ...tc,
    problem_id: insertedProblemIds[0],
    score: 20,  // 每个用例20分
    created_at: new Date(),
    updated_at: new Date()
}));

// 数组排序测试用例
const sortTestCases = [
    { input: "5\n3 1 4 2 5", output: "1 2 3 4 5", is_sample: true },
    { input: "3\n1 2 3", output: "1 2 3", is_sample: true },
    { input: "1\n42", output: "42", is_sample: false },
    { input: "4\n4 3 2 1", output: "1 2 3 4", is_sample: false },
    { input: "5\n5 5 5 5 5", output: "5 5 5 5 5", is_sample: false }
].map(tc => ({
    ...tc,
    problem_id: insertedProblemIds[1],
    score: 20,
    time_limit: 3000,  // 排序题给更多时间
    created_at: new Date(),
    updated_at: new Date()
}));

// 字符串反转测试用例
const reverseTestCases = [
    { input: "hello", output: "olleh", is_sample: true },
    { input: "world", output: "dlrow", is_sample: true },
    { input: "a", output: "a", is_sample: false },
    { input: "12345", output: "54321", is_sample: false },
    { input: "racecar", output: "racecar", is_sample: false }
].map(tc => ({
    ...tc,
    problem_id: insertedProblemIds[2],
    score: 20,
    created_at: new Date(),
    updated_at: new Date()
}));

// 插入所有测试用例
const allTestCases = [...apbTestCases, ...sortTestCases, ...reverseTestCases];
const testCaseInsertResult = db.test_cases.insertMany(allTestCases);
print(`✓ 创建了 ${Object.keys(testCaseInsertResult.insertedIds).length} 个测试用例`);
print(`  - A+B Problem: 5个用例 (2个示例 + 3个隐藏)`);
print(`  - 数组排序: 5个用例 (2个示例 + 3个隐藏)`);
print(`  - 字符串反转: 5个用例 (2个示例 + 3个隐藏)\n`);

// ============================================
// 5. 数据验证
// ============================================
print("5. 数据验证...");

const userCount = db.users.countDocuments({ username: { $in: ["test_user_1", "test_user_2"] } });
const problemCount = db.problems.countDocuments({ title: { $in: testProblemTitles } });
const testCaseCount = db.test_cases.countDocuments({ problem_id: { $in: insertedProblemIds } });

print(`✓ 用户数量: ${userCount} (预期: 2)`);
print(`✓ 题目数量: ${problemCount} (预期: 3)`);
print(`✓ 测试用例数量: ${testCaseCount} (预期: 15)`);

if (userCount === 2 && problemCount === 3 && testCaseCount === 15) {
    print("\n✅ 所有测试数据初始化成功!");
} else {
    print("\n⚠️  警告: 数据数量不匹配，请检查!");
}

// ============================================
// 6. 输出测试数据信息
// ============================================
print("\n========================================");
print("测试数据信息");
print("========================================\n");

print("📝 测试账号:");
print("  用户名: test_user_1");
print("  密码: test123");
print("  邮箱: user1@test.com\n");

print("  用户名: test_user_2");
print("  密码: test123");
print("  邮箱: user2@test.com\n");

print("📚 测试题目:");
testProblemsData.forEach((problem, index) => {
    const problemId = insertedProblemIds[index];
    print(`  ${index + 1}. ${problem.title}`);
    print(`     ID: ${problemId}`);
    print(`     难度: ${problem.difficulty}`);
    print(`     时间限制: ${problem.time_limit}ms`);
    print(`     内存限制: ${problem.memory_limit}MB`);
    print(`     测试用例: 5个 (2个示例 + 3个隐藏)\n`);
});

print("========================================");
print("初始化完成!");
print("========================================");

