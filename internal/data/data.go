package data

import (
	"github.com/topcms/kratos-template/internal/conf"
	"github.com/topcms/kratos-template/internal/data/query"

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
	q     *query.Query // gen 生成的类型安全 Query，通过 query.Use(db) 绑定
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	// 1) init mysql
	db, err := infraMySQL.NewDB(infraMySQL.Config{
		DSN:             c.Database.Source,
		MaxIdleConns:    int(c.Database.MaxIdleConns),
		MaxOpenConns:    int(c.Database.MaxOpenConns),
		ConnMaxLifetime: c.Database.ConnMaxLifetime.AsDuration(),
	})
	if err != nil {
		return nil, nil, err
	}
	// 2) init redis
	rd := infraRedis.NewClient(infraRedis.Config{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})

	d := &Data{
		db:    db,
		redis: rd,
		q:     query.Use(db), // 将 *gorm.DB 绑定到 gen Query，复用同一连接池
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
