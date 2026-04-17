package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	commonv1 "github.com/topcms/kratos-template/api/common/v1"
	userv1 "github.com/topcms/kratos-template/api/user/v1"
	"github.com/topcms/kratos-template/internal/biz"
	"github.com/topcms/kratos-template/internal/conf"
	"github.com/topcms/kratos-template/internal/data/model"

	infralogging "github.com/topcms/kratos-infra/middleware/logging"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	goredis "github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type userRepo struct {
	data                  *Data
	log                   *log.Helper
	remoteUserServiceName string
	remoteDialTimeout     time.Duration
	discovery             registry.Discovery
}

const (
	// 用于标记“这是服务间内部拉取”，避免服务自己递归再次触发 discovery 拉取。
	internalDiscoveryFetchMetaKey = "x-internal-discovery-fetch"
	userCacheTTL                  = 5 * time.Minute
)

// NewUserRepo .
func NewUserRepo(data *Data, logger log.Logger, discovery registry.Discovery, c *conf.Data) biz.UserRepo {
	remote := (*conf.RemoteUserService)(nil)
	if c != nil && c.Remote != nil {
		remote = c.Remote.UserService
	}

	var serviceName string
	var dialTimeout time.Duration
	if remote != nil {
		serviceName = remote.ServiceName
		dialTimeout = remote.DialTimeout.AsDuration()
	}
	if dialTimeout <= 0 {
		dialTimeout = 2 * time.Second // fallback：确保 WithTimeout 不会瞬间超时
	}

	return &userRepo{
		data:                  data,
		log:                   log.NewHelper(logger),
		remoteUserServiceName: serviceName,
		remoteDialTimeout:     dialTimeout,
		discovery:             discovery,
	}
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*biz.User, error) {
	// 内部拉取（由 discovery 客户端触发）时，禁止再次发起远程拉取，避免递归。
	if isInternalDiscoveryFetch(ctx) {
		return r.findLocal(ctx, id)
	}

	// 1) redis cache
	key := fmt.Sprintf("user:%d", id)
	if s, err := r.data.redis.Get(ctx, key).Result(); err == nil && s != "" {
		var u biz.User
		if err := json.Unmarshal([]byte(s), &u); err == nil {
			return &u, nil
		}
		r.log.Warnf("redis unmarshal user failed: id=%d", id)
	} else if err != nil && !errors.Is(err, goredis.Nil) {
		r.log.Warnf("redis get failed: id=%d err=%v", id, err)
	}

	// 2) mysql
	u, err := r.findLocalFromDB(ctx, id)
	if err != nil {
		return nil, err
	}
	if u != nil {
		// 回填缓存（demo：固定 TTL）
		r.cacheUser(ctx, key, u)
		return u, nil
	}

	// 3) discovery remote fetch
	if r.discovery == nil || r.remoteUserServiceName == "" {
		return nil, nil
	}

	traceID := infralogging.TraceIDFromContext(ctx)
	outCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"X-Request-Id", traceID,
		internalDiscoveryFetchMetaKey, "true",
	))

	ctxDial, cancel := context.WithTimeout(ctx, r.remoteDialTimeout)
	defer cancel()

	conn, err := kratosgrpc.DialInsecure(
		ctxDial,
		kratosgrpc.WithEndpoint("discovery:///"+r.remoteUserServiceName),
		kratosgrpc.WithDiscovery(r.discovery),
	)
	if err != nil {
		r.log.Warnf("dial discovery user service failed: service=%s err=%v", r.remoteUserServiceName, err)
		return nil, nil
	}
	defer func() { _ = conn.Close() }()

	client := userv1.NewUserServiceClient(conn)
	rsp, err := client.GetUser(outCtx, &userv1.ReqUserDetail{
		Id:       id,
		Metadata: &commonv1.Metadata{TraceId: traceID},
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, nil
		}
		r.log.Warnf("remote GetUser failed: id=%d err=%v", id, err)
		return nil, nil
	}
	if rsp == nil || rsp.User == nil {
		return nil, nil
	}

	u = &biz.User{
		ID:          rsp.User.Id,
		UserCode:    rsp.User.UserCode,
		UserName:    rsp.User.UserName,
		UserType:    rsp.User.UserType,
		RealName:    rsp.User.RealName,
		UserDesc:    rsp.User.UserDesc,
		Email:       rsp.User.Email,
		Gender:      rsp.User.Gender,
		Tel:         rsp.User.Tel,
		Mobile:      rsp.User.Mobile,
		DeptCode:    rsp.User.DeptCode,
		Avatar:      rsp.User.Avatar,
		AvatarThumb: rsp.User.AvatarThumb,
		IsAccount:   rsp.User.IsAccount,
		State:       rsp.User.State,
		CreatedBy:   rsp.User.CreatedBy,
		CreatedAt:   rsp.User.CreatedAt,
		UpdatedBy:   rsp.User.UpdatedBy,
		UpdatedAt:   rsp.User.UpdatedAt,
	}
	r.cacheUser(ctx, key, u)
	return u, nil
}

func isInternalDiscoveryFetch(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	v := md.Get(internalDiscoveryFetchMetaKey)
	return len(v) > 0 && (v[0] == "true" || v[0] == "1")
}

func (r *userRepo) findLocal(ctx context.Context, id int64) (*biz.User, error) {
	// 内部拉取只回源：redis -> mysql，不再触发 remote。
	key := fmt.Sprintf("user:%d", id)
	if s, err := r.data.redis.Get(ctx, key).Result(); err == nil && s != "" {
		var u biz.User
		if err := json.Unmarshal([]byte(s), &u); err == nil {
			return &u, nil
		}
	}
	return r.findLocalFromDB(ctx, id)
}

func (r *userRepo) findLocalFromDB(ctx context.Context, id int64) (*biz.User, error) {
	u := r.data.q.User
	result, err := u.WithContext(ctx).Where(u.ID.Eq(int32(id))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toBizUser(result), nil
}

func (r *userRepo) Create(ctx context.Context, user *biz.User) (*biz.User, error) {
	if user == nil {
		return nil, nil
	}
	entity := toModelUser(user, true)
	if err := r.data.q.User.WithContext(ctx).Create(entity); err != nil {
		return nil, err
	}
	created, err := r.findLocalFromDB(ctx, int64(entity.ID))
	if err != nil {
		return nil, err
	}
	if created != nil {
		r.cacheUser(ctx, fmt.Sprintf("user:%d", created.ID), created)
	}
	return created, nil
}

func (r *userRepo) Update(ctx context.Context, user *biz.User) (*biz.User, error) {
	if user == nil {
		return nil, nil
	}
	entity := toModelUser(user, false)
	u := r.data.q.User
	_, err := u.WithContext(ctx).Where(u.ID.Eq(int32(user.ID))).Updates(entity)
	if err != nil {
		return nil, err
	}
	updated, err := r.findLocalFromDB(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if updated != nil {
		r.cacheUser(ctx, fmt.Sprintf("user:%d", updated.ID), updated)
	}
	return updated, nil
}

func (r *userRepo) List(ctx context.Context, page, size int64) ([]*biz.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	u := r.data.q.User
	do := u.WithContext(ctx)
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	offset := int((page - 1) * size)
	rows, err := do.Order(u.ID.Desc()).Offset(offset).Limit(int(size)).Find()
	if err != nil {
		return nil, 0, err
	}
	list := make([]*biz.User, 0, len(rows))
	for _, row := range rows {
		list = append(list, toBizUser(row))
	}
	return list, total, nil
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

func formatTime(p *time.Time) string {
	if p == nil {
		return ""
	}
	return p.Format(time.RFC3339)
}

func strPtr(v string) *string {
	return &v
}

func int32Ptr(v int32) *int32 {
	return &v
}

func toBizUser(m *model.User) *biz.User {
	if m == nil {
		return nil
	}
	return &biz.User{
		ID:          int64(m.ID),
		UserCode:    derefStr(m.UserCode),
		UserName:    derefStr(m.UserName),
		UserType:    derefStr(m.UserType),
		RealName:    derefStr(m.RealName),
		UserDesc:    derefStr(m.UserDesc),
		Email:       derefStr(m.Email),
		Gender:      derefInt32(m.Gender),
		Tel:         derefStr(m.Tel),
		Mobile:      derefStr(m.Mobile),
		DeptCode:    derefStr(m.DeptCode),
		Avatar:      derefStr(m.Avatar),
		AvatarThumb: derefStr(m.AvatarThumb),
		IsAccount:   derefInt32(m.IsAccount),
		State:       derefInt32(m.State),
		CreatedBy:   derefStr(m.CreatedBy),
		CreatedAt:   formatTime(m.CreatedAt),
		UpdatedBy:   derefStr(m.UpdatedBy),
		UpdatedAt:   formatTime(m.UpdatedAt),
	}
}

func toModelUser(user *biz.User, isCreate bool) *model.User {
	if user == nil {
		return nil
	}
	now := time.Now()
	entity := &model.User{
		ID:          int32(user.ID),
		UserCode:    strPtr(user.UserCode),
		UserName:    strPtr(user.UserName),
		UserType:    strPtr(user.UserType),
		RealName:    strPtr(user.RealName),
		UserDesc:    strPtr(user.UserDesc),
		Email:       strPtr(user.Email),
		Gender:      int32Ptr(user.Gender),
		Tel:         strPtr(user.Tel),
		Mobile:      strPtr(user.Mobile),
		DeptCode:    strPtr(user.DeptCode),
		Avatar:      strPtr(user.Avatar),
		AvatarThumb: strPtr(user.AvatarThumb),
		IsAccount:   int32Ptr(user.IsAccount),
		State:       int32Ptr(user.State),
		UpdatedBy:   strPtr(user.UpdatedBy),
		UpdatedAt:   &now,
	}
	if isCreate {
		entity.CreatedBy = strPtr(user.CreatedBy)
		entity.CreatedAt = &now
	}
	return entity
}

func (r *userRepo) cacheUser(ctx context.Context, key string, u *biz.User) {
	if u == nil {
		return
	}
	b, err := json.Marshal(u)
	if err != nil {
		return
	}
	_ = r.data.redis.Set(ctx, key, string(b), userCacheTTL).Err()
}
