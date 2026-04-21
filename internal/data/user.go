package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	userv1 "github.com/topcms/kratos-template/api/user/v1"
	"github.com/topcms/kratos-template/internal/biz"
	"github.com/topcms/kratos-template/internal/conf"
	"github.com/topcms/kratos-template/internal/data/model"
	"github.com/topcms/kratos-template/internal/mapper"

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

func (r *userRepo) Get(ctx context.Context, id int64) (*biz.User, error) {
	// 内部拉取（由 discovery 客户端触发）时，禁止再次发起远程拉取，避免递归。
	if isInternalDiscoveryFetch(ctx) {
		u, err := r.findLocal(ctx, id)
		return mapper.ModelUserToBiz(u), err
	}

	// 1) redis cache
	key := fmt.Sprintf("user:%d", id)
	if s, err := r.data.redis.Get(ctx, key).Result(); err == nil && s != "" {
		var u model.User
		if err := json.Unmarshal([]byte(s), &u); err == nil {
			return mapper.ModelUserToBiz(&u), nil
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
		return mapper.ModelUserToBiz(u), nil
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
		Id: id,
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

	u = mapper.ProtoToModelUser(rsp.User)
	r.cacheUser(ctx, key, u)
	return mapper.ModelUserToBiz(u), nil
}

func isInternalDiscoveryFetch(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	v := md.Get(internalDiscoveryFetchMetaKey)
	return len(v) > 0 && (v[0] == "true" || v[0] == "1")
}

func (r *userRepo) findLocal(ctx context.Context, id int64) (*model.User, error) {
	// 内部拉取只回源：redis -> mysql，不再触发 remote。
	key := fmt.Sprintf("user:%d", id)
	if s, err := r.data.redis.Get(ctx, key).Result(); err == nil && s != "" {
		var u model.User
		if err := json.Unmarshal([]byte(s), &u); err == nil {
			return &u, nil
		}
	}
	return r.findLocalFromDB(ctx, id)
}

func (r *userRepo) findLocalFromDB(ctx context.Context, id int64) (*model.User, error) {
	u := r.data.q.User
	result, err := u.WithContext(ctx).Where(u.UserID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *userRepo) Create(ctx context.Context, user *biz.User) (*biz.User, error) {
	if user == nil {
		return nil, nil
	}

	m := mapper.BizUserToModel(user)
	if err := r.data.q.User.WithContext(ctx).Create(m); err != nil {
		return nil, err
	}
	created, err := r.findLocalFromDB(ctx, m.UserID)
	if err != nil {
		return nil, err
	}
	if created != nil {
		r.cacheUser(ctx, fmt.Sprintf("user:%d", created.UserID), created)
	}
	return mapper.ModelUserToBiz(created), nil
}

func (r *userRepo) Update(ctx context.Context, id int64, fields map[string]interface{}) (*biz.User, error) {
	if id <= 0 {
		return nil, nil
	}
	if len(fields) == 0 {
		return nil, nil
	}
	u := r.data.q.User
	_, err := u.WithContext(ctx).Where(u.UserID.Eq(id)).Updates(fields)
	if err != nil {
		return nil, err
	}
	updated, err := r.findLocalFromDB(ctx, id)
	if err != nil {
		return nil, err
	}
	if updated != nil {
		r.cacheUser(ctx, fmt.Sprintf("user:%d", updated.UserID), updated)
	}
	return mapper.ModelUserToBiz(updated), nil
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
	rows, err := do.Order(u.UserID.Desc()).Offset(offset).Limit(int(size)).Find()
	if err != nil {
		return nil, 0, err
	}
	list := make([]*biz.User, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapper.ModelUserToBiz(row))
	}
	return list, total, nil
}

func (r *userRepo) cacheUser(ctx context.Context, key string, u *model.User) {
	if u == nil {
		return
	}
	b, err := json.Marshal(u)
	if err != nil {
		return
	}
	_ = r.data.redis.Set(ctx, key, string(b), userCacheTTL).Err()
}
