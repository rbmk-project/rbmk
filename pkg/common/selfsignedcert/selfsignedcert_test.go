// SPDX-License-Identifier: GPL-3.0-or-later

package selfsignedcert_test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/bassosimone/runtimex"
	"github.com/rbmk-project/rbmk/pkg/common/selfsignedcert"
)

func TestSelfSignedCert(t *testing.T) {
	// 1. generate the certificate and private key
	cert := selfsignedcert.New(selfsignedcert.NewConfigExampleCom())
	cert.WriteFiles("testdata")

	// 2. create a suitable TLS listener
	serverConfig := &tls.Config{Certificates: []tls.Certificate{
		runtimex.PanicOnError1(tls.X509KeyPair(cert.CertPEM, cert.KeyPEM)),
	}}
	listener := runtimex.PanicOnError1(tls.Listen("tcp", "127.0.0.1:0", serverConfig))
	defer listener.Close()

	// 3. create a listening HTTP server using the testdata files
	expectByes := []byte("Bonsoir, Elliot!\n")
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(expectByes)
		}),
	}
	go srv.Serve(listener)

	// 4. create a suitable HTTP client
	pool := x509.NewCertPool()
	runtimex.Assert(pool.AppendCertsFromPEM(cert.CertPEM))
	clientConfig := &tls.Config{RootCAs: pool}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientConfig,
		},
	}

	// 5. perform an HTTP round trip
	URL := &url.URL{Scheme: "https", Host: listener.Addr().String(), Path: "/"}
	resp, err := client.Get(URL.String())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// 6. make sure the response is correct
	if resp.StatusCode != http.StatusOK {
		t.Fatal("expected 200, got", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expectByes, body) {
		t.Fatal("expected", expectByes, ", got", body)
	}
}
