import json
import requests
import sys
from typing import Dict, List, Any

# 配置
API_BASE_URL = "http://8.138.184.24:8080/api/v1"
TOKEN = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjhmZjc1Y2NkNjI1MDg2OGUxN2QyZjUxIiwidXNlcm5hbWUiOiJ0ZXN0Iiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzYyNTg0MDg0LCJpYXQiOjE3NjI0OTc2ODR9.OIwhKqF2xe71fIwt8wUSwiBrLzhFIs2uf2SKLhAYwm0"
HEADERS = {
    "Authorization": TOKEN,
    "Content-Type": "application/json"
}

# 创建题目
def create_problem(problem_data: Dict[str, Any]) -> str:
    """
    创建题目并返回题目ID
    """
    url = f"{API_BASE_URL}/problem"
    
    # 处理time_limit，从字符串转换为整数毫秒
    time_limit_str = problem_data.get("time_limit", "1000ms")
    if isinstance(time_limit_str, str):
        if time_limit_str.endswith('s'):
            time_limit = int(float(time_limit_str[:-1]) * 1000)  # 秒转毫秒
        elif time_limit_str.endswith('ms'):
            time_limit = int(time_limit_str[:-2])  # 已经是毫秒
        else:
            time_limit = int(time_limit_str)  # 假设已经是数字
    else:
        time_limit = int(time_limit_str)
    
    # 处理memory_limit，从字符串转换为整数MB
    memory_limit_str = problem_data.get("memory_limit", "256MB")
    if isinstance(memory_limit_str, str):
        if memory_limit_str.endswith('MB'):
            memory_limit = int(memory_limit_str[:-2])
        elif memory_limit_str.endswith('KB'):
            memory_limit = int(memory_limit_str[:-2]) // 1024  # KB转MB
        elif memory_limit_str.endswith('GB'):
            memory_limit = int(memory_limit_str[:-2]) * 1024  # GB转MB
        else:
            memory_limit = int(memory_limit_str)  # 假设已经是数字
    else:
        memory_limit = int(memory_limit_str)
    
    # 构建请求体
    request_data = {
        "title": problem_data.get("title", ""),
        "description": problem_data.get("description", ""),
        "difficulty": problem_data.get("difficulty", "medium"),
        "input": problem_data.get("input_format", ""),
        "output": problem_data.get("output_format", ""),
        "sample_input": problem_data.get("sample_input", ""),
        "sample_output": problem_data.get("sample_output", ""),
        "time_limit": time_limit,
        "memory_limit": memory_limit,
        "tags": problem_data.get("tags", []),
        "status": "published",  # 设置为已发布状态
        "is_public": True       # 设置为公开状态
    }
    
    print(f"创建题目: {request_data['title']}")
    print(f"时间限制: {time_limit}ms, 内存限制: {memory_limit}MB")
    
    try:
        response = requests.post(url, headers=HEADERS, json=request_data)
        
        # 添加详细错误信息
        if response.status_code != 200:
            print(f"HTTP错误代码: {response.status_code}")
            print(f"响应内容: {response.text}")
            
        response.raise_for_status()
        result = response.json()
        
        # 检查响应格式，API返回的是 {'code': 200, 'message': 'success', 'data': {...}}
        if result.get("code") == 200 and result.get("message") == "success":
            problem_id = result.get("data", {}).get("id")
            print(f"题目创建成功，ID: {problem_id}")
            return problem_id
        else:
            print(f"题目创建失败: {result.get('message', '未知错误')}")
            return None
    except Exception as e:
        print(f"创建题目时发生错误: {str(e)}")
        return None

# 创建测试用例
def create_test_cases(problem_id: str, test_cases: List[Dict[str, Any]]) -> bool:
    """
    为指定题目创建测试用例
    """
    if not test_cases:
        print("没有测试用例需要创建")
        return True
    
    url = f"{API_BASE_URL}/testcase/batch"
    
    # 构建请求体
    request_data = {
        "problem_id": problem_id,
        "test_cases": []
    }
    
    for i, test_case in enumerate(test_cases):
        tc_data = {
            "input": test_case.get("input", ""),
            "output": test_case.get("expected", ""),
            "is_sample": test_case.get("is_hidden", False) == False,  # 如果不隐藏，则为示例
            "score": 10  # 默认每个测试用例10分
        }
        request_data["test_cases"].append(tc_data)
    
    print(f"创建 {len(test_cases)} 个测试用例")
    
    try:
        response = requests.post(url, headers=HEADERS, json=request_data)
        response.raise_for_status()
        result = response.json()
        
        # 检查响应格式，API返回的是 {'code': 200, 'message': 'success', 'data': {...}}
        if result.get("code") == 200 and result.get("message") == "success":
            print(f"测试用例创建成功")
            return True
        else:
            print(f"测试用例创建失败: {result.get('message', '未知错误')}")
            return False
    except Exception as e:
        print(f"创建测试用例时发生错误: {str(e)}")
        return False

# 处理JSONL文件
def process_jsonl_file(file_path: str) -> None:
    """
    处理JSONL文件中的题目数据
    """
    success_count = 0
    fail_count = 0
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            for line_num, line in enumerate(f, 1):
                try:
                    problem_data = json.loads(line.strip())
                    
                    # 创建题目
                    problem_id = create_problem(problem_data)
                    
                    if problem_id:
                        # 创建测试用例
                        test_cases = problem_data.get("test_cases", [])
                        if create_test_cases(problem_id, test_cases):
                            success_count += 1
                            print(f"第 {line_num} 行题目及其测试用例创建成功")
                        else:
                            fail_count += 1
                            print(f"第 {line_num} 行题目创建成功，但测试用例创建失败")
                    else:
                        fail_count += 1
                        print(f"第 {line_num} 行题目创建失败")
                    
                    print("-" * 50)
                    
                except json.JSONDecodeError as e:
                    print(f"第 {line_num} 行JSON解析错误: {str(e)}")
                    fail_count += 1
                    continue
                except Exception as e:
                    print(f"处理第 {line_num} 行时发生错误: {str(e)}")
                    fail_count += 1
                    continue
                    
    except FileNotFoundError:
        print(f"文件未找到: {file_path}")
        return
    except Exception as e:
        print(f"处理文件时发生错误: {str(e)}")
        return
    
    print(f"处理完成! 成功: {success_count}, 失败: {fail_count}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("用法: python import_problems.py <jsonl文件路径>")
        sys.exit(1)
    
    jsonl_file_path = sys.argv[1]
    process_jsonl_file(jsonl_file_path)