package utils

import (
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment               string        `mapstructure:"ENVIRONMENT"`
	DBSource1                 string        `mapstructure:"DB_SOURCE_1"`
	DBSource2                 string        `mapstructure:"DB_SOURCE_2"`
	TokenSymmetricKey         string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	HTTPServerAddress         string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	MonolithUrl               string        `mapstructure:"MONOLITH_URL"`
	AccessTokenDuration       time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration      time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSMTPHost             string        `mapstructure:"EMAIL_SMTP_HOST"`
	EmailSMTPPort             int           `mapstructure:"EMAIL_SMTP_PORT"`
	EmailAUTHUsername         string        `mapstructure:"EMAIL_AUTH_USERNAME"`
	EmailAUTHPassword         string        `mapstructure:"EMAIL_AUTH_PASSWORD"`
	EmailSenderName           string        `mapstructure:"EMAIL_SENDER_NAME"`
	GoogleAuthClientId        string        `mapstructure:"GOOGLE_AUTH_CLIENT_ID"`
	GoogleAuthClientSecret    string        `mapstructure:"GOOGLE_AUTH_CLIENT_SECRET"`
	GoogleAuthRedirectUrl     string        `mapstructure:"GOOGLE_AUTH_REDIRECT_URL"`
	FacebookAuthClientId      string        `mapstructure:"FACEBOOK_AUTH_CLIENT_ID"`
	FacebookAuthClientSecret  string        `mapstructure:"FACEBOOK_AUTH_CLIENT_SECRET"`
	FacebookAuthRedirectUrl   string        `mapstructure:"FACEBOOK_AUTH_REDIRECT_URL"`
	MicrosoftAuthClientId     string        `mapstructure:"MICROSOFT_AUTH_CLIENT_ID"`
	MicrosoftAuthClientSecret string        `mapstructure:"MICROSOFT_AUTH_CLIENT_SECRET"`
	MicrosoftAuthRedirectUrl  string        `mapstructure:"MICROSOFT_AUTH_REDIRECT_URL"`
	MicrosoftAuthTenantId     string        `mapstructure:"MICROSOFT_AUTH_TENANT_ID"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("DB_SOURCE_1", "root:password@tcp(127.0.0.1:3306)/dolphin?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetDefault("DB_SOURCE_2", "")
	viper.SetDefault("HTTP_SERVER_ADDRESS", "0.0.0.0:8080")

	viper.SetDefault("TOKEN_SYMMETRIC_KEY", "super_secret_keys")
	viper.SetDefault("ACCESS_TOKEN_DURATION", "24h")
	viper.SetDefault("REFRESH_TOKEN_DURATION", "8760h")

	viper.SetDefault("EMAIL_SMTP_HOST", "sandbox.smtp.mailtrap.io")
	viper.SetDefault("EMAIL_SMTP_PORT", "587")
	viper.SetDefault("EMAIL_AUTH_USERNAME", "123455av123231")
	viper.SetDefault("EMAIL_AUTH_PASSWORD", "12312asdas1231")
	viper.SetDefault("EMAIL_SENDER_NAME", "Dolphin <dolphin@62teknologi.com>")

	viper.SetDefault("GoogleAuthClientId", "1234abcd!@#$1234")
	viper.SetDefault("GoogleAuthClientSecret", "1234abcd!@#$1234")
	viper.SetDefault("GoogleAuthRedirectUrl", "https://dolphin.com/auth/callback/google")
	viper.SetDefault("FacebookAuthClientId", "1234abcd!@#$1234")
	viper.SetDefault("FacebookAuthClientSecret", "1234abcd!@#$1234")
	viper.SetDefault("FacebookAuthRedirectUrl", "https://dolphin.com/auth/callback/facebook")
	viper.SetDefault("MicrosoftAuthClientId", "1234abcd!@#$1234")
	viper.SetDefault("MicrosoftAuthClientSecret", "1234abcd!@#$1234")
	viper.SetDefault("MicrosoftAuthRedirectUrl", "https://dolphin.com/auth/callback/microsoft")
	viper.SetDefault("MicrosoftAuthTenantId", "1234abcd!@#$1234")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
