package sources

import (
	"context"
	"net/http"
	"time"
)

func init() {
	addSources(
		Source{
			Seed:      true,
			name:      "free-proxy-list",
			Homepage:  "https://free-proxy-list.net",
			Frequency: 30 * time.Minute,
			Feed:      httpProxyRegexFeed("https://free-proxy-list.net", "Free Proxy List"),
		}, Source{
			Seed:      true,
			name:      "anonymous-free-proxy",
			Homepage:  "https://free-proxy-list.net/anonymous-proxy.html",
			Frequency: 30 * time.Minute,
			Feed:      httpProxyRegexFeed("https://free-proxy-list.net/anonymous-proxy.html", "Anonymous Proxy"),
		}, Source{
			name:      "uk-proxy",
			Seed:      true,
			Homepage:  "https://free-proxy-list.net/uk-proxy.html",
			Frequency: 30 * time.Minute,
			Feed:      httpProxyRegexFeed("https://free-proxy-list.net/uk-proxy.html", "UK Proxy List"),
		}, Source{
			name:      "ssl-proxy",
			Seed:      true,
			Homepage:  "https://www.sslproxies.org/",
			Frequency: 30 * time.Minute,
			Feed: func(ctx context.Context, h *http.Client) Src {
				return gen(regexFeed(ctx, h, "https://www.sslproxies.org/", "https", "SSL Proxy"))
			},
		}, Source{
			Seed:      true,
			name:      "socks-proxy-list",
			Homepage:  "https://www.socks-proxy.net/",
			Frequency: 30 * time.Minute,
			Feed: func(ctx context.Context, h *http.Client) Src {
				return gen(regexFeed(ctx, h, "https://www.socks-proxy.net/", "socks4", "Socks Proxy"))
			},
		})
}
