package vault

import (
	"context"
	"fmt"
	"log"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/moby/buildkit/session/secrets"
)

type Client interface {
	ReadWithContext(ctx context.Context, path string) (*vault.Secret, error)
}

var _ secrets.SecretStore = (*SecretStore)(nil)

type SecretStore struct {
	Prefix      string
	VaultClient Client
	Logger      *log.Logger
}

func NewSecretStore(client Client, logger *log.Logger, opts ...Option) *SecretStore {
	options := &options{}

	for _, opt := range opts {
		opt(options)
	}

	return &SecretStore{
		Prefix:      options.prefix,
		VaultClient: client,
		Logger:      logger,
	}
}

type options struct {
	prefix string
}

type Option func(*options)

func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

func (s *SecretStore) GetSecret(ctx context.Context, lookup string) ([]byte, error) {
	s.Logger.Printf("Got request for: %q\n", lookup)

	vaultPath, vaultField, err := s.vaultPath(lookup)
	if err != nil {
		return nil, err
	}
	s.Logger.Printf("Looking for field %q in path %q\n", vaultField, vaultPath)

	return s.readSecretField(ctx, vaultPath, vaultField)
}

func (s *SecretStore) vaultPath(lookup string) (path string, field string, err error) {
	fullLookup := strings.Join(append(strings.Split(s.Prefix, "/"), lookup), "/")
	fullLookup = strings.TrimLeft(fullLookup, "/")

	pathAndField := strings.SplitN(fullLookup, ".", 2)
	if len(pathAndField) != 2 {
		return "", "", fmt.Errorf("invalid input: %s", fullLookup)
	}

	pathParts := strings.Split(pathAndField[0], "/")

	// insert "data" after the first item in the path
	return strings.Join(append([]string{pathParts[0], "data"}, pathParts[1:]...), "/"), pathAndField[1], nil
}

func (s *SecretStore) readSecretField(ctx context.Context, path, field string) ([]byte, error) {
	secret, err := s.VaultClient.ReadWithContext(ctx, path)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, secrets.ErrNotFound
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("malformed secret data")
	}

	value, ok := data[field].(string)
	if !ok {
		return nil, fmt.Errorf("malformed secret value")
	}

	return []byte(value), nil
}
