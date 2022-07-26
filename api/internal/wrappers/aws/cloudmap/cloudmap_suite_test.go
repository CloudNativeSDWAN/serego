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

package cloudmap_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/aws/cloudmap"
	"github.com/aws/aws-sdk-go-v2/aws"
	sd "github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	namespaces []types.NamespaceSummary
	services   []types.ServiceSummary
	endpoints  []types.InstanceSummary
	tags       [][]types.Tag
	metas      []map[string]string

	serv types.ServiceSummary
	endp types.InstanceSummary

	listTagsForResource func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error)
	listNamespaces      func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error)
	listServices        func(ctx context.Context, params *sd.ListServicesInput, optFns ...func(*sd.Options)) (*sd.ListServicesOutput, error)
	listInstances       func(ctx context.Context, params *sd.ListInstancesInput, optFns ...func(*sd.Options)) (*sd.ListInstancesOutput, error)
)

func TestCloudmap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloudmap Suite")
}

var _ = BeforeSuite(func() {
	namespaces = []types.NamespaceSummary{}
	services = []types.ServiceSummary{}
	endpoints = []types.InstanceSummary{}
	tags = [][]types.Tag{}
	metas = []map[string]string{}
	for i := 1; i < 5; i++ {
		namespaces = append(namespaces, types.NamespaceSummary{
			Name: aws.String(fmt.Sprintf("ns-name-%d", i)),
			Arn:  aws.String(fmt.Sprintf("ns-arn-%d", i)),
			Id:   aws.String(fmt.Sprintf("ns-id-%d", i)),
		})

		services = append(services, types.ServiceSummary{
			Name: aws.String(fmt.Sprintf("serv-name-%d", i)),
			Arn:  aws.String(fmt.Sprintf("serv-arn-%d", i)),
			Id:   aws.String(fmt.Sprintf("serv-id-%d", i)),
		})

		endpoints = append(endpoints, types.InstanceSummary{
			Id: aws.String(fmt.Sprintf("endp-name-%d", i)),
			Attributes: map[string]string{
				"AWS_INSTANCE_IPV4":      fmt.Sprintf("%d0.%d1.%d2.%d3", i, i, i, i),
				"AWS_INSTANCE_PORT":      "80",
				fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
				"another":                "value",
			},
		})

		tags = append(tags, []types.Tag{
			{
				Key:   aws.String(fmt.Sprintf("key-%d", i)),
				Value: aws.String(fmt.Sprintf("val-%d", i)),
			},
			{
				Key:   aws.String("another"),
				Value: aws.String("value"),
			},
		})

		metas = append(metas, map[string]string{
			fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
			"another":                "value",
		})
	}

	listTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
		switch aws.ToString(params.ResourceARN) {
		case "ns-arn-1", "serv-arn-1":
			return &sd.ListTagsForResourceOutput{
				Tags: tags[0],
			}, nil
		case "ns-arn-2", "serv-arn-2":
			return &sd.ListTagsForResourceOutput{
				Tags: tags[1],
			}, nil
		case "ns-arn-3", "serv-arn-3":
			return &sd.ListTagsForResourceOutput{
				Tags: tags[2],
			}, nil
		case "ns-arn-4", "serv-arn-4":
			return &sd.ListTagsForResourceOutput{
				Tags: tags[3],
			}, nil
		default:
			return nil, &types.ResourceNotFoundException{}
		}
	}
	listNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
		return &sd.ListNamespacesOutput{
			Namespaces: namespaces,
		}, nil
	}
	listServices = func(ctx context.Context, params *sd.ListServicesInput, optFns ...func(*sd.Options)) (*sd.ListServicesOutput, error) {
		return &sd.ListServicesOutput{
			Services: services,
		}, nil
	}
	listInstances = func(ctx context.Context, params *sd.ListInstancesInput, optFns ...func(*sd.Options)) (*sd.ListInstancesOutput, error) {
		return &sd.ListInstancesOutput{
			Instances: endpoints,
		}, nil
	}
	cloudmap.PollTick = time.Microsecond
})
