package config

import (
	"github.com/golang/glog"
	"github.com/orensimple/otus_events_scheduler/internal/domain/errors"

	"github.com/spf13/viper"
)

// Init config
func Init(path string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			glog.Errorf("Config file not found %s", err.Error())
		} else {
			glog.Errorf("Config file was found but another error was produced %s", err.Error())
		}
	}
	err := Validate()
	if err != nil {
		glog.Errorf(err.Error())
	}
	return viper.ReadInConfig()
}

// Validate check config params
func Validate() error {
	if len(viper.GetString("log_level.file")) <= 0 {
		glog.Errorf("Cannot read log_level.file in config")
		return errors.ErrConfigWrangParams
	}
	if len(viper.GetString("log_level.command")) <= 0 {
		glog.Errorf("Cannot read log_level.ficommandle in config")
		return errors.ErrConfigWrangParams
	}
	if len(viper.GetString("log_file")) <= 0 {
		glog.Errorf("Cannot read log_file in config")
		return errors.ErrConfigWrangParams
	}
	if len(viper.GetString("http_listen.ip")) <= 0 {
		glog.Errorf("Cannot read http_listen.ip in config")
		return errors.ErrConfigWrangParams
	}
	if len(viper.GetString("http_listen.port")) <= 0 {
		glog.Errorf("Cannot read lhttp_listen.port in config")
		return errors.ErrConfigWrangParams
	}
	return nil
}
