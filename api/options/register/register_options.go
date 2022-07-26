// Copyright (c) 2022 Cisco Systems, Inc. and its affiliates
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package register

import (
	"math"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"k8s.io/apimachinery/pkg/util/validation"
)

// RegisterMode defines the way Register should behave. Look at the examples
// and the accepted values.
type RegisterMode uint

const (
	// CreateOrUpdateRegisterMode instructs List to create the object if it
	// doesn't exist or update it otherwise.
	CreateOrUpdateMode RegisterMode = iota
	// CreateRegisterMode instructs List to create the object only, and return
	// an error if the object already exists.
	CreateMode
	// UpdateRegisterMode instructs List to update the object only, and return
	// an error if the object does not exist.
	UpdateMode

	minPortNumber int32 = 1
	maxPortNumber int32 = math.MaxUint16
)

type Options struct {
	// RegisterMode defines the behavior of the Register function.
	// Accepted values are:
	// - CreateOrUpdateRegisterMode (the default)
	// - CreateRegisterMode, to only create the object
	// - UpdateRegisterMode, to only update the object
	RegisterMode
	// Metadata to register with the object.
	Metadata map[string]string
	// ReplaceMetadata instructs the Register function to replace every single
	// metadata already existing with ones included in Metadata field,
	// including deleting all keys not present there.
	//
	// Therefore, this only has effect if the object already exists.
	ReplaceMetadata bool
	// Address to register. If nil, no address will be registered and the one
	// already existing will be retained if the object already exists, or an
	// empty string will be registered otherwise.
	Address *string
	// Port to register. If nil, no port will be registered and the one
	// already existing will be retained if the object already exists, or 0
	// will be registered otherwise.
	Port *int32
	// GenerateName instructs Register to generate a name for the endpoint,
	// starting from its parent's service name. This option is only considered
	// when the endpoint operation is defined with no name, otherwise it is
	// ignored.
	GenerateName bool
}

type Option func(*Options) error

// WithUpdateMode instructs Register to only update the resource and return an
// error if the resource does not exist.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Register(
//		register.WithKV("commit", "adf6h45bc"),
// 		register.WithUpdateMode())
func WithUpdateMode() Option {
	return func(ro *Options) error {
		if ro == nil {
			return srerr.NoOptionsProvided
		}

		ro.RegisterMode = UpdateMode
		return nil
	}
}

// WithCreateMode instructs Register to only create the resource and return an
// error if the resource already exists.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Register(
//		register.WithKV("commit", "adf6h45bc"),
// 		register.WithCreateMode())
func WithCreateMode() Option {
	return func(ro *Options) error {
		if ro == nil {
			return srerr.NoOptionsProvided
		}

		ro.RegisterMode = CreateMode
		return nil
	}
}

// WithMetadata registers the provided metadata to the resource.
//
// Note that if the resource already exists, each key-value you provide here
// will be added or replaced to the ones already registered, and all the others
// will be kept. If you want to replace *all* existing metadata with the ones
// you provide here, then you will have to use WithReplaceMetadata along with
// this.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Register(
//		register.WithKV("commit", "adf6h45bc"),
// 		register.WithCreateMode())
func WithMetadata(metadata map[string]string) Option {
	return func(ro *Options) error {
		if ro == nil {
			return srerr.NoOptionsProvided
		}

		if ro.Metadata == nil {
			ro.Metadata = map[string]string{}
		}

		// First check for empty keys
		for k := range metadata {
			if k == "" {
				return srerr.EmptyMetadataKey
			}

			// TODO: check for long values, as well.
		}

		// Then push the kvs
		for k, v := range metadata {
			ro.Metadata[k] = v
		}

		for k, v := range metadata {
			ro.Metadata[k] = v
		}

		return nil
	}
}

// Shortcut for WithMetadata when providing just one key value.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Register(
//		register.WithMetadataKeyValue("commit", "adf6h45bc"),
// 		register.WithMetadataKeyValue("maintainer", "john.smith"))
//
// It is the same as:
// 	sd.Namespace("hr").Service("payroll").Register(
//		register.WithMetadata(map[string]string{
//			"commit": "adf6h45bc",
//			"maintainer": "john.smith",
// 		})
func WithMetadataKeyValue(key, value string) Option {
	return WithMetadata(map[string]string{key: value})
}

// Shortcut for WithMetadataKeyValue.
func WithKV(key, value string) Option {
	return WithMetadataKeyValue(key, value)
}

// WithReplaceMetadata instructs the register operation to remove all existing
// metadata and replace them with the provided ones.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Register(
//		register.WithMetadata(map[string]string{
//			"commit": "adf6h45bc",
//			"maintainer": "john.smith",
// 		},
//		register.WithReplaceMetadata())
func WithReplaceMetadata() Option {
	return func(ro *Options) error {
		if ro == nil {
			return srerr.NoOptionsProvided
		}

		ro.ReplaceMetadata = true
		return nil
	}
}

// WithAddress registers the provided IP address to the endpoint, and is thus
// ignored when registering a namespace or a service.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Endpoint("payroll-internal").
// 		Register(
// 			register.WithAddress("10.10.10.22"),
//			register.WithPort(9876))
func WithAddress(address string) Option {
	return func(ro *Options) error {
		if ro == nil {
			return srerr.NoOptionsProvided
		}

		if address != "" && len(validation.IsValidIP(address)) > 0 {
			return srerr.InvalidAddress
		}

		ro.Address = &address
		return nil
	}
}

// WithPort registers the provided port to the endpoint, and is thus
// ignored when registering a namespace or a service.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Endpoint("payroll-internal").
// 		Register(
// 			register.WithAddress("10.10.10.22"),
//			register.WithPort(9876))
func WithPort(port int32) Option {
	return func(ro *Options) error {
		if ro == nil {
			return srerr.NoOptionsProvided
		}

		if port != 0 && (port < minPortNumber || port > maxPortNumber) {
			return srerr.InvalidPort
		}

		ro.Port = &port
		return nil
	}
}

// WithGenerateName instructs the endpoint operation to generate a name for
// this endpoint before creating it, and thus it is ignored when registering a
// namespace or a service and will return an error if WithUpdateMode is also
// provided.
//
// This option is useful when you don't really care about its name or when
// having multiple instances of a service, and thus multiple endpoints that
// have just the same purpose.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Endpoint("").
// 		Register(
// 			register.WithAddress("10.10.10.22"),
//			register.WithPort(9876),
// 			register.WithGenerateName())
//
// The resulted endpoint will have a random name based on its parent service's
// name. In this example, its name will resemble something like:
// 	payroll-ab78ss02ff
func WithGenerateName() Option {
	return func(ro *Options) error {
		if ro == nil {
			return srerr.NoOptionsProvided
		}

		ro.GenerateName = true
		return nil
	}
}
