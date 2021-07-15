// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package factory

//nolint:gci
import (
	"fmt"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"google.golang.org/grpc/encoding"
	_ "google.golang.org/grpc/encoding/proto" // imported to override default gRPC codec
)

// Importing factory module implicitly overrides default gRPC codec, and as the factory is imported
// in any server code, vtprotobuf marshaling will be enabled for all Talos gRPC servers.

// Using here the original golang/protobuf package so that we can continue serializing messages
// from our dependencies, particularly from etcd and containerd.

// Name is the name registered for the proto compressor.
const Name = "proto"

// VTProtoCodec implements optimized marshaling for vtprotobuf-enabled messages.
type VTProtoCodec struct{}

type vtprotoMessage interface {
	MarshalVT() ([]byte, error)
	UnmarshalVT([]byte) error
}

// Marshal implements encodings.Codec.
func (VTProtoCodec) Marshal(v interface{}) ([]byte, error) {
	if vt, ok := v.(vtprotoMessage); ok {
		return vt.MarshalVT()
	}

	if vv, ok := v.(proto.Message); ok {
		return proto.Marshal(vv)
	}

	return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
}

// Unmarshal implements encodings.Codec.
func (VTProtoCodec) Unmarshal(data []byte, v interface{}) error {
	if vt, ok := v.(vtprotoMessage); ok {
		return vt.UnmarshalVT(data)
	}

	if vv, ok := v.(proto.Message); ok {
		return proto.Unmarshal(data, vv)
	}

	return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
}

// Name implements encodings.Codec.
func (VTProtoCodec) Name() string {
	return Name
}

func init() {
	encoding.RegisterCodec(VTProtoCodec{})
}
