package sources

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func init() {
	addSources(
		github("TheSpeedX/PROXY-List", map[string]string{
			"http":   "/master/http.txt",
			"socks4": "/master/socks4.txt",
			"socks5": "/master/socks5.txt",
		}, 3*time.Hour),
		github("andigwandi/free-proxy", map[string]string{
			"http": "/main/proxy_list.txt",
		}, 2*time.Hour),
		github("aslisk/proxyhttps", map[string]string{
			"https":  "/main/https.txt",
			"socks4": "/main/socks4.txt",
		}, 24*time.Hour),
		github("fahimscirex/proxybd", map[string]string{
			"http":   "/master/proxylist/http.txt",
			"socks4": "/master/proxylist/socks4.txt",
			"socks5": "/master/proxylist/socks5.txt",
		}, 4*time.Hour),
		github("hookzof/socks5_list", map[string]string{
			"socks5": "/master/proxy.txt",
		}, 5*time.Minute),
		github("mmpx12/proxy-list", map[string]string{
			"http":   "/master/http.txt",
			"https":  "/master/https.txt",
			"socks4": "/master/socks4.txt",
			"socks5": "/master/socks5.txt",
		}, 1*time.Hour),
		github("monosans/proxy-list", map[string]string{
			"http":   "/main/proxies/http.txt",
			"socks4": "/main/proxies/socks4.txt",
			"socks5": "/main/proxies/socks5.txt",
		}, 30*time.Minute),
		github("ObcbO/getproxy", map[string]string{
			"https":  "/master/file/https.txt",
			"http":   "/master/file/http.txt",
			"socks4": "/master/file/socks4.txt",
			"socks5": "/master/file/socks5.txt",
		}, 6*time.Hour),
		github("officialputuid/KangProxy", map[string]string{
			"http":   "/KangProxy/https/http.txt",
			"https":  "/KangProxy/https/https.txt",
			"socks4": "/KangProxy/socks4/socks4.txt",
			"socks5": "/KangProxy/socks5/socks5.txt",
		}, 2*time.Hour),
		github("proxy4parsing/proxy-list", map[string]string{
			"http": "/main/http.txt",
		}, 15*time.Minute),
		github("roosterkid/openproxylist", map[string]string{
			"https":  "/main/HTTPS_RAW.txt",
			"socks4": "/main/SOCKS4_RAW.txt",
			"socks5": "/main/SOCKS5_RAW.txt",
		}, 30*time.Minute),
		github("ShiftyTR/Proxy-List", map[string]string{
			"http":   "/master/http.txt",
			"https":  "/master/https.txt",
			"socks4": "/master/socks4.txt",
			"socks5": "/master/socks5.txt",
		}, 10*time.Minute),
		github("Zaeem20/FREE_PROXIES_LIST", map[string]string{
			"http":   "/master/http.txt",
			"https":  "/master/https.txt",
			"socks4": "/master/socks4.txt",
			"socks5": "/master/socks5.txt",
		}, 1*time.Hour),
		github("zevtyardt/proxy-list", map[string]string{
			"http":   "/main/http.txt",
			"socks4": "/main/socks4.txt",
			"socks5": "/main/socks5.txt",
		}, 2*time.Hour),
		github("zloi-user/hideip.me", map[string]string{
			"http":   "/main/http.txt",
			"https":  "/main/https.txt",
			"socks4": "/main/socks4.txt",
			"socks5": "/main/socks5.txt",
		}, 8*time.Minute),
		github("ErcinDedeoglu", map[string]string{
			"http":   "/proxies/main/proxies/http.txt",
			"https":  "/proxies/main/proxies/https.txt",
			"socks4": "/proxies/main/proxies/socks4.txt",
			"socks5": "/proxies/main/proxies/socks5.txt",
		}, 8*time.Minute),
		github("prxchk/proxy-list", map[string]string{
			"http":   "/main/http.txt",
			"socks4": "/main/socks4.txt",
			"socks5": "/main/socks5.txt",
		}, 10*time.Minute),
		github("proxifly/free-proxy-list", map[string]string{
			"http":   "/main/proxies/protocols/http/data.txt",
			"socks4": "/main/proxies/protocols/http/socks4.txt",
			"socks5": "/main/proxies/protocols/http/socks5.txt",
		}, 10*time.Minute),
	)
}

func github(repo string, files map[string]string, freq time.Duration) Source {
	ownerSplit := strings.Split(repo, "/")
	prefix := fmt.Sprintf("https://raw.githubusercontent.com/%s", repo)
	return Source{
		name:      ownerSplit[0],
		Homepage:  fmt.Sprintf("https://github.com/%s", repo),
		UrlPrefix: prefix,
		Frequency: freq,
		Seed:      true,
		Feed: func(ctx context.Context, h *http.Client) Src {
			f := regexFeedBase(ctx, h, prefix, ":")
			m := merged()
			for t, loc := range files {
				m.refresh(f(loc, t))
			}
			return m
		},
	}
}
