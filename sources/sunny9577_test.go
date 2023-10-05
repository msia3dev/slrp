package sources

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSunny9577(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("./testdata/sunny9577")))
	defer server.Close()
	sunnyUrl = fmt.Sprintf("%s/page", server.URL)
	testSource(t, func(ctx context.Context) Src {
		return ByName("sunny9577").Feed(ctx, http.DefaultClient)
	}, 5)
}
