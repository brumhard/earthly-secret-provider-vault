package provider

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/brumhard/earthly-secret-provider-vault/pkg/vault"

	"github.com/earthly/earthly/util/cliutil"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"github.com/moby/buildkit/session/secrets"
	"gopkg.in/yaml.v3"
)

const vaultConfigFile = "vault.yml"

var CfgFilePath = filepath.Join(cliutil.GetEarthlyDir(), vaultConfigFile)

var ErrInvalidConfig = errors.New("invalid config")

type Config struct {
	// Token is a token that is used to authenticate with Vault.
	Token string `yaml:"token,omitempty"`
	// Address is the address of the Vault server.
	Address string `yaml:"address,omitempty"`
	// Prefix will be prepended to any secret that is passed in
	Prefix string `yaml:"prefix,omitempty"`
}

func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token is required: %w", ErrInvalidConfig)
	}

	if c.Address == "" {
		return fmt.Errorf("address is required: %w", ErrInvalidConfig)
	}

	if _, err := url.ParseRequestURI(c.Address); err != nil {
		return fmt.Errorf("address %q should be a valid URL: %w", c.Address, ErrInvalidConfig)
	}

	return nil
}

type Provider struct {
	Logger *log.Logger
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) LoadSecretStore() (secrets.SecretStore, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	// reading default vault token location (https://github.com/hashicorp/vault/blob/c18dd63a9ff0291b38b5765471ae83e93fbd2ff6/command/token/helper_internal.go#L35)
	// ignore error since if that file doesn't exist, still try to read from earthly config dir
	token, _ := os.ReadFile(filepath.Join(homeDir, ".vault-token"))
	config := Config{Token: string(token)}

	cfgFile, err := os.Open(CfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", ErrInvalidConfig)
	}
	defer cfgFile.Close()

	if err := yaml.NewDecoder(cfgFile).Decode(&config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = config.Address

	client, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}
	client.SetToken(config.Token)

	return vault.NewSecretStore(client.Logical(), p.Logger, vault.WithPrefix(config.Prefix)), nil
}

func (p *Provider) SetConfigKey(key, value string) error {
	// read the config and create the file if it doesn't exist yet
	cfgFile, err := os.OpenFile(CfgFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed opening or creating config file: %w", err)
	}
	defer cfgFile.Close()

	config := Config{}
	err = yaml.NewDecoder(cfgFile).Decode(&config)
	// if it's an EOF error just proceed since the file probably just got created and is empty
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed reading config file: %w", err)
	}

	switch key {
	case "token":
		config.Token = value
	case "address":
		config.Address = value
	case "prefix":
		config.Prefix = value
	default:
		return fmt.Errorf("key %q is not supported: %w", key, ErrInvalidConfig)
	}

	// delete file content and reset I/O offset to 0
	if err := cfgFile.Truncate(0); err != nil {
		return fmt.Errorf("failed truncating config file before write: %w", err)
	}

	if _, err := cfgFile.Seek(0, 0); err != nil {
		return err
	}

	// write new config to now empty file
	if err := yaml.NewEncoder(cfgFile).Encode(&config); err != nil {
		return fmt.Errorf("failed writing config file: %w", err)
	}

	return nil
}
