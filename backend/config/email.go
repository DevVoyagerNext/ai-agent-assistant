package config

type Email struct {
	From     string `mapstructure:"from" json:"from" yaml:"from"`             // 发送人邮箱
	Nickname string `mapstructure:"nickname" json:"nickname" yaml:"nickname"` // 发送人昵称
	Secret   string `mapstructure:"secret" json:"secret" yaml:"secret"`       // 授权码
	Host     string `mapstructure:"host" json:"host" yaml:"host"`             // SMTP服务器地址
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`             // 端口
	IsSSL    bool   `mapstructure:"is-ssl" json:"is-ssl" yaml:"is-ssl"`       // 是否使用SSL
}
