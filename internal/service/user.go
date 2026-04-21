package service

import (
	"context"
	"errors"

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
		User: mapper.BizUserToProto(u),
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, in *v1.ReqUserCreate) (*v1.RspUserCreate, error) {
	created, err := s.uc.CreateUser(ctx, mapper.ProtoToBizUser(in.GetUser()))
	if err != nil {
		return nil, err
	}
	return &v1.RspUserCreate{
		User: mapper.BizUserToProto(created),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *v1.ReqUserUpdate) (*v1.RspUserUpdate, error) {
	fields, err := mapper.ReqUserUpdateToDBFields(in)
	if err != nil {
		if errors.Is(err, mapper.ErrNoFieldsToUpdate) {
			return nil, kerrors.BadRequest("INVALID_UPDATE_FIELDS", "no fields to update")
		}
		return nil, err
	}

	updated, err := s.uc.UpdateUser(ctx, in.GetUserId(), fields)
	if err != nil {
		return nil, err
	}
	return &v1.RspUserUpdate{
		User: mapper.BizUserToProto(updated),
	}, nil
}

func (s *UserService) ListUser(ctx context.Context, in *v1.ReqUserList) (*v1.RspUserList, error) {
	users, total, err := s.uc.ListUser(ctx, in.GetPage(), in.GetSize())
	if err != nil {
		return nil, err
	}
	list := make([]*v1.UserInfo, 0, len(users))
	for _, item := range users {
		list = append(list, mapper.BizUserToProto(item))
	}
	return &v1.RspUserList{
		Total: total,
		List:  list,
	}, nil
}
