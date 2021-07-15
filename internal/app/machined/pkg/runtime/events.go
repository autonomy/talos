// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package runtime

import (
	"fmt"
	"time"

	"github.com/rs/xid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/talos-systems/talos/pkg/grpc/factory"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
)

// Event is what is sent on the wire.
type Event struct {
	TypeURL string
	ID      xid.ID
	Payload proto.Message
}

// WatchFunc defines the watcher callback function.
type WatchFunc func(<-chan Event)

// WatchOptions defines options for the watch call.
//
// Only one of TailEvents, TailID or TailDuration should be non-zero.
type WatchOptions struct {
	// Return that many past events.
	//
	// If TailEvents is negative, return all the events available.
	TailEvents int
	// Start at ID > specified.
	TailID xid.ID
	// Start at timestamp Now() - TailDuration.
	TailDuration time.Duration
}

// WatchOptionFunc defines the options for the watcher.
type WatchOptionFunc func(opts *WatchOptions) error

// WithTailEvents sets up Watcher to return specified number of past events.
//
// If number is negative, all the available past events are returned.
func WithTailEvents(number int) WatchOptionFunc {
	return func(opts *WatchOptions) error {
		if !opts.TailID.IsNil() || opts.TailDuration != 0 {
			return fmt.Errorf("WithTailEvents can't be specified at the same time with WithTailID or WithTailDuration")
		}

		opts.TailEvents = number

		return nil
	}
}

// WithTailID sets up Watcher to return events with ID > TailID.
func WithTailID(id xid.ID) WatchOptionFunc {
	return func(opts *WatchOptions) error {
		if opts.TailEvents != 0 || opts.TailDuration != 0 {
			return fmt.Errorf("WithTailID can't be specified at the same time with WithTailEvents or WithTailDuration")
		}

		opts.TailID = id

		return nil
	}
}

// WithTailDuration sets up Watcher to return events with timestamp >= (now - tailDuration).
func WithTailDuration(dur time.Duration) WatchOptionFunc {
	return func(opts *WatchOptions) error {
		if opts.TailEvents != 0 || !opts.TailID.IsNil() {
			return fmt.Errorf("WithTailDuration can't be specified at the same time with WithTailEvents or WithTailID")
		}

		opts.TailDuration = dur

		return nil
	}
}

// Watcher defines a runtime event watcher.
type Watcher interface {
	Watch(WatchFunc, ...WatchOptionFunc) error
}

// Publisher defines a runtime event publisher.
type Publisher interface {
	Publish(proto.Message)
}

// EventStream defines the runtime event stream.
type EventStream interface {
	Watcher
	Publisher
}

// ToMachineEvent serializes Event as proto message machine.Event.
func (event *Event) ToMachineEvent() (*machine.Event, error) {
	value, err := factory.VTProtoCodec{}.Marshal(event.Payload)
	if err != nil {
		return nil, err
	}

	return &machine.Event{
		Data: &anypb.Any{
			TypeUrl: event.TypeURL,
			Value:   value,
		},
		Id: event.ID.String(),
	}, nil
}
