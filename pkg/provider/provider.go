package provider

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/earthly/earthly/util/cliutil"
	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("secret not found")

const vaultConfigFile = "vault.yml"

type VaultConfig struct {
	// Token is a token that is used to authenticate with Vault.
	Token string `yaml:"token"`
	// Address is the address of the Vault server.
	Address string `yaml:"address"`
	// Prefix will be prepended to any secret that is passed in
	Prefix string `yaml:"prefix"`
}

func (c *VaultConfig) Validate() error {
	if c.Token == "" {
		return errors.New("token is required")
	}

	if c.Address == "" {
		return errors.New("address is required")
	}

	return nil
}

type Provider struct {
	Out io.Writer
	Err io.Writer
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) PrintSecret(args []string) error {
	homeDir, err := homedir.Dir()
	if err != nil {
		return err
	}

	// reading default vault token location (https://github.com/hashicorp/vault/blob/c18dd63a9ff0291b38b5765471ae83e93fbd2ff6/command/token/helper_internal.go#L35)
	// ignore error since if that file doesn't exist, still try to read from earthly config dir
	token, _ := os.ReadFile(filepath.Join(homeDir, ".vault-token"))
	config := VaultConfig{Token: string(token)}

	cfgFile, err := os.Open(filepath.Join(cliutil.GetEarthlyDir(), vaultConfigFile))
	if err != nil {
		return err
	}

	if err := yaml.NewDecoder(cfgFile).Decode(&config); err != nil {
		return err
	}

	if err := config.Validate(); err != nil {
		return err
	}

	apiConfig := vault.DefaultConfig()
	apiConfig.Address = config.Address

	client, err := vault.NewClient(apiConfig)
	if err != nil {
		return err
	}
	client.SetToken(config.Token)

	lookup := args[0]
	fmt.Fprintf(p.Err, "got request for: %s\n", lookup)

	fullLookup := strings.Join(append(strings.Split(config.Prefix, "/"), lookup), "/")
	fullLookup = strings.TrimPrefix(fullLookup, "/")
	fmt.Fprintf(p.Err, "full lookup path with prefix: %s\n", fullLookup)

	pathAndField := strings.SplitN(fullLookup, ".", 2)
	if len(pathAndField) != 2 {
		return fmt.Errorf("invalid input: %s", fullLookup)
	}

	pathParts := strings.Split(pathAndField[0], "/")

	// insert "data" after the first item in the path
	vaultPath := strings.Join(append([]string{pathParts[0], "data"}, pathParts[1:]...), "/")

	secret, err := client.Logical().Read(vaultPath)
	if err != nil {
		return err
	}

	if secret == nil {
		return ErrNotFound
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("malformed secret data")
	}

	value, ok := data[pathAndField[1]].(string)
	if !ok {
		return fmt.Errorf("malformed secret value")
	}

	fmt.Fprint(p.Out, value)

	return nil
}
