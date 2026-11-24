package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Discover reads the issuer's OpenID configuration and returns its userinfo endpoint.
func Discover(ctx context.Context, issuerURL string) (string, error) {
	var config struct {
		UserInfoEndpoint string `json:"userinfo_endpoint"`
	}
	if err := getJSON(ctx, strings.TrimSuffix(issuerURL, "/")+"/.well-known/openid-configuration", "", &config); err != nil {
		return "", fmt.Errorf("failed to discover issuer %s: %w", issuerURL, err)
	}

	if config.UserInfoEndpoint == "" {
		return "", fmt.Errorf("issuer %s has no userinfo endpoint", issuerURL)
	}

	return config.UserInfoEndpoint, nil
}

// FetchUserInfo returns the claims the userinfo endpoint reports for the access token, which is passed as the
// Authorization header value.
func FetchUserInfo(ctx context.Context, userInfoURL, accessToken string) (map[string]any, error) {
	var userInfo map[string]any
	if err := getJSON(ctx, userInfoURL, accessToken, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func getJSON(ctx context.Context, url, authorization string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
