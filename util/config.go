package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config struct hold the configuration of the application
type Config struct {
	DBDriver string `mapstructure:"DB_DRIVER"` 
	DBSource string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
	GRPCDomain string `mapstructure:"GRPC_DOMAIN"`
	MigrationUrl string `mapstructure:"MIGRATION_URL"`
	TokenSymmetricKey string `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	RedisURL string `mapstructure:"REDIS_URL"`
	ENV string `mapstructure:"ENV"`
}

// LoadConfig function loads the configuration from the file
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

