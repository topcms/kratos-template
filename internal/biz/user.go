package biz

import (
	"context"
	"fmt"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/topcms/kratos-template/internal/data/model"
)

// UserRepo 仓储接口（由 data 实现）。
type UserRepo interface {
	Get(context.Context, int64) (*model.User, error)
	Create(context.Context, *model.User) (*model.User, error)
	Update(context.Context, int64, map[string]interface{}) (*model.User, error)
	List(context.Context, int64, int64) ([]*model.User, int64, error)
}

// UserUseCase 用户用例层。
type UserUseCase struct {
	repo UserRepo
}

// NewUserUseCase new a User use case.
func NewUserUseCase(repo UserRepo) *UserUseCase {
	return &UserUseCase{repo: repo}
}

// GetUser 根据 ID 查询用户。
func (uc *UserUseCase) GetUser(ctx context.Context, id int64) (*model.User, error) {
	if id <= 0 {
		return nil, kerrors.BadRequest("INVALID_USER_ID", "user id must be greater than 0")
	}
	u, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, kerrors.NotFound("USER_NOT_FOUND", fmt.Sprintf("user %d not found", id))
	}
	return u, nil
}

func (uc *UserUseCase) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, kerrors.BadRequest("INVALID_USER", "user is required")
	}
	if user.UserName == "" {
		return nil, kerrors.BadRequest("INVALID_USER_NAME", "user_name is required")
	}
	return uc.repo.Create(ctx, user)
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, id int64, fields map[string]interface{}) (*model.User, error) {
	if len(fields) == 0 {
		return nil, kerrors.BadRequest("INVALID_USER", "user is required")
	}
	if id <= 0 {
		return nil, kerrors.BadRequest("INVALID_USER_ID", "user id must be greater than 0")
	}
	return uc.repo.Update(ctx, id, fields)
}

func (uc *UserUseCase) ListUser(ctx context.Context, page, size int64) ([]*model.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return uc.repo.List(ctx, page, size)
}
