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

package core_test

import (
	"context"
	"fmt"

	"github.com/CloudNativeSDWAN/serego/api/core"
	"github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/deregister"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
)

func ExampleServiceOperation_Endpoint() {
	// Start an operation on the main endpoint of service payroll
	// inside the hr namespace.
	endpOp := servReg.Namespace("hr").Service("payroll").Endpoint("main")

	// Perform an operation, for example Get:
	endp, err := endpOp.Get(context.TODO())
	if err != nil {
		// Check the error here. Look at the other examples.
		return
	}

	// Do something with the endpoint.
	fmt.Printf("endpoint has metadata %+v\n", endp.Metadata)
}

func ExampleEndpointOperation_Get() {
	endp, err := servReg.Namespace("hr").Service("payroll").Endpoint("main").
		Get(context.TODO())
	if err != nil {
		// Here errors is the project's errors package, not the one
		// included in golang.
		if errors.IsNotFound(err) {
			fmt.Println("endpoint does not exist")
		} else {
			// This is the same error as returned by the service registry.
			fmt.Println("error while getting the endpoint:", err)
		}

		return
	}

	fmt.Printf("endpoint %s has can be reached at %s:%d", endp.Name, endp.Address, endp.Port)
}

func ExampleEndpointOperation_Register() {
	// This will create the endpoint if it does not exist, or update it
	// otherwise.

	err := servReg.Namespace("hr").Service("payroll").Endpoint("payroll-TCP").
		Register(context.TODO(),
			register.WithAddress("10.10.10.2"),
			register.WithPort(8080),
			register.WithMetadata(map[string]string{
				"env":            "production",
				"protocol":       "TCP",
				"authentication": "token",
			}),
		)
	if err != nil {
		fmt.Println("could not register endpoint:", err)
		return
	}
}

func ExampleEndpointOperation_Register_updateMode() {
	// Update the port of the service.
	// WithKV is just a shortcut for WithMetadataKeyValue.

	err := servReg.Namespace("hr").Service("payroll").Endpoint("payroll-TCP").
		Register(context.TODO(),
			register.WithPort(443),
			register.WithKV("authentication", "jwt"),
		)
	if err != nil {
		fmt.Println("could not register endpoint:", err)
		return
	}
}

func ExampleEndpointOperation_Deregister() {
	// Deregister does not return any errors in case the endpoint
	// does not exist. WithFailIfNotExists overrides this.

	err := servReg.Namespace("hr").Service("payroll").Endpoint("payroll-TCP").
		Deregister(context.Background(),
			deregister.WithFailIfNotExists(),
		)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("endpoint does not exist")
		} else {
			fmt.Println("could not deregister endpoint:", err)
		}

		return
	}

	fmt.Println("endpoint deregistered successfully")
}

func ExampleEndpointOperation_List() {
	// Get all endpoints of service "payroll" inside namespace "hr" and for
	// each of them update their metadata.

	endpIterator := servReg.Namespace("hr").Service("payroll").
		Endpoint(core.Any).List()

	for {
		endp, endpOp, err := endpIterator.Next(context.Background())
		if err != nil {
			if !errors.IsIteratorDone(err) {
				fmt.Println("could not get next endpoint:", err)
			}

			break
		}

		fmt.Printf("updating metadata for endpoint %s...\n", endp.Name)

		endpOp.Register(context.Background(),
			register.WithKV("protocol", "UDP"),
			// WithUpdateMode is not necessary here, as register will figure
			// this out on its own, but we include anyways to show you a usage
			// example.
			register.WithUpdateMode(),
		)
	}
}

func ExampleEndpointOperation_List_withFilters() {
	// Get only specific endpoints.

	serviceName := "payroll"
	endpIterator := servReg.Namespace("hr").Service(serviceName).
		Endpoint(core.Any).List(
		list.WithCIDR("10.10.10.0/24"),
		list.WithPortIn(8080, 443),
	)

	fmt.Println("service", serviceName, "can be reached through these endpoints:")
	for {
		endp, _, err := endpIterator.Next(context.Background())
		if err != nil {
			if !errors.IsIteratorDone(err) {
				fmt.Println("could not get next endpoint:", err)
			}

			break
		}

		fmt.Printf("- %s (%s:%d)\n", endp.Name, endp.Address, endp.Port)
	}
}
