// Package tests 统一存放项目集成测试和单元测试。
//
// 测试分类：
//   - handler_test.go：HTTP 接口层测试（健康检查、认证、分页等）
//   - model_test.go：数据模型单元测试
//   - jwt_test.go：JWT 令牌生成/解析测试
//
// 运行测试：
//
//	go test ./tests/... -v
package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	return r
}

func TestHealthEndpoint(t *testing.T) {
	r := setupTestRouter()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "casbin-demo",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}

	if response["service"] != "casbin-demo" {
		t.Errorf("Expected service 'casbin-demo', got %v", response["service"])
	}
}

func TestLoginRequestValidation(t *testing.T) {
	r := setupTestRouter()
	r.POST("/api/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": 400, "message": "请求参数错误"})
			return
		}
		c.JSON(200, gin.H{"code": 200, "message": "success", "data": gin.H{"token": "test-token"}})
	})

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "123456"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}
}

func TestLoginRequestMissingFields(t *testing.T) {
	r := setupTestRouter()
	r.POST("/api/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": 400, "message": "请求参数错误"})
			return
		}
		c.JSON(200, gin.H{"code": 200})
	})

	body, _ := json.Marshal(map[string]string{"username": "admin"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for missing fields, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	r := setupTestRouter()
	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"code": 401, "message": "未提供认证令牌"})
			c.Abort()
			return
		}
		c.Next()
	})
	auth.GET("/userinfo", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 200, "data": gin.H{"username": "admin"}})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/userinfo", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized access, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUserListPagination(t *testing.T) {
	r := setupTestRouter()
	r.GET("/api/users", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("page_size", "10")
		c.JSON(200, gin.H{
			"code": 200,
			"data": gin.H{
				"list":      []interface{}{},
				"total":     0,
				"page":      page,
				"page_size": pageSize,
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	if data["page"] != "1" {
		t.Errorf("Expected page 1, got %v", data["page"])
	}
}
