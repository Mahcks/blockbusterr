package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mahcks/blockbusterr/pkg/utils"
	"github.com/spf13/viper"
)

type Config struct {
	Timestamp time.Time `mapstructure:"timestamp" json:"timestamp"`

	Twitch struct {
		Bot struct {
			Prefix    string `mapstructure:"prefix" json:"prefix"`
			Channel   string `mapstructure:"channel" json:"channel"`
			ChannelID string `mapstructure:"channel_id" json:"channel_id"`
			Username  string `mapstructure:"username" json:"username"`
			OAuth     string `mapstructure:"oauth" json:"oauth"`
		} `mapstructure:"bot" json:"bot"`

		Helix struct {
			ClientID       string `mapstructure:"client_id" json:"client_id"`
			ClientSecret   string `mapstructure:"client_secret" json:"client_secret"`
			EventSubSecret string `mapstructure:"eventsub_secret" json:"eventsub_secret"`
		} `mapstructure:"helix" json:"helix"`
	} `mapstructure:"twitch" json:"twitch"`

	Turso struct {
		URL string `mapstructure:"url" json:"url"`
	} `mapstructure:"turso" json:"turso"`

	APIKeys struct {
		LastFM string `mapstructure:"lastfm" json:"lastfm"`
	} `mapstructure:"api_keys" json:"api_keys"`
}

//nolint:gocritic
func New(Version string, Timestamp time.Time) (*Config, error) {
	config := viper.New()

	config.SetConfigType("yaml")
	config.AddConfigPath("./config")
	config.AddConfigPath("./src/config")
	config.AddConfigPath("./app/config")

	// Use the dev config file if the version is dev
	if Version == "dev" {
		config.SetConfigName("config.dev.yaml")
	} else {
		config.SetConfigName("config.yaml")
	}

	err := config.ReadInConfig()
	if err != nil {
		fmt.Println("Fatal error config file: default \n", err)
		return nil, err
	}

	// Environment
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AllowEmptyEnv(false)

	c := &Config{}

	c.Timestamp = Timestamp

	errUnmarshal := config.Unmarshal(&c)
	if errUnmarshal != nil {
		return nil, err
	}

	// Check for empty values in the Config struct
	configValue := reflect.ValueOf(c).Elem()
	configType := configValue.Type()
	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldType := configType.Field(i)

		// Check for nested fields (e.g., Redis and Twitch in this example)
		if field.Kind() == reflect.Struct {
			for j := 0; j < field.NumField(); j++ {
				nestedField := field.Field(j)
				nestedFieldType := fieldType.Type.Field(j)

				if utils.IsEmptyValue(nestedField) {
					return nil, fmt.Errorf(
						"empty value for config field: %s.%s",
						fieldType.Name,
						nestedFieldType.Name,
					)
				}
			}
		} else {
			if utils.IsEmptyValue(field) {
				return nil, fmt.Errorf("empty value for config field: %s", fieldType.Name)
			}
		}
	}

	return c, nil
}
