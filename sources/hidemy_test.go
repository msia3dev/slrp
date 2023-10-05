package sources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHidemy(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("./testdata/hidemy")))
	defer server.Close()
	testSource(t, func(ctx context.Context) Src {
		return ByName("hidemy.io").Feed(ctx, http.DefaultClient)
	}, 5)
}
