package gohttpclient

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// WaitForHttpServer attempts to establish a http TCP connection to listenAddress
// in a given amount of time. It returns upon a successful connection;
// otherwise exits with an error.
func WaitForHttpServer(url string, waitDuration time.Duration, numRetries int) {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	for i := 0; i < numRetries; i++ {
		resp, err := httpClient.Get(url)

		if err != nil {
			if i > 0 {
				fmt.Printf("\nWaitForHttpServer: httpClient.Get(%s) retry:[%d], %v\n", url, i, err)
			}
			time.Sleep(waitDuration)
			continue
		}
		// All seems is good
		fmt.Printf("OK: Server responded after %d retries, with status code %d ", i, resp.StatusCode)
		return
	}
	log.Fatalf("Server %s not ready up after %d attempts", url, numRetries)
}
