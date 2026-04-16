package biz

import (
	"context"
	"fmt"

	kerrors "github.com/go-kratos/kratos/v2/errors"
)

// User 用户领域模型。
type User struct {
	ID     int64
	Name   string
	Avatar string
}

// UserRepo 仓储接口（由 data 实现）。
type UserRepo interface {
	FindByID(context.Context, int64) (*User, error)
}

// UserUsecase 用户用例层。
type UserUsecase struct {
	repo UserRepo
}

// NewUserUsecase new a User usecase.
func NewUserUsecase(repo UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// GetUser 根据 ID 查询用户。
func (uc *UserUsecase) GetUser(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, kerrors.BadRequest("INVALID_USER_ID", "user id must be greater than 0")
	}
	u, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, kerrors.NotFound("USER_NOT_FOUND", fmt.Sprintf("user %d not found", id))
	}
	return u, nil
}
