package main

import (
	"fmt"
	"log/slog"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Listen          string     `envconfig:"listen" default:":8080"`
	LogLevel        slog.Level `envconfig:"log_level" default:"debug"`
	DynamoEndpoint  string     `envconfig:"dynamo_endpoint" default:""` // Must be empty to on AWS
	MedicationTable string     `envconfig:"medication_table" default:"medication"`
}

func NewConfig() (Config, error) {
	c := Config{}
	err := envconfig.Process("med", &c)
	if err != nil {
		return Config{}, fmt.Errorf("unable to parse config: %w", err)
	}
	return c, nil
}

func MustNewConfig() Config {
	c, err := NewConfig()
	if err != nil {
		panic(err)
	}
	return c
}
