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

	servicedirectory "cloud.google.com/go/servicedirectory/apiv1"
	"github.com/CloudNativeSDWAN/serego/api/core"
	"github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
)

func Example() {
	var (
		client *servicedirectory.RegistrationClient
	)

	// Initialize the client and the settings (look at other examples) here...
	sd, err := core.NewServiceRegistryFromServiceDirectory(
		client,
		wrapper.WithProjectID("my-project-id"),
		wrapper.WithRegion("us-east1"),
	)
	if err != nil {
		// check the error...
		return
	}

	// List all services inside a namespace called "sales", with a certain metadata:
	servIterator := sd.Namespace("sales").
		Service(core.Any).
		List(list.WithMetadataKeyValue("maintainer", "alice.smith@company.com"))

	// Loop through all found results
	for {
		service, _, err := servIterator.Next(context.TODO())
		if err != nil {
			if errors.IsIteratorDone(err) {
				// Finished iterating.
				fmt.Println("finished.")
			} else {
				fmt.Println("could not get next service:", err)
			}

			break
		}

		fmt.Printf("found service with name %s and metadata %+v\n", service.Name, service.Metadata)
	}
}
