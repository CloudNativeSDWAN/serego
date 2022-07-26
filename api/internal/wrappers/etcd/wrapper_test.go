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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clientv3 "go.etcd.io/etcd/client/v3"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/etcd"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
)

var _ = Describe("Wrapper", func() {
	Describe("Creating an etcd wrapper", func() {
		It("should return the wrapper", func() {
			cl := &clientv3.Client{}
			e, err := etcd.NewEtcdWrapper(cl, &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheCleanUpTime})
			Expect(err).NotTo(HaveOccurred())
			Expect(e).NotTo(BeNil())
		})

		Context("with a nil client", func() {
			It("returns an error", func() {
				_, err := etcd.NewEtcdWrapper(nil, &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheCleanUpTime})
				Expect(err).To(Equal(srerr.NoClientProvided))
			})
		})
	})
})
