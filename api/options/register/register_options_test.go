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

package register_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
)

var _ = Describe("Register Options", func() {
	var (
		addr       = "10.10.10.10"
		port int32 = 80
		opts *register.Options
	)
	BeforeEach(func() {
		opts = &register.Options{}
	})

	It("applies the correct mode", func() {
		By("setting create mode", func() {
			err := register.WithCreateMode()(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = register.WithCreateMode()(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&register.Options{
				RegisterMode: register.CreateMode,
			}))
		})
		By("setting update mode", func() {
			err := register.WithUpdateMode()(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = register.WithUpdateMode()(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&register.Options{
				RegisterMode: register.UpdateMode,
			}))
		})
	})

	It("pushes metadata correctly", func() {
		By("using WithReplaceMetadata", func() {
			err := register.WithReplaceMetadata()(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = register.WithReplaceMetadata()(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&register.Options{
				ReplaceMetadata: true,
			}))
		})
		By("using WithMetadataKeyValue", func() {
			err := register.WithMetadataKeyValue("key", "value")(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = register.WithMetadataKeyValue("key", "value")(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&register.Options{
				ReplaceMetadata: true,
				Metadata:        map[string]string{"key": "value"},
			}))
		})
		By("using WithKV", func() {
			err := register.WithKV("key-1", "val-1")(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = register.WithKV("key-1", "val-1")(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&register.Options{
				ReplaceMetadata: true,
				Metadata: map[string]string{
					"key":   "value",
					"key-1": "val-1",
				},
			}))

			err = register.WithKV("", "val-1")(opts)
			Expect(err).To(Equal(srerr.EmptyMetadataKey))
		})
		By("using WithMetadata", func() {
			err := register.WithMetadata(map[string]string{
				"key-1":   "edited-val-1",
				"new-key": "new-val",
			})(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = register.WithMetadata(map[string]string{
				"key-1":   "edited-val-1",
				"new-key": "new-val",
			})(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&register.Options{
				ReplaceMetadata: true,
				Metadata: map[string]string{
					"key":     "value",
					"key-1":   "edited-val-1",
					"new-key": "new-val",
				},
			}))
		})
	})

	It("sets the correct address", func() {
		err := register.WithAddress("")(nil)
		Expect(err).To(Equal(srerr.NoOptionsProvided))

		err = register.WithAddress("10.10.10")(opts)
		Expect(err).To(Equal(srerr.InvalidAddress))

		err = register.WithAddress(addr)(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(&register.Options{
			Address: &addr,
		}))
	})

	It("sets the correct port", func() {
		err := register.WithPort(0)(nil)
		Expect(err).To(Equal(srerr.NoOptionsProvided))

		err = register.WithPort(-1)(opts)
		Expect(err).To(Equal(srerr.InvalidPort))

		err = register.WithPort(math.MaxUint16 + 1)(opts)
		Expect(err).To(Equal(srerr.InvalidPort))

		err = register.WithPort(port)(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(&register.Options{
			Port: &port,
		}))
	})

	It("sets generate name correctly", func() {
		err := register.WithGenerateName()(nil)
		Expect(err).To(Equal(srerr.NoOptionsProvided))

		err = register.WithGenerateName()(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(&register.Options{
			GenerateName: true,
		}))
	})
})
