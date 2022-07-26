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
	"time"

	"github.com/CloudNativeSDWAN/serego/api/core"
	"github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/deregister"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
)

func ExampleNamespaceOperation_Service() {
	// Start an operation on a service called payroll inside the hr namespace.
	servOp := servReg.Namespace("hr").Service("payroll")

	// Perform an operation, for example Get:
	serv, err := servOp.Get(context.TODO())
	if err != nil {
		// Here errors is the project's errors package, not the one
		// included in golang.
		if errors.IsNotFound(err) {
			fmt.Println("service does not exist")
		} else {
			// This is the same error as returned by the service registry.
			fmt.Println("error while getting the service:", err)
		}
		return
	}

	// Do something with the service.
	fmt.Printf("service has metadata %+v\n", serv.Metadata)
}

func ExampleServiceOperation_Get() {
	serv, err := servReg.Namespace("hr").Service("payroll").Get(context.TODO())

	if err != nil {
		// Here errors is the project's errors package, not the one
		// included in golang.
		if errors.IsNotFound(err) {
			fmt.Println("service does not exist")
		} else {
			// This is the same error as returned by the service registry.
			fmt.Println("error while getting the service:", err)
		}

		return
	}

	fmt.Println("service has metadata", serv.Metadata)
}

func ExampleServiceOperation_Register() {
	// This will create the service if it does not exist, or update it
	// otherwise.
	err := servReg.Namespace("hr").Service("payroll").Register(
		context.TODO(),
		register.WithMetadata(map[string]string{
			"env":             "production",
			"version":         "v1.2.2",
			"maintainer":      "team22@it.company.com",
			"traffic-profile": "standard",
		}),
	)
	if err != nil {
		fmt.Println("could not register namespace:", err)
		return
	}
}

func ExampleServiceOperation_Register_updateMode() {
	// Update the version of the service.
	// WithKV is just a shortcut for WithMetadataKeyValue.
	err := servReg.Namespace("hr").Service("payroll").Register(
		context.TODO(),
		// Force an update.
		register.WithUpdateMode(),
		register.WithKV("building", "bulding 12"),
	)
	if err != nil {
		fmt.Println("could not register service:", err)
		return
	}
}

func ExampleServiceOperation_Deregister() {
	// Deregister does not return any errors in case the service
	// does not exist. WithFailIfNotExists overrides this.
	err := servReg.Namespace("hr").Service("payroll").Deregister(
		context.TODO(),
		deregister.WithFailIfNotExists(),
	)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("service does not exist")
		} else {
			fmt.Println("could not deregister service:", err)
		}

		return
	}

	fmt.Println("service deregistered successfully")
}

func ExampleServiceOperation_List() {
	// Get all services inside namespace "hr" and for each of them update
	// their metadata.

	servIterator := servReg.Namespace("hr").Service(core.Any).List()

	for {
		serv, servOp, err := servIterator.Next(context.Background())
		if err != nil {
			if !errors.IsIteratorDone(err) {
				fmt.Println("could not get next service:", err)
			}

			break
		}

		fmt.Printf("updating metadata for service %s...\n", serv.Name)

		servOp.Register(
			context.Background(),
			register.WithKV("last-update", time.Now().String()),
			// WithUpdateMode is not necessary here, as register will figure
			// this out on its own, but we include anyways to show you a usage
			// example.
			register.WithUpdateMode(),
		)
	}
}

func ExampleServiceOperation_List_withFilters() {
	// Get only specific services, and that are in a production stage.
	servIterator := servReg.Namespace("hr").Service(core.Any).List(
		list.WithNameIn("payroll", "profiles", "support"),
		list.WithKV("stage", "prod"),
	)

	for {
		service, _, err := servIterator.Next(context.Background())
		if err != nil {
			if !errors.IsIteratorDone(err) {
				fmt.Println("could not get next service:", err)
			}

			break
		}

		fmt.Println("service", service.Name, "is in production and has metadata", service.Metadata)
	}
}
