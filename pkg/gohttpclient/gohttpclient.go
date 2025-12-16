package gohttpclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// WaitForHttpServer attempts to establish a http TCP connection to listenAddress
// in a given amount of time. It returns upon a successful connection;
// otherwise exits with an error.
func WaitForHttpServer(url string, waitDuration time.Duration, numRetries int, l *slog.Logger) {
	l.Debug("WaitForHttpServer waiting for server", "url", url, "waitSeconds", waitDuration.Seconds(), "numRetries", numRetries)
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	for i := 0; i < numRetries; i++ {
		resp, err := httpClient.Get(url)

		if err != nil {
			if i > 0 {
				l.Info("WaitForHttpServer httpClient.Get retry", "url", url, "retry", i, "error", err)
			}
			time.Sleep(waitDuration)
			continue
		}
		// All seems is good
		l.Info("Server responded OK", "retries", i, "statusCode", resp.StatusCode)
		return
	}
	l.Error("Server not ready after retries, exiting", "url", url, "attempts", numRetries)
	panic(fmt.Sprintf("Server %s not ready after %d attempts", url, numRetries))
}

func GetJsonFromUrlWithBearerAuth(url string, token string, caCert []byte, allowInsecure bool, readTimeout time.Duration, l *slog.Logger) (string, error) {
	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + token

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		l.Error("Error on http.NewRequest", "error", err)
		return "", err
	}

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: allowInsecure,
		},
	}
	// Send req using http Client
	client := &http.Client{
		Transport: tr,
		Timeout:   readTimeout,
	}
	resp, err := client.Do(req)

	if err != nil {
		l.Error("GetJsonFromUrlWithBearerAuth error on response", "error", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Error("GetJsonFromUrlWithBearerAuth error on Body.Close", "error", err)
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		l.Error("Error on response StatusCode is not OK", "statusCode", resp.StatusCode)
		return "", errors.New(fmt.Sprintf("Error on response StatusCode:%d\n", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("GetJsonFromUrlWithBearerAuth error reading response bytes", "error", err)
		return "", err
	}
	return string(body), nil
}
