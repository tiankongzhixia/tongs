package config

type Redis struct {
	Addr     string `json:"addr" yaml:"addr" mapstructure:"addr"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	MaxIdl   int    `json:"max-idl" yaml:"max-idl" mapstructure:"max-idl"`
}
