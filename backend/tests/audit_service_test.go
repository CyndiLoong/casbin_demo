package tests

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"casbin-demo/internal/model"
	"casbin-demo/internal/repository"
	"casbin-demo/internal/service"
	"casbin-demo/pkg/cache"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.AuditApplication{}, &model.SysMessage{})

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
	return db, cleanup
}

func setupTestAuditService(t *testing.T, db *gorm.DB) (*service.AuditService, *repository.AuditRepository, *repository.UserRepository) {
	t.Helper()
	auditRepo := repository.NewAuditRepository(db)
	userRepo := repository.NewUserRepository(db)
	cacheClient := cache.NewClient(nil)
	auditService := service.NewAuditService(auditRepo, userRepo, cacheClient, nil, nil, nil)
	return auditService, auditRepo, userRepo
}

func createTestUser(db *gorm.DB, username string, isAdmin bool) *model.User {
	user := &model.User{
		UUID:     uuid.New().String(),
		Username: username,
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		Nickname: username,
		Status:   1,
	}
	db.Create(user)
	if isAdmin {
		db.Exec("INSERT INTO roles (name, label, status) VALUES (?, ?, 1)", "admin", "管理员")
		var role model.Role
		db.Where("name = ?", "admin").First(&role)
		db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID)
	}
	return user
}

func TestSubmitApplication(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "testuser", false)

	req := &model.CreateAuditRequest{
		ResourceName:   "GPT-4 API",
		ResourceType:   "llm",
		APIName:        "gpt-4",
		APIDescription: "GPT-4 large language model",
		Purpose:        "Internal development and testing",
		ExpectedQPS:    10,
		ContactInfo:    "test@example.com",
	}

	ctx := context.Background()
	resp, err := svc.SubmitApplication(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}
	if resp == nil {
		t.Fatal("SubmitApplication returned nil response")
	}
	if resp.ResourceName != req.ResourceName {
		t.Errorf("expected ResourceName %s, got %s", req.ResourceName, resp.ResourceName)
	}
	if resp.Status != model.AuditStatusPending {
		t.Errorf("expected status %d, got %d", model.AuditStatusPending, resp.Status)
	}
	if resp.ApplicantID != user.ID {
		t.Errorf("expected ApplicantID %d, got %d", user.ID, resp.ApplicantID)
	}
	if !resp.CanWithdraw {
		t.Error("expected CanWithdraw to be true for new application")
	}
	if resp.WithdrawRemain <= 0 {
		t.Error("expected WithdrawRemain to be positive")
	}
}

func TestSubmitApplication_UserNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)

	req := &model.CreateAuditRequest{
		ResourceName: "Test API",
		ResourceType: "llm",
		APIName:      "test",
		Purpose:      "test",
	}

	ctx := context.Background()
	_, err := svc.SubmitApplication(ctx, 9999, req)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestReviewApplication_Approve(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "applicant", false)
	admin := createTestUser(db, "admin", true)

	req := &model.CreateAuditRequest{
		ResourceName: "GPT-4 API",
		ResourceType: "llm",
		APIName:      "gpt-4",
		Purpose:      "testing",
		ExpectedQPS:  5,
	}

	ctx := context.Background()
	resp, err := svc.SubmitApplication(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}

	reviewReq := &model.ReviewAuditRequest{
		Approved: true,
		Comment:  "Approved for production use",
	}

	err = svc.ReviewApplication(ctx, admin.ID, resp.ID, reviewReq)
	if err != nil {
		t.Fatalf("ReviewApplication failed: %v", err)
	}

	app, err := svc.GetApplication(resp.ID)
	if err != nil {
		t.Fatalf("GetApplication failed: %v", err)
	}
	if app.Status != model.AuditStatusApproved {
		t.Errorf("expected status %d, got %d", model.AuditStatusApproved, app.Status)
	}
	if app.ReviewerName != admin.Nickname {
		t.Errorf("expected ReviewerName %s, got %s", admin.Nickname, app.ReviewerName)
	}
	if app.ReviewComment != reviewReq.Comment {
		t.Errorf("expected ReviewComment %s, got %s", reviewReq.Comment, app.ReviewComment)
	}
	if app.CanWithdraw {
		t.Error("expected CanWithdraw to be false after review")
	}
}

