package reader

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// The api reader pulls a bill from an HTTP endpoint instead of a file. The
// import argument is the source (an address, an account id) and fills
// {source} in the URL; {env.NAME} reads an environment variable so keys
// never live in a template. The response is then read as json (or xml)
// with the usual records/columns contract.
//
//	reader:
//	  format: api
//	  url: "https://api.etherscan.io/v2/api?chainid=1&module=account&action=txlist&address={source}&apikey={env.ETHERSCAN_KEY}"
//	  records: "//result/*"
//	  columns: { hash: hash, time: timeStamp, value: value }

var placeholderPattern = regexp.MustCompile(`\{(source|env\.[A-Za-z_][A-Za-z0-9_]*)\}`)

// ExpandURL fills {source} and {env.X} placeholders. Missing environment
// variables are an error so a bad key never becomes a silent 401.
func ExpandURL(template, source string) (string, error) {
	var missing []string
	out := placeholderPattern.ReplaceAllStringFunc(template, func(m string) string {
		key := m[1 : len(m)-1]
		if key == "source" {
			return url.QueryEscape(source)
		}
		name := strings.TrimPrefix(key, "env.")
		value, ok := os.LookupEnv(name)
		if !ok || value == "" {
			missing = append(missing, name)
			return ""
		}
		return url.QueryEscape(value)
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("api reader: environment variable %s is not set", strings.Join(missing, ", "))
	}
	return out, nil
}

func fetchAPI(source string, cfg Config) ([]byte, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("api reader requires `url`")
	}
	endpoint, err := ExpandURL(cfg.URL, source)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for name, value := range cfg.Headers {
		expanded, err := ExpandURL(value, source)
		if err != nil {
			return nil, err
		}
		req.Header.Set(name, expanded)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json, application/xml;q=0.9, */*;q=0.5")
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api reader: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api reader: %s returned %s: %s", redactURL(endpoint), resp.Status, strings.TrimSpace(string(bytes.TrimSpace(body)[:min(len(body), 200)])))
	}
	return body, nil
}

// readAPIResponse parses a fetched body as the configured tree format.
func readAPIResponse(data []byte, cfg Config) (Table, error) {
	inner := cfg
	inner.Format = NormalizeFormat(cfg.Response)
	if inner.Format == "csv" { // Response unset: default to json
		inner.Format = "json"
	}
	switch inner.Format {
	case "json":
		return readJSON(data, inner)
	case "xml":
		return readXML(data, inner)
	default:
		return Table{}, fmt.Errorf("api reader: response must be json or xml, got %q", cfg.Response)
	}
}

// redactURL hides query values so keys do not end up in error messages.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	for k := range q {
		q.Set(k, "…")
	}
	u.RawQuery = q.Encode()
	return u.String()
}
