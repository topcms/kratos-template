package mapper

import (
	"time"

	v1 "github.com/topcms/kratos-template/api/user/v1"
	"github.com/topcms/kratos-template/internal/data/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ModelUserToProto converts model.User to proto UserInfo.
func ModelUserToProto(u *model.User) *v1.UserInfo {
	if u == nil {
		return nil
	}
	return &v1.UserInfo{
		UserId:       u.UserID,
		UserCode:     u.UserCode,
		UserName:     u.UserName,
		Avatar:       u.Avatar,
		Gender:       u.Gender,
		Introduction: u.Introduction,
		RegType:      u.RegType,
		RegTime:      timeToProtoTime(u.RegTime),
		RegIp:        u.RegIP,
		Country:      u.Country,
		Province:     u.Province,
		City:         u.City,
		Lang:         u.Lang,
		CreatedAt:    timeToProtoTime(u.CreatedAt),
		UpdatedAt:    timeToProtoTime(u.UpdatedAt),
		IsDel:        u.IsDel,
	}
}

// ProtoToModelUser converts proto UserInfo to model.User.
func ProtoToModelUser(u *v1.UserInfo) *model.User {
	if u == nil {
		return nil
	}
	return &model.User{
		UserID:       u.UserId,
		UserCode:     u.UserCode,
		UserName:     u.UserName,
		Avatar:       u.Avatar,
		Gender:       u.Gender,
		Introduction: u.Introduction,
		RegType:      u.RegType,
		RegTime:      protoTimeToTime(u.RegTime),
		RegIP:        u.RegIp,
		Country:      u.Country,
		Province:     u.Province,
		City:         u.City,
		Lang:         u.Lang,
		CreatedAt:    protoTimeToTime(u.CreatedAt),
		UpdatedAt:    protoTimeToTime(u.UpdatedAt),
		IsDel:        u.IsDel,
	}
}

func timeToProtoTime(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func protoTimeToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
