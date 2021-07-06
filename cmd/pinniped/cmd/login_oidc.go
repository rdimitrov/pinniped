// Copyright 2020-2021 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
	"k8s.io/client-go/transport"
	"k8s.io/klog/v2/klogr"

	"go.pinniped.dev/internal/execcredcache"
	"go.pinniped.dev/internal/groupsuffix"
	"go.pinniped.dev/internal/plog"
	"go.pinniped.dev/pkg/conciergeclient"
	"go.pinniped.dev/pkg/oidcclient"
	"go.pinniped.dev/pkg/oidcclient/filesession"
	"go.pinniped.dev/pkg/oidcclient/oidctypes"
)

//nolint: gochecknoinits
func init() {
	loginCmd.AddCommand(oidcLoginCommand(oidcLoginCommandRealDeps()))
}

type oidcLoginCommandDeps struct {
	lookupEnv     func(string) (string, bool)
	login         func(string, string, ...oidcclient.Option) (*oidctypes.Token, error)
	exchangeToken func(context.Context, *conciergeclient.Client, string) (*clientauthv1beta1.ExecCredential, error)
}

func oidcLoginCommandRealDeps() oidcLoginCommandDeps {
	return oidcLoginCommandDeps{
		lookupEnv: os.LookupEnv,
		login:     oidcclient.Login,
		exchangeToken: func(ctx context.Context, client *conciergeclient.Client, token string) (*clientauthv1beta1.ExecCredential, error) {
			return client.ExchangeToken(ctx, token)
		},
	}
}

type oidcLoginFlags struct {
	issuer                       string
	clientID                     string
	listenPort                   uint16
	scopes                       []string
	skipBrowser                  bool
	sessionCachePath             string
	caBundlePaths                []string
	caBundleData                 []string
	debugSessionCache            bool
	requestAudience              string
	conciergeEnabled             bool
	conciergeAuthenticatorType   string
	conciergeAuthenticatorName   string
	conciergeEndpoint            string
	conciergeCABundle            string
	conciergeAPIGroupSuffix      string
	credentialCachePath          string
	upstreamIdentityProviderName string
	upstreamIdentityProviderType string
}

func oidcLoginCommand(deps oidcLoginCommandDeps) *cobra.Command {
	var (
		cmd = &cobra.Command{
			Args:         cobra.NoArgs,
			Use:          "oidc --issuer ISSUER",
			Short:        "Login using an OpenID Connect provider",
			SilenceUsage: true,
		}
		flags              oidcLoginFlags
		conciergeNamespace string // unused now
	)
	cmd.Flags().StringVar(&flags.issuer, "issuer", "", "OpenID Connect issuer URL")
	cmd.Flags().StringVar(&flags.clientID, "client-id", "pinniped-cli", "OpenID Connect client ID")
	cmd.Flags().Uint16Var(&flags.listenPort, "listen-port", 0, "TCP port for localhost listener (authorization code flow only)")
	cmd.Flags().StringSliceVar(&flags.scopes, "scopes", []string{oidc.ScopeOfflineAccess, oidc.ScopeOpenID, "pinniped:request-audience"}, "OIDC scopes to request during login")
	cmd.Flags().BoolVar(&flags.skipBrowser, "skip-browser", false, "Skip opening the browser (just print the URL)")
	cmd.Flags().StringVar(&flags.sessionCachePath, "session-cache", filepath.Join(mustGetConfigDir(), "sessions.yaml"), "Path to session cache file")
	cmd.Flags().StringSliceVar(&flags.caBundlePaths, "ca-bundle", nil, "Path to TLS certificate authority bundle (PEM format, optional, can be repeated)")
	cmd.Flags().StringSliceVar(&flags.caBundleData, "ca-bundle-data", nil, "Base64 encoded TLS certificate authority bundle (base64 encoded PEM format, optional, can be repeated)")
	cmd.Flags().BoolVar(&flags.debugSessionCache, "debug-session-cache", false, "Print debug logs related to the session cache")
	cmd.Flags().StringVar(&flags.requestAudience, "request-audience", "", "Request a token with an alternate audience using RFC8693 token exchange")
	cmd.Flags().BoolVar(&flags.conciergeEnabled, "enable-concierge", false, "Use the Concierge to login")
	cmd.Flags().StringVar(&conciergeNamespace, "concierge-namespace", "pinniped-concierge", "Namespace in which the Concierge was installed")
	cmd.Flags().StringVar(&flags.conciergeAuthenticatorType, "concierge-authenticator-type", "", "Concierge authenticator type (e.g., 'webhook', 'jwt')")
	cmd.Flags().StringVar(&flags.conciergeAuthenticatorName, "concierge-authenticator-name", "", "Concierge authenticator name")
	cmd.Flags().StringVar(&flags.conciergeEndpoint, "concierge-endpoint", "", "API base for the Concierge endpoint")
	cmd.Flags().StringVar(&flags.conciergeCABundle, "concierge-ca-bundle-data", "", "CA bundle to use when connecting to the Concierge")
	cmd.Flags().StringVar(&flags.conciergeAPIGroupSuffix, "concierge-api-group-suffix", groupsuffix.PinnipedDefaultSuffix, "Concierge API group suffix")
	cmd.Flags().StringVar(&flags.credentialCachePath, "credential-cache", filepath.Join(mustGetConfigDir(), "credentials.yaml"), "Path to cluster-specific credentials cache (\"\" disables the cache)")
	cmd.Flags().StringVar(&flags.upstreamIdentityProviderName, "upstream-identity-provider-name", "", "The name of the upstream identity provider used during login with a Supervisor")
	cmd.Flags().StringVar(&flags.upstreamIdentityProviderType, "upstream-identity-provider-type", "oidc", "The type of the upstream identity provider used during login with a Supervisor (e.g. 'oidc', 'ldap', 'activedirectory')")

	mustMarkHidden(cmd, "debug-session-cache")
	mustMarkRequired(cmd, "issuer")
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return runOIDCLogin(cmd, deps, flags) }

	mustMarkDeprecated(cmd, "concierge-namespace", "not needed anymore")
	mustMarkHidden(cmd, "concierge-namespace")

	return cmd
}

