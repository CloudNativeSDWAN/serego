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
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	cm "github.com/CloudNativeSDWAN/serego/api/internal/wrappers/aws/cloudmap"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Wrapper", func() {
	Describe("Creating Cloud Map wrapper", func() {
		Context("with a nil client", func() {
			It("should return an error", func() {
				_, err := cm.NewCloudMapWrapper(nil, &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime})
				Expect(err).To(Equal(srerr.NoClientProvided))
			})
		})

		It("should return the wrapper", func() {
			_, err := cm.NewCloudMapWrapper(&servicediscovery.Client{}, &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
