package main

import (
	//_ "embed"
	"flag"
	"github.com/jaeha-choi/Proj_Coconut_Desktop/internal/client"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"github.com/jaeha-choi/Proj_Coconut_Utility/util"
	"os"
)

//var uiString []byte

func main() {
	// Command line arguments (flags) overrides configuration file, if exist

	// Double dash arguments (e.g. --config-path) is not possible with "flag" package it seems like. Consider
	// Using "getopt" package.
	confPath := flag.String("config-path", "./config/config.yml", "Configuration file path")
	logLevelArg := flag.String("log-level", "warning", "Logging level")

	serverHostFlag := flag.String("host", "", "Server address")
	serverPortFlag := flag.Int("port", 0, "Server port")
	localPortFlag := flag.Int("local-port", 0, "Local port")
	keyPathFlag := flag.String("cert-path", "", "Key pair path")
	dataPathFlag := flag.String("data-path", "", "Data path")

	flag.Parse()

	// Setup logger
	var logLevel log.LoggingMode
	switch *logLevelArg {
	case "debug":
		logLevel = log.DEBUG
	case "info":
		logLevel = log.INFO
	case "warning":
		logLevel = log.WARNING
	case "error":
		logLevel = log.ERROR
	case "fatal":
		logLevel = log.FATAL
	default:
		logLevel = log.WARNING
	}

	logger := log.NewLogger(os.Stdout, logLevel, "CLIENT ")
	var cli *client.Client
	var err error

	// Read configurations
	if cli, err = client.ReadConfig(*confPath, logger); err != nil {
		logger.Warning("Could not read config, trying default config")
		cli = client.InitConfig(logger)
		if err := util.WriteConfig(*confPath, cli); err != nil {
			logger.Debug(err)
			logger.Warning("Could not save config")
		}
	}

	// Override configurations if arguments are provided
	if *serverHostFlag != "" {
		cli.ServerHost = *serverHostFlag
	}
	if *keyPathFlag != "" {
		cli.KeyPath = *keyPathFlag
	}
	if *dataPathFlag != "" {
		cli.DataPath = *dataPathFlag
	}
	if 0 < *serverPortFlag && *serverPortFlag < 65536 {
		cli.ServerPort = uint16(*serverPortFlag)
	} else if *serverPortFlag != 0 {
		logger.Fatal("Provided port out of range")
		os.Exit(1)
	}
	if 0 < *localPortFlag && *localPortFlag < 65536 {
		cli.LocalPort = uint16(*localPortFlag)
	} else if *localPortFlag != 0 {
		logger.Fatal("Provided port out of range")
		os.Exit(1)
	}

	client.Start("./data/ui/UI.glade", cli)
	//client.Start(string(uiString))
}
