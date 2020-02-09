package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

type cloudflare struct {
	baseEndpoint string
	token        string
	client       *http.Client
	logger       *log.Logger
}

const (
	defaultCloudflareEndpoint = "https://api.cloudflare.com/"
	cacheEndpointFormat       = "client/v4/zones/%s/purge_cache"
)

var (
	errNoZone  = errors.New("zone ID not set")
	errNoToken = errors.New("token not set")
)

func newCloudflare(baseEndpoint, token string) *cloudflare {
	return &cloudflare{
		baseEndpoint: baseEndpoint,
		token:        token,
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
			},
		},
		logger: nil,
	}
}

func (c *cloudflare) debug(debug bool) {
	if debug {
		c.logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		c.logger = nil
	}
}

func (c *cloudflare) log(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(format, args...)
	}
}

func (c *cloudflare) logIf(a func() string) {
	if c.logger == nil {
		return
	}

	if s := a(); s != "" {
		c.log("%s", s)
	}
}

func (c *cloudflare) clearCache(zone string) error {
	if zone == "" {
		return errNoZone
	}

	base := c.baseEndpoint
	if base == "" {
		base = defaultCloudflareEndpoint
	}

	if c.token == "" {
		return errNoToken
	}

	endpoint := fmt.Sprintf(base+cacheEndpointFormat, zone)
	c.log("Setting endpoint to be: %s", endpoint)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(`{"purge_everything":true}`))
	if err != nil {
		return &reqresError{kind: "build request", endpoint: endpoint, original: err}
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	c.logIf(func() string {
		req, _ := httputil.DumpRequestOut(req, true)
		return fmt.Sprintf("Request:\n%s", string(req))
	})

	res, err := c.client.Do(req)
	if err != nil {
		return &reqresError{kind: "execute request", endpoint: endpoint, original: err}
	}

	c.logIf(func() string {
		req, _ := httputil.DumpResponse(res, true)
		return fmt.Sprintf("Response:\n%s", string(req))
	})

	if res.StatusCode != http.StatusOK {
		c.log("Status code for request isn't 200, got: %d %s", res.StatusCode, res.Status)
		return &requestError{endpoint: endpoint, statusCode: res.StatusCode}
	}

	return nil
}

type reqresError struct {
	original error
	endpoint string
	kind     string
}

func (rr *reqresError) Error() string {
	return fmt.Sprintf("unable to %s to %q: %s", rr.kind, rr.endpoint, rr.original.Error())
}

type requestError struct {
	endpoint   string
	statusCode int
}

func (a *requestError) Error() string {
	var reason string

	switch a.statusCode {
	case http.StatusUnauthorized:
		reason = "token is not authorized to perform the action, token needs permission \"Zone > Cache Purge\""
	case http.StatusForbidden:
		reason = "invalid token provided: token is incorrect or no longer valid"
	default:
		reason = fmt.Sprintf("%d %s", a.statusCode, http.StatusText(a.statusCode))
	}

	return fmt.Sprintf("unable to perform request to %q: %s", a.endpoint, reason)
}
