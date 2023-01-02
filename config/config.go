package config

type Config struct {
	Server Server `mapstructure:"server" json:"server" yaml:"server"`
	Tongs  Tongs  `mapstructure:"tongs" json:"tongs" yaml:"tongs"`
}
