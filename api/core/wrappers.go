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

package core

import (
	"fmt"

	servicedirectory "cloud.google.com/go/servicedirectory/apiv1"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	cmw "github.com/CloudNativeSDWAN/serego/api/internal/wrappers/aws/cloudmap"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/etcd"
	sdw "github.com/CloudNativeSDWAN/serego/api/internal/wrappers/gcp/servicedirectory"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// NewServiceRegistryFromServiceDirectory starts a new ServiceRegistry wrapper
// on top of Google Service Directory.
//
// You must provide a valid client for this to work correctly, along with
// non-nil and valid settings. Optionally, you can also fine tune the behavior
// of the API by providing Wrapper options as well.
//
// Note that this needs a registration client, *not* a lookup client.
//
// NOTE: this *needs* a region and project ID, please look at the example.
func NewServiceRegistryFromServiceDirectory(client *servicedirectory.RegistrationClient, option ...wrapper.Option) (*ServiceRegistry, error) {
	wopts := &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime}
	for _, wo := range option {
		if err := wo(wopts); err != nil {
			return nil, err
		}
	}

	wrapper, err := sdw.NewServiceDirectoryWrapper(client, wopts)
	if err != nil {
		return nil, fmt.Errorf("could not get wrapper for Service Directory: %w", err)
	}

	return &ServiceRegistry{
		wrapper: wrapper,
	}, nil
}

// NewServiceRegistryFromCloudMap starts a new ServiceRegistry wrapper on top
// of AWS Cloud Map.
//
// It returns an error if the client is nil.
func NewServiceRegistryFromCloudMap(client *servicediscovery.Client, option ...wrapper.Option) (*ServiceRegistry, error) {
	wopts := &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime}
	for _, wo := range option {
		if err := wo(wopts); err != nil {
			return nil, err
		}
	}

	wrapper, err := cmw.NewCloudMapWrapper(client, wopts)
	if err != nil {
		return nil, fmt.Errorf("could not get wrapper for Cloud Map: %w", err)
	}

	return &ServiceRegistry{
		wrapper: wrapper,
	}, nil
}

// NewServiceRegistryFromEtcd uses etcd to create a service registry.
//
// It returns an error if the client is nil or if the connection to etcd fails.
//
// NOTE: you will have to provide a prefix for etcd, please look at the
// example.
func NewServiceRegistryFromEtcd(client *clientv3.Client, option ...wrapper.Option) (*ServiceRegistry, error) {
	wopts := &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime}
	for _, wo := range option {
		if err := wo(wopts); err != nil {
			return nil, err
		}
	}

	wrapper, err := etcd.NewEtcdWrapper(client, wopts)
	if err != nil {
		return nil, fmt.Errorf("could not get wrapper for etcd: %w", err)
	}

	return &ServiceRegistry{
		wrapper: wrapper,
	}, nil
}

// NewServiceRegistryFromWrapper returns a ServiceRegistry wrapper with a
// generic service registry.
//
// You should not use this function to create a ServiceRegistry wrapper, but
// rather use one of the other provided functions as this one is mostly used
// for testing and may be deprecated or removed in future.
func NewServiceRegistryFromWrapper(wrapper ops.ServiceRegistryWrapper, wopt ...wrapper.Option) (*ServiceRegistry, error) {
	if wrapper == nil {
		return nil, srerr.NoOperationSet
	}

	sr := &ServiceRegistry{
		wrapper: wrapper,
	}

	return sr, nil
}
