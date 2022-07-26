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

package etcd_test

import (
	"context"
	"fmt"
	"testing"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

var (
	namespaces    []*coretypes.Namespace
	kvsNamespaces []*mvccpb.KeyValue
	services      []*coretypes.Service
	kvsServices   []*mvccpb.KeyValue
	endpoints     []*coretypes.Endpoint
	kvsEndpoints  []*mvccpb.KeyValue
	ctx           context.Context
)

var _ = BeforeSuite(func() {
	ctx = context.TODO()
	namespaces = []*coretypes.Namespace{}
	kvsNamespaces = []*mvccpb.KeyValue{}
	for i := 1; i < 5; i++ {
		n := &coretypes.Namespace{
			Name: fmt.Sprintf("ns-%d", i),
			Metadata: map[string]string{
				fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
				"another-key":            "another-value",
			},
		}
		kvsNamespaces = append(kvsNamespaces, &mvccpb.KeyValue{
			Key: []byte("/" + n.Name),
			Value: func() []byte {
				v, _ := yaml.Marshal(n)
				return v
			}(),
		})
		namespaces = append(namespaces, n)
		namespaces[i-1].OriginalObject = kvsNamespaces[i-1]

		s := &coretypes.Service{
			Name:      fmt.Sprintf("serv-%d", i),
			Namespace: n.Name,
			Metadata: map[string]string{
				fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
				"another-key":            "another-value",
			},
		}
		kvsServices = append(kvsServices, &mvccpb.KeyValue{
			Key: []byte("/" + s.Name),
			Value: func() []byte {
				v, _ := yaml.Marshal(s)
				return v
			}(),
		})
		services = append(services, s)
		services[i-1].OriginalObject = kvsServices[i-1]

		e := &coretypes.Endpoint{
			Name:      fmt.Sprintf("endp-%d", i),
			Namespace: n.Name,
			Service:   s.Name,
			Metadata: map[string]string{
				fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
				"another-key":            "another-value",
			},
			Address: fmt.Sprintf("%d0.%d1.%d2.%d3", i, i, i, i),
			Port:    80 + int32(i),
		}
		kvsEndpoints = append(kvsEndpoints, &mvccpb.KeyValue{
			Key: []byte("/" + e.Name),
			Value: func() []byte {
				v, _ := yaml.Marshal(e)
				return v
			}(),
		})
		endpoints = append(endpoints, e)
		endpoints[i-1].OriginalObject = kvsEndpoints[i-1]
	}
})

func TestEtcd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Etcd Suite")
}

type fakeKV struct {
	_Put     func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	_Get     func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	_Delete  func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
	_Compact func(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error)
	_Do      func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error)
	_Txn     func(ctx context.Context) clientv3.Txn
}

func (f *fakeKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return f._Put(ctx, key, val, opts...)
}

func (f *fakeKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return f._Get(ctx, key, opts...)
}

func (f *fakeKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return f._Delete(ctx, key, opts...)
}

func (f *fakeKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return f._Compact(ctx, rev, opts...)
}

func (f *fakeKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return f._Do(ctx, op)
}

func (f *fakeKV) Txn(ctx context.Context) clientv3.Txn {
	return f._Txn(ctx)
}
