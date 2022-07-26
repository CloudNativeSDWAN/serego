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

package get

import (
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
)

// Options to fine tune the behavior of the Get operation.
type Options struct {
	// ForceRefresh forces Get to bypass cache and retrieve the object from
	// the service registry.
	ForceRefresh bool
}

type Option func(*Options) error

// WithForceRefresh forces Get to bypass cache and retrieve the object from
// the service registry.
//
// NOTE: if you started the API with wrapper.WithNoCache() then this option
// will have no effect as it is the default behavior anyways.
//
// Example:
// 	ns, err := sr.Namespace("hr").Get(ctx, get.WithForceRefresh())
// 	if err != nil {
//		// errors is serego's errors package
//		if errors.IsNotFound(err) {
//			fmt.Println("namespace not found")
//			return
// 		}
//
//		fmt.Println("could not get namespace:", err)
// 		return
// 	}
//
// 	fmt.Printf("namespace has metadata %+v\n", ns.Metadata)
func WithForceRefresh() Option {
	return func(gopts *Options) error {
		if gopts == nil {
			return srerr.NoOptionsProvided
		}

		gopts.ForceRefresh = true
		return nil
	}
}
