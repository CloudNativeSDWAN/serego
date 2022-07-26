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

package servicedirectory_test

import (
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/gcp/servicedirectory"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sd "cloud.google.com/go/servicedirectory/apiv1"
)

var _ = Describe("Wrapper", func() {
	dumbCl := &sd.RegistrationClient{}

	Describe("Creating Service Directory wrapper", func() {
		Context("with a nil client", func() {
			It("should return an error", func() {
				var cl *sd.RegistrationClient
				_, err := servicedirectory.NewServiceDirectoryWrapper(cl, &wrapper.Options{})
				Expect(err).To(Equal(srerr.NoClientProvided))
			})
		})

		Context("with no project id", func() {
			It("should returns a error", func() {
				_, err := servicedirectory.NewServiceDirectoryWrapper(dumbCl, &wrapper.Options{})
				Expect(err).To(Equal(srerr.NoProjectIDSet))
			})
		})

		Context("with no default region", func() {
			It("should return an error", func() {
				_, err := servicedirectory.NewServiceDirectoryWrapper(dumbCl, &wrapper.Options{ProjectID: "project-id"})
				Expect(err).To(Equal(srerr.NoLocationSet))
			})
		})

		Context("with all parameters", func() {
			It("should return the wrapper", func() {
				_, err := servicedirectory.NewServiceDirectoryWrapper(dumbCl, &wrapper.Options{ProjectID: "project-id", Region: "us-west-2"})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
