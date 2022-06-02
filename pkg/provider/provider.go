package provider

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/earthly/earthly/util/cliutil"
	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("secret not found")

const vaultConfigFile = "vault.yml"

var cfgFilePath = filepath.Join(cliutil.GetEarthlyDir(), vaultConfigFile)

type Config struct {
	// Token is a token that is used to authenticate with Vault.
	Token string `yaml:"token"`
	// Address is the address of the Vault server.
	Address string `yaml:"address"`
	// Prefix will be prepended to any secret that is passed in
	Prefix string `yaml:"prefix"`
}

func (c *Config) Validate() error {
	if c.Token == "" {
		return errors.New("token is required")
	}

	if c.Address == "" {
		return errors.New("address is required")
	}

	return nil
}

type Provider struct {
	Logger *log.Logger
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) LoadSecretFetcher() (*SecretFetcher, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	// reading default vault token location (https://github.com/hashicorp/vault/blob/c18dd63a9ff0291b38b5765471ae83e93fbd2ff6/command/token/helper_internal.go#L35)
	// ignore error since if that file doesn't exist, still try to read from earthly config dir
	token, _ := os.ReadFile(filepath.Join(homeDir, ".vault-token"))
	config := Config{Token: string(token)}

	cfgFile, err := os.Open(cfgFilePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.NewDecoder(cfgFile).Decode(&config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	apiConfig := vault.DefaultConfig()
	apiConfig.Address = config.Address

	client, err := vault.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}
	client.SetToken(config.Token)

	return &SecretFetcher{
		VaultClient: client.Logical(),
		Prefix:      config.Prefix,
		Logger:      p.Logger,
	}, nil
}