func TestReviewApplication_Reject(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "applicant2", false)
	admin := createTestUser(db, "admin2", true)

	req := &model.CreateAuditRequest{
		ResourceName: "Test API",
		ResourceType: "llm",
		APIName:      "test-api",
		Purpose:      "test purpose",
	}

	ctx := context.Background()
	resp, err := svc.SubmitApplication(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}

	reviewReq := &model.ReviewAuditRequest{
		Approved: false,
		Comment:  "Insufficient justification",
	}

	err = svc.ReviewApplication(ctx, admin.ID, resp.ID, reviewReq)
	if err != nil {
		t.Fatalf("ReviewApplication failed: %v", err)
	}

	app, err := svc.GetApplication(resp.ID)
	if err != nil {
		t.Fatalf("GetApplication failed: %v", err)
	}
	if app.Status != model.AuditStatusRejected {
		t.Errorf("expected status %d, got %d", model.AuditStatusRejected, app.Status)
	}
}

func TestReviewApplication_NotPending(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "applicant3", false)
	admin := createTestUser(db, "admin3", true)

	req := &model.CreateAuditRequest{
		ResourceName: "Test API",
		ResourceType: "llm",
		APIName:      "test",
		Purpose:      "test",
	}

	ctx := context.Background()
	resp, err := svc.SubmitApplication(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}

	reviewReq := &model.ReviewAuditRequest{Approved: true}
	err = svc.ReviewApplication(ctx, admin.ID, resp.ID, reviewReq)
	if err != nil {
		t.Fatalf("first review failed: %v", err)
	}

	err = svc.ReviewApplication(ctx, admin.ID, resp.ID, reviewReq)
	if err == nil {
		t.Error("expected error when reviewing already reviewed application")
	}
}

func TestWithdrawApplication(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "withdrawuser", false)

	req := &model.CreateAuditRequest{
		ResourceName: "Test API",
		ResourceType: "llm",
		APIName:      "test-withdraw",
		Purpose:      "test withdraw",
	}

	ctx := context.Background()
	resp, err := svc.SubmitApplication(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}

	wdReq := &model.WithdrawAuditRequest{Reason: "Changed my mind"}
	err = svc.WithdrawApplication(ctx, user.ID, resp.ID, wdReq)
	if err != nil {
		t.Fatalf("WithdrawApplication failed: %v", err)
	}

	app, err := svc.GetApplication(resp.ID)
	if err != nil {
		t.Fatalf("GetApplication failed: %v", err)
	}
	if app.Status != model.AuditStatusWithdrawn {
		t.Errorf("expected status %d, got %d", model.AuditStatusWithdrawn, app.Status)
	}
	if app.WithdrawReason != wdReq.Reason {
		t.Errorf("expected WithdrawReason %s, got %s", wdReq.Reason, app.WithdrawReason)
	}
	if app.CanWithdraw {
		t.Error("expected CanWithdraw to be false after withdraw")
	}
}

func TestWithdrawApplication_NotOwner(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user1 := createTestUser(db, "user1", false)
	user2 := createTestUser(db, "user2", false)

	req := &model.CreateAuditRequest{
		ResourceName: "Test API",
		ResourceType: "llm",
		APIName:      "test",
		Purpose:      "test",
	}

	ctx := context.Background()
	resp, err := svc.SubmitApplication(ctx, user1.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}

	wdReq := &model.WithdrawAuditRequest{}
	err = svc.WithdrawApplication(ctx, user2.ID, resp.ID, wdReq)
	if err == nil {
		t.Error("expected error when non-owner tries to withdraw")
	}
}

func TestWithdrawApplication_AfterReview(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "user4", false)
	admin := createTestUser(db, "admin4", true)

	req := &model.CreateAuditRequest{
		ResourceName: "Test API",
		ResourceType: "llm",
		APIName:      "test",
		Purpose:      "test",
	}

	ctx := context.Background()
	resp, err := svc.SubmitApplication(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}

	reviewReq := &model.ReviewAuditRequest{Approved: true}
	err = svc.ReviewApplication(ctx, admin.ID, resp.ID, reviewReq)
	if err != nil {
		t.Fatalf("ReviewApplication failed: %v", err)
	}

	wdReq := &model.WithdrawAuditRequest{}
	err = svc.WithdrawApplication(ctx, user.ID, resp.ID, wdReq)
	if err == nil {
		t.Error("expected error when withdrawing reviewed application")
	}
}

func TestGetMyApplications(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "listuser", false)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		req := &model.CreateAuditRequest{
			ResourceName: "API " + string(rune('A'+i)),
			ResourceType: "llm",
			APIName:      "api-" + string(rune('a'+i)),
			Purpose:      "test",
		}
		_, err := svc.SubmitApplication(ctx, user.ID, req)
		if err != nil {
			t.Fatalf("SubmitApplication %d failed: %v", i, err)
		}
	}

	list, total, err := svc.ListMyApplications(user.ID, 1, 10, nil)
	if err != nil {
		t.Fatalf("GetMyApplications failed: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 items, got %d", len(list))
	}
}

