package sources

import (
	"context"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/nfx/go-htmltable"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nfx/slrp/pmux"
	"github.com/rs/zerolog/log"
)

var hidemyNamePages []string
var hidemyUrl = "https://hidemy.io/en/proxy-list/?anon=34"

func init() {
	addSources(Source{
		name:      "hidemy.io",
		Homepage:  "https://hidemy.io",
		Frequency: 1 * time.Hour,
		Feed:      simpleGen(hidemyName),
		Seed:      true,
	})
}

// Scrapes http://hidemy.name/
func hidemyName(ctx context.Context, h *http.Client) (found []pmux.Proxy, err error) {
	launch := launcher.
		NewUserMode().
		Set("headless", "new").
		UserDataDir("tmp/t").
		MustLaunch()

	// Launch a new browser with default options, and connect to it.
	browser := rod.New().ControlURL(launch).NoDefaultDevice().MustConnect()

	// Even you forget to close, rod will close it after main process ends.
	defer browser.MustClose()

	page := stealth.MustPage(browser)

	// new page clear cookie，inject stealth.JS
	go browser.EachEvent(func(e *proto.TargetTargetCreated) {
		if e.TargetInfo.Type != proto.TargetTargetInfoTypePage {
			return
		}
		browser.MustPageFromTargetID(e.TargetInfo.TargetID).MustEvalOnNewDocument(stealth.JS)
	})()

	// Create a new page
	wait := page.MustNavigate(hidemyUrl).
		Timeout(20 * time.Second).
		MustWaitLoad().
		MustWaitDOMStable().
		MustWaitIdle().
		MustWaitRequestIdle()
	wait()
	body := page.MustHTML()

	if !strings.Contains(body, "Online database of proxy lists") {
		browser.MustClose()
		return nil, fmt.Errorf("failed to bypass cloudflare")
	}

	// Parse the HTML content using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, nil
	}

	paginationDiv := doc.Find("div.pagination")
	listItems := paginationDiv.Find("li")
	secondToLastListItem := listItems.Eq(listItems.Length() - 2) // Get the second-to-last li element
	lastUrl, exists := secondToLastListItem.Find("a").Attr("href")
	lastPage := 64
	if exists {
		u, err := url.Parse(lastUrl)
		if err != nil {
			log.Error().Err(err).Msg("")
		} else {
			m, _ := url.ParseQuery(u.RawQuery)
			lastPage, _ = strconv.Atoi(m["start"][0])
		}
	}

	// 64 per page
	pattern := "https://hidemy.io/en/proxy-list/?anon=34&start=%d"
	for i := 0; i < lastPage+1; i += 64 {
		url := fmt.Sprintf(pattern, i)
		hidemyNamePages = append(hidemyNamePages, url)
	}

	fetch := func(url string) (f []pmux.Proxy, e error) {
		b := page.MustNavigate(url).MustWaitLoad().MustHTML()

		p, _ := htmltable.NewFromString(b)
		if p.Len() == 0 {
			log.Error().Str("url", url).Msg("no tables found")
			return
		}
		e = p.Each3("IP address", "Port", "Type", func(host, port, types string) error {
			for _, v := range strings.Split(types, ",") {
				v = strings.ToLower(strings.TrimSpace(v))
				proxy := pmux.NewProxy(fmt.Sprintf("%s:%s", host, port), v)
				f = append(f, proxy)
			}
			return nil
		})
		if e != nil {
			e = skipErr(e, strEC{"url", url})
		}
		return
	}

	for _, scrapeUrl := range hidemyNamePages {
		proxies, _ := fetch(scrapeUrl)
		found = append(found, proxies...)
	}
	return
}
