package libmaxprocs

type Config struct {
	Engine EngineEnum `yaml:"engine" env:"ENGINE" env-default:"disabled" env-description:"Engine to use (disabled, auto, direct)."`

	Auto   AutoConfig   `yaml:"auto"   env-prefix:"AUTO__"`
	Direct DirectConfig `yaml:"direct" env-prefix:"DIRECT__"`
}

type AutoConfig struct {
	RuntimeOverhead int `yaml:"runtimeOverhead" env:"RUNTIME_OVERHEAD" env-default:"0" env-description:"Overhead of system threads. This values subtracted from floor rounded CPU quota."`
}

type DirectConfig struct {
	Value int `yaml:"value" env:"VALUE" env-default:"0" env-description:"Set GOMAXPROCS directly."`
}
