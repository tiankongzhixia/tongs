package initialize

import (
	"github.com/go-redis/redis"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
	"tongs/config"
	"tongs/global"
	"tongs/tong"
)

func InitDatabase() {
	var orm *gorm.DB
	if global.CONFIG.Server.Database.Pgsql.Host != "" {
		orm = initPostGreSql(global.CONFIG.Server.Database.Pgsql)
	} else if global.CONFIG.Server.Database.Mysql.Host != "" {
		orm = initMysqlGreSql(global.CONFIG.Server.Database.Mysql)
	} else {
		panic("不支持的数据库")
	}
	global.DB = orm
}

func initMysqlGreSql(config config.Mysql) *gorm.DB {
	if config.Host == "" {
		panic("PG数据库连接为空")
	}
	mysqlConfig := mysql.Config{
		DSN:                       config.Dsn(), // DSN data source name
		DefaultStringSize:         191,          // string 类型字段的默认长度
		SkipInitializeWithVersion: false,        // 根据版本自动配置

	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), generateGormConfig(config.Prefix, config.Singular)); err != nil {
		return nil
	} else {
		db.InstanceSet("gorm:table_options", "ENGINE="+config.Engine)
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
		return db
	}
}

func initPostGreSql(config config.Pgsql) *gorm.DB {
	if config.Host == "" {
		panic("PG数据库连接为空")
	}
	pgsqlConfig := postgres.Config{
		DSN:                  config.Dsn(), // DSN data source name
		PreferSimpleProtocol: false,
	}
	orm, err := gorm.Open(postgres.New(pgsqlConfig), generateGormConfig(config.Prefix, config.Singular))
	if err != nil {
		panic("PG数据库连接失败")
	}
	db, _ := orm.DB()

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	db.SetMaxIdleConns(config.MaxIdleConns)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	db.SetMaxOpenConns(config.MaxOpenConns)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	db.SetConnMaxLifetime(time.Hour)
	return orm
}

func InitRedis() {
	rdb := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     global.CONFIG.Server.Redis.Addr,
		Password: global.CONFIG.Server.Redis.Password, // no password set
		DB:       0,                                   // use default DB
	})
	//Tongs的redis
	if global.CONFIG.Tongs.Redis.Addr != "" {
		tong.Redis = redis.NewClient(&redis.Options{
			Network:  "tcp",
			Addr:     global.CONFIG.Tongs.Redis.Addr,
			Password: global.CONFIG.Tongs.Redis.Password, // no password set
			DB:       0,                                  // use default DB
		})
	} else {
		tong.Redis = rdb
	}
	if global.CONFIG.Tongs.Bloom.Open {
		if global.CONFIG.Tongs.Bloom.Redis.Addr != "" {
			tong.BloomRedis = redis.NewClient(&redis.Options{
				Network:  "tcp",
				Addr:     global.CONFIG.Tongs.Bloom.Redis.Addr,
				Password: global.CONFIG.Tongs.Bloom.Redis.Password, // no password set
				DB:       0,                                        // use default DB
			})
		} else {
			tong.BloomRedis = tong.Redis
		}
	}
	global.Redis = rdb
}

// Config gorm 自定义配置
func generateGormConfig(prefix string, singular bool) *gorm.Config {
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   prefix,
			SingularTable: singular,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	return config
}
