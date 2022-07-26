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

package deregister

import (
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
)

// Options to fine tune the behavior of the Deregister operation.
type Options struct {
	// FailNotExists instructs the Deregister operation to return an error if
	// the object does not exist.
	FailNotExists bool
}

type Option func(*Options) error

// WithFailIfNotExists instructs the Deregister operation to return an error if
// the object does not exist. If you try to deregister a resource that does not
// exist and this option is not provided, then the Deregister operation will
// *not* return any errors.
//
// Example:
// 	err := sr.Namespace("hr").Delete(ctx, deregister.WithFailIfNotExists())
func WithFailIfNotExists() Option {
	return func(do *Options) error {
		if do == nil {
			return srerr.NoOptionsProvided
		}

		do.FailNotExists = true
		return nil
	}
}
