package config

type Qiniu struct {
	AccessKey string `mapstructure:"access-key" json:"accessKey" yaml:"access-key"`
	SecretKey string `mapstructure:"secret-key" json:"secretKey" yaml:"secret-key"`
	Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket"`
	Domain    string `mapstructure:"domain" json:"domain" yaml:"domain"`
	Zone      string `mapstructure:"zone" json:"zone" yaml:"zone"`
	UseHTTPS  bool   `mapstructure:"use-https" json:"useHttps" yaml:"use-https"`
}
