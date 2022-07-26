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

package types_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/CloudNativeSDWAN/serego/api/core/types"
)

var _ = Describe("Objects", func() {
	const (
		nsName   string = "ns"
		servName string = "serv"
		endpName string = "endp"
	)

	Describe("Testing namespaces", func() {
		Context("cloning a namespace", func() {
			It("returns an equal namespace", func() {
				ns := &types.Namespace{
					Name: nsName,
					Metadata: map[string]string{
						"key": "val",
					},
				}

				Expect(ns.Clone()).To(Equal(&types.Namespace{
					Name: nsName,
					Metadata: map[string]string{
						"key": "val",
					},
				}))
			})
		})
	})

	Describe("Testing services", func() {
		Context("cloning a service", func() {
			It("returns an equal service", func() {
				serv := &types.Service{
					Name:      servName,
					Namespace: nsName,
					Metadata: map[string]string{
						"key": "val",
					},
				}

				Expect(serv.Clone()).To(Equal(&types.Service{
					Name:      servName,
					Namespace: nsName,
					Metadata: map[string]string{
						"key": "val",
					},
				}))
			})
		})
	})

	Describe("Testing endpoints", func() {
		Context("cloning a endpoint", func() {
			It("returns an equal endpoint", func() {
				endp := &types.Endpoint{
					Name:      endpName,
					Service:   servName,
					Namespace: nsName,
					Address:   "10.10.10.10",
					Port:      8080,
					Metadata: map[string]string{
						"key": "val",
					},
				}

				Expect(endp.Clone()).To(Equal(&types.Endpoint{
					Name:      endpName,
					Service:   servName,
					Namespace: nsName,
					Address:   "10.10.10.10",
					Port:      8080,
					Metadata: map[string]string{
						"key": "val",
					},
				}))
			})
		})
	})
})
