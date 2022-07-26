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

package errors_test

import (
	"context"
	"fmt"

	"github.com/CloudNativeSDWAN/serego/api/core"
	"github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
)

func ExampleIsIteratorDone() {
	var sr *core.ServiceRegistry
	// Define sr here... (look at the documentation)

	// Only get endpoints on a certain network
	it := sr.Namespace("production").
		Service("payroll").
		Endpoint(core.Any).
		List(
			list.WithCIDR("10.11.12.0/24"),
			list.WithPortIn(443, 6443),
		)

	for {
		endp, endpOp, err := it.Next(context.Background())
		if err != nil {
			if errors.IsIteratorDone(err) {
				fmt.Println("no more elements, goodbye!")
			} else {
				fmt.Println("could not get the next element:", err)
			}

			return
		}

		// Do something with the endpoint
		_, _ = endp, endpOp
	}
}

func ExampleIsNotFound() {
	var sr *core.ServiceRegistry
	// Define sr here... (look at the documentation)

	_, err := sr.Namespace("production").Get(context.TODO())
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("namespace 'production' does not exist!")
		} else {
			fmt.Println("could not get namespace", err)
		}
	} else {
		fmt.Println("namespace 'production' does exist!")
	}
}

func ExampleIsPermissionsError() {
	var sr *core.ServiceRegistry
	// Define sr here... (look at the documentation)

	_, err := sr.Namespace("production").Get(context.TODO())
	if err != nil {
		switch {
		case errors.IsNotFound(err):
			fmt.Println("namespace does not exist, do you want to create it?")
		case errors.IsPermissionsError(err):
			fmt.Println("permission denied:", err)
			return
		default:
			// some other error happened which is dependent of the underlying service registry
			fmt.Println("error while getting namespace:", err)
			return
		}
	} else {
		fmt.Println("namespace 'production' does exist!")
	}
}

func ExampleIsAlreadyExists() {
	var sr *core.ServiceRegistry

	err := sr.Namespace("production").Register(context.Background(), register.WithCreateMode())
	if err != nil {
		switch {
		case errors.IsAlreadyExists(err):
			fmt.Println("namespace already exists")
		case errors.IsPermissionsError(err):
			fmt.Println("permission denied:", err)
		default:
			// some other error happened which is dependent of the underlying service registry
			fmt.Println("error while getting namespace:", err)
		}

		return
	}

	fmt.Println("namespace 'production' was created exist!")
}
