// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v7"

	"github.com/goharbor/harbor/src/lib/errors"
)

var (
	// ErrRetryTimeout timeout error for retrying
	ErrRetryTimeout = errors.New("retry timeout")
)

type abort struct {
	cause error
}

func (a *abort) Error() string {
	if a.cause != nil {
		return fmt.Sprintf("retry abort, error: %v", a.cause)
	}

	return "retry abort"
}

// Abort wrap err to stop the Retry function
func Abort(err error) error {
	return &abort{cause: err}
}

// Options options for the retry functions
type Options struct {
	InitialInterval time.Duration                        // the initial interval for retring after failure, default 100 milliseconds
	MaxInterval     time.Duration                        // the max interval for retring after failure, default 1 second
	Timeout         time.Duration                        // the total time before returning if something is wrong, default 1 minute
	Callback        func(err error, sleep time.Duration) // the callback function for Retry when the f called failed
	Backoff         bool
}

// Option ...
type Option func(*Options)

// InitialInterval set initial interval
func InitialInterval(initial time.Duration) Option {
	return func(opts *Options) {
		opts.InitialInterval = initial
	}
}

// MaxInterval set max interval
func MaxInterval(maxInterval time.Duration) Option {
	return func(opts *Options) {
		opts.MaxInterval = maxInterval
	}
}

// Timeout set timeout interval
func Timeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

// Callback set callback
func Callback(callback func(err error, sleep time.Duration)) Option {
	return func(opts *Options) {
		opts.Callback = callback
	}
}

// Backoff set backoff
func Backoff(backoff bool) Option {
	return func(opts *Options) {
		opts.Backoff = backoff
	}
}

// Retry retry until f run successfully or timeout
//
// NOTE: This function will use exponential backoff and jitter for retrying, see
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/ for more information
func Retry(f func() error, options ...Option) error {
	opts := &Options{
		Backoff: true,
	}

	for _, o := range options {
		o(opts)
	}

	if opts.InitialInterval <= 0 {
		opts.InitialInterval = time.Millisecond * 100
	}

	if opts.MaxInterval <= 0 {
		opts.MaxInterval = time.Second
	}

	if opts.Timeout <= 0 {
		opts.Timeout = time.Minute
	}

	var b backoff.BackOff = &backoff.ZeroBackOff{}

	if opts.Backoff {
		exp := backoff.NewExponentialBackOff()
		exp.InitialInterval = opts.InitialInterval
		exp.MaxInterval = opts.MaxInterval
		exp.Multiplier = 2
		exp.RandomizationFactor = 0.5
		b = exp
	}

	_, retryErr := backoff.Retry(context.Background(), func() (struct{}, error) {
		err := f()
		if err == nil {
			return struct{}{}, nil
		}

		var ab *abort
		if errors.As(err, &ab) {
			return struct{}{}, backoff.Permanent(ab.cause)
		}

		return struct{}{}, err
	},
		backoff.WithBackOff(b),
		backoff.WithMaxElapsedTime(opts.Timeout),
		backoff.WithNotify(func(err error, sleep time.Duration) {
			if opts.Callback != nil {
				opts.Callback(err, sleep)
			}
		}),
	)

	if retryErr == nil {
		return nil
	}

	if re := backoff.AsRetryError(retryErr); re != nil {
		if errors.Is(re.Cause, backoff.ErrPermanent) {
			return re.LastErr
		}
		return errors.New(ErrRetryTimeout).WithCause(re.LastErr)
	}

	return retryErr
}
