package config

type JWT struct {
	SigningKey         string `mapstructure:"signing-key" json:"signing-key" yaml:"signing-key"`                            // JWT签名
	ExpiresTime        int64  `mapstructure:"expires-time" json:"expires-time" yaml:"expires-time"`                         // 短Token过期时间(秒)
	RefreshExpiresTime int64  `mapstructure:"refresh-expires-time" json:"refresh-expires-time" yaml:"refresh-expires-time"` // 长Token过期时间(秒)
	BufferTime         int64  `mapstructure:"buffer-time" json:"buffer-time" yaml:"buffer-time"`                            // 缓冲时间
	Issuer             string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`                                           // 签发者
}
