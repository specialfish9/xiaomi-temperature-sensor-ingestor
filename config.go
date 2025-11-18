package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/specialfish9/confuso"
)

type Config struct {
	DBAddress  string `confuso:"db_address" validate:"required"`
	DBUser     string `confuso:"db_user" validate:"required"`
	DBPassword string `confuso:"db_password" validate:"required"`
	DBName     string `confuso:"db_name" validate:"required"`
	Devices    string `confuso:"devices" validate:"required"`
	Tick       int    `confuso:"tick" validate:"required,min=1"`
}

func NewConfig(fileName string) (*Config, error) {
	var config Config

	if err := confuso.LoadConf(fileName, &config); err != nil {
		return nil, fmt.Errorf("config: loading config file: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config: validating config: %w", err)
	}

	return &config, nil
}
