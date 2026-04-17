package biz

import (
	"context"
	"fmt"

	kerrors "github.com/go-kratos/kratos/v2/errors"
)

// User 用户领域模型。
type User struct {
	ID          int64
	UserCode    string
	UserName    string
	UserType    string
	RealName    string
	UserDesc    string
	Email       string
	Gender      int32
	Tel         string
	Mobile      string
	DeptCode    string
	Avatar      string
	AvatarThumb string
	IsAccount   int32
	State       int32
	CreatedBy   string
	CreatedAt   string
	UpdatedBy   string
	UpdatedAt   string
}

// UserRepo 仓储接口（由 data 实现）。
type UserRepo interface {
	FindByID(context.Context, int64) (*User, error)
	Create(context.Context, *User) (*User, error)
	Update(context.Context, *User) (*User, error)
	List(context.Context, int64, int64) ([]*User, int64, error)
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

func (uc *UserUsecase) CreateUser(ctx context.Context, user *User) (*User, error) {
	if user == nil {
		return nil, kerrors.BadRequest("INVALID_USER", "user is required")
	}
	if user.UserName == "" {
		return nil, kerrors.BadRequest("INVALID_USER_NAME", "user_name is required")
	}
	return uc.repo.Create(ctx, user)
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, user *User) (*User, error) {
	if user == nil {
		return nil, kerrors.BadRequest("INVALID_USER", "user is required")
	}
	if user.ID <= 0 {
		return nil, kerrors.BadRequest("INVALID_USER_ID", "user id must be greater than 0")
	}
	return uc.repo.Update(ctx, user)
}

func (uc *UserUsecase) ListUser(ctx context.Context, page, size int64) ([]*User, int64, error) {
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
