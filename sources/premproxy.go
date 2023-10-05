package sources

import (
	"context"
	"errors"
	"fmt"
	"github.com/nfx/slrp/pmux"
	"net/http"
	"regexp"
	"strings"

	"github.com/dop251/goja"
)

func init() {
	//Sources = append(Sources, Source{
	//	ID:        5,
	//	Homepage:  "https://premproxy.com/list/",
	//	UrlPrefix: "https://premproxy.com/",
	//	Frequency: 6 * time.Hour,
	//	Feed:      premproxy,
	//})
}

func deobfuscatePorts(script string) (map[string]string, error) {
	vm := goja.New()
	_, err := vm.RunString(`
	var document = null;
	var readyCb = null;
	var sets = {};
	function $(param) {
		this.ready = function(cb) {
			readyCb = cb;
		}
		this.html = function(value) {
			sets[param] = value;
		}
		return this
	}`)
	if err != nil {
		return nil, err
	}
	_, err = vm.RunString(script)
	if err != nil {
		return nil, err
	}
	_, err = vm.RunString(`readyCb()`)
	if err != nil {
		return nil, err
	}
	var sets map[string]string
	err = vm.ExportTo(vm.Get("sets"), &sets)
	return sets, err
}

func permproxyMapping(ctx context.Context, h *http.Client, html []byte, referer string) (map[string]string, error) {
	split := strings.Split(referer, "/")
	match := premproxyPortMappingScriptRE.FindSubmatch(html)
	if len(match) == 0 {
		return nil, errors.New("cannot find script location")
	}
	scriptUrl := fmt.Sprintf("%s//%s%s", split[0], split[2], string(match[1]))
	packedJS, serial, err := req{
		URL:              scriptUrl,
		ExpectInResponse: "function(p,a,c,k,e,d)",
		Headers: map[string]string{
			"Referer": referer,
		},
	}.Do(ctx, h)
	if err != nil {
		return nil, err
	}
	// TODO: perhaps we should do "parent serial"?...
	mapping, err := deobfuscatePorts(string(packedJS))
	if err != nil {
		return nil, wrapError(err, intEC{"serial", serial})
	}
	return mapping, err
}

func premproxyFetchPage(ctx context.Context, h *http.Client, url string) ([]byte, map[string]string, error) {
	html, serial, err := req{
		URL:              url,
		ExpectInResponse: "Proxies on this list",
		Headers: map[string]string{
			"Referer": "https://premproxy.com/",
		},
	}.Do(ctx, h)
	if err != nil {
		return nil, nil, err
	}
	mapping, err := permproxyMapping(ctx, h, html, url)
	if err != nil {
		return nil, nil, skipErr(err, intEC{"parentSerial", serial})
	}
	return html, mapping, nil
}

var premproxyPortMappingScriptRE = regexp.MustCompile(`(?m)<script src="(/(js|js-socks)/[^\.]+.js)">`)
var premproxyObfuscatedAddrRE = regexp.MustCompile(`(?m)\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}\|[^"]+`)
var premproxyObfuscatedSocksAddrRE = regexp.MustCompile(`(?m)(\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}\|[^"]+).*(SOCKS[4|5])`)

func premproxyHttpPage(ctx context.Context, h *http.Client, url string) func() ([]pmux.Proxy, error) {
	return func() (found []pmux.Proxy, err error) {
		html, mapping, err := premproxyFetchPage(ctx, h, url)
		if err != nil {
			return nil, err
		}
		for _, match := range premproxyObfuscatedAddrRE.FindAllString(string(html), -1) {
			split := strings.Split(match, "|")
			if len(split) != 2 {
				continue
			}
			port, ok := mapping["."+split[1]]
			if !ok {
				continue
			}
			addr := fmt.Sprintf("%s:%s", split[0], port)
			found = append(found, pmux.HttpsProxy(addr))
		}
		return found, nil
	}
}

func premproxySocksPage(ctx context.Context, h *http.Client, url string) func() ([]pmux.Proxy, error) {
	return func() (found []pmux.Proxy, err error) {
		html, mapping, err := premproxyFetchPage(ctx, h, url)
		if err != nil {
			return nil, err
		}
		for _, match := range premproxyObfuscatedSocksAddrRE.FindAllStringSubmatch(string(html), -1) {
			split := strings.Split(match[1], "|")
			if len(split) != 2 {
				continue
			}
			port, ok := mapping["."+split[1]]
			if !ok {
				continue
			}
			addr := fmt.Sprintf("%s:%s", split[0], port)
			found = append(found, pmux.NewProxy(addr, strings.ToLower(match[2])))
		}
		return found, nil
	}
}

var premProxyHttpPages []string
var premProxySocksPages []string

func init() {
	list := "https://premproxy.com/list/%02d.htm"
	for i := 1; i <= 5; i++ {
		url := fmt.Sprintf(list, i)
		premProxyHttpPages = append(premProxyHttpPages, url)
	}
	list = "https://premproxy.com/socks-list/%02d.htm"
	for i := 1; i <= 13; i++ {
		url := fmt.Sprintf(list, i)
		premProxySocksPages = append(premProxySocksPages, url)
	}
}

func premproxy(ctx context.Context, h *http.Client) Src {
	m := merged()
	for _, url := range premProxyHttpPages {
		m.refresh(premproxyHttpPage(ctx, h, url))
	}
	for _, url := range premProxySocksPages {
		m.refresh(premproxySocksPage(ctx, h, url))
	}
	return m
}
