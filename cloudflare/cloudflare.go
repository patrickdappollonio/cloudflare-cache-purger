package cloudflare

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

// Cache is the cache object that holds communication
// procedures against the Cloudflare Cache API
type Cache struct {
	baseEndpoint string
	token        string
	client       *http.Client
	logger       *log.Logger
}

// DefaultEndpoint holds the default endpoint
const DefaultEndpoint = "https://api.cloudflare.com/"

const cacheEndpointFormat = "client/v4/zones/%s/purge_cache"

var (
	// ErrNoZone is returned when no zone ID is specified
	ErrNoZone = errors.New("zone ID not set")

	// ErrZoneTooShort is returned when the zone ID provided is too short
	ErrZoneTooShort = errors.New("zone ID must be 32 characters long")

	// ErrNoToken is returned when there's no token specified
	ErrNoToken = errors.New("token not set")
)

// New creates a new Cloudflare cache clear object
func New(token string) *Cache {
	return &Cache{
		token: token,
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

// SetEndpoint allows you to change the default endpoint
// for Cloudflare (set with DefaultEndpoint) to any other
func (c *Cache) SetEndpoint(endpoint string) {
	c.baseEndpoint = endpoint
}

// SetDebug enables debug information
func (c *Cache) SetDebug(debug bool) {
	if debug {
		c.logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		c.logger = nil
	}
}

// SetDebugOutput allows to redirect the output of the cache
// purger to another destination other than stdout
func (c *Cache) SetDebugOutput(w io.Writer) {
	if c.logger == nil {
		return
	}

	c.logger.SetOutput(w)
}

func (c *Cache) log(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(format, args...)
	}
}

func (c *Cache) logIf(a func() string) {
	if c.logger == nil {
		return
	}

	if s := a(); s != "" {
		c.log("%s", s)
	}
}

// Clear clears the cache by making the request to the Cloudflare
// API and issuing a request to delete them
func (c *Cache) Clear(zone string) error {
	if zone == "" {
		return ErrNoZone
	}

	if len(zone) != 32 {
		return ErrZoneTooShort
	}

	base := c.baseEndpoint
	if base == "" {
		base = DefaultEndpoint
	}

	if c.token == "" {
		return ErrNoToken
	}

	endpoint := fmt.Sprintf(base+cacheEndpointFormat, zone)
	c.log("Setting endpoint to be: %s", endpoint)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(`{"purge_everything":true}`))
	if err != nil {
		return &HTTPClientError{kind: "build request", endpoint: endpoint, original: err}
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	c.logIf(func() string {
		req, _ := httputil.DumpRequestOut(req, true)
		return fmt.Sprintf("Request:\n%s", string(req))
	})

	res, err := c.client.Do(req)
	if err != nil {
		return &HTTPClientError{kind: "execute request", endpoint: endpoint, original: err}
	}

	c.logIf(func() string {
		req, _ := httputil.DumpResponse(res, true)
		return fmt.Sprintf("Response:\n%s", string(req))
	})

	if res.StatusCode != http.StatusOK {
		c.log("Status code for request isn't 200, got: %d %s", res.StatusCode, res.Status)
		return &HTTPStatusCodeError{endpoint: endpoint, statusCode: res.StatusCode}
	}

	fmt.Fprintln(os.Stdout, "Cache deleted for zone:", zone)
	return nil
}

// HTTPClientError encapsulates any unexpected error message sent by
// the Cloudflare API
type HTTPClientError struct {
	original error
	endpoint string
	kind     string
}

// Error implements the error interface
func (rr *HTTPClientError) Error() string {
	return fmt.Sprintf("unable to %s to %q: %s", rr.kind, rr.endpoint, rr.original.Error())
}

// HTTPStatusCodeError encapsulates an incorrect HTTP status code
// returned by the Cloudflare API
type HTTPStatusCodeError struct {
	endpoint   string
	statusCode int
}

func (a *HTTPStatusCodeError) Error() string {
	var reason string

	switch a.statusCode {
	case http.StatusUnauthorized:
		reason = "token is not authorized to perform the action, token needs permission \"Zone > Cache Purge\""
	case http.StatusForbidden:
		reason = "invalid token provided: token is incorrect or no longer valid"
	case http.StatusBadRequest:
		reason = "Cloudflare rejected the request since it's malformed, most likely the Zone ID is incorrect"
	default:
		reason = fmt.Sprintf("%d %s", a.statusCode, http.StatusText(a.statusCode))
	}

	return fmt.Sprintf("unable to perform request to %q: %s", a.endpoint, reason)
}
