// SPDX-License-Identifier: GPL-3.0-or-later

package httpconntrace_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/rbmk-project/rbmk/pkg/common/httpconntrace"
)

func Example() {
	// Create a test server that just echoes back
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	defer ts.Close()

	// Create and send request
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		fmt.Printf("failed to create request: %s\n", err)
		return
	}

	// Use Do instead of client.Do to get connection endpoints
	resp, endpoints, err := httpconntrace.Do(http.DefaultClient, req)
	if err != nil {
		fmt.Printf("request failed: %s\n", err)
		return
	}
	defer resp.Body.Close()

	// Print the endpoints we collected
	fmt.Printf("Local: %v\n", endpoints.LocalAddr.IsValid())
	fmt.Printf("Remote: %v\n", endpoints.RemoteAddr.IsValid())

	// Output:
	// Local: true
	// Remote: true
}
