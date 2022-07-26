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

package etcd

import (
	"context"
	"strings"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func getOne(ctx context.Context, kv clientv3.KV, name string) (*mvccpb.KeyValue, error) {
	resp, err := kv.Get(ctx, prependSlash(name))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, srerr.NotFound
	}

	return resp.Kvs[0], nil
}

func getList(ctx context.Context, kv clientv3.KV, name string, limit int32) ([]*mvccpb.KeyValue, error) {
	etcdLimit := int64(limit)
	resp, err := kv.Get(ctx, prependSlash(name), clientv3.WithFromKey(), clientv3.WithLimit(etcdLimit))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return []*mvccpb.KeyValue{}, nil
	}

	return resp.Kvs, nil
}

func prependSlash(name string) string {
	if strings.HasPrefix(name, "/") {
		return name
	}

	return "/" + name
}
