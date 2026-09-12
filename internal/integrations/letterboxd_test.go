package integrations

import (
	"net/http"
	"testing"
)

func TestLetterboxdScrapesPublicListWithDirectTMDBIDs(t *testing.T) {
	client := NewLetterboxd(LetterboxdConfig{})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/max/list/weekend/":
			return jsonResponse(http.StatusOK, `<ul><li class="posteritem"><div data-film-id="a" data-target-link="/film/movie/" data-item-name="Movie &amp; More" data-item-full-display-name="Movie &amp; More (2025)"></div></li><li class="griditem"><div data-film-id="b" data-target-link="/film/show/" data-item-name="Show" data-item-full-display-name="Show (2024)"></div></li></ul>`), nil
		case "/film/movie/":
			return jsonResponse(http.StatusOK, `<a href="https://www.themoviedb.org/movie/10/" data-track-action="TMDB">TMDB</a>`), nil
		case "/film/show/":
			return jsonResponse(http.StatusOK, `<a data-track-action="TMDB" href="https://www.themoviedb.org/tv/20/">TMDB</a>`), nil
		default:
			t.Fatalf("path = %s", request.URL.Path)
			return nil, nil
		}
	})
	items, err := client.GetListItems(t.Context(), "max", "weekend", false, "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if items.Movies[0].Title != "Movie & More" || items.Movies[0].IDs.TMDB != 10 || items.Shows[0].IDs.TMDB != 20 {
		t.Fatalf("items = %#v", items)
	}
}

func TestLetterboxdFailsClosedWhenMarkupChanges(t *testing.T) {
	client := NewLetterboxd(LetterboxdConfig{})
	client.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `<html><body>Cloudflare challenge</body></html>`), nil
	})
	if _, err := client.GetListItems(t.Context(), "max", "weekend", false, "", 10); err == nil {
		t.Fatal("expected changed-markup error")
	}
}

func TestLetterboxdRequiresNormalizedIdentifiers(t *testing.T) {
	client := NewLetterboxd(LetterboxdConfig{})
	if _, err := client.GetListItems(t.Context(), "", "weekend", false, "", 10); err == nil {
		t.Fatal("expected owner error")
	}
}

func TestLetterboxdSkipsUnmappedItemsAndFollowsNextPage(t *testing.T) {
	client := NewLetterboxd(LetterboxdConfig{})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/max/list/weekend/":
			return jsonResponse(http.StatusOK, `<li class="posteritem"><div data-target-link="/film/unmapped/" data-item-name="Bad"></div></li><a rel="next" href="/max/list/weekend/page/2/">Next</a>`), nil
		case "/film/unmapped/":
			return jsonResponse(http.StatusOK, `<html></html>`), nil
		case "/max/list/weekend/page/2/":
			return jsonResponse(http.StatusOK, `<li class="posteritem"><div data-target-link="/film/good/" data-item-name="Good"></div></li>`), nil
		case "/film/good/":
			return jsonResponse(http.StatusOK, `<a href="https://www.themoviedb.org/movie/22/">TMDB</a>`), nil
		default:
			t.Fatalf("path=%s", request.URL.Path)
			return nil, nil
		}
	})
	items, err := client.GetListItems(t.Context(), "max", "weekend", false, "movie", 1)
	if err != nil || len(items.Movies) != 1 || items.Movies[0].IDs.TMDB != 22 || len(items.Warnings) != 1 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
