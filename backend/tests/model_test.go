package tests

import (
	"testing"

	"casbin-demo/internal/model"
)

func TestUserModel(t *testing.T) {
	user := model.User{
		Username: "testuser",
		Nickname: "Test User",
		Email:    "test@example.com",
		Status:   1,
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", user.Username)
	}

	if user.Status != 1 {
		t.Errorf("Expected status 1, got %d", user.Status)
	}

	if user.TableName() != "users" {
		t.Errorf("Expected table name 'users', got %s", user.TableName())
	}
}

func TestRoleModel(t *testing.T) {
	role := model.Role{
		Name:        "editor",
		Label:       "编辑",
		Description: "内容编辑角色",
		Status:      1,
	}

	if role.Name != "editor" {
		t.Errorf("Expected name 'editor', got %s", role.Name)
	}

	if role.TableName() != "roles" {
		t.Errorf("Expected table name 'roles', got %s", role.TableName())
	}
}

func TestPermissionModel(t *testing.T) {
	perm := model.Permission{
		Name:   "article:edit",
		Label:  "编辑文章",
		Path:   "/api/articles/:id",
		Method: "PUT",
	}

	if perm.Method != "PUT" {
		t.Errorf("Expected method 'PUT', got %s", perm.Method)
	}

	if perm.TableName() != "permissions" {
		t.Errorf("Expected table name 'permissions', got %s", perm.TableName())
	}
}

func TestLoginRequestModelValidation(t *testing.T) {
	req := model.LoginRequest{
		Username: "admin",
		Password: "123456",
	}

	if req.Username == "" {
		t.Error("Username should not be empty")
	}

	if req.Password == "" {
		t.Error("Password should not be empty")
	}
}