func TestGetAllApplications(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user1 := createTestUser(db, "user_a", false)
	user2 := createTestUser(db, "user_b", false)

	ctx := context.Background()
	for _, u := range []*model.User{user1, user2} {
		req := &model.CreateAuditRequest{
			ResourceName: u.Username + "'s API",
			ResourceType: "llm",
			APIName:      u.Username + "-api",
			Purpose:      "test",
		}
		_, err := svc.SubmitApplication(ctx, u.ID, req)
		if err != nil {
			t.Fatalf("SubmitApplication for %s failed: %v", u.Username, err)
		}
	}

	list, total, err := svc.ListAllApplications(1, 10, nil, "")
	if err != nil {
		t.Fatalf("GetAllApplications failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 items, got %d", len(list))
	}
}

func TestStatusText(t *testing.T) {
	cases := []struct {
		status   int
		expected string
	}{
		{model.AuditStatusPending, "待审核"},
		{model.AuditStatusApproved, "已通过"},
		{model.AuditStatusRejected, "已驳回"},
		{model.AuditStatusWithdrawn, "已撤回"},
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()
	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "statustext", false)

	ctx := context.Background()
	for i, tc := range cases[:1] {
		t.Run(tc.expected, func(t *testing.T) {
			req := &model.CreateAuditRequest{
				ResourceName: "Status Test " + string(rune('A'+i)),
				ResourceType: "llm",
				APIName:      "status-test",
				Purpose:      "test",
			}
			resp, err := svc.SubmitApplication(ctx, user.ID, req)
			if err != nil {
				t.Fatalf("SubmitApplication failed: %v", err)
			}
			if resp.StatusText != tc.expected {
				t.Errorf("status %d: expected %s, got %s", tc.status, tc.expected, resp.StatusText)
			}
		})
	}
}

func TestGetApplication(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "detailuser", false)

	req := &model.CreateAuditRequest{
		ResourceName:   "Detail Test API",
		ResourceType:   "llm",
		APIName:        "detail-api",
		APIDescription: "Detailed API description",
		Purpose:        "Testing detail endpoint",
		ExpectedQPS:    15,
		ContactInfo:    "detail@test.com",
	}

	ctx := context.Background()
	submitResp, err := svc.SubmitApplication(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}

	detail, err := svc.GetApplication(submitResp.ID)
	if err != nil {
		t.Fatalf("GetApplication failed: %v", err)
	}

	if detail.ID != submitResp.ID {
		t.Errorf("expected ID %d, got %d", submitResp.ID, detail.ID)
	}
	if detail.ResourceName != req.ResourceName {
		t.Errorf("expected ResourceName %s, got %s", req.ResourceName, detail.ResourceName)
	}
	if detail.APIName != req.APIName {
		t.Errorf("expected APIName %s, got %s", req.APIName, detail.APIName)
	}
	if detail.Purpose != req.Purpose {
		t.Errorf("expected Purpose %s, got %s", req.Purpose, detail.Purpose)
	}
	if detail.ExpectedQPS != req.ExpectedQPS {
		t.Errorf("expected ExpectedQPS %d, got %d", req.ExpectedQPS, detail.ExpectedQPS)
	}
}

func TestGetApplication_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	svc, _, _ := setupTestAuditService(t, db)

	_, err := svc.GetApplication(99999)
	if err == nil {
		t.Error("expected error for non-existent application")
	}
}

func TestCountPending(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc, _, _ := setupTestAuditService(t, db)
	user := createTestUser(db, "pendinguser", false)
	admin := createTestUser(db, "pendingadmin", true)

	ctx := context.Background()
	count, err := svc.GetPendingCount()
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 pending, got %d", count)
	}

	req := &model.CreateAuditRequest{
		ResourceName: "Pending Test",
		ResourceType: "llm",
		APIName:      "pending-api",
		Purpose:      "test pending count",
	}
	resp, err := svc.SubmitApplication(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("SubmitApplication failed: %v", err)
	}

	count, err = svc.GetPendingCount()
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 pending, got %d", count)
	}

	reviewReq := &model.ReviewAuditRequest{Approved: true}
	err = svc.ReviewApplication(ctx, admin.ID, resp.ID, reviewReq)
	if err != nil {
		t.Fatalf("ReviewApplication failed: %v", err)
	}

	count, err = svc.GetPendingCount()
	if err != nil {
		t.Fatalf("GetPendingCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 pending after review, got %d", count)
	}
}
