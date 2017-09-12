package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultJSONLog  = false
	DefaultLogLevel = "debug"
	DefaultLogFile  = false
	DefaultPort     = "7780"
)

// Provider defines a set of read-only methods for accessing the application
// configuration params as defined in one of the config files.
type Provider interface {
	ConfigFileUsed() string
	Get(key string) interface{}
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt64(key string) int64
	GetSizeInBytes(key string) uint
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	InConfig(key string) bool
	IsSet(key string) bool
}

var defaultConfig *viper.Viper

func Config() Provider {
	return defaultConfig
}

func LoadConfigProvider(appName, fileConfig string) Provider {
	return readViperConfig(appName)
}

func init() {
	defaultConfig = initViper("ingestor")
}

func globalConfig() *viper.Viper {
	v := viper.New()

	configPath := flag.String("config", "", "The file path to a config file")
	jsonLog := flag.Bool("json-log", DefaultJSONLog, "Enable JSON formatting of log")
	logLevel := flag.String("log-level", DefaultLogLevel, "Log level")
	logFile := flag.Bool("log-file", DefaultLogFile, "Write log messages into a file")
	port := flag.String("port", DefaultPort, "Port")

	flag.Parse()

	// global defaults

	v.SetDefault("ConfigPath", *configPath)
	v.SetDefault("JSONLog", *jsonLog)
	v.SetDefault("LogLevel", *logLevel)
	v.SetDefault("LogFile", *logFile)
	v.SetDefault("Port", *port)

	return v
}

func initViper(appName string) *viper.Viper {
	v := globalConfig()

	v.SetConfigType("json")

	if path := v.GetString("ConfigPath"); path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(fmt.Sprintf("/etc/%s", strings.ToLower(appName)))
		v.BindEnv("awsId")
		v.BindEnv("awsSecret")
	}

	err := v.ReadInConfig()
	if err != nil {
		fmt.Printf("[Fatal] Unable to read config file: %s", err.Error())
		os.Exit(1)
	}

	return v
}

func readViperConfig(appName string) *viper.Viper {
	v := globalConfig()

	v.SetEnvPrefix(appName)
	v.AutomaticEnv()

	return v
}
