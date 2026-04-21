package mapper

import (
	"errors"
	"time"

	v1 "github.com/topcms/kratos-template/api/user/v1"
	"github.com/topcms/kratos-template/internal/biz"
	"github.com/topcms/kratos-template/internal/data/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ModelUserToBiz maps persistence model to domain entity.
func ModelUserToBiz(m *model.User) *biz.User {
	if m == nil {
		return nil
	}
	return &biz.User{
		UserID:       m.UserID,
		UserCode:     m.UserCode,
		UserName:     m.UserName,
		Avatar:       m.Avatar,
		Gender:       m.Gender,
		Introduction: m.Introduction,
		RegType:      m.RegType,
		RegTime:      m.RegTime,
		RegIP:        m.RegIP,
		Country:      m.Country,
		Province:     m.Province,
		City:         m.City,
		Lang:         m.Lang,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		IsDel:        m.IsDel,
	}
}

// BizUserToModel maps domain entity to persistence model.
func BizUserToModel(b *biz.User) *model.User {
	if b == nil {
		return nil
	}
	return &model.User{
		UserID:       b.UserID,
		UserCode:     b.UserCode,
		UserName:     b.UserName,
		Avatar:       b.Avatar,
		Gender:       b.Gender,
		Introduction: b.Introduction,
		RegType:      b.RegType,
		RegTime:      b.RegTime,
		RegIP:        b.RegIP,
		Country:      b.Country,
		Province:     b.Province,
		City:         b.City,
		Lang:         b.Lang,
		CreatedAt:    b.CreatedAt,
		UpdatedAt:    b.UpdatedAt,
		IsDel:        b.IsDel,
	}
}

// BizUserToProto converts biz.User to proto UserInfo.
func BizUserToProto(u *biz.User) *v1.UserInfo {
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

// ProtoToBizUser converts proto UserInfo to biz.User.
func ProtoToBizUser(u *v1.UserInfo) *biz.User {
	if u == nil {
		return nil
	}
	return &biz.User{
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

// ModelUserToProto converts model.User to proto UserInfo.
func ModelUserToProto(u *model.User) *v1.UserInfo {
	return BizUserToProto(ModelUserToBiz(u))
}

// ProtoToModelUser converts proto UserInfo to model.User.
func ProtoToModelUser(u *v1.UserInfo) *model.User {
	return BizUserToModel(ProtoToBizUser(u))
}

// ErrNoFieldsToUpdate 表示除 updated_at 外没有任何可持久化的字段。
var ErrNoFieldsToUpdate = errors.New("no fields to update")

// ReqUserUpdateToDBFields 将更新请求转为 GORM Updates 使用的列名映射（含 updated_at）。
func ReqUserUpdateToDBFields(in *v1.ReqUserUpdate) (map[string]interface{}, error) {
	if in == nil {
		return nil, ErrNoFieldsToUpdate
	}
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
		return nil, ErrNoFieldsToUpdate
	}
	return fields, nil
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
