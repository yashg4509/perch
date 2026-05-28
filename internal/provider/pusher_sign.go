package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PusherSignedGET performs a signed Channels HTTP API GET (auth_key + HMAC signature).
func PusherSignedGET(ctx context.Context, client *http.Client, baseURL, path, appKey, appSecret string) (int, error) {
	if client == nil {
		client = HTTPClientForAPI()
	}
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"auth_key":       appKey,
		"auth_timestamp": ts,
		"auth_version":   "1.0",
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, k+"="+params[k])
	}
	query := strings.Join(pairs, "&")
	signPayload := strings.Join([]string{"GET", path, query}, "\n")
	mac := hmac.New(sha256.New, []byte(appSecret))
	_, _ = mac.Write([]byte(signPayload))
	sig := hex.EncodeToString(mac.Sum(nil))
	fullURL := strings.TrimSuffix(baseURL, "/") + path + "?" + query + "&auth_signature=" + sig

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

// PusherChannelsPath returns the list-channels path for an app id.
func PusherChannelsPath(appID string) string {
	return "/apps/" + strings.TrimSpace(appID) + "/channels"
}

// PusherAPIBaseURL returns the Channels REST host for a cluster (e.g. us2 → api-us2.pusher.com).
func PusherAPIBaseURL(cluster string) string {
	cluster = strings.TrimSpace(cluster)
	if cluster == "" {
		return "https://api.pusher.com"
	}
	return "https://api-" + cluster + ".pusher.com"
}

// PusherProbeOK reports whether a signed channels list call succeeded.
func PusherProbeOK(statusCode int, err error) (bool, string) {
	if err != nil {
		return false, truncate([]byte(err.Error()), 80)
	}
	if statusCode >= 200 && statusCode < 300 {
		return true, ""
	}
	return false, fmt.Sprintf("pusher http %d", statusCode)
}
