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

package get_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
)

var _ = Describe("Get Options", func() {
	var opts *get.Options
	BeforeEach(func() {
		opts = &get.Options{
			ForceRefresh: false,
		}
	})

	It("applies correct options", func() {
		err := get.WithForceRefresh()(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(&get.Options{
			ForceRefresh: true,
		}))
	})

	Context("when providing nil options", func() {
		It("returns an error", func() {
			By("providing them from outside", func() {
				err := get.WithForceRefresh()(nil)
				Expect(err).To(Equal(srerr.NoOptionsProvided))
			})
		})
	})
})
