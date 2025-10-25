/*
@Author: omenkk7
@Date: 2025/10/9
@Description: 用户模块集成测试
*/

package server

import (
	"net/http"
	"testing"

	"zk-code-arena-server/pkg/app/api-server/server/testutils"
)

// TestUserRegisterAndLogin 测试用户注册和登录流程
func TestUserRegisterAndLogin(t *testing.T) {
	// 设置测试服务器
	s, cleanup := testutils.SetupTestServer(t)
	defer cleanup()

	app := s.GetApp()

	t.Run("用户注册成功", func(t *testing.T) {
		registerBody := map[string]interface{}{
			"student_id": "2021001",
			"username":   "test_student",
			"password":   "password123",
			"email":      "test@student.com",
			"role":       "student",
		}

		w := testutils.MakeGinRequest(t, app, "POST", "/api/v1/user/register", registerBody, nil)

		testutils.AssertStatusCode(t, w, http.StatusOK)
		testutils.AssertResponseCode(t, w, 200)
	})

	t.Run("重复注册失败", func(t *testing.T) {
		// 先注册一个用户
		registerBody := map[string]interface{}{
			"student_id": "2021002",
			"username":   "duplicate_user",
			"password":   "password123",
			"email":      "duplicate@test.com",
		}

		w := testutils.MakeGinRequest(t, app, "POST", "/api/v1/user/register", registerBody, nil)
		testutils.AssertStatusCode(t, w, http.StatusOK)

		// 尝试用相同用户名再次注册
		w = testutils.MakeGinRequest(t, app, "POST", "/api/v1/user/register", registerBody, nil)
		testutils.AssertStatusCode(t, w, http.StatusInternalServerError)
	})

	t.Run("用户登录成功", func(t *testing.T) {
		// 先注册
		registerBody := map[string]interface{}{
			"student_id": "2021003",
			"username":   "login_test",
			"password":   "password123",
			"email":      "login@test.com",
		}

		w := testutils.MakeGinRequest(t, app, "POST", "/api/v1/user/register", registerBody, nil)
		testutils.AssertStatusCode(t, w, http.StatusOK)

		// 登录
		loginBody := map[string]interface{}{
			"student_id": "2021003",
			"password":   "password123",
		}

		w = testutils.MakeGinRequest(t, app, "POST", "/api/v1/user/login", loginBody, nil)
		testutils.AssertStatusCode(t, w, http.StatusOK)

		var loginResp struct {
			Code int `json:"code"`
			Data struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		testutils.ParseResponse(t, w, &loginResp)

		if loginResp.Data.Token == "" {
			t.Error("Expected token in response, got empty string")
		}
	})

	t.Run("错误密码登录失败", func(t *testing.T) {
		// 先注册
		registerBody := map[string]interface{}{
			"student_id": "2021004",
			"username":   "wrong_pass_test",
			"password":   "correct_password",
			"email":      "wrongpass@test.com",
		}

		w := testutils.MakeGinRequest(t, app, "POST", "/api/v1/user/register", registerBody, nil)
		testutils.AssertStatusCode(t, w, http.StatusOK)

		// 使用错误密码登录
		loginBody := map[string]interface{}{
			"student_id": "2021004",
			"password":   "wrong_password",
		}

		w = testutils.MakeGinRequest(t, app, "POST", "/api/v1/user/login", loginBody, nil)
		testutils.AssertStatusCode(t, w, http.StatusUnauthorized)
	})
}

// TestUserProfile 测试用户资料获取和更新
//func TestUserProfile(t *testing.T) {
//	s, cleanup := testutils.SetupTestServer(t)
//	defer cleanup()
//
//	app := s.GetRouter()
//
//	// 创建测试用户并获取 token
//	token := testutils.CreateTestUser(t, app, "2021005", "password123", "student")
//
//	t.Run("获取用户资料成功", func(t *testing.T) {
//		headers := testutils.WithAuth(token)
//		w := testutils.MakeGinRequest(t, app, "GET", "/api/v1/user/profile", nil, headers)
//
//		testutils.AssertStatusCode(t, w, http.StatusOK)
//		testutils.AssertResponseCode(t, w, 200)
//
//		var resp struct {
//			Code int `json:"code"`
//			Data struct {
//				StudentID string `json:"student_id"`
//				Username  string `json:"username"`
//			} `json:"data"`
//		}
//		testutils.ParseResponse(t, w, &resp)
//
//		if resp.Data.StudentID != "2021005" {
//			t.Errorf("Expected student_id 2021005, got %s", resp.Data.StudentID)
//		}
//	})
//
//	t.Run("未认证访问用户资料失败", func(t *testing.T) {
//		w := testutils.MakeGinRequest(t, app, "GET", "/api/v1/user/profile", nil, nil)
//
//		testutils.AssertStatusCode(t, w, http.StatusUnauthorized)
//	})
//
//	t.Run("更新用户资料成功", func(t *testing.T) {
//		updateBody := map[string]interface{}{
//			"real_name": "测试用户",
//			"school":    "仲恺农业工程学院",
//			"major":     "计算机科学",
//		}
//
//		headers := testutils.WithAuth(token)
//		w := testutils.MakeGinRequest(t, app, "PUT", "/api/v1/user/profile", updateBody, headers)
//
//		testutils.AssertStatusCode(t, w, http.StatusOK)
//		testutils.AssertResponseCode(t, w, 200)
//	})
//}

// TestUserPermissions 测试用户权限控制
//func TestUserPermissions(t *testing.T) {
//	s, cleanup := testutils.SetupTestServer(t)
//	defer cleanup()
//
//	app := s.GetRouter()
//
//	// 创建普通学生
//	studentToken := testutils.CreateTestUser(t, app, "2021006", "password123", "student")
//
//	// 创建管理员
//	adminToken := testutils.CreateTestUser(t, app, "admin001", "admin123", "admin")
//
//	t.Run("普通用户获取用户列表失败", func(t *testing.T) {
//		headers := testutils.WithAuth(studentToken)
//		w := testutils.MakeGinRequest(t, app, "GET", "/api/v1/user?page=1&page_size=10", nil, headers)
//
//		testutils.AssertStatusCode(t, w, http.StatusForbidden)
//	})
//
//	t.Run("管理员获取用户列表成功", func(t *testing.T) {
//		headers := testutils.WithAuth(adminToken)
//		w := testutils.MakeGinRequest(t, app, "GET", "/api/v1/user?page=1&page_size=10", nil, headers)
//
//		testutils.AssertStatusCode(t, w, http.StatusOK)
//		testutils.AssertResponseCode(t, w, 200)
//	})
//}
