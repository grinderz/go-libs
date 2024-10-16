package libzap

type OutputFileConfig struct {
	Dir        string `yaml:"dir"        env:"DIR"         env-default:"logs"`
	TimeLayout string `yaml:"timeLayout" env:"TIME_LAYOUT" env-default:"2006-01-02"`
}

type PresetConfig struct {
	Level             string              `yaml:"level"             env:"LEVEL"               env-default:""`
	DisableCaller     bool                `yaml:"disableCaller"     env:"DISABLE_CALLER"      env-default:"false"`
	DisableStacktrace bool                `yaml:"disableStacktrace" env:"DISABLE_STACKTRACE"  env-default:"false"`
	LevelEncoder      string              `yaml:"levelEncoder"      env:"LEVEL_ENCODER"       env-default:""`
	TimeEncoder       string              `yaml:"timeEncoder"       env:"TIME_ENCODER"        env-default:""`
	TimeLayout        string              `yaml:"timeLayout"        env:"TIME_LAYOUT"         env-default:""`
	DurationEncoder   string              `yaml:"durationEncoder"   env:"DURATION_ENCODER"    env-default:"string"`
	CallerEncoder     string              `yaml:"callerEncoder"     env:"CALLER_ENCODER"      env-default:""`
	Outputs           map[OutputEnum]bool `yaml:"outputs"           env:"OUTPUTS"             env-default:""`
	OutputFile        OutputFileConfig    `yaml:"outputFile"        env-prefix:"OUTPUT_FILE_"`
	JSONTimeKey       string              `yaml:"jsonTimeKey"       env:"JSON_TIME_KEY"       env-default:"ts"`
	JSONLevelKey      string              `yaml:"jsonLevelKey"      env:"JSON_LEVEL_KEY"      env-default:"level"`
	JSONNameKey       string              `yaml:"jsonNameKey"       env:"JSON_NAME_KEY"       env-default:"logger"`
	JSONCallerKey     string              `yaml:"jsonCallerKey"     env:"JSON_CALLER_KEY"     env-default:"caller"`
	JSONFunctionKey   string              `yaml:"jsonFunctionKey"   env:"JSON_FUNCTION_KEY"   env-default:""`
	JSONMessageKey    string              `yaml:"jsonMessageKey"    env:"JSON_MESSAGE_KEY"    env-default:"msg"`
	JSONStacktraceKey string              `yaml:"jsonStacktraceKey" env:"JSON_STACKTRACE_KEY" env-default:"stacktrace"`
	SkipLineEnding    bool                `yaml:"skipLineEnding"    env:"SKIP_LINE_ENDING"    env-default:"false"`
	LineEnding        string              `yaml:"lineEnding"        env:"LINE_ENDING"         env-default:""`
	Encoding          EncodingEnum        `yaml:"encoding"          env:"ENCODING"            env-default:""`
	Development       bool                `yaml:"development"       env:"DEVELOPMENT"         env-default:"false"`
	ConsoleSeparator  string              `yaml:"consoleSeparator"  env:"CONSOLE_SEPARATOR"   env-default:""`
}

type Config struct {
	Preset PresetEnum `yaml:"preset" env:"PRESET" env-default:"production"`

	Development PresetConfig `yaml:"development" env-prefix:"DEVELOPMENT_"`
	Production  PresetConfig `yaml:"production"  env-prefix:"PRODUCTION_"`
}
