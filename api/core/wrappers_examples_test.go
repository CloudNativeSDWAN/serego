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
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	etcdns "go.etcd.io/etcd/client/v3/namespace"
	"google.golang.org/api/option"
)

func ExampleNewServiceRegistryFromServiceDirectory() {
	// First, get a registration client from service directory.
	// Note that this is just one way to do it, refer to service directory's
	// documentation and API to learn alternative ways.
	cl, err := servicedirectory.NewRegistrationClient(
		context.Background(),
		option.WithCredentialsFile("path/to/the/service-account.json"),
	)
	if err != nil {
		fmt.Println("could not get service directory client:", err, ". Exiting...")
		return
	}
	defer cl.Close()

	// NewServiceRegistryFromServiceDirectory returns an error only if the
	// registration client is nil or if project ID and region are not
	// provided.
	sr, err := core.NewServiceRegistryFromServiceDirectory(
		cl,
		wrapper.WithProjectID("my-project-id"),
		wrapper.WithRegion("us-east1"),
		wrapper.WithCacheExpirationTime(10*time.Minute),
	)
	if err != nil {
		// check for any errors here....
	}

	// You can now start doing operations: look at the other examples.
	service, err := sr.Namespace("hr").Service("payroll").Get(context.TODO())
	if err != nil {
		// check for any errors here...
		return
	}

	fmt.Println("service", service.Name, "has metadata", service.Metadata)
}

func ExampleNewServiceRegistryFromCloudMap() {
	// First, get a client for Cloud Map. This is just an example:
	// refer to Cloud Map's documentation to learn more.
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-west-2"))
	if err != nil {
		fmt.Println("could not get client for AWS Cloud Map:", err, ". Exiting...")
		return
	}

	// NewServiceRegistryFromCloudMap returns an error only if the provided
	// client is nil.
	sd, _ := core.NewServiceRegistryFromCloudMap(servicediscovery.NewFromConfig(cfg))

	// You can now start doing operations: look at the other examples.
	service, err := sd.Namespace("hr").Service("payroll").Get(context.TODO())
	if err != nil {
		// check for any errors here...
		return
	}

	fmt.Println("service", service.Name, "has metadata", service.Metadata)
}

func ExampleNewServiceRegistryFromEtcd() {
	// First, get a client for Cloud Map. This is just an example:
	// refer to etcd's documentation to learn more.
	cfg := clientv3.Config{
		Endpoints:   []string{"http://localhost:2379"},
		Username:    "my-username",
		Password:    "my-password",
		DialTimeout: 5 * time.Second,
	}

	cl, err := clientv3.New(cfg)
	if err != nil {
		fmt.Println("could not get client for etcd:", err, ". Exiting...")
		return
	}

	// Define a prefix that all objects will have. Here, for example, all
	// objects will have the "/service-registry" prefix. Serego will use this
	// prefix and each object will be added as "/name". Examples:
	// - "/service-registry/namespaces/my-namespace"
	// - "/service-registry/namespaces/hr/services/payroll"
	cl.KV = etcdns.NewKV(cl.KV, "/service-registry")

	sr, err := core.NewServiceRegistryFromEtcd(cl)
	if err != nil {
		// check for any errors here...
		return
	}

	// You can now start doing operations: look at the other examples.
	service, err := sr.Namespace("hr").Service("payroll").Get(context.TODO())
	if err != nil {
		// check for any errors here...
		return
	}

	fmt.Println("service", service.Name, "has metadata", service.Metadata)
}
