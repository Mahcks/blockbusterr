package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Version string `mapstructure:"version" yaml:"version,omitempty"`

	Trakt struct {
		ClientID     string `mapstructure:"client_id" json:"client_id" yaml:"client_id"`
		ClientSecret string `mapstructure:"client_secret" json:"client_secret" yaml:"client_secret"`
	} `mapstructure:"trakt" json:"trakt" yaml:"trakt"`

	Radarr struct {
		URL            string `mapstructure:"url" json:"url" yaml:"url"`
		APIKey         string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
		QualityProfile int    `mapstructure:"quality_profile" json:"quality_profile" yaml:"quality_profile"`
		RootFolder     string `mapstructure:"root_folder" json:"root_folder" yaml:"root_folder"`
	} `mapstructure:"radarr" json:"radarr" yaml:"radarr"`

	Sonarr struct {
		URL            string `mapstructure:"url" json:"url" yaml:"url"`
		APIKey         string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
		QualityProfile int    `mapstructure:"quality_profile" json:"quality_profile" yaml:"quality_profile"`
		RootFolder     string `mapstructure:"root_folder" json:"root_folder" yaml:"root_folder"`
	} `mapstructure:"sonarr" json:"sonarr" yaml:"sonarr"`

	Jobs struct {
		SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval"`

		TrendingMovies struct {
			Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int  `mapstructure:"limit" json:"limit" yaml:"limit"`
		} `mapstructure:"trending_movies" json:"trending_movies" yaml:"trending_movies"`

		TrendingShows struct {
			Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int  `mapstructure:"limit" json:"limit" yaml:"limit"`
		} `mapstructure:"trending_shows" json:"trending_shows" yaml:"trending_shows"`

		PopularMovies struct {
			Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int  `mapstructure:"limit" json:"limit" yaml:"limit"`
		} `mapstructure:"popular_movies" json:"popular_movies" yaml:"popular_movies"`

		PopularShows struct {
			Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int  `mapstructure:"limit" json:"limit" yaml:"limit"`
		} `mapstructure:"popular_shows" json:"popular_shows" yaml:"popular_shows"`

		BoxOffice struct {
			Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int  `mapstructure:"limit" json:"limit" yaml:"limit"`
		} `mapstructure:"box_office" json:"box_office" yaml:"box_office"`

		FavoritedMovies struct {
			Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period  string `mapstructure:"period" json:"period" yaml:"period"`
		} `mapstructure:"favorited_movies" json:"favorited_movies" yaml:"favorited_movies"`

		PlayedMovies struct {
			Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period  string `mapstructure:"period" json:"period" yaml:"period"`
		} `mapstructure:"played_movies" json:"played_movies" yaml:"played_movies"`

		WatchedMovies struct {
			Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period  string `mapstructure:"period" json:"period" yaml:"period"`
		} `mapstructure:"watched_movies" json:"watched_movies" yaml:"watched_movies"`

		CollectedMovies struct {
			Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period  string `mapstructure:"period" json:"period" yaml:"period"`
		} `mapstructure:"collected_movies" json:"collected_movies" yaml:"collected_movies"`

		AnticipatedMovies struct {
			Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int  `mapstructure:"limit" json:"limit" yaml:"limit"`
		} `mapstructure:"anticipated_movies" json:"anticipated_movies" yaml:"anticipated_movies"`

		FavoritedShows struct {
			Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period  string `mapstructure:"period" json:"period" yaml:"period"`
		} `mapstructure:"favorited_shows" json:"favorited_shows" yaml:"favorited_shows"`

		PlayedShows struct {
			Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period  string `mapstructure:"period" json:"period" yaml:"period"`
		} `mapstructure:"played_shows" json:"played_shows" yaml:"played_shows"`

		WatchedShows struct {
			Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period  string `mapstructure:"period" json:"period" yaml:"period"`
		} `mapstructure:"watched_shows" json:"watched_shows" yaml:"watched_shows"`

		CollectedShows struct {
			Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period  string `mapstructure:"period" json:"period" yaml:"period"`
		} `mapstructure:"collected_shows" json:"collected_shows" yaml:"collected_shows"`

		AnticipatedShows struct {
			Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit   int  `mapstructure:"limit" json:"limit" yaml:"limit"`
		} `mapstructure:"anticipated_shows" json:"anticipated_shows" yaml:"anticipated_shows"`
	} `mapstructure:"jobs" json:"jobs" yaml:"jobs"`

	// Internal field to track config file path
	ConfigFilePath string `mapstructure:"-" json:"-" yaml:"-"`
}

// New creates a new Config instance with the given settings
func New(version string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.AddConfigPath("./config")
	v.AddConfigPath("/home/nonroot/config")

	var configFileName string
	if version == "dev" {
		configFileName = "config.dev.yaml"
		v.SetConfigName(configFileName)
	} else {
		configFileName = "config.yaml"
		v.SetConfigName(configFileName)
	}

	var configFilePath string
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found, using only ENV variables")
			configFilePath = "" // No file found
		} else {
			return nil, fmt.Errorf("config load error: %w", err)
		}
	} else {
		configFilePath = v.ConfigFileUsed()
	}

	c := &Config{
		ConfigFilePath: configFilePath,
	}
	if err := v.Unmarshal(c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return c, nil
}

// Save writes the current config back to the config file
func (c *Config) Save() error {
	if c.ConfigFilePath == "" {
		return fmt.Errorf("no config file path set, cannot save")
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(c.ConfigFilePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
