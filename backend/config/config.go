package config

type Server struct {
	Mysql Mysql `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	JWT   JWT   `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Redis Redis `mapstructure:"redis" json:"redis" yaml:"redis"`
	Email Email `mapstructure:"email" json:"email" yaml:"email"`
	Log   Log   `mapstructure:"log" json:"log" yaml:"log"`
	AI    AI    `mapstructure:"ai" json:"ai" yaml:"ai"`
}
