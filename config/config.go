package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	DefaultJSONLog   = false
	DefaultLogLevel  = "debug"
	DefaultLogFile   = false
	DefaultLogFolder = "/etc/ingestor"
	DefaultPort      = "7780"
)

type Provider = viper.Viper

var defaultConfig *viper.Viper

func init() {
	defaultConfig = initViper("ingestor")
}

func Config() *Provider {
	return defaultConfig
}

func initViper(appName string) *Provider {
	v := viper.New()

	configPath := flag.String("config", "", "The file path to a config file")
	jsonLog := flag.Bool("json-log", DefaultJSONLog, "Enable JSON formatting of log")
	logLevel := flag.String("log-level", DefaultLogLevel, "Log level")
	logFile := flag.Bool("log-file", DefaultLogFile, "Write log messages into a file")
	logFolder := flag.String("log-folder", DefaultLogFolder, "Place log files")
	port := flag.String("port", DefaultPort, "Port")

	flag.Parse()

	// global defaults

	v.SetDefault("ConfigPath", *configPath)
	v.SetDefault("JSONLog", *jsonLog)
	v.SetDefault("LogLevel", *logLevel)
	v.SetDefault("LogFile", *logFile)
	v.SetDefault("LogFolder", *logFolder)
	v.SetDefault("Port", *port)

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
