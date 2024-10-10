// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stream

import (
	"fmt"
	"time"
)

const (
	// MaxConcurrency is the max number of concurrent go routines (for message handling) for a single stream consumer.
	MaxConcurrency = 64

	// MaxMaxRetries is the max number of retries of a message for a single consumer group.
	MaxMaxRetries = 64

	// MinIdleTimeout is the minimum time that can be configured as idle timeout for a stream consumer.
	MinIdleTimeout = 5 * time.Second
)

// ConsumerOption is used to configure consumers.
type ConsumerOption interface {
	apply(*ConsumerConfig)
}

// consumerOptionFunc allows to have functions implement the ConsumerOption interface.
type consumerOptionFunc func(*ConsumerConfig)

// Apply calls f(config).
func (f consumerOptionFunc) apply(config *ConsumerConfig) {
	f(config)
}

// WithConcurrency sets up the concurrency of the stream consumer.
func WithConcurrency(concurrency int) ConsumerOption {
	if concurrency < 1 || concurrency > MaxConcurrency {
		// misconfiguration - panic to keep options clean
		panic(fmt.Sprintf("provided concurrency %d is invalid - has to be between 1 and %d",
			concurrency, MaxConcurrency))
	}
	return consumerOptionFunc(func(c *ConsumerConfig) {
		c.Concurrency = concurrency
	})
}

// WithHandlerOptions sets up the default handler options of a stream consumer.
func WithHandlerOptions(opts ...HandlerOption) ConsumerOption {
	return consumerOptionFunc(func(c *ConsumerConfig) {
		for _, opt := range opts {
			opt.apply(&c.DefaultHandlerConfig)
		}
	})
}

// HandlerOption is used to configure the handler consuming a single stream.
type HandlerOption interface {
	apply(*HandlerConfig)
}

// handlerOptionFunc allows to have functions implement the HandlerOption interface.
type handlerOptionFunc func(*HandlerConfig)

// Apply calls f(config).
func (f handlerOptionFunc) apply(config *HandlerConfig) {
	f(config)
}

// WithMaxRetries can be used to set the max retry count for a specific handler.
func WithMaxRetries(maxRetries int) HandlerOption {
	if maxRetries < 0 || maxRetries > MaxMaxRetries {
		// misconfiguration - panic to keep options clean
		panic(fmt.Sprintf("provided maxRetries %d is invalid - has to be between 0 and %d", maxRetries, MaxMaxRetries))
	}
	return handlerOptionFunc(func(c *HandlerConfig) {
		c.maxRetries = maxRetries
	})
}

// WithIdleTimeout can be used to set the idle timeout for a specific handler.
func WithIdleTimeout(timeout time.Duration) HandlerOption {
	if timeout < MinIdleTimeout {
		// misconfiguration - panic to keep options clean
		panic(fmt.Sprintf("provided timeout %d is invalid - has to be longer than %s", timeout, MinIdleTimeout))
	}
	return handlerOptionFunc(func(c *HandlerConfig) {
		c.idleTimeout = timeout
	})
}
