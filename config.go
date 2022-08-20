package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbHost     string
	DbPort     string
	DbName     string
	DbUser     string
	DbPassword string

	R2AccountId       string
	R2AccessKeyId     string
	R2AccessKeySecret string

	JobKey string

	AppEnv string
}

const (
	AppEnvDevelopment = "development"
	AppEnvProduction  = "production"
)

func (c *Config) GetDbConnectionString() string {
	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.DbUser, c.DbPassword, c.DbHost, c.DbPort, c.DbName)
	return psqlInfo
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load env: %w", err)
	}
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = AppEnvDevelopment
	} else {
		if appEnv != AppEnvDevelopment && appEnv != AppEnvProduction {
			return nil, fmt.Errorf("failed to validate APP_ENV: invalid value %q", appEnv)
		}
	}
	return &Config{
		DbHost:            os.Getenv("DB_HOST"),
		DbPort:            os.Getenv("DB_PORT"),
		DbName:            os.Getenv("DB_NAME"),
		DbUser:            os.Getenv("DB_USER"),
		DbPassword:        os.Getenv("DB_PASSWORD"),
		R2AccountId:       os.Getenv("R2_ACCOUNTID"),
		R2AccessKeyId:     os.Getenv("R2_ACCESSKEYID"),
		R2AccessKeySecret: os.Getenv("R2_ACCESSKEYSECRET"),
		JobKey:            os.Getenv("JOB_KEY"),
		AppEnv:            os.Getenv("APP_ENV"),
	}, nil
}
