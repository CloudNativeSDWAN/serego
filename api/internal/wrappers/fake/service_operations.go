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

package fake

import (
	"context"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
)

type ServiceOperation struct {
	Name_     string
	Get_      func(context.Context, *get.Options) (*coretypes.Service, error)
	Create_   func(context.Context, map[string]string) (*coretypes.Service, error)
	Update_   func(context.Context, map[string]string) (*coretypes.Service, error)
	Delete_   func(context.Context) error
	List_     func(*list.Options) ops.ServiceLister
	Endpoint_ func(name string) ops.EndpointOperation
}

func (s *ServiceOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Service, error) {
	return s.Get_(ctx, opts)
}

func (s *ServiceOperation) Create(ctx context.Context, metadata map[string]string) (*coretypes.Service, error) {
	return s.Create_(ctx, metadata)
}

func (s *ServiceOperation) Update(ctx context.Context, metadata map[string]string) (*coretypes.Service, error) {
	return s.Update_(ctx, metadata)
}

func (s *ServiceOperation) Delete(ctx context.Context) error {
	return s.Delete_(ctx)
}

func (s *ServiceOperation) List(opts *list.Options) ops.ServiceLister {
	return s.List_(opts)
}

type FakeServiceIterator struct {
	Next_ func(ctx context.Context) (*coretypes.Service, ops.ServiceOperation, error)
}

func (s *FakeServiceIterator) Next(ctx context.Context) (*coretypes.Service, ops.ServiceOperation, error) {
	return s.Next_(ctx)
}

func (s *ServiceOperation) Endpoint(name string) ops.EndpointOperation {
	return s.Endpoint_(name)
}
