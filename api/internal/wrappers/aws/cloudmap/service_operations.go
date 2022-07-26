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
	"path"
	"reflect"
	"time"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	"github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
)

type serviceWithTags struct {
	service *types.Service
	tags    []types.Tag
}

type cmServiceOperation struct {
	wrapper  *AwsCloudMapWrapper
	parentOp *cmNamespaceOperation
	pathName string
	name     string
}

func (s *cmServiceOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Service, error) {
	if !opts.ForceRefresh {
		if serv := s.wrapper.getFromCache(s.pathName); serv != nil {
			return serv.(*coretypes.Service), nil
		}
	}

	if servID := s.wrapper.getFromCache(path.Join(s.pathName, pathID)); servID != nil {
		// We already have its ID, nice! It means we can load it instantly.
		return s.getByID(ctx, servID.(*string))
	}

	// We don't have its ID, it means we have to search for it...
	sl := s.List(&list.Options{})
	serv, _, err := sl.Next(ctx)
	if err != nil {
		if errors.IsIteratorDone(err) {
			return nil, errors.ServiceNotFound
		}

		return nil, err
	}

	return serv, nil
}

func (s *cmServiceOperation) getByID(ctx context.Context, servID *string) (*coretypes.Service, error) {
	out, err := s.wrapper.client.GetService(ctx, &servicediscovery.GetServiceInput{
		Id: servID,
	})
	if err != nil {
		s.deleteFromCache()
		return nil, err
	}

	outTags, err := s.wrapper.client.ListTagsForResource(ctx, &servicediscovery.ListTagsForResourceInput{
		ResourceARN: out.Service.Arn,
	})
	if err != nil {
		// This occurs when someone deletes the service just after
		// we retrieved it, making tags disappear as well.
		// So we throw an error because the namespace doesn't exist anymore
		// either.
		return nil, fmt.Errorf("cannot get service from ID %s: %w", *servID, err)
	}

	service := toCoreService(s.parentOp.name, out.Service, outTags.Tags)
	s.putOnCache(service)

	return service, nil
}

func (s *cmServiceOperation) getID(ctx context.Context) (*string, error) {
	if servID := s.wrapper.getFromCache(path.Join(s.pathName, pathID)); servID != nil {
		return servID.(*string), nil
	}

	serv, err := s.Get(ctx, &get.Options{})
	if err != nil {
		return nil, err
	}

	return serv.OriginalObject.(*types.Service).Id, nil
}

func (s *cmServiceOperation) deleteFromCache() {
	s.wrapper.cache.Delete(path.Join(s.pathName))
	s.wrapper.cache.Delete(path.Join(s.pathName, pathID))
	s.wrapper.cache.Delete(path.Join(s.pathName, pathARN))
}

func (s *cmServiceOperation) putOnCache(service *coretypes.Service) {
	s.wrapper.cache.SetDefault(s.pathName, service)

	original := service.OriginalObject.(*types.Service)
	s.wrapper.cache.Set(path.Join(s.pathName, pathID), original.Id, time.Hour)
	s.wrapper.cache.Set(path.Join(s.pathName, pathARN), original.Arn, time.Hour)
}

func (s *cmServiceOperation) Create(ctx context.Context, metadata map[string]string) (*coretypes.Service, error) {
	namespaceID, err := s.parentOp.getID(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get ID of parent namespace: %w", err)
	}

	out, err := s.wrapper.client.CreateService(ctx, &servicediscovery.CreateServiceInput{
		Name:        aws.String(s.name),
		NamespaceId: namespaceID,
		Tags:        fromMapToTagsSlice(metadata),
		Type:        types.ServiceTypeOptionHttp,
	})
	if err != nil {
		return nil, err
	}

	// We get it again, so we can return any change in the original object as
	// well, e.g. last update time, etc..
	return s.getByID(ctx, out.Service.Id)
}

func (s *cmServiceOperation) Update(ctx context.Context, metadata map[string]string) (*coretypes.Service, error) {
	// You cannot edit the name of service, so the only thing you can update
	// is its name
	var serv *types.Service
	{
		service, err := s.Get(ctx, &get.Options{})
		if err != nil {
			return nil, fmt.Errorf("error while checking if service exists: %w", err)
		}

		serv = service.OriginalObject.(*types.Service)
	}

	if err := updateTags(ctx, s.wrapper.client, *serv.Arn, metadata); err != nil {
		return nil, err
	}

	return s.getByID(ctx, serv.Id)
}

