package conf

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/tcardonne/restic-controller/restic"
)

// ExporterConfig contains general configuration for the Prometheus exporter
type ExporterConfig struct {
	BindAddress string `mapstructure:"bind_address"`
}

// Repository contains configuration for one repository
type Repository struct {
	Name     string `mapstructure:"name" validate:"required"`
	URL      string `mapstructure:"url" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Check    struct {
		Schedule     string `mapstructure:"schedule" validate:"required"`
		RunOnStartup bool   `mapstructure:"run_on_startup"`
	} `mapstructure:"check" validate:"required,dive"`
	Retention struct {
		Schedule     string               `mapstructure:"schedule" validate:"required"`
		RunOnStartup bool                 `mapstructure:"run_on_startup"`
		Policy       *restic.ForgetPolicy `mapstructure:"policy" validate:"required,dive"` // A changer
	} `mapstructure:"retention" validate:"required,dive"`
}

// Configuration is the root of the configuration
type Configuration struct {
	Log          LogConfig      `mapstructure:"log"`
	Exporter     ExporterConfig `mapstructure:"exporter"`
	Repositories []*Repository  `mapstructure:"repositories" validate:"required,dive"`
}

// LoadConfiguration loads and validates the configuration from a file
func LoadConfiguration(configFile string) (*Configuration, error) {
	configuration := Configuration{
		Exporter: ExporterConfig{
			BindAddress: "127.0.0.1:8080",
		},
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	err := viper.Unmarshal(&configuration)
	if err != nil {
		return nil, err
	}

	if err := validateConfiguration(&configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}

func validateConfiguration(config *Configuration) error {
	validate := validator.New()

	return validate.Struct(config)
}
