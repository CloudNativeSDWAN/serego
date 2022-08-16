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

package wrapper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
)

var _ = Describe("Wrapper Options", func() {
	var options *wrapper.Options
	BeforeEach(func() {
		options = &wrapper.Options{}
	})

	It("sets correct cache option time", func() {
		err := wrapper.WithCacheExpirationTime(time.Minute)(options)
		Expect(err).NotTo(HaveOccurred())
		Expect(options).To(Equal(&wrapper.Options{
			CacheExpirationTime: time.Minute,
		}))

		err = wrapper.WithCacheExpirationTime(0)(options)
		Expect(err).To(Equal(srerr.InvalidCacheExpirationTime))

		err = wrapper.WithCacheExpirationTime(-1)(options)
		Expect(err).To(Equal(srerr.InvalidCacheExpirationTime))

		err = wrapper.WithNoCache()(options)
		Expect(err).NotTo(HaveOccurred())
		Expect(options).To(Equal(&wrapper.Options{
			CacheExpirationTime: 0,
		}))
	})

	It("sets correct region option", func() {
		err := wrapper.WithRegion("my-region")(options)
		Expect(err).NotTo(HaveOccurred())
		Expect(options).To(Equal(&wrapper.Options{
			Region: "my-region",
		}))
	})

	It("sets correct project option", func() {
		err := wrapper.WithProjectID("my-project")(options)
		Expect(err).NotTo(HaveOccurred())
		Expect(options).To(Equal(&wrapper.Options{
			ProjectID: "my-project",
		}))
	})
})
