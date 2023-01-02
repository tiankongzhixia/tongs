package config

type Server struct {
	Port     string   `mapstructure:"port" json:"port" yaml:"port"`
	Database Database `mapstructure:"database" json:"database" yaml:"database"`
	Redis    Redis    `mapstructure:"redis" json:"redis" yaml:"redis"`
	Zap      Zap      `mapstructure:"zap" json:"zap" yaml:"zap"`
}
