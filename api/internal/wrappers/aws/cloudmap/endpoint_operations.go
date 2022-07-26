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

package cloudmap

import (
	"context"
	"fmt"
	"net"
	"path"
	"reflect"
	"strconv"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	"github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
)

type cmEndpointOperation struct {
	wrapper  *AwsCloudMapWrapper
	parentOp *cmServiceOperation
	name     string
	pathName string
}

func (e *cmEndpointOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Endpoint, error) {
	if !opts.ForceRefresh {
		if endp := e.wrapper.getFromCache(e.pathName); endp != nil {
			return endp.(*coretypes.Endpoint), nil
		}
	}

	serviceID, err := e.parentOp.getID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error while getting parent service: %w", err)
	}

	out, err := e.wrapper.client.GetInstance(ctx, &servicediscovery.GetInstanceInput{
		InstanceId: aws.String(e.name),
		ServiceId:  serviceID,
	})
	if err != nil {
		return nil, err
	}

	endpoint := toCoreEndpoint(e.parentOp.parentOp.name, e.parentOp.name, out.Instance)
	e.putOnCache(endpoint)

	return endpoint, nil
}

func (e *cmEndpointOperation) Create(ctx context.Context, address string, port int32, metadata map[string]string) (*coretypes.Endpoint, error) {
	// We copy it, so we don't modify the one provided.
	metadataToCreate := map[string]string{}
	for k, v := range metadata {
		metadataToCreate[k] = v
	}

	if address != "" {
		ip := net.ParseIP(address)
		if ip.To4() != nil {
			metadataToCreate["AWS_INSTANCE_IPV4"] = address
		} else {
			metadataToCreate["AWS_INSTANCE_IPV6"] = address
		}
	} else {
		delete(metadataToCreate, "AWS_INSTANCE_IPV4")
		delete(metadataToCreate, "AWS_INSTANCE_IPV6")
	}

	if port != 0 {
		metadataToCreate["AWS_INSTANCE_PORT"] = strconv.Itoa(int(port))
	} else {
		delete(metadataToCreate, "AWS_INSTANCE_PORT")
	}

	serviceID, err := e.parentOp.getID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error while getting parent service: %w", err)
	}

	out, err := e.wrapper.client.RegisterInstance(ctx, &servicediscovery.RegisterInstanceInput{
		InstanceId: aws.String(e.name),
		ServiceId:  serviceID,
		Attributes: metadataToCreate,
	})
	if err != nil {
		return nil, err
	}

	_, err = pollOperationStatus(ctx, e.wrapper.client, aws.ToString(out.OperationId))
	if err != nil {
		return nil, fmt.Errorf("error while checking operation status: %w", err)
	}

	return e.Get(ctx, &get.Options{})
}

func (e *cmEndpointOperation) Update(ctx context.Context, address string, port int32, metadata map[string]string) (*coretypes.Endpoint, error) {
	return e.Create(ctx, address, port, metadata)
}

func (e *cmEndpointOperation) Delete(ctx context.Context) error {
	defer e.deleteFromCache()

	serviceID, err := e.parentOp.getID(ctx)
	if err != nil {
		return fmt.Errorf("error while getting parent service: %w", err)
	}

	out, err := e.wrapper.client.DeregisterInstance(ctx, &servicediscovery.DeregisterInstanceInput{
		InstanceId: aws.String(e.name),
		ServiceId:  serviceID,
	})
	if err != nil {
		return err
	}

	_, err = pollOperationStatus(ctx, e.wrapper.client, aws.ToString(out.OperationId))
	if err != nil {
		return fmt.Errorf("error while checking operation status: %w", err)
	}

	return nil
}

func (e *cmEndpointOperation) deleteFromCache() {
	e.wrapper.cache.Delete(path.Join(e.pathName))
}

func (e *cmEndpointOperation) putOnCache(endpoint *coretypes.Endpoint) {
	e.wrapper.cache.SetDefault(e.pathName, endpoint)
}

func (e *cmEndpointOperation) List(opts *list.Options) ops.EndpointLister {
	return &cloudMapEndpointsIterator{
		wrapper:  e.wrapper,
		parentOp: e.parentOp,
		options:  opts,
		hasMore:  true,
	}
}

