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
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"go.etcd.io/etcd/api/v3/mvccpb"
	pb "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
)

func ExampleNamespace_convert() {
	// Get the namespace here in some way...
	// Here we just use a sample.
	ns := &Namespace{}

	// In case you are using Google Service Directory, convert it like this:
	// NOTE: pb is from "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	sdNamespace := ns.OriginalObject.(*pb.Namespace)
	_ = sdNamespace

	// In case you are using AWS Cloud Map, convert it like this:
	// NOTE: types is from "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	cmNamespace := ns.OriginalObject.(*types.Namespace)
	_ = cmNamespace

	// In case you are using etcd, convert it like this:
	// NOTE: mvccp is from "go.etcd.io/etcd/api/v3/mvccpb"
	etcdNamespace := ns.OriginalObject.(*mvccpb.KeyValue)
	_ = etcdNamespace
}

func ExampleService_convert() {
	// Get the service here in some way...
	// Here we just use a sample.
	service := &Service{}

	// In case you are using Google Service Directory, convert it like this:
	// NOTE: pb is from "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	sdService := service.OriginalObject.(*pb.Service)
	_ = sdService

	// In case you are using AWS Cloud Map, convert it like this:
	// NOTE: types is from "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	cmService := service.OriginalObject.(*types.Service)
	_ = cmService

	// In case you are using etcd, convert it like this:
	// NOTE: mvccp is from "go.etcd.io/etcd/api/v3/mvccpb"
	etcdService := service.OriginalObject.(*mvccpb.KeyValue)
	_ = etcdService
}

func ExampleEndpoint_convert() {
	// Get the endpoint here in some way...
	// Here we just use a sample.
	endpoint := &Endpoint{}

	// In case you are using Google Service Directory, convert it like this:
	// NOTE: pb is from "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	sdEndpoint := endpoint.OriginalObject.(*pb.Endpoint)
	_ = sdEndpoint

	// In case you are using AWS Cloud Map, convert it like this:
	// NOTE: types is from "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	cmEndpoint := endpoint.OriginalObject.(*types.Instance)
	_ = cmEndpoint

	// In case you are using etcd, convert it like this:
	// NOTE: mvccp is from "go.etcd.io/etcd/api/v3/mvccpb"
	etcdEndpoint := endpoint.OriginalObject.(*mvccpb.KeyValue)
	_ = etcdEndpoint
}
