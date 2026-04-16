package data

import (
	"github.com/topcms/kratos-template/internal/conf"

	infraMySQL "github.com/topcms/kratos-infra/mysql"
	infraRedis "github.com/topcms/kratos-infra/redis"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo)

// Data 聚合数据访问依赖（数据库、缓存客户端等可在此扩展）。
type Data struct {
	db    *gorm.DB
	redis *goredis.Client
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	// 1) init mysql
	db, err := infraMySQL.NewDB(infraMySQL.Config{
		DSN:             c.Database.Source,
		MaxIdleConns:    c.Database.MaxIdleConns,
		MaxOpenConns:    c.Database.MaxOpenConns,
		ConnMaxLifetime: c.Database.ConnMaxLifetime,
	})
	if err != nil {
		return nil, nil, err
	}
	// 2) init redis
	rd := infraRedis.NewClient(infraRedis.Config{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           c.Redis.DB,
		DialTimeout:  c.Redis.DialTimeout,
		ReadTimeout:  c.Redis.ReadTimeout,
		WriteTimeout: c.Redis.WriteTimeout,
	})

	// 3) migrate (demo：仅包含 user 表)
	if err := db.AutoMigrate(&userModel{}); err != nil {
		return nil, nil, err
	}

	d := &Data{
		db:    db,
		redis: rd,
	}

	cleanup := func() {
		log.Info("closing the data resources")
		if d.db != nil {
			sqlDB, err := d.db.DB()
			if err == nil {
				_ = sqlDB.Close()
			}
		}
		if d.redis != nil {
			_ = d.redis.Close()
		}
	}
	return d, cleanup, nil
}