type cloudMapEndpointsIterator struct {
	wrapper    *AwsCloudMapWrapper
	options    *list.Options
	parentOp   *cmServiceOperation
	parentServ *coretypes.Service
	parentID   *string
	currIndex  int
	nextToken  *string
	elements   []types.InstanceSummary
	hasMore    bool
}

func (ei *cloudMapEndpointsIterator) Next(ctx context.Context) (*coretypes.Endpoint, ops.EndpointOperation, error) {
	client := ei.wrapper.client

	if ei.parentOp.name == "" || ei.parentOp.parentOp.name == "" {
		return nil, nil, fmt.Errorf("cannot get the next element: %w", errors.MissingName)
	}

	if ei.parentServ == nil && ei.hasMore {
		serv, err := ei.parentOp.Get(ctx, &get.Options{})
		if err != nil {
			ei.hasMore = false
			return nil, nil, fmt.Errorf("error while getting parent service: %w", err)
		}
		ei.parentServ = serv
		ei.parentID = ei.parentServ.OriginalObject.(*types.Service).Id
	}

	for i := ei.currIndex; i < len(ei.elements); i++ {
		elemsToFilter := []types.InstanceSummary{ei.elements[i]}
		if ei.options.AddressFilters != nil {
			// If user has address filters we need to test this with different
			// IPs.
			elemsToFilter = []types.InstanceSummary{}
			for _, val := range []string{"AWS_INSTANCE_IPV4", "AWS_INSTANCE_IPV6"} {
				elemsToFilter = append(elemsToFilter, types.InstanceSummary{
					Id: ei.elements[i].Id,
					Attributes: func() map[string]string {
						cpyMap := map[string]string{}
						for k, v := range ei.elements[i].Attributes {
							if k != val {
								cpyMap[k] = v
							}
						}

						return cpyMap
					}(),
				})
			}
		}

		for j := range elemsToFilter {
			inst := toCoreEndpoint(ei.parentServ.Namespace, ei.parentServ.Name, &elemsToFilter[j])

			if passed, _ := ei.options.Filter(inst); passed {
				newOp := ei.parentOp.Endpoint(inst.Name).(*cmEndpointOperation)
				newOp.putOnCache(inst)
				ei.currIndex = i + 1
				return inst, newOp, nil
			}
		}
	}

	if ei.hasMore {
		out, err := client.ListInstances(ctx, &servicediscovery.ListInstancesInput{
			ServiceId:  ei.parentID,
			MaxResults: aws.Int32(ei.options.Results),
			NextToken:  ei.nextToken,
		})
		if err != nil {
			ei.hasMore = false
			return nil, nil, fmt.Errorf("error while getting new resources: %w", err)
		}

		ei.elements = append(ei.elements, out.Instances...)
		if out.NextToken != nil {
			ei.nextToken = out.NextToken
			ei.hasMore = true
		} else {
			ei.nextToken = nil
			ei.hasMore = false
		}

		return ei.Next(ctx)
	}

	return nil, nil, errors.IteratorDone
}

func toCoreEndpoint(namespace, service string, inst interface{}) *coretypes.Endpoint {
	instValue := reflect.ValueOf(inst).Elem()
	attributes := instValue.FieldByName("Attributes").
		Interface().(map[string]string)

	address := attributes["AWS_INSTANCE_IPV4"]
	if address == "" {
		address = attributes["AWS_INSTANCE_IPV6"]
	}

	var port int32 = 0
	if attributes["AWS_INSTANCE_PORT"] != "" {
		p, err := strconv.Atoi(attributes["AWS_INSTANCE_PORT"])
		if err == nil {
			port = int32(p)
		}
	}

	removeAttr := map[string]bool{
		"AWS_INSTANCE_IPV4": true,
		"AWS_INSTANCE_IPV6": true,
		"AWS_INSTANCE_PORT": true,
	}
	metadata := map[string]string{}
	for k, v := range attributes {
		if _, exists := removeAttr[k]; !exists {
			metadata[k] = v
		}
	}

	return &coretypes.Endpoint{
		Name:      instValue.FieldByName("Id").Elem().String(),
		Namespace: namespace,
		Service:   service,
		Port:      port,
		Address:   address,
		Metadata:  metadata,
		OriginalObject: func() *types.Instance {
			if summary, ok := inst.(*types.InstanceSummary); ok {
				return fromSummaryToInstance(summary)
			}

			return inst.(*types.Instance)
		}(),
	}
}
