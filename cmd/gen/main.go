// 代码生成器，只在开发/CI 阶段运行，不进 server 二进制。
//
// 仅从真实 MySQL 反向生成 model 与 query；表前缀与 DSN 使用 Kratos config 加载（与 cmd/server 一致）。
//
// 用法（在项目根目录执行，与 cmd/server 一致）：
//
//	go run ./cmd/gen -tables ts_user,ts_admin
//	go run ./cmd/gen -tables user,admin   # 短名会拼 data.database.table_prefix
//	go run ./cmd/gen -all
//
// -conf 默认为 configs，与 cmd/server 一致；自定义配置时再传 -conf。
// 连接串优先环境变量 GEN_DSN（覆盖配置文件中的 data.database.source）。
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/topcms/kratos-template/internal/conf"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func resolveTableName(name, prefix string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, prefix) {
		return name
	}
	return prefix + name
}

func main() {
	confPath := flag.String("conf", "configs", "配置路径（目录或文件），与 cmd/server -conf 一致")
	tablesFlag := flag.String("tables", "", "逗号分隔的表名；可写全名 ts_user 或短名 user（会拼 table_prefix）；与 -all 二选一")
	allFlag := flag.Bool("all", false, "生成当前库中全部表；与 -tables 二选一")
	flag.Parse()

	if *tablesFlag != "" && *allFlag {
		fmt.Fprintln(os.Stderr, "gen: -tables 与 -all 不能同时使用")
		os.Exit(2)
	}
	if *tablesFlag == "" && !*allFlag {
		fmt.Fprintln(os.Stderr, "gen: 必须指定 -tables 或 -all")
		fmt.Fprintln(os.Stderr, "示例: go run ./cmd/gen -tables user,admin")
		fmt.Fprintln(os.Stderr, "      go run ./cmd/gen -all")
		os.Exit(2)
	}

	cfgPath := strings.TrimSpace(*confPath)

	// 与 cmd/server 一致仅使用 file 源；勿使用无前缀 env 源，否则会混入大量环境变量导致 Scan 失败。
	c := config.New(
		config.WithSource(
			file.NewSource(cfgPath),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic("gen: 加载配置失败: " + err.Error())
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic("gen: 解析配置失败: " + err.Error())
	}
	if bc.Data == nil || bc.Data.Database == nil {
		panic("gen: 配置缺少 data.database")
	}
	dbCfg := bc.Data.Database

	tablePrefix := strings.TrimSpace(dbCfg.TablePrefix)
	if tablePrefix == "" {
		panic("gen: 请在配置文件中设置 data.database.table_prefix")
	}

	dsn := strings.TrimSpace(os.Getenv("GEN_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(dbCfg.Source)
	}
	if dsn == "" {
		panic("gen: 请设置环境变量 GEN_DSN，或在配置文件中填写 data.database.source")
	}

	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic("gen: 连接数据库失败: " + err.Error())
	}

	genCfg := gen.Config{
		OutPath:      "internal/data/query",
		ModelPkgPath: "model",

		Mode: gen.WithDefaultQuery | gen.WithQueryInterface,

		FieldNullable:     true,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: false,
		FieldWithTypeTag:  true,
	}
	genCfg.WithModelNameStrategy(func(tableName string) string {
		base := strings.TrimPrefix(tableName, tablePrefix)
		if base == "" {
			return tableName
		}
		r, size := utf8.DecodeRuneInString(base)
		if r == utf8.RuneError {
			return base
		}
		return string(unicode.ToUpper(r)) + base[size:]
	})
	genCfg.WithFileNameStrategy(func(tableName string) string {
		return strings.TrimPrefix(tableName, tablePrefix)
	})

	g := gen.NewGenerator(genCfg)
	g.UseDB(db)

	if *allFlag {
		models := g.GenerateAllTable()
		g.ApplyBasic(models...)
	} else {
		parts := strings.Split(*tablesFlag, ",")
		var metas []interface{}
		for _, p := range parts {
			tn := resolveTableName(p, tablePrefix)
			if tn == "" {
				continue
			}
			metas = append(metas, g.GenerateModel(tn))
		}
		if len(metas) == 0 {
			panic("gen: -tables 解析后没有有效表名")
		}
		g.ApplyBasic(metas...)
	}

	g.Execute()
}
