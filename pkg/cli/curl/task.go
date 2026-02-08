// SPDX-License-Identifier: GPL-3.0-or-later

package curl

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/rbmk-project/rbmk/internal/netcore"
	"github.com/rbmk-project/rbmk/internal/testablenet"
	"github.com/rbmk-project/rbmk/pkg/common/closepool"
	"github.com/rbmk-project/rbmk/pkg/common/dialonce"
)

// Task runs the curl task.
type Task struct {
	// LogsWriter is where we write structured logs
	LogsWriter io.Writer

	// MaxTime is the maximum time to wait for the operation to finish.
	MaxTime time.Duration

	// Method is the HTTP method to use
	Method string

	// Output is where we write the response body
	Output io.Writer

	// ResolveMap maps HOST:PORT to IP address
	ResolveMap map[string]string

	// URL is the URL to fetch
	URL string

	// VerboseOutput is where we write the verbose output
	VerboseOutput io.Writer
}

// Run executes the curl task
func (task *Task) Run(ctx context.Context) error {
	// Setup the overall operation timeout using the context
	ctx, cancel := context.WithTimeout(ctx, task.MaxTime)
	defer cancel()

	// Set up the JSON logger for writing the measurements
	logger := slog.New(slog.NewJSONHandler(task.LogsWriter, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create a pool containing closers
	pool := &closepool.Pool{}
	defer pool.Close()

	// Create netcore network instance making sure we dial the
	// endpoint at most once, thus avoiding infinite dialing loops such
	// as the one occurring with https://avdox.globalvoices.org/.
	netx := netcore.NewNetwork()
	netx.DialContextFunc = dialonce.Wrap(testablenet.DialContext.Get())
	netx.Logger = logger

	// Also, honor the `--resolve` flag.
	netx.SplitHostPort = func(endpoint string) (string, string, error) {
		hostname, port, err := net.SplitHostPort(endpoint)
		if err != nil {
			return "", "", err
		}
		if match, ok := task.ResolveMap[endpoint]; ok {
			hostname = match
		}
		return hostname, port, nil
	}

	// Create the HTTP client to use and make sure we're using
	// an overall operation timeout for the transfer
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Note: [httpDoAndLog] assumes we don't follow redirects. Changing
			// this would break connection tracking and logging.
			//
			// While this may seem technical debt, we'll most likely want to
			// perform requests one at a time, in the future, when we will be
			// following redirects, to observe interim bodies and generate
			// additional structured logs pertaining to the redirects.
			//
			// Also, for measuring, the main use case is that of supplying
			// this command with the address to use via `--resolve`.
			return http.ErrUseLastResponse
		},
		Timeout:   task.MaxTime, // ensure the overall operation is bounded
		Transport: newHTTPLogTransport(netx, pool),
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, task.Method, task.URL, http.NoBody)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	// Print the request, if verbose
	//
	// TODO(bassosimone): when we'll enable redirects with `--location` we
	// won't be able to print the intermediate steps anymore.
	fmt.Fprintf(task.VerboseOutput, "> %s %s HTTP/%d.%d\n",
		req.Method, req.URL.RequestURI(),
		req.ProtoMajor, req.ProtoMinor)
	fmt.Fprintf(task.VerboseOutput, "> Host: %s\n", req.Host)
	task.printHeaders(req.Header, ">")
	fmt.Fprintf(task.VerboseOutput, ">\n")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Print the response, if verbose
	fmt.Fprintf(task.VerboseOutput, "< HTTP/%d.%d %d %s\n",
		resp.ProtoMajor, resp.ProtoMinor,
		resp.StatusCode, resp.Status)
	task.printHeaders(resp.Header, "<")
	fmt.Fprintf(task.VerboseOutput, "<\n")

	// Copy the response body
	//
	// TODO(bassosimone): maybe we should use [iox.CopyContext] here.
	if _, err := io.Copy(task.Output, resp.Body); err != nil {
		return fmt.Errorf("reading or writing response body: %w", err)
	}

	// Explicitly close the connections in the pool
	pool.Close()
	return nil
}

// printHeaders prints HTTP headers with the given prefix
func (task *Task) printHeaders(headers http.Header, prefix string) {
	for name, values := range headers {
		for _, value := range values {
			fmt.Fprintf(task.VerboseOutput, "%s %s: %s\n", prefix, name, value)
		}
	}
}