func runOIDCLogin(cmd *cobra.Command, deps oidcLoginCommandDeps, flags oidcLoginFlags) error { //nolint:funlen
	pLogger, err := SetLogLevel(deps.lookupEnv)
	if err != nil {
		plog.WarningErr("Received error while setting log level", err)
	}

	// Initialize the session cache.
	var sessionOptions []filesession.Option

	// If the hidden --debug-session-cache option is passed, log all the errors from the session cache with klog.
	if flags.debugSessionCache {
		logger := klogr.New().WithName("session")
		sessionOptions = append(sessionOptions, filesession.WithErrorReporter(func(err error) {
			logger.Error(err, "error during session cache operation")
		}))
	}
	sessionCache := filesession.New(flags.sessionCachePath, sessionOptions...)

	// Initialize the login handler.
	opts := []oidcclient.Option{
		oidcclient.WithContext(cmd.Context()),
		oidcclient.WithLogger(klogr.New()),
		oidcclient.WithScopes(flags.scopes),
		oidcclient.WithSessionCache(sessionCache),
	}

	if flags.listenPort != 0 {
		opts = append(opts, oidcclient.WithListenPort(flags.listenPort))
	}

	if flags.requestAudience != "" {
		opts = append(opts, oidcclient.WithRequestAudience(flags.requestAudience))
	}

	if flags.upstreamIdentityProviderName != "" {
		opts = append(opts, oidcclient.WithUpstreamIdentityProvider(
			flags.upstreamIdentityProviderName, flags.upstreamIdentityProviderType))
	}

	switch flags.upstreamIdentityProviderType {
	case "oidc":
		// this is the default, so don't need to do anything
	case "ldap":
		opts = append(opts, oidcclient.WithCLISendingCredentials())
	case "activedirectory":
		opts = append(opts, oidcclient.WithCLISendingCredentials())
	default:
		// Surprisingly cobra does not support this kind of flag validation. See https://github.com/spf13/pflag/issues/236
		return fmt.Errorf(
			"--upstream-identity-provider-type value not recognized: %s (supported values: oidc, ldap, activedirectory)",
			flags.upstreamIdentityProviderType)
	}

	var concierge *conciergeclient.Client
	if flags.conciergeEnabled {
		var err error
		concierge, err = conciergeclient.New(
			conciergeclient.WithEndpoint(flags.conciergeEndpoint),
			conciergeclient.WithBase64CABundle(flags.conciergeCABundle),
			conciergeclient.WithAuthenticator(flags.conciergeAuthenticatorType, flags.conciergeAuthenticatorName),
			conciergeclient.WithAPIGroupSuffix(flags.conciergeAPIGroupSuffix),
		)
		if err != nil {
			return fmt.Errorf("invalid Concierge parameters: %w", err)
		}
	}

	// --skip-browser replaces the default "browser open" function with one that prints to stderr.
	if flags.skipBrowser {
		opts = append(opts, oidcclient.WithBrowserOpen(func(url string) error {
			cmd.PrintErr("Please log in: ", url, "\n")
			return nil
		}))
	}

	if len(flags.caBundlePaths) > 0 || len(flags.caBundleData) > 0 {
		client, err := makeClient(flags.caBundlePaths, flags.caBundleData)
		if err != nil {
			return err
		}
		opts = append(opts, oidcclient.WithClient(client))
	}
	// Look up cached credentials based on a hash of all the CLI arguments and the cluster info.
	cacheKey := struct {
		Args        []string                   `json:"args"`
		ClusterInfo *clientauthv1beta1.Cluster `json:"cluster"`
	}{
		Args:        os.Args[1:],
		ClusterInfo: loadClusterInfo(),
	}
	var credCache *execcredcache.Cache
	if flags.credentialCachePath != "" {
		credCache = execcredcache.New(flags.credentialCachePath)
		if cred := credCache.Get(cacheKey); cred != nil {
			pLogger.Debug("using cached cluster credential.")
			return json.NewEncoder(cmd.OutOrStdout()).Encode(cred)
		}
	}

	pLogger.Debug("Performing OIDC login", "issuer", flags.issuer, "client id", flags.clientID)
	// Do the basic login to get an OIDC token.
	token, err := deps.login(flags.issuer, flags.clientID, opts...)
	if err != nil {
		return fmt.Errorf("could not complete Pinniped login: %w", err)
	}
	cred := tokenCredential(token)

	// If the concierge was configured, exchange the credential for a separate short-lived, cluster-specific credential.
	if concierge != nil {
		pLogger.Debug("Exchanging token for cluster credential", "endpoint", flags.conciergeEndpoint, "authenticator type", flags.conciergeAuthenticatorType, "authenticator name", flags.conciergeAuthenticatorName)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cred, err = deps.exchangeToken(ctx, concierge, token.IDToken.Token)
		if err != nil {
			return fmt.Errorf("could not complete Concierge credential exchange: %w", err)
		}
		pLogger.Debug("Successfully exchanged token for cluster credential.")
	} else {
		pLogger.Debug("No concierge configured, skipping token credential exchange")
	}

	// If there was a credential cache, save the resulting credential for future use.
	if credCache != nil {
		pLogger.Debug("caching cluster credential for future use.")
		credCache.Put(cacheKey, cred)
	}
	return json.NewEncoder(cmd.OutOrStdout()).Encode(cred)
}

