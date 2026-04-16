package service

import (
	"context"

	commonv1 "github.com/topcms/kratos-template/api/common/v1"
	v1 "github.com/topcms/kratos-template/api/user/v1"
	"github.com/topcms/kratos-template/internal/biz"
)

// UserService 实现用户查询服务。
type UserService struct {
	v1.UnimplementedUserServiceServer

	uc *biz.UserUsecase
}

// NewUserService new a user service.
func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

// GetUser implements api.user.v1.UserServiceServer.
func (s *UserService) GetUser(ctx context.Context, in *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	u, err := s.uc.GetUser(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.GetUserResponse{
		Meta: &commonv1.Reply{
			Code:    0,
			Message: "ok",
		},
		User: &v1.User{
			Id:     u.ID,
			Name:   u.Name,
			Avatar: u.Avatar,
		},
	}, nil
}
