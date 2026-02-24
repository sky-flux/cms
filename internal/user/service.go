package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/sky-flux/cms/internal/pkg/mail"
)

type Service struct {
	repo         UserRepository
	roleRepo     RoleRepository
	urRepo       UserRoleRepository
	tokenRevoker TokenRevoker
	auditLog     audit.Logger
	mailer       mail.Sender
	siteName     string
}

func NewService(
	repo UserRepository,
	roleRepo RoleRepository,
	urRepo UserRoleRepository,
	tokenRevoker TokenRevoker,
	auditLog audit.Logger,
	mailer mail.Sender,
	siteName string,
) *Service {
	return &Service{
		repo:         repo,
		roleRepo:     roleRepo,
		urRepo:       urRepo,
		tokenRevoker: tokenRevoker,
		auditLog:     auditLog,
		mailer:       mailer,
		siteName:     siteName,
	}
}

func (s *Service) ListUsers(ctx context.Context, f ListFilter) ([]UserResp, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}

	users, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	resps := make([]UserResp, len(users))
	for i := range users {
		roleSlug, _ := s.urRepo.GetRoleSlug(ctx, users[i].ID)
		resps[i] = ToUserResp(&users[i], roleSlug)
	}
	return resps, total, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*UserResp, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	roleSlug, _ := s.urRepo.GetRoleSlug(ctx, user.ID)
	resp := ToUserResp(user, roleSlug)
	return &resp, nil
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserReq) (*UserResp, error) {
	role, err := s.roleRepo.GetBySlug(ctx, req.Role)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, apperror.Conflict("email already exists", nil)
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, apperror.Internal("hash password failed", err)
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: hash,
		DisplayName:  req.DisplayName,
		Status:       model.UserStatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := s.urRepo.Assign(ctx, user.ID, role.ID); err != nil {
		return nil, fmt.Errorf("assign role to new user: %w", err)
	}

	// Send welcome email asynchronously.
	go func() {
		html, err := mail.RenderWelcome(s.siteName, req.Email, req.Password)
		if err != nil {
			slog.Error("render welcome email failed", "error", err, "email", req.Email)
			return
		}
		msg := mail.Message{
			To:      req.Email,
			Subject: "Welcome to " + s.siteName,
			HTML:    html,
		}
		if err := s.mailer.Send(context.Background(), msg); err != nil {
			slog.Error("send welcome email failed", "error", err, "email", req.Email)
		}
	}()

	resp := ToUserResp(user, req.Role)
	return &resp, nil
}

func (s *Service) UpdateUser(ctx context.Context, id string, req *UpdateUserReq) (*UserResp, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := user.Status

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	// Handle role change.
	roleSlug := ""
	if req.Role != nil {
		role, err := s.roleRepo.GetBySlug(ctx, *req.Role)
		if err != nil {
			return nil, err
		}
		if err := s.urRepo.Assign(ctx, user.ID, role.ID); err != nil {
			return nil, fmt.Errorf("update user role: %w", err)
		}
		roleSlug = *req.Role
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	// If status changed to disabled, revoke tokens and send notification.
	if req.Status != nil && *req.Status == model.UserStatusDisabled && oldStatus != model.UserStatusDisabled {
		go func() {
			html, err := mail.RenderDisabled(s.siteName, user.Email)
			if err != nil {
				slog.Error("render disabled email failed", "error", err, "email", user.Email)
				return
			}
			msg := mail.Message{
				To:      user.Email,
				Subject: "Account Disabled",
				HTML:    html,
			}
			if err := s.mailer.Send(context.Background(), msg); err != nil {
				slog.Error("send disabled email failed", "error", err, "email", user.Email)
			}
			if err := s.tokenRevoker.RevokeAllForUser(context.Background(), user.ID); err != nil {
				slog.Error("revoke tokens after disable failed", "error", err, "user_id", user.ID)
			}
		}()
	}

	// Audit log the update.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "user",
		ResourceID:   user.ID,
	}); err != nil {
		slog.Error("audit log user update failed", "error", err)
	}

	if roleSlug == "" {
		roleSlug, _ = s.urRepo.GetRoleSlug(ctx, user.ID)
	}

	resp := ToUserResp(user, roleSlug)
	return &resp, nil
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	// Self-delete check.
	callerID, _ := ctx.Value("user_id").(string)
	if callerID == id {
		return apperror.Forbidden("cannot delete yourself", nil)
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if deleting the last active super admin.
	roleSlug, _ := s.urRepo.GetRoleSlug(ctx, user.ID)
	if roleSlug == "super" {
		count, err := s.countActiveSuperAdmins(ctx)
		if err != nil {
			return fmt.Errorf("delete user count super: %w", err)
		}
		if count <= 1 {
			return apperror.Forbidden("cannot delete the last super admin", nil)
		}
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if err := s.tokenRevoker.RevokeAllForUser(ctx, id); err != nil {
		slog.Error("revoke tokens after delete failed", "error", err, "user_id", id)
	}

	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "user",
		ResourceID:   id,
	}); err != nil {
		slog.Error("audit log user delete failed", "error", err)
	}

	return nil
}

// countActiveSuperAdmins counts active users with the "super" role.
func (s *Service) countActiveSuperAdmins(ctx context.Context) (int64, error) {
	return s.urRepo.CountActiveByRoleSlug(ctx, "super")
}
