package data

import (
	"context"

	"github.com/topcms/kratos-template/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo .
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*biz.User, error) {
	_ = ctx
	// 模板示例：用内存映射模拟数据源，后续可替换为 DB/Redis 查询。
	mockUsers := map[int64]*biz.User{
		1: {ID: 1, Name: "Alice", Avatar: "https://example.com/avatar/alice.png"},
		2: {ID: 2, Name: "Bob", Avatar: "https://example.com/avatar/bob.png"},
	}
	u, ok := mockUsers[id]
	if !ok {
		r.log.Warnf("FindByID user not found: %d", id)
		return nil, nil
	}
	r.log.Infof("FindByID success: id=%d name=%s", id, u.Name)
	return u, nil
}
