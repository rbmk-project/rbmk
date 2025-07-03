// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"

	"github.com/rbmk-project/common/mocks"
	"github.com/stretchr/testify/assert"
)

// tlsEngineMock is a mocked [TLSEngine] used for testing.
type tlsEngineMock struct{}

var _ TLSEngine = &tlsEngineMock{}

// Name implements [TLSEngine].
func (t *tlsEngineMock) Name() string {
	return "mock"
}

// NewClientConn implements [TLSEngine].
func (t *tlsEngineMock) NewClientConn(conn net.Conn, config *tls.Config) TLSConn {
	return &mocks.TLSConn{}
}

// Parrot implements [TLSEngine].
func (t *tlsEngineMock) Parrot() string {
	return "parrot"
}

func TestNetwork_newTLSEngine(t *testing.T) {
	t.Run("when we have a custom TLSEngine", func(t *testing.T) {
		// setup with mocked engine
		nx := &Network{
			TLSEngine: &tlsEngineMock{},
		}

		// obtain the engine and check on its type
		engine := nx.newTLSEngine()
		_, ok := engine.(*tlsEngineMock)
		assert.True(t, ok)

		// verify that the name is correct
		assert.Equal(t, engine.Name(), "mock")

		// verify that the conn is of the correct type
		conn := engine.NewClientConn(&mocks.Conn{}, &tls.Config{})
		_, ok = conn.(*mocks.TLSConn)
		assert.True(t, ok)

		// verify that the engine parrot is correct
		assert.Equal(t, engine.Parrot(), "parrot")
	})

	t.Run("when we have a custom NewTLSClientConn", func(t *testing.T) {
		// setup with mocked func
		nx := &Network{
			NewTLSClientConn: func(conn net.Conn, config *tls.Config) TLSConn {
				return &mocks.TLSConn{
					MockHandshakeContext: func(ctx context.Context) error {
						return errors.New("mocked error")
					},
				}
			},
		}

		// obtain the engine - we cannot check the type
		// of the engine since it's a function
		engine := nx.newTLSEngine()

		// verify that the name is correct
		name := engine.Name()
		assert.Equal(t, name, "unknown")

		// verify that the conn handshake returns an error
		conn := engine.NewClientConn(&mocks.Conn{}, &tls.Config{})
		err := conn.HandshakeContext(context.Background())
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "mocked error")

		// verify that the engine parrot is correct
		assert.Equal(t, engine.Parrot(), "unknown")
	})

	t.Run("when we have both TLSEngine and NewTLSClientConn", func(t *testing.T) {
		// It suffices to ensure that the custom TLSEngine is used
		// since we already extensively test it above
		nx := &Network{
			TLSEngine: &tlsEngineMock{},
			NewTLSClientConn: func(conn net.Conn, config *tls.Config) TLSConn {
				panic("unimplemented")
			},
		}
		engine := nx.newTLSEngine()
		_, ok := engine.(*tlsEngineMock)
		assert.True(t, ok)
	})

	t.Run("when nothing has been configured", func(t *testing.T) {
		// setup an empty network
		nx := &Network{}

		// obtain the engine and check on its type
		engine := nx.newTLSEngine()
		_, ok := engine.(*TLSEngineStdlib)
		assert.True(t, ok)

		// verify that the name is correct
		assert.Equal(t, engine.Name(), "stdlib")

		// verify that the conn is of the correct type
		conn := engine.NewClientConn(&mocks.Conn{}, &tls.Config{})
		_, ok = conn.(*tls.Conn)

		// verify that the engine parrot is correct
		assert.Equal(t, engine.Parrot(), "")
	})
}
