package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"regexp"
	"strings"
	"testing"
)

func TestCloudflareCacheClear(t *testing.T) {
	const goodToken = "HsSgD4SnMnpkrZWk6MFYP24tDCFDfcJjvaFzmnTr"
	const goodZoneID = "JdgHwtCDFN3aEsjprrvEUZMy5Vja3uhR"

	reZoneURL := regexp.MustCompile(`\/client\/v4\/zones\/(\w+)\/purge_cache`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr, _ := httputil.DumpRequest(r, true)
		log.Println(string(rr))

		// Handle token, this will yield an empty string if the header wasn't there
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		switch token {
		case "":
			http.Error(w, "no token sent", http.StatusForbidden)
			return
		case goodToken:
			// handle below
		default:
			http.Error(w, "invalid token sent", http.StatusUnauthorized)
			return
		}

		// Find if the endpoint for clear cache was hit
		if matches := reZoneURL.FindStringSubmatch(r.URL.Path); len(matches) == 2 {
			// If zone doesn't match, return 404
			if matches[1] != goodZoneID {
				http.Error(w, "unknown zone id: "+matches[1], http.StatusNotFound)
				return
			}

			// Otherwise return 200
			w.WriteHeader(http.StatusOK)
			return
		}

		// If we're hitting any other endpoint, simply error out with 404
		http.Error(w, "404: unknown endpoint hit", http.StatusNotFound)
		return
	}))

	defer srv.Close()

	cases := []struct {
		name  string
		zone  string
		token string
		err   error
	}{
		{
			name:  "good run",
			zone:  goodZoneID,
			token: goodToken,
		},
		{
			name:  "empty token",
			zone:  goodZoneID,
			token: "",
			err:   errNoToken,
		},
		{
			name:  "empty zone",
			zone:  "",
			token: goodToken,
			err:   errNoZone,
		},
		{
			name:  "zone id less than 32 characters",
			zone:  "thisislessthan32chars",
			token: goodToken,
			err:   errZoneTooShort,
		},
		{
			name:  "fake token",
			zone:  goodZoneID,
			token: "my-fake-token",
			err:   &requestError{statusCode: http.StatusUnauthorized},
		},
		{
			name:  "unknown zone",
			zone:  "C63R6bm5uyGyv2skaSbj2YmvrJXmgjS3",
			token: goodToken,
			err:   &requestError{statusCode: http.StatusNotFound},
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(tt *testing.T) {
			clf := newCloudflare(srv.URL+"/", v.token)

			err := clf.clearCache(v.zone)
			if err != nil {
				if fmt.Sprintf("%T", err) != fmt.Sprintf("%T", v.err) {
					tt.Fatalf("expecting error to be of type %T, but got %T: %s", v.err, err, err.Error())
				}

				e1, m1 := err.(*requestError)
				e2, m2 := v.err.(*requestError)

				if m1 && m2 && e1.statusCode != e2.statusCode {
					tt.Fatalf("expecting status code on error to be %d, got %d", e2.statusCode, e1.statusCode)
				}
			}

			if err == nil && v.err != nil {
				tt.Fatalf("expecting function to fail with error of type %T but got no error", v.err)
			}
		})
	}
}
