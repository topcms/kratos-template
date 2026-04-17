package service

import (
	"context"
	"time"

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
func (s *UserService) GetUser(ctx context.Context, in *v1.ReqUserDetail) (*v1.RspUserDetail, error) {
	u, err := s.uc.GetUser(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.RspUserDetail{
		Meta: &commonv1.Reply{
			Code:    0,
			Message: "ok",
		},
		User: toProtoUserInfo(u),
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, in *v1.ReqUserCreate) (*v1.RspUserCreate, error) {
	created, err := s.uc.CreateUser(ctx, toBizUser(in.GetUser()))
	if err != nil {
		return nil, err
	}
	return &v1.RspUserCreate{
		Meta: &commonv1.Reply{
			Code:    0,
			Message: "ok",
		},
		User: toProtoUserInfo(created),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *v1.ReqUserUpdate) (*v1.RspUserUpdate, error) {
	updated, err := s.uc.UpdateUser(ctx, toBizUser(in.GetUser()))
	if err != nil {
		return nil, err
	}
	return &v1.RspUserUpdate{
		Meta: &commonv1.Reply{
			Code:    0,
			Message: "ok",
		},
		User: toProtoUserInfo(updated),
	}, nil
}

func (s *UserService) ListUser(ctx context.Context, in *v1.ReqUserList) (*v1.RspUserList, error) {
	users, total, err := s.uc.ListUser(ctx, in.GetPage(), in.GetSize())
	if err != nil {
		return nil, err
	}
	list := make([]*v1.UserInfo, 0, len(users))
	for _, item := range users {
		list = append(list, toProtoUserInfo(item))
	}
	return &v1.RspUserList{
		Meta: &commonv1.Reply{
			Code:    0,
			Message: "ok",
		},
		Total: total,
		List:  list,
	}, nil
}

func toProtoUserInfo(u *biz.User) *v1.UserInfo {
	if u == nil {
		return nil
	}
	return &v1.UserInfo{
		Id:          u.ID,
		UserCode:    u.UserCode,
		UserName:    u.UserName,
		UserType:    u.UserType,
		RealName:    u.RealName,
		UserDesc:    u.UserDesc,
		Email:       u.Email,
		Gender:      u.Gender,
		Tel:         u.Tel,
		Mobile:      u.Mobile,
		DeptCode:    u.DeptCode,
		Avatar:      u.Avatar,
		AvatarThumb: u.AvatarThumb,
		IsAccount:   u.IsAccount,
		State:       u.State,
		CreatedBy:   u.CreatedBy,
		CreatedAt:   defaultNow(u.CreatedAt),
		UpdatedBy:   u.UpdatedBy,
		UpdatedAt:   defaultNow(u.UpdatedAt),
	}
}

func toBizUser(u *v1.UserInfo) *biz.User {
	if u == nil {
		return nil
	}
	return &biz.User{
		ID:          u.Id,
		UserCode:    u.UserCode,
		UserName:    u.UserName,
		UserType:    u.UserType,
		RealName:    u.RealName,
		UserDesc:    u.UserDesc,
		Email:       u.Email,
		Gender:      u.Gender,
		Tel:         u.Tel,
		Mobile:      u.Mobile,
		DeptCode:    u.DeptCode,
		Avatar:      u.Avatar,
		AvatarThumb: u.AvatarThumb,
		IsAccount:   u.IsAccount,
		State:       u.State,
		CreatedBy:   u.CreatedBy,
		CreatedAt:   u.CreatedAt,
		UpdatedBy:   u.UpdatedBy,
		UpdatedAt:   u.UpdatedAt,
	}
}

func defaultNow(v string) string {
	if v != "" {
		return v
	}
	return time.Now().Format(time.RFC3339)
}
