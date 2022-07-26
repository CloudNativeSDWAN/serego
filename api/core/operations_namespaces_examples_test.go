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

	servicedirectory "cloud.google.com/go/servicedirectory/apiv1"
	"github.com/CloudNativeSDWAN/serego/api/core"
	"github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/deregister"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
)

var (
	servReg *core.ServiceRegistry
)

func ExampleServiceRegistry_Namespace() {
	// Here we get the service directory registration client and we define
	// its settings...
	var sdClient *servicedirectory.RegistrationClient

	// Get the registration client: refer to service directory's documentation
	// and http://localhost:6060/pkg/github.com/CloudNativeSDWAN/serego/api/core/#NewServiceRegistryFromServiceDirectory
	_ = sdClient

	// Now we get the wrapper.
	sd, err := core.NewServiceRegistryFromServiceDirectory(sdClient,
		wrapper.WithProjectID("my-project-id"),
		wrapper.WithRegion("us-east1"),
	)
	if err != nil {
		fmt.Println("could not get the wrapper:", err)
		return
	}

	namespace, err := sd.Namespace("my-namespace-name").Get(context.TODO())
	if err != nil {
		// Check the error here. Look at the other examples.
		return
	}

	// Do something with the namespace.
	fmt.Printf("Namespace has metadata %+v\n", namespace.Metadata)
}

func ExampleNamespaceOperation_Get() {
	// get.WithForceRefresh() instructs the Get operation to ignore cache and
	// get the namespace directly from the service registry.
	// Most of the times you won't need this and
	// you can omit that. Here it is included to show the example.

	ns, err := servReg.Namespace("sales").Get(
		context.TODO(),
		get.WithForceRefresh(),
	)
	if err != nil {
		// Here errors is the project's errors package, not the one
		// included in golang.
		if errors.IsNotFound(err) {
			fmt.Println("namespace does not exist")
		} else {
			// This is the same error as returned by the service registry.
			fmt.Println("error while getting the namespace:", err)
		}

		return
	}

	fmt.Println("namespace has metadata", ns.Metadata)
}

func ExampleNamespaceOperation_Register() {
	// This will create the namespace if it does not exist, or update it
	// otherwise.

	err := servReg.Namespace("sales").Register(context.TODO(),
		register.WithMetadata(map[string]string{
			"contact":  "sales@company.com",
			"manager":  "john.smith@sales.company.com",
			"building": "building 24",
		}),
	)
	if err != nil {
		fmt.Println("could not register namespace:", err)
		return
	}
}

func ExampleNamespaceOperation_Register_updateMode() {
	// Update the contact method and building.
	// WithKV is just a shortcut for WithMetadataKeyValue.

	err := servReg.Namespace("sales").Register(context.TODO(),
		// Force an update.
		register.WithUpdateMode(),
		register.WithMetadataKeyValue("contact", "sales.department@company.com"),
		register.WithKV("building", "bulding 12"),
	)
	if err != nil {
		fmt.Println("could not register namespace:", err)
		return
	}
}

func ExampleNamespaceOperation_Deregister() {
	// Deregister does not return any errors in case the namespace
	// does not exist. WithFailIfNotExists overrides this.

	nsName := "sales"
	err := servReg.Namespace(nsName).Deregister(context.TODO(),
		deregister.WithFailIfNotExists(),
	)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("namespace does not exist")
		} else {
			fmt.Println("could not deregister namespace:", err)
		}

		return
	}

	fmt.Println("namespace deregistered successfully")
}

func ExampleNamespaceOperation_List() {
	// Get all namespaces and for each of them update their metadata.

	nsIterator := servReg.Namespace(core.Any).List()

	for {
		ns, nsOp, err := nsIterator.Next(context.Background())
		if err != nil {
			if !errors.IsIteratorDone(err) {
				fmt.Println("could not get next namespace:", err)
			}

			break
		}

		fmt.Printf("updating metadata for namespace %s...\n", ns.Name)

		nsOp.Register(context.Background(),
			register.WithKV("last-update", time.Now().String()),
			// WithUpdateMode is not necessary here, as register will figure
			// this out on its own, but we include it to show you a usage
			// example.
			register.WithUpdateMode(),
		)
	}
}

func ExampleNamespaceOperation_List_withFilters() {
	// Get only specific namespaces, and that are in a production stage.

	nsIterator := servReg.Namespace(core.Any).List(
		list.WithNameIn("sales", "it", "hr"),
		list.WithKV("stage", "prod"),
	)

	for {
		ns, _, err := nsIterator.Next(context.Background())
		if err != nil {
			if !errors.IsIteratorDone(err) {
				fmt.Println("could not get next namespace:", err)
			}

			break
		}

		fmt.Println("namespace", ns.Name, "is in production and has metadata", ns.Metadata)
	}
}
