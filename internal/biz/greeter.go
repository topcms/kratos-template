package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Greeter 领域模型示例。
type Greeter struct {
	Hello string
}

// GreeterRepo 仓储接口（由 data 实现）。
type GreeterRepo interface {
	Save(context.Context, *Greeter) (*Greeter, error)
	Update(context.Context, *Greeter) (*Greeter, error)
	FindByID(context.Context, int64) (*Greeter, error)
	ListByHello(context.Context, string) ([]*Greeter, error)
	ListAll(context.Context) ([]*Greeter, error)
}

// GreeterUsecase 用例层。
type GreeterUsecase struct {
	repo GreeterRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewGreeterUsecase(repo GreeterRepo) *GreeterUsecase {
	return &GreeterUsecase{repo: repo}
}

// CreateGreeter 创建 Greeter 示例逻辑。
func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, g *Greeter) (*Greeter, error) {
	log.Infof("CreateGreeter: %v", g.Hello)
	return uc.repo.Save(ctx, g)
}
