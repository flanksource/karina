/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const defaultTimeout = 10 * time.Second

// Dialer creates connections using Kubernetes API Server port-forwarding
type Dialer struct {
	proxy          Proxy
	clientset      *kubernetes.Clientset
	proxyTransport http.RoundTripper
	upgrader       spdy.Upgrader
	timeout        time.Duration
}

// NewDialer creates a new dialer for a given API server scope
func NewDialer(p Proxy, clientset *kubernetes.Clientset, config *rest.Config, options ...func(*Dialer) error) (*Dialer, error) {
	if p.Port == 0 {
		return nil, errors.New("port required")
	}

	dialer := &Dialer{
		proxy: p,
	}

	for _, option := range options {
		err := option(dialer)
		if err != nil {
			return nil, err
		}
	}

	if dialer.timeout == 0 {
		dialer.timeout = defaultTimeout
	}

	proxyTransport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}
	dialer.proxyTransport = proxyTransport
	dialer.upgrader = upgrader
	dialer.clientset = clientset
	return dialer, nil
}

// DialContextWithAddr is a GO grpc compliant dialer construct
func (d *Dialer) DialContextWithAddr(ctx context.Context, addr string) (net.Conn, error) {
	return d.DialContext(ctx, scheme, addr)
}

// DialContext creates proxied port-forwarded connections.
// ctx is currently unused, but fulfils the type signature used by GRPC.
func (d *Dialer) DialContext(_ context.Context, network string, addr string) (net.Conn, error) {
	req := d.clientset.CoreV1().RESTClient().
		Post().
		Resource(d.proxy.Kind).
		Namespace(d.proxy.Namespace).
		Name(d.proxy.ResourceName).
		SubResource("portforward")

	dialer := spdy.NewDialer(d.upgrader, &http.Client{Transport: d.proxyTransport}, "POST", req.URL())

	p, _, err := dialer.Dial(portforward.PortForwardProtocolV1Name)
	if err != nil {
		return nil, errors.Wrap(err, "error upgrading connection for: "+req.URL().String())
	}
	headers := http.Header{}
	headers.Set(corev1.StreamType, corev1.StreamTypeError)
	headers.Set(corev1.PortHeader, fmt.Sprintf("%d", d.proxy.Port))
	// We only create a single stream over the connection
	headers.Set(corev1.PortForwardRequestIDHeader, "0")
	errorStream, err := p.CreateStream(headers)
	if err != nil {
		return nil, err
	}

	if err := errorStream.Close(); err != nil {
		return nil, err
	}

	headers.Set(corev1.StreamType, corev1.StreamTypeData)
	dataStream, err := p.CreateStream(headers)
	if err != nil {
		return nil, errors.Wrap(err, "error creating forwarding stream")
	}

	c := NewConn(dataStream)

	return c, nil
}

// DialTimeout sets the timeout
func DialTimeout(duration time.Duration) func(*Dialer) error {
	return func(d *Dialer) error {
		return d.setTimeout(duration)
	}
}

func (d *Dialer) setTimeout(duration time.Duration) error {
	d.timeout = duration
	return nil
}
