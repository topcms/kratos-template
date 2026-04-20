package service

import (
	"context"
	"time"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	v1 "github.com/topcms/kratos-template/api/user/v1"
	"github.com/topcms/kratos-template/internal/biz"
	"github.com/topcms/kratos-template/internal/mapper"
)

// UserService 实现用户查询服务。
type UserService struct {
	v1.UnimplementedUserServiceServer

	uc *biz.UserUseCase
}

// NewUserService new a user service.
func NewUserService(uc *biz.UserUseCase) *UserService {
	return &UserService{uc: uc}
}

// GetUser implements api.user.v1.UserServiceServer.
func (s *UserService) GetUser(ctx context.Context, in *v1.ReqUserDetail) (*v1.RspUserDetail, error) {
	u, err := s.uc.GetUser(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.RspUserDetail{
		User: mapper.ModelUserToProto(u),
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, in *v1.ReqUserCreate) (*v1.RspUserCreate, error) {
	created, err := s.uc.CreateUser(ctx, mapper.ProtoToModelUser(in.GetUser()))
	if err != nil {
		return nil, err
	}
	return &v1.RspUserCreate{
		User: mapper.ModelUserToProto(created),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *v1.ReqUserUpdate) (*v1.RspUserUpdate, error) {
	fields := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if in.UserCode != nil {
		fields["user_code"] = in.GetUserCode()
	}
	if in.UserName != nil {
		fields["user_name"] = in.GetUserName()
	}
	if in.Avatar != nil {
		fields["avatar"] = in.GetAvatar()
	}
	if in.Gender != nil {
		fields["gender"] = in.GetGender()
	}
	if in.Introduction != nil {
		fields["introduction"] = in.GetIntroduction()
	}
	if in.RegType != nil {
		fields["reg_type"] = in.GetRegType()
	}
	if in.GetRegTime() != nil {
		fields["reg_time"] = in.GetRegTime().AsTime()
	}
	if in.RegIp != nil {
		fields["reg_ip"] = in.GetRegIp()
	}
	if in.Country != nil {
		fields["country"] = in.GetCountry()
	}
	if in.Province != nil {
		fields["province"] = in.GetProvince()
	}
	if in.City != nil {
		fields["city"] = in.GetCity()
	}
	if in.Lang != nil {
		fields["lang"] = in.GetLang()
	}
	if in.IsDel != nil {
		fields["is_del"] = in.GetIsDel()
	}

	if len(fields) == 1 {
		return nil, kerrors.BadRequest("INVALID_UPDATE_FIELDS", "no fields to update")
	}

	updated, err := s.uc.UpdateUser(ctx, in.GetUserId(), fields)
	if err != nil {
		return nil, err
	}
	return &v1.RspUserUpdate{
		User: mapper.ModelUserToProto(updated),
	}, nil
}

func (s *UserService) ListUser(ctx context.Context, in *v1.ReqUserList) (*v1.RspUserList, error) {
	users, total, err := s.uc.ListUser(ctx, in.GetPage(), in.GetSize())
	if err != nil {
		return nil, err
	}
	list := make([]*v1.UserInfo, 0, len(users))
	for _, item := range users {
		list = append(list, mapper.ModelUserToProto(item))
	}
	return &v1.RspUserList{
		Total: total,
		List:  list,
	}, nil
}
