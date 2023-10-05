package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nfx/slrp/pmux"
	"github.com/rs/zerolog/log"
)

var sunnyUrl = "https://sunny9577.github.io/proxy-scraper/proxies.json"

type sunnyRecord struct {
	IP             string `json:"ip"`
	Port           string `json:"port"`
	Protocols      string `json:"type"`
	AnonymityLevel string `json:"anonymity"`
	Country        string `json:"country"`
}

func (r *sunnyRecord) Proxies() (proxies []pmux.Proxy) {
	addr := fmt.Sprintf("%s:%s", r.IP, r.Port)
	protocols := strings.Split(r.Protocols, "/")
	for _, protocol := range protocols {
		protocol = strings.ToLower(protocol)
	}

	var allowedProtocols = []string{
		"http", "https", "socks4", "socks5",
	}

	for _, v := range protocols {
		found := false
		for _, str := range allowedProtocols {
			if str == v {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		proxies = append(proxies, pmux.NewProxy(addr, v))
	}
	return
}

func init() {
	addSources(Source{
		name:      "sunny9577",
		Homepage:  "https://github.com/sunny9577/proxy-scraper",
		Frequency: 3 * time.Hour,
		Seed:      true,
		Feed:      simpleGen(sunny9577),
	})
}

func sunny9577(ctx context.Context, h *http.Client) (found []pmux.Proxy, err error) {
	var results []sunnyRecord
	log.Info().Msg("Loading sunny9577 database")
	r, err := h.Get(sunnyUrl)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &results)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return
	}
	for _, d := range results {
		found = append(found, d.Proxies()...)
	}
	return
}
