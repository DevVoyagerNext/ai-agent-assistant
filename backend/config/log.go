package config

type Log struct {
	Enabled           bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Level             string   `mapstructure:"level" json:"level" yaml:"level"`
	Format            string   `mapstructure:"format" json:"format" yaml:"format"`
	OutputPaths       []string `mapstructure:"output-paths" json:"outputPaths" yaml:"output-paths"`
	ErrorOutputPaths  []string `mapstructure:"error-output-paths" json:"errorOutputPaths" yaml:"error-output-paths"`
	DisableCaller     bool     `mapstructure:"disable-caller" json:"disableCaller" yaml:"disable-caller"`
	DisableStacktrace bool     `mapstructure:"disable-stacktrace" json:"disableStacktrace" yaml:"disable-stacktrace"`
}
