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

type NamespaceOperation struct {
	Name_    string
	Get_     func(context.Context, *get.Options) (*coretypes.Namespace, error)
	Create_  func(context.Context, map[string]string) (*coretypes.Namespace, error)
	Update_  func(context.Context, map[string]string) (*coretypes.Namespace, error)
	Delete_  func(context.Context) error
	List_    func(*list.Options) ops.NamespaceLister
	Service_ func(string) ops.ServiceOperation
}

func (n *NamespaceOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Namespace, error) {
	return n.Get_(ctx, opts)
}

func (n *NamespaceOperation) Create(ctx context.Context, metadata map[string]string) (*coretypes.Namespace, error) {
	return n.Create_(ctx, metadata)
}

func (n *NamespaceOperation) Update(ctx context.Context, metadata map[string]string) (*coretypes.Namespace, error) {
	return n.Update_(ctx, metadata)
}

func (n *NamespaceOperation) Delete(ctx context.Context) error {
	return n.Delete_(ctx)
}

func (n *NamespaceOperation) List(opts *list.Options) ops.NamespaceLister {
	return n.List_(opts)
}

type FakeNamespaceIterator struct {
	Next_ func(context.Context) (*coretypes.Namespace, ops.NamespaceOperation, error)
}

func (ni *FakeNamespaceIterator) Next(ctx context.Context) (*coretypes.Namespace, ops.NamespaceOperation, error) {
	return ni.Next_(ctx)
}

func (n *NamespaceOperation) Service(name string) ops.ServiceOperation {
	return n.Service_(name)
}
