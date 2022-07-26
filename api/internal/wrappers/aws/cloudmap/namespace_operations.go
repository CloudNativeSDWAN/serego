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
	"github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
)

type cmNamespaceOperation struct {
	wrapper  *AwsCloudMapWrapper
	name     string
	pathName string
}

func (n *cmNamespaceOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Namespace, error) {
	if !opts.ForceRefresh {
		if ns := n.wrapper.getFromCache(n.pathName); ns != nil {
			return ns.(*coretypes.Namespace), nil
		}
	}

	if nsID := n.wrapper.getFromCache(path.Join(n.pathName, pathID)); nsID != nil {
		// We already have its ID, nice! It means we can load it instantly.
		return n.getByID(ctx, nsID.(*string))
	}

	// We don't have its ID, it means we have to search for it...
	nsit := n.List(&list.Options{})
	ns, _, err := nsit.Next(ctx)
	if err != nil {
		if errors.IsIteratorDone(err) {
			return nil, errors.NamespaceNotFound
		}

		return nil, err
	}

	return ns, nil
}

func (n *cmNamespaceOperation) getByID(ctx context.Context, nsID *string) (*coretypes.Namespace, error) {
	out, err := n.wrapper.client.GetNamespace(ctx, &servicediscovery.GetNamespaceInput{
		Id: nsID,
	})
	if err != nil {
		n.deleteFromCache()
		return nil, fmt.Errorf("cannot get namespace from ID %s: %w", *nsID, err)
	}

	outTags, err := n.wrapper.client.ListTagsForResource(ctx, &servicediscovery.ListTagsForResourceInput{
		ResourceARN: out.Namespace.Arn,
	})
	if err != nil {
		// This occurs when someone deletes the namespace just after
		// we retrieved it, making tags disappear as well.
		// So we throw an error because the namespace doesn't exist anymore
		// either.
		return nil, fmt.Errorf("cannot get namespace from ID %s: %w", *nsID, err)
	}

	namespace := toCoreNamespace(out.Namespace, outTags.Tags)
	n.putOnCache(namespace)

	return namespace, nil
}

func (n *cmNamespaceOperation) getID(ctx context.Context) (*string, error) {
	if nsID := n.wrapper.getFromCache(path.Join(n.pathName, pathID)); nsID != nil {
		return nsID.(*string), nil
	}

	namespace, err := n.Get(ctx, &get.Options{})
	if err != nil {
		return nil, err
	}

	return namespace.OriginalObject.(*types.Namespace).Id, nil
}

func (n *cmNamespaceOperation) deleteFromCache() {
	n.wrapper.cache.Delete(path.Join(n.pathName))
	n.wrapper.cache.Delete(path.Join(n.pathName, pathID))
	n.wrapper.cache.Delete(path.Join(n.pathName, pathARN))
}

func (n *cmNamespaceOperation) putOnCache(namespace *coretypes.Namespace) {
	n.wrapper.cache.SetDefault(n.pathName, namespace)

	original := namespace.OriginalObject.(*types.Namespace)
	n.wrapper.cache.Set(path.Join(n.pathName, pathID), original.Id, time.Hour)
	n.wrapper.cache.Set(path.Join(n.pathName, pathARN), original.Arn, time.Hour)
}

func (n *cmNamespaceOperation) Create(ctx context.Context, metadata map[string]string) (*coretypes.Namespace, error) {
	out, err := n.wrapper.client.CreateHttpNamespace(ctx, &servicediscovery.CreateHttpNamespaceInput{
		Name: aws.String(n.name),
		Tags: fromMapToTagsSlice(metadata),
	})
	if err != nil {
		return nil, err
	}

	awsOperation, err := pollOperationStatus(ctx, n.wrapper.client, aws.ToString(out.OperationId))
	if err != nil {
		return nil,
			fmt.Errorf("error while checking operation status %s: %w",
				aws.ToString(out.OperationId), err)
	}

	nsID := awsOperation.Targets["NAMESPACE"]

	// We get it again, so we can return any change in the original object as
	// well, e.g. last update time, etc..
	return n.getByID(ctx, &nsID)
}

func (n *cmNamespaceOperation) Update(ctx context.Context, metadata map[string]string) (*coretypes.Namespace, error) {
	// You cannot edit the name of namespace, so the only thing you can update
	// in a namespace is its tags.
	var ns *types.Namespace
	{
		namespace, err := n.Get(ctx, &get.Options{})
		if err != nil {
			return nil, fmt.Errorf("error while checking if namespace exists: %w", err)
		}

		ns = namespace.OriginalObject.(*types.Namespace)
	}

	if err := updateTags(ctx, n.wrapper.client, *ns.Arn, metadata); err != nil {
		return nil, err
	}

	return n.getByID(ctx, ns.Id)
}

