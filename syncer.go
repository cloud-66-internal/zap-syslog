// Copyright (c) 2017 Timon Wong
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zapsyslog

import (
	"crypto/tls"
	"net"
	"time"

	"go.uber.org/zap/zapcore"
)

var (
	_ zapcore.WriteSyncer = &ConnSyncer{}
)

// ConnSyncer describes connection sink for syslog.
type ConnSyncer struct {
	network   string
	raddr     string
	conn      net.Conn
	tlsConfig *tls.Config
	timeout   *time.Duration
}

// NewConnSyncer returns a new conn sink for syslog. Pass nil as tlsConfig to disable TLS.
func NewConnSyncer(network, raddr string, tlsConfig *tls.Config, timeout *time.Duration) (*ConnSyncer, error) {
	s := &ConnSyncer{
		network:   network,
		raddr:     raddr,
		tlsConfig: tlsConfig,
		timeout:   timeout,
	}

	err := s.connect()
	if err != nil {
		return nil, err
	}

	return s, nil
}

// connect makes a connection to the syslog server.
func (s *ConnSyncer) connect() error {
	if s.conn != nil {
		// ignore err from close, it makes sense to continue anyway
		s.conn.Close()
		s.conn = nil
	}

	var c net.Conn
	var err error
	if s.tlsConfig != nil {
		var dialer *tls.Dialer
		if s.timeout != nil {
			dialer = &tls.Dialer{
				NetDialer: &net.Dialer{
					Timeout: *s.timeout,
				},
			}
		} else {
			dialer = &tls.Dialer{}
		}
		c, err = dialer.Dial(s.network, s.raddr)
	} else {
		var dialer *net.Dialer
		if s.timeout != nil {
			dialer = &net.Dialer{
				Timeout: *s.timeout,
			}
		} else {
			dialer = &net.Dialer{}
		}
		c, err = dialer.Dial(s.network, s.raddr)
	}
	if err != nil {
		return err
	}

	s.conn = c
	return nil
}

// Write writes to syslog with retry.
func (s *ConnSyncer) Write(p []byte) (n int, err error) {
	if s.conn != nil {
		if n, err := s.conn.Write(p); err == nil {
			return n, err
		}
	}
	if err := s.connect(); err != nil {
		return 0, err
	}

	return s.conn.Write(p)
}

// Sync implements zapcore.WriteSyncer interface.
func (s *ConnSyncer) Sync() error {
	return nil
}
