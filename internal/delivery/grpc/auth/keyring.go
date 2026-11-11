package serviceauth

import (
	"context"

	"github.com/dockloud/service-auth/core"
)

type Keyring map[string]serviceauth.Key

func (k Keyring) ResolveKey(_ context.Context, keyID string) (serviceauth.Key, error) {
	key, ok := k[keyID]
	if !ok {
		return serviceauth.Key{}, serviceauth.ErrUnknownKey
	}
	return key, nil
}
