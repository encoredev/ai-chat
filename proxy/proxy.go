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
			panic(errors.Wrap(err, "ngrok.Connect"))
		}
		cfg := config.HTTPEndpoint()
		if secrets.NGrokDomain != "" {
			cfg = config.HTTPEndpoint(config.WithDomain(secrets.NGrokDomain))
		}
		f, err := session.ListenAndForward(ctx, &baseURL, cfg)
		if err != nil {
			panic(errors.Wrap(err, "ngrok.ListenAndForward"))
		}
		proxyURL, err := url.Parse(f.URL())
		rlog.Info("ngrok is started", "url", proxyURL.String())
		if err != nil {
			panic(errors.Wrap(err, "url.Parse"))
		}

		return *proxyURL
	}
	return baseURL
}()
