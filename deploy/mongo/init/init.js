// MongoDB 初始化脚本
// 创建数据库和集合
db = db.getSiblingDB('zk_code_arena');

// 创建用户集合索引
db.users.createIndex({ "username": 1 }, { unique: true });
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "student_id": 1 }, { unique: true, sparse: true });

// 创建题目集合索引
db.problems.createIndex({ "title": "text", "description": "text" });
db.problems.createIndex({ "difficulty": 1 });
db.problems.createIndex({ "tags": 1 });
db.problems.createIndex({ "status": 1 });
db.problems.createIndex({ "is_public": 1 });
db.problems.createIndex({ "created_by": 1 });

// 创建提交集合索引
db.submits.createIndex({ "user_id": 1 });
db.submits.createIndex({ "problem_id": 1 });
db.submits.createIndex({ "status": 1 });
db.submits.createIndex({ "created_at": -1 });
db.submits.createIndex({ "user_id": 1, "problem_id": 1 });

// 创建测试用例集合索引
db.test_cases.createIndex({ "problem_id": 1 });

// 插入默认管理员用户
db.users.insertOne({
    "username": "admin",
    "email": "admin@zk.edu.cn",
    "real_name": "系统管理员",
    "role": "admin",
    "is_active": true,
    "created_at": new Date(),
    "updated_at": new Date()
});

print("数据库初始化完成");
