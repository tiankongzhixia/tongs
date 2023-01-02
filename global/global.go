package global

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"tongs/config"
	"tongs/tong"
)

var (
	DB           *gorm.DB
	Redis        *redis.Client
	Log          *zap.Logger
	CONFIG       config.Config
	VP           *viper.Viper
	TongsManager = &tong.Manager{}
)
