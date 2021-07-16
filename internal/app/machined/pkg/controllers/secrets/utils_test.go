// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package secrets_test

import (
	"context"
	"testing"
	"time"

	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/stretchr/testify/assert"

	"github.com/talos-systems/talos/internal/app/machined/pkg/controllers/secrets"
)

func TestRateLimitEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inCh := make(chan controller.ReconcileEvent)
	outCh := secrets.RateLimitEvents(ctx, inCh, time.Second)

	inputs := 0
	outputs := 0

	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()

LOOP:
	for {
		select {
		case <-timer.C:
			break LOOP
		case <-outCh:
			outputs++
		case inCh <- controller.ReconcileEvent{}:
			inputs++
		}
	}

	assert.Less(t, outputs, 5)
	assert.Greater(t, inputs, 15)
}