func (s *cmServiceOperation) Delete(ctx context.Context) error {
	defer s.deleteFromCache()

	var serv *types.Service
	{
		service, err := s.Get(ctx, &get.Options{})
		if err != nil {
			return fmt.Errorf("error while checking if service exists: %w", err)
		}

		serv = service.OriginalObject.(*types.Service)
	}

	_, err := s.wrapper.client.DeleteService(ctx, &servicediscovery.DeleteServiceInput{
		Id: serv.Id,
	})
	return err
}

func (s *cmServiceOperation) List(opts *list.Options) ops.ServiceLister {
	if s.name != "" {
		if opts == nil {
			opts = &list.Options{}
		}

		// Add the name as a filter
		if opts.NameFilters == nil {
			opts.NameFilters = &list.NameFilters{}
		}

		opts.NameFilters.In = append(opts.NameFilters.In, s.name)
	}

	if opts.Results == 0 {
		opts.Results = list.DefaultListResultsNumber
	}

	return &cloudMapServicesIterator{
		wrapper:  s.wrapper,
		parentOp: s.parentOp,
		options:  opts,
		hasMore:  true,
	}
}

type cloudMapServicesIterator struct {
	wrapper   *AwsCloudMapWrapper
	options   *list.Options
	parentOp  *cmNamespaceOperation
	parentNs  *coretypes.Namespace
	parentID  *string
	currIndex int
	nextToken *string
	elements  []types.ServiceSummary
	hasMore   bool
}

func (si *cloudMapServicesIterator) Next(ctx context.Context) (*coretypes.Service, ops.ServiceOperation, error) {
	client := si.wrapper.client

	if si.parentOp.name == "" {
		return nil, nil, fmt.Errorf("cannot load next element: %w", errors.EmptyNamespaceName)
	}

	if si.parentNs == nil && si.hasMore {
		ns, err := si.parentOp.Get(ctx, &get.Options{})
		if err != nil {
			si.hasMore = false
			return nil, nil, fmt.Errorf("error while getting parent namespace: %w", err)
		}
		si.parentNs = ns
		si.parentID = si.parentNs.OriginalObject.(*types.Namespace).Id
	}

	for i := si.currIndex; i < len(si.elements); i++ {
		serv := toCoreService(si.parentNs.Name, &si.elements[i], []types.Tag{})
		serv.OriginalObject.(*types.Service).NamespaceId = si.parentID

		outTags, err := client.ListTagsForResource(ctx, &servicediscovery.ListTagsForResourceInput{
			ResourceARN: si.elements[i].Arn,
		})
		if err != nil {
			if errors.IsNotFound(err) {
				// This means that the namespace was deleted just after we
				// got it, and it is a very rare case, but let's cover this
				// anyways.
				continue
			}

			// Keep empty tags
		} else {
			serv.Metadata = fromTagsSliceToMap(outTags.Tags)
		}

		if passed, _ := si.options.Filter(serv); passed {
			si.currIndex = i + 1
			newWrapper := si.parentOp.Service(serv.Name).(*cmServiceOperation)
			newWrapper.putOnCache(serv)

			return serv, newWrapper, nil
		}
	}

	if si.hasMore {
		out, err := client.ListServices(ctx, &servicediscovery.ListServicesInput{
			Filters: []types.ServiceFilter{
				{
					Name:      types.ServiceFilterNameNamespaceId,
					Condition: types.FilterConditionEq,
					Values:    []string{*si.parentID},
				},
			},
			MaxResults: aws.Int32(si.options.Results),
			NextToken:  si.nextToken,
		})
		if err != nil {
			si.hasMore = false
			return nil, nil, fmt.Errorf("error while getting new resources: %w", err)
		}

		si.elements = append(si.elements, out.Services...)
		if out.NextToken != nil {
			si.nextToken = out.NextToken
			si.hasMore = true
		} else {
			si.nextToken = nil
			si.hasMore = false
		}

		return si.Next(ctx)
	}

	return nil, nil, errors.IteratorDone
}

func (s *cmServiceOperation) Endpoint(name string) ops.EndpointOperation {
	pathName := path.Join(s.pathName, pathEndpoints, name)

	return &cmEndpointOperation{
		wrapper:  s.wrapper,
		parentOp: s,
		name:     name,
		pathName: pathName,
	}
}

func toCoreService(namespaceName string, serv interface{}, tags []types.Tag) *coretypes.Service {
	// serv is either a *types.Service or *types.ServiceSummary.
	servValue := reflect.ValueOf(serv).Elem()
	return &coretypes.Service{
		Name:      servValue.FieldByName("Name").Elem().String(),
		Namespace: namespaceName,
		Metadata:  fromTagsSliceToMap(tags),
		OriginalObject: func() *types.Service {
			if s, ok := serv.(*types.ServiceSummary); ok {
				return fromSummaryToService(s)
			}

			return serv.(*types.Service)
		}(),
	}
}
