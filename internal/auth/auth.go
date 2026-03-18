package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/dsswift/cli-exchange/internal/config"
)

var scopes = []string{
	"https://graph.microsoft.com/Mail.ReadWrite",
	"https://graph.microsoft.com/Calendars.Read",
	"https://graph.microsoft.com/User.Read",
}

type fileCache struct {
	path string
}

func (c *fileCache) Replace(ctx context.Context, unmarshaler cache.Unmarshaler, hints cache.ReplaceHints) error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil
	}
	return unmarshaler.Unmarshal(data)
}

func (c *fileCache) Export(ctx context.Context, marshaler cache.Marshaler, hints cache.ExportHints) error {
	data, err := marshaler.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0600)
}

// Authenticator handles MSAL device code flow authentication for Microsoft Graph API.
type Authenticator struct {
	client    public.Client
	cachePath string
}

// New creates an Authenticator configured with the given ExchangeConfig.
func New(cfg *config.ExchangeConfig) (*Authenticator, error) {
	fc := &fileCache{path: cfg.TokenCachePath}
	client, err := public.New(cfg.ClientID,
		public.WithAuthority(cfg.Authority),
		public.WithCache(fc),
	)
	if err != nil {
		return nil, fmt.Errorf("creating MSAL client: %w", err)
	}
	return &Authenticator{client: client, cachePath: cfg.TokenCachePath}, nil
}

// GetAccessToken returns a valid access token, using cached tokens when possible
// and falling back to device code flow when necessary.
func (a *Authenticator) GetAccessToken(ctx context.Context) (string, error) {
	accounts, err := a.client.Accounts(ctx)
	if err == nil && len(accounts) > 0 {
		result, err := a.client.AcquireTokenSilent(ctx, scopes, public.WithSilentAccount(accounts[0]))
		if err == nil {
			return result.AccessToken, nil
		}
	}

	dc, err := a.client.AcquireTokenByDeviceCode(ctx, scopes)
	if err != nil {
		return "", fmt.Errorf("initiating device code flow: %w", err)
	}

	// Print the device code message to stderr so it doesn't pollute command output.
	fmt.Fprintln(os.Stderr, dc.Result.Message)

	result, err := dc.AuthenticationResult(ctx)
	if err != nil {
		return "", fmt.Errorf("completing device code flow: %w", err)
	}

	return result.AccessToken, nil
}

// ClearCache removes the persisted token cache file.
func (a *Authenticator) ClearCache() error {
	if err := os.Remove(a.cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing token cache: %w", err)
	}
	return nil
}
