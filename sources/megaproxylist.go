package sources

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/nfx/slrp/pmux"
	"github.com/rs/zerolog/log"
)

func init() {
	//Sources = append(Sources, Source{
	//	ID:        65,
	//	Homepage:  "https://www.megaproxylist.net",
	//	Frequency: 24 * time.Hour,
	//	Feed:      simpleGen(megaproxylist),
	//})
}

var megaproxylistUrl = "https://www.megaproxylist.net/download-proxy-list.aspx"
var downloadUrlRegex = regexp.MustCompile(`/download/megaproxylist-.*-.*_(?P<hash>.*).zip`)

func unzipInMemory(body []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to open file as zip file: %w", err)
	}

	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		if zipFile.Name != "megaproxylist.csv" {
			continue
		}

		unzippedFileBytes, err := readZipFile(zipFile)
		if err != nil {
			return nil, fmt.Errorf("zip: can't read megaproxylist.csv")
		}
		return unzippedFileBytes, nil

	}
	return nil, fmt.Errorf("zip: can't find desired file")
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func getIpAddr(ctx context.Context, address string) (string, error) {
	if net.ParseIP(address) != nil {
		return address, nil
	}
	addrs, err := net.DefaultResolver.LookupIP(ctx, "ip4", address)
	if err != nil {
		return "", fmt.Errorf("Failed to resolve domain %s: %w", address, err)
	}
	return addrs[0].String(), nil
}

// Scrapes https://www.megaproxylist.net
func megaproxylist(ctx context.Context, h *http.Client) (found []pmux.Proxy, err error) {
	log.Info().Msg("Loading proxy Megaproxy database")

	resp, err := h.Get(megaproxylistUrl)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, err
	}
	if resp.Body == nil {
		log.Error().Err(err).Msg("")
		return nil, fmt.Errorf("empty body")
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	matches := downloadUrlRegex.FindStringSubmatch(string(body))
	if len(matches) == 0 {
		return nil, fmt.Errorf("download url not found")
	}
	hashIndex := downloadUrlRegex.SubexpIndex("hash")
	hash := matches[hashIndex]

	downloadUrl := fmt.Sprintf("https://www.megaproxylist.net/download/megaproxylist-csv-%s_%s.zip", time.Now().Format("20060102"), hash)

	resp, err = h.Get(downloadUrl)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, fmt.Errorf("empty body")
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	csvData, err := unzipInMemory(body)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewBuffer(csvData))
	r.Comma = ';'
	r.TrimLeadingSpace = true

	// trick to skip header
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if len(record) != 4 {
			log.Info().Msg("length skipping")
			continue
		}

		addr, err := getIpAddr(ctx, record[0])
		if err != nil {
			log.Error().Msg(err.Error())
			continue
		}

		found = append(found,
			pmux.NewProxy(fmt.Sprintf("%s:%s", addr, record[1]),
				"http"))
	}

	return found, nil
}