func (n *cmNamespaceOperation) Delete(ctx context.Context) error {
	defer n.deleteFromCache()

	var ns *types.Namespace
	{
		namespace, err := n.Get(ctx, &get.Options{})
		if err != nil {
			return fmt.Errorf("error while checking if namespace exists: %w", err)
		}

		ns = namespace.OriginalObject.(*types.Namespace)
	}

	out, err := n.wrapper.client.DeleteNamespace(ctx, &servicediscovery.DeleteNamespaceInput{
		Id: ns.Id,
	})
	if err != nil {
		return err
	}

	_, err = pollOperationStatus(ctx, n.wrapper.client, aws.ToString(out.OperationId))
	if err != nil {
		return fmt.Errorf("error while checking operation status: %w", err)
	}

	return nil
}

func (n *cmNamespaceOperation) List(opts *list.Options) operations.NamespaceLister {
	if n.name != "" {
		if opts == nil {
			opts = &list.Options{}
		}

		// Add the name as a filter
		if opts.NameFilters == nil {
			opts.NameFilters = &list.NameFilters{}
		}

		opts.NameFilters.In = append(opts.NameFilters.In, n.name)
	}

	if opts.Results == 0 {
		opts.Results = list.DefaultListResultsNumber
	}

	return &cloudMapNamespaceIterator{
		wrapper:   n.wrapper,
		hasMore:   true,
		currIndex: 0,
		elements:  []types.NamespaceSummary{},
		options:   opts,
	}
}

type cloudMapNamespaceIterator struct {
	wrapper   *AwsCloudMapWrapper
	options   *list.Options
	currIndex int
	nextToken *string
	elements  []types.NamespaceSummary
	hasMore   bool
}

func (ni *cloudMapNamespaceIterator) Next(ctx context.Context) (*coretypes.Namespace, operations.NamespaceOperation, error) {
	client := ni.wrapper.client

	for i := ni.currIndex; i < len(ni.elements); i++ {
		ns := toCoreNamespace(&ni.elements[i], []types.Tag{})

		outTags, err := client.ListTagsForResource(ctx, &servicediscovery.ListTagsForResourceInput{
			ResourceARN: ni.elements[i].Arn,
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
			ns.Metadata = fromTagsSliceToMap(outTags.Tags)
		}

		if passed, _ := ni.options.Filter(ns); passed {
			ni.currIndex = i + 1
			newWrapper := ni.wrapper.Namespace(ns.Name).(*cmNamespaceOperation)
			newWrapper.putOnCache(ns)
			return ns, newWrapper, nil
		}
	}

	if ni.hasMore {
		out, err := client.ListNamespaces(ctx, &servicediscovery.ListNamespacesInput{
			MaxResults: aws.Int32(ni.options.Results),
			NextToken:  ni.nextToken,
		})
		if err != nil {
			ni.hasMore = false
			return nil, nil, fmt.Errorf("error while getting next page: %w", err)
		}

		ni.elements = append(ni.elements, out.Namespaces...)
		if out.NextToken != nil {
			ni.nextToken = out.NextToken
			ni.hasMore = true
		} else {
			ni.nextToken = nil
			ni.hasMore = false
		}

		return ni.Next(ctx)
	}

	return nil, nil, errors.IteratorDone
}

func (n *cmNamespaceOperation) Service(name string) operations.ServiceOperation {
	var (
		pathName string
	)

	if name != "" {
		pathName = path.Join(n.pathName, pathServices, name)
	}

	return &cmServiceOperation{
		wrapper:  n.wrapper,
		parentOp: n,
		name:     name,
		pathName: pathName,
	}
}

func toCoreNamespace(ns interface{}, tags []types.Tag) *coretypes.Namespace {
	// ns is either a *types.Namespace or *types.NamespaceSummary.
	nsValue := reflect.ValueOf(ns).Elem()
	return &coretypes.Namespace{
		Name:     nsValue.FieldByName("Name").Elem().String(),
		Metadata: fromTagsSliceToMap(tags),
		OriginalObject: func() *types.Namespace {
			if summary, ok := ns.(*types.NamespaceSummary); ok {
				return fromSummaryToNamespace(summary)
			}

			return ns.(*types.Namespace)
		}(),
	}
}
