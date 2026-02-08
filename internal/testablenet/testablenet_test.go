// SPDX-License-Identifier: GPL-3.0-or-later

package testablenet

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDialContextProvider(t *testing.T) {
	// When the provider has zero value, Get returns a non-nil function
	// that dials using the standard library.
	t.Run("zero value returns standard library dialer", func(t *testing.T) {
		dcp := &DialContextProvider{}
		fx := dcp.Get()
		require.NotNil(t, fx)
	})

	// When we Set a custom function, Get returns a function that
	// delegates to the custom one.
	t.Run("set overrides the dial function", func(t *testing.T) {
		expectedErr := errors.New("custom dialer called")
		dcp := &DialContextProvider{}

		dcp.Set(func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, expectedErr
		})

		fx := dcp.Get()
		require.NotNil(t, fx)

		conn, err := fx(context.Background(), "tcp", "127.0.0.1:80")
		assert.Nil(t, conn)
		assert.ErrorIs(t, err, expectedErr)
	})

	// When we Set back to nil, Get returns the default standard
	// library dialer again.
	t.Run("set nil reverts to default", func(t *testing.T) {
		dcp := &DialContextProvider{}

		dcp.Set(func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, errors.New("should not persist")
		})

		dcp.Set(nil)

		fx := dcp.Get()
		require.NotNil(t, fx)
	})
}

func TestLookupHostProvider(t *testing.T) {
	// When the provider has zero value, Get returns a non-nil function
	// that resolves using the standard library.
	t.Run("zero value returns standard library resolver", func(t *testing.T) {
		lhp := &LookupHostProvider{}
		fx := lhp.Get()
		require.NotNil(t, fx)
	})

	// When we Set a custom function, Get returns a function that
	// delegates to the custom one.
	t.Run("set overrides the lookup function", func(t *testing.T) {
		expectedAddrs := []string{"1.2.3.4"}
		lhp := &LookupHostProvider{}

		lhp.Set(func(ctx context.Context, domain string) ([]string, error) {
			return expectedAddrs, nil
		})

		fx := lhp.Get()
		require.NotNil(t, fx)

		addrs, err := fx(context.Background(), "example.com")
		require.NoError(t, err)
		assert.Equal(t, expectedAddrs, addrs)
	})

	// When we Set a custom function that returns an error, Get returns
	// a function that propagates the error.
	t.Run("set override propagates errors", func(t *testing.T) {
		expectedErr := errors.New("lookup failed")
		lhp := &LookupHostProvider{}

		lhp.Set(func(ctx context.Context, domain string) ([]string, error) {
			return nil, expectedErr
		})

		fx := lhp.Get()
		require.NotNil(t, fx)

		addrs, err := fx(context.Background(), "example.com")
		assert.Nil(t, addrs)
		assert.ErrorIs(t, err, expectedErr)
	})

	// When we Set back to nil, Get returns the default standard
	// library resolver again.
	t.Run("set nil reverts to default", func(t *testing.T) {
		lhp := &LookupHostProvider{}

		lhp.Set(func(ctx context.Context, domain string) ([]string, error) {
			return nil, errors.New("should not persist")
		})

		lhp.Set(nil)

		fx := lhp.Get()
		require.NotNil(t, fx)
	})
}

func TestRootCAsProvider(t *testing.T) {
	// When the provider has zero value, Get returns nil (meaning
	// the system root CAs will be used).
	t.Run("zero value returns nil", func(t *testing.T) {
		rcp := &RootCAsProvider{}
		pool := rcp.Get()
		assert.Nil(t, pool)
	})

	// When we Set a custom pool, Get returns that pool.
	t.Run("set overrides the cert pool", func(t *testing.T) {
		expectedPool := x509.NewCertPool()
		rcp := &RootCAsProvider{}

		rcp.Set(expectedPool)

		pool := rcp.Get()
		assert.Same(t, expectedPool, pool)
	})

	// When we Set back to nil, Get returns nil again.
	t.Run("set nil reverts to default", func(t *testing.T) {
		rcp := &RootCAsProvider{}

		rcp.Set(x509.NewCertPool())
		rcp.Set(nil)

		pool := rcp.Get()
		assert.Nil(t, pool)
	})
}
