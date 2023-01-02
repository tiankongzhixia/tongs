package config

type Tongs struct {
	Ua        []UserAgent `json:"ua,omitempty" yaml:"ua" mapstructure:"ua"`                      //ua列表
	AutoUa    bool        `json:"auto-ua" yaml:"auto-ua" mapstructure:"auto-ua"`                 //自动设置ua
	AutoDelay bool        `json:"auto-delay" yaml:"auto-delay" mapstructure:"auto-delay"`        //自动设置随机delay
	Save      Save        `json:"save" yaml:"save" mapstructure:"save"`                          //设置保存item
	MaxDepth  int         `json:"max-depth,omitempty" yaml:"max-depth" mapstructure:"max-depth"` //最大深度
	Bloom     Bloom       `json:"bloom,omitempty" yaml:"bloom" mapstructure:"bloom"`             //布隆过滤器
	Redis     Redis       `mapstructure:"redis" json:"redis" yaml:"redis"`                       //存储请求、队列等信息的redis客户端 为空则向上查找
}

// UserAgent 请求头
type UserAgent struct {
	Label  string   `json:"label,omitempty" yaml:"label" mapstructure:"label"`   //分组名称
	Values []string `json:"values,omitempty" yaml:"label" mapstructure:"values"` //组内所有请求头
}

type Bloom struct {
	Open  bool  `json:"open" yaml:"open" mapstructure:"open"`              //开启redis布隆过滤器
	Alone bool  `json:"alone,omitempty" yaml:"alone" mapstructure:"alone"` //true: 每个Task独立使用过滤器 false: 一个Tongs内的Task使用一个
	Redis Redis `mapstructure:"redis" json:"redis" yaml:"redis"`           //支持布隆过滤器的redis客户端 为空则向上查找
}

type Save struct {
	Open  bool `json:"open,omitempty" yml:"open" mapstructure:"open"`    //是否开启保存内存 开启后默认强制Count为True
	Count bool `json:"count,omitempty" yml:"count" mapstructure:"count"` //是否开启计数
}
