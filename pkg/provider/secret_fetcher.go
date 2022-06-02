package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

type VaultClient interface {
	ReadWithContext(ctx context.Context, path string) (*vault.Secret, error)
}

type SecretFetcher struct {
	Prefix      string
	VaultClient VaultClient
	Logger      *log.Logger
}

func (s *SecretFetcher) Fetch(ctx context.Context, lookup string) (string, error) {
	s.Logger.Printf("got request for: %q\n", lookup)

	vaultPath, vaultField, err := s.vaultPath(lookup)
	if err != nil {
		return "", err
	}
	s.Logger.Printf("looking for field %q in path %q\n", vaultField, vaultPath)

	return s.readSecretField(ctx, vaultPath, vaultField)
}

func (s *SecretFetcher) vaultPath(lookup string) (path string, field string, err error) {
	fullLookup := strings.Join(append(strings.Split(s.Prefix, "/"), lookup), "/")
	fullLookup = strings.TrimPrefix(fullLookup, "/")

	pathAndField := strings.SplitN(fullLookup, ".", 2)
	if len(pathAndField) != 2 {
		return "", "", fmt.Errorf("invalid input: %s", fullLookup)
	}

	pathParts := strings.Split(pathAndField[0], "/")

	// insert "data" after the first item in the path
	return strings.Join(append([]string{pathParts[0], "data"}, pathParts[1:]...), "/"), pathAndField[1], nil
}

func (s *SecretFetcher) readSecretField(ctx context.Context, path, field string) (string, error) {
	secret, err := s.VaultClient.ReadWithContext(ctx, path)
	if err != nil {
		return "", err
	}

	if secret == nil {
		return "", ErrNotFound
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("malformed secret data")
	}

	value, ok := data[field].(string)
	if !ok {
		return "", fmt.Errorf("malformed secret value")
	}

	return value, nil
}