func makeClient(caBundlePaths []string, caBundleData []string) (*http.Client, error) {
	pool := x509.NewCertPool()
	for _, p := range caBundlePaths {
		pem, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("could not read --ca-bundle: %w", err)
		}
		pool.AppendCertsFromPEM(pem)
	}
	for _, d := range caBundleData {
		pem, err := base64.StdEncoding.DecodeString(d)
		if err != nil {
			return nil, fmt.Errorf("could not read --ca-bundle-data: %w", err)
		}
		pool.AppendCertsFromPEM(pem)
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				RootCAs:    pool,
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	client.Transport = transport.DebugWrappers(client.Transport)
	return client, nil
}

func tokenCredential(token *oidctypes.Token) *clientauthv1beta1.ExecCredential {
	cred := clientauthv1beta1.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ExecCredential",
			APIVersion: "client.authentication.k8s.io/v1beta1",
		},
		Status: &clientauthv1beta1.ExecCredentialStatus{
			Token: token.IDToken.Token,
		},
	}
	if !token.IDToken.Expiry.IsZero() {
		cred.Status.ExpirationTimestamp = &token.IDToken.Expiry
	}
	return &cred
}

func SetLogLevel(lookupEnv func(string) (string, bool)) (*plog.PLogger, error) {
	debug, _ := lookupEnv("PINNIPED_DEBUG")
	if debug == "true" {
		err := plog.ValidateAndSetLogLevelGlobally(plog.LevelDebug)
		if err != nil {
			return nil, err
		}
	}
	logger := plog.New("Pinniped login: ")
	return &logger, nil
}

// mustGetConfigDir returns a directory that follows the XDG base directory convention:
//   $XDG_CONFIG_HOME defines the base directory relative to which user specific configuration files should
//   be stored. If $XDG_CONFIG_HOME is either not set or empty, a default equal to $HOME/.config should be used.
// [1] https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func mustGetConfigDir() string {
	const xdgAppName = "pinniped"

	if path := os.Getenv("XDG_CONFIG_HOME"); path != "" {
		return filepath.Join(path, xdgAppName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, ".config", xdgAppName)
}
