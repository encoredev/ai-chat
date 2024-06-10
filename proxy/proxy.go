package proxy

import (
	"context"
	"net/url"

	"github.com/cockroachdb/errors"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"

	encore "encore.dev"
	"encore.dev/rlog"
)

// This uses Encore's built-in secrets manager, learn more: https://encore.dev/docs/primitives/secrets
var secrets struct {
	NGrokToken  string
	NGrokDomain string
}

var BaseURL = func() url.URL {
	ctx := context.Background()
	baseURL := encore.Meta().APIBaseURL
	if encore.Meta().Environment.Cloud == encore.CloudLocal {
		if secrets.NGrokToken == "" {
			rlog.Warn("NGrokToken or NGrokDomain is not set, skipping ngrok")
			return baseURL
		}
		rlog.Info("Starting ngrok")
		session, err := ngrok.Connect(ctx, ngrok.WithAuthtoken(secrets.NGrokToken))
		if err != nil {
			rlog.Warn("Failed to start ngrok", "error", errors.Wrap(err, "ngrok.Connect"))
			return baseURL
		}
		cfg := config.HTTPEndpoint()
		rlog.Info("Started ngrok")
		if secrets.NGrokDomain != "" {
			cfg = config.HTTPEndpoint(config.WithDomain(secrets.NGrokDomain))
		}
		f, err := session.ListenAndForward(ctx, &baseURL, cfg)
		if err != nil {
			rlog.Warn("Failed to start ngrok", "error", errors.Wrap(err, "ngrok.Connect"))
			return baseURL
		}
		proxyURL, err := url.Parse(f.URL())
		rlog.Info("ngrok is started", "url", proxyURL.String())
		if err != nil {
			rlog.Warn("Failed to start ngrok", "error", errors.Wrap(err, "ngrok.Connect"))
			return baseURL
		}
		rlog.Info("Started ngrok")
		return *proxyURL
	}
	return baseURL
}()
