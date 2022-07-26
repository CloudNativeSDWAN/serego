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

package types

import (
	"reflect"
)

// Endpoint represents the combination of address:port where to contact the
// service. This is the actual "place" where you can reach a
// service/application.
type Endpoint struct {
	// Name of this endpoint.
	Name string `json:"name" yaml:"name"`
	// Service is the name of the parent service that this endpoint belongs to.
	Service string `json:"service" yaml:"service"`
	// Namespace is the name of the namespace that this endpoint - and its
	// parent service - belongs to.
	Namespace string `json:"namespace" yaml:"namespace"`
	// Address where to reach the parent service/application.
	// It can be either an empty string, an IPv4 or an IPv6 address.
	// An empty string can be interpreted as an unknown address that will
	// be filled later.
	Address string `json:"address" yaml:"address"`
	// Port where to reach the parent service, and thus can have a value
	// between 0 and 65535.
	// The zero value can be interpreted as an empty or temporary value that
	// can be filled later.
	Port int32 `json:"port" yaml:"port"`
	// Metadata is a map of key-value pairs that add more context or
	// information about this endpoint.
	//
	// Check out the main documentation for examples.
	Metadata map[string]string `json:"metadata" yaml:"metadata"`
	// OriginalObject is a pointer to the endpoint object as it is stored on
	// the service registry and is provided in case you need data or
	// information that is specific or unique to that service registry and is
	// not covered by this project.
	//
	// You will need to cast this object appropriately: please check Get
	// operations examples for a way to do this.
	OriginalObject interface{} `json:"-" yaml:"-"`
}

// DeepEqualTo compares the endpoint with the one provided as argument and
// returns true if they are equal. Note that this will not compare pointers
// but only the values of the fields inside the two endpoints.
//
// Two endpoints are considered equal if all the following conditions apply
// at the same time:
//
// 	- they have the same name
//	- they belong to the same service
// 	- they belong to the same namespace
// 	- they have the same address
// 	- they have the same port
// 	- they have the same combination of keys and values in their metadata,
// 	  including the number of keys but excluding the order.
//
// Note that this will *not* compare the OriginalObject field and, therefore,
// you will have to do that on your own.
func (e *Endpoint) DeepEqualTo(ep *Endpoint) bool {
	if ep == nil {
		return false
	}

	return e.Name == ep.Name &&
		e.Namespace == ep.Namespace &&
		e.Service == ep.Service &&
		e.Address == ep.Address &&
		e.Port == ep.Port &&
		reflect.DeepEqual(e.Metadata, ep.Metadata)
}

// Clone returns a pointer to a *new* endpoint object that is equivalent to
// this endpoint, which is used as source.
// To learn what equivalent means, read DeepEqualTo.
//
// Note that the OriginalObject will *not* be copied, as the newly cloned
// object will just have a reference to the source OriginalObject.
func (e *Endpoint) Clone() *Endpoint {
	return &Endpoint{
		Name:           e.Name,
		Service:        e.Service,
		Namespace:      e.Namespace,
		Address:        e.Address,
		Port:           e.Port,
		Metadata:       deepCopyMap(e.Metadata),
		OriginalObject: e.OriginalObject,
	}
}
