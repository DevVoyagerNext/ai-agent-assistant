package config

type AI struct {
	BaseURL string `mapstructure:"base-url" json:"baseUrl" yaml:"base-url"`
	APIKey  string `mapstructure:"api-key" json:"apiKey" yaml:"api-key"`
	Model   string `mapstructure:"model" json:"model" yaml:"model"`
}
