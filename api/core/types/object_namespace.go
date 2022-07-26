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

// Namespace represents a namespace in the service registry, which you can
// think of as a group that contains different services/applications.
type Namespace struct {
	// Name of the namespace.
	Name string `json:"name" yaml:"name"`
	// Metadata is a map of key-value pairs that add more context or
	// information about this namespace.
	//
	// Check out the main documentation for examples.
	Metadata map[string]string `json:"metadata" yaml:"metadata"`
	// OriginalObject is a pointer to the namespace object as it is stored on
	// the service registry and is provided in case you need data or
	// information that is specific or unique to that service registry and is
	// not covered by this project.
	//
	// You will need to cast this object appropriately: please check Get
	// operations examples for a way to do this.
	OriginalObject interface{} `json:"-" yaml:"-"`
}

// DeepEqualTo compares the namespace with the one provided as argument and
// returns true if they are equal. Note that this will not compare pointers
// but only the values of the fields inside the two namespaces.
//
// Two namespaces are considered equal if all the following conditions apply
// at the same time:
//
// 	- they have the same name
// 	- they have the same combination of keys and values in their metadata,
// 	  including the number of keys but excluding the order.
//
// Note that this will *not* compare the OriginalObject field and, therefore,
// you will have to do that on your own.
func (n *Namespace) DeepEqualTo(namespace *Namespace) bool {
	return n.Name == namespace.Name &&
		reflect.DeepEqual(n.Metadata, namespace.Metadata)
}

// Clone returns a pointer to a *new* namespace object that is equivalent to
// this namespace, which is used as source.
// To learn what equivalent means, read DeepEqualTo.
//
// Note that the OriginalObject will *not* be copied, as the newly cloned
// object will just have a reference to the source OriginalObject.
func (n *Namespace) Clone() *Namespace {
	return &Namespace{
		Name:           n.Name,
		Metadata:       deepCopyMap(n.Metadata),
		OriginalObject: n.OriginalObject,
	}
}
