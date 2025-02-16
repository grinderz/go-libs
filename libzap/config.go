package libzap

type OutputFileConfig struct {
	Dir        string `yaml:"dir"        env:"DIR"         env-default:"logs"       env-description:"Set the output dir for logs."`
	TimeLayout string `yaml:"timeLayout" env:"TIME_LAYOUT" env-default:"2006-01-02" env-description:"Set the time layout for file name (appID-time.log)."`
}

type PresetConfig struct {
	Level             string              `yaml:"level"             env:"LEVEL"               env-default:""           env-description:"Set the log level."`
	DisableCaller     bool                `yaml:"disableCaller"     env:"DISABLE_CALLER"      env-default:"false"      env-description:"Don't log the callers info."`
	DisableStacktrace bool                `yaml:"disableStacktrace" env:"DISABLE_STACKTRACE"  env-default:"false"      env-description:"Don't log the stack trace info."`
	LevelEncoder      string              `yaml:"levelEncoder"      env:"LEVEL_ENCODER"       env-default:""           env-description:"Override the level encoder."`
	TimeEncoder       string              `yaml:"timeEncoder"       env:"TIME_ENCODER"        env-default:""           env-description:"Override the time encoder."`
	TimeLayout        string              `yaml:"timeLayout"        env:"TIME_LAYOUT"         env-default:""           env-description:"Override the time layout."`
	DurationEncoder   string              `yaml:"durationEncoder"   env:"DURATION_ENCODER"    env-default:"string"     env-description:"Set the duration encoder."`
	CallerEncoder     string              `yaml:"callerEncoder"     env:"CALLER_ENCODER"      env-default:""           env-description:"Override the caller encoder."`
	Outputs           map[OutputEnum]bool `yaml:"outputs"           env:"OUTPUTS"             env-default:""           env-description:"The outputs override (stdout, stderr, file)."`
	JSONTimeKey       string              `yaml:"jsonTimeKey"       env:"JSON_TIME_KEY"       env-default:"ts"         env-description:"Set the key used for time log entry. If key is empty, the entry is omitted."`
	JSONLevelKey      string              `yaml:"jsonLevelKey"      env:"JSON_LEVEL_KEY"      env-default:"level"      env-description:"Set the key used for level log entry. If key is empty, the entry is omitted."`
	JSONNameKey       string              `yaml:"jsonNameKey"       env:"JSON_NAME_KEY"       env-default:"logger"     env-description:"Set the key used for name log entry. If key is empty, the entry is omitted."`
	JSONCallerKey     string              `yaml:"jsonCallerKey"     env:"JSON_CALLER_KEY"     env-default:"caller"     env-description:"Set the key used for caller log entry. If key is empty, the entry is omitted."`
	JSONFunctionKey   string              `yaml:"jsonFunctionKey"   env:"JSON_FUNCTION_KEY"   env-default:""           env-description:"Set the key used for function log entry. If key is empty, the entry is omitted."`
	JSONMessageKey    string              `yaml:"jsonMessageKey"    env:"JSON_MESSAGE_KEY"    env-default:"msg"        env-description:"Set the key used for message log entry. If key is empty, the entry is omitted."`
	JSONStacktraceKey string              `yaml:"jsonStacktraceKey" env:"JSON_STACKTRACE_KEY" env-default:"stacktrace" env-description:"Set the key used for stack trace log entry. If key is empty, the entry is omitted."`
	SkipLineEnding    bool                `yaml:"skipLineEnding"    env:"SKIP_LINE_ENDING"    env-default:"false"      env-description:"Disable adding newline characters between the log statements."`
	LineEnding        string              `yaml:"lineEnding"        env:"LINE_ENDING"         env-default:""           env-description:"Override the Unix-style default"`
	Encoding          EncodingEnum        `yaml:"encoding"          env:"ENCODING"            env-required:"true"      env-description:"Set the logger encoder (console, json)."`
	Development       bool                `yaml:"development"       env:"DEVELOPMENT"         env-default:"false"      env-description:"Puts the logger in development mode, which makes DPanic-level logs panic instead of simply logging an error."`
	ConsoleSeparator  string              `yaml:"consoleSeparator"  env:"CONSOLE_SEPARATOR"   env-default:""           env-description:"Override the separator used for console encoding."`

	OutputFile OutputFileConfig `yaml:"outputFile" env-prefix:"OUTPUT_FILE_"`
}

type Config struct {
	Preset PresetEnum `yaml:"preset" env:"PRESET" env-default:"production" env-description:"Override the preset (development, production)."`

	Development PresetConfig `yaml:"development" env-prefix:"DEVELOPMENT_"`
	Production  PresetConfig `yaml:"production"  env-prefix:"PRODUCTION_"`
}
