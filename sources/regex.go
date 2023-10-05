package sources

import (
	"context"
	"net/http"
	"time"
)

func init() {
	addSources(
		regexSource("https://free-proxy-list.net", "Proxy List", map[string]string{
			"http": "/",
		}, true, 30*time.Minute),
		regexSource("https://sslproxies.org", "SSL Proxy List", map[string]string{
			"https": "/",
		}, true, 30*time.Minute),
		regexSource("https://us-proxy.org", "US Proxy List", map[string]string{
			"http": "/",
		}, true, 30*time.Minute),
		regexSource("https://openproxylist.xyz", ":", map[string]string{
			"http":   "/http.txt",
			"socks4": "/socks4.txt",
			"socks5": "/socks5.txt",
		}, true, 1*time.Hour),
		regexSource("https://api.proxyscrape.com/v2/?request=getproxies&protocol=", ":", map[string]string{
			"http":   "http",
			"socks4": "socks4",
			"socks5": "socks5",
		}, true, 1*time.Hour),
		regexSource("https://proxyspace.pro", ":", map[string]string{
			"http":   "/http.txt",
			"https":  "/https.txt",
			"socks4": "/socks4.txt",
			"socks5": "/socks5.txt",
		}, true, 1*time.Hour),
		regexSource("https://rootjazz.com/proxies", ":", map[string]string{
			"http": "/proxies.txt",
		}, true, 1*time.Hour),
		regexSource("https://www.juproxy.com", ":", map[string]string{
			"http": "/free_api",
		}, true, 1*time.Hour),
		regexSource("https://openproxy.space", "Proxy List", map[string]string{
			"http":   "/list/http",
			"socks4": "/list/socks4",
			"socks5": "/list/socks5",
		}, false, 24*time.Hour),
		regexSource("https://proxypedia.org", "Proxy List", map[string]string{
			"http": "/",
		}, true, 10*time.Minute),
	)
}

func regexSource(home, expect string, files map[string]string, seed bool, freq time.Duration) Source {
	return Source{
		Homepage:  home,
		Frequency: freq,
		Seed:      seed,
		Feed: func(ctx context.Context, h *http.Client) Src {
			f := regexFeedBase(ctx, h, home, expect)
			m := merged()
			for t, loc := range files {
				m.refresh(f(loc, t))
			}
			return m
		},
	}
}
