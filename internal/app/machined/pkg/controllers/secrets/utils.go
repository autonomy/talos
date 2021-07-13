// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package secrets

import (
	"context"
	"time"

	"github.com/cosi-project/runtime/pkg/controller"
	"golang.org/x/time/rate"
)

// RateLimitEvents to reduce the rate of reconcile events.
//
// RateLimitEvents makes sure that reconcile events are not coming faster than interval.
// Any reconcile events which come during the waiting delay are coalesced with the original events.
//
//nolint:gocyclo
func RateLimitEvents(ctx context.Context, in <-chan controller.ReconcileEvent, interval time.Duration) <-chan controller.ReconcileEvent {
	limiter := rate.NewLimiter(rate.Every(interval), 1)
	ch := make(chan controller.ReconcileEvent)

	go func() {
		for {
			var event controller.ReconcileEvent

			// wait for an actual reconcile event
			select {
			case <-ctx.Done():
				return
			case event = <-in:
			}

			// figure out if the event can be delivered immediately
			reservation := limiter.Reserve()
			delay := reservation.Delay()

			if delay != 0 {
				timer := time.NewTimer(delay)
				defer timer.Stop()

			WAIT:
				for {
					select {
					case <-ctx.Done():
						reservation.Cancel()

						return
					case <-in:
						// coalesce extra events while waiting
					case <-timer.C:
						break WAIT
					}
				}
			}

			// deliver rate-limited coalesced event
			select {
			case <-ctx.Done():
				return
			case ch <- event:
			}
		}
	}()

	return ch
}
