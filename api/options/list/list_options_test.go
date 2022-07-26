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

package list_test

import (
	"math"
	"syscall"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List Options", func() {
	var (
		addr       = "10.10.10.10"
		port int32 = 80
		_, _       = addr, port
		opts *list.Options
	)
	BeforeEach(func() {
		opts = &list.Options{}
	})

	Context("name filters", func() {
		It("applies name prefix", func() {
			prefix := "prod-"
			err := list.WithNamePrefix(prefix)(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = list.WithNamePrefix("")(opts)
			Expect(err).To(Equal(srerr.InvalidNamePrefixFilter))

			err = list.WithNamePrefix(prefix)(opts)
			Expect(opts).To(Equal(&list.Options{
				NameFilters: &list.NameFilters{
					Prefix: prefix,
				},
			}))

			err = list.WithNamePrefix(prefix)(&list.Options{
				NameFilters: &list.NameFilters{
					In: []string{"one"},
				},
			})
			Expect(err).To(Equal(srerr.IncompatibleNameFilters))
		})
		It("applies name in filters", func() {
			names := []string{"name-1", "name-2"}
			err := list.WithNameIn(names...)(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = list.WithNameIn()(opts)
			Expect(err).To(Equal(srerr.EmptyNameInFilter))

			err = list.WithNameIn(names...)(opts)
			Expect(opts).To(Equal(&list.Options{
				NameFilters: &list.NameFilters{
					In: names,
				},
			}))

			err = list.WithNameIn(names...)(&list.Options{
				NameFilters: &list.NameFilters{
					Prefix: "prod-",
				},
			})
			Expect(err).To(Equal(srerr.IncompatibleNameFilters))
		})
	})

	Context("metadata filters", func() {
		It("applies metadata filters", func() {
			By("using keys", func() {
				err := list.WithKs("whatever")(nil)
				Expect(err).To(Equal(srerr.NoOptionsProvided))

				err = list.WithKs()(opts)
				Expect(err).To(Equal(srerr.EmptyMetadataKeysFilter))

				err = list.WithKs("key-1", "")(opts)
				Expect(err).To(Equal(srerr.EmptyMetadataKey))

				err = list.WithKs("key-1", "key-2", "key-3")(opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts).To(Equal(&list.Options{
					MetadataFilters: &list.MetadataFilters{
						Metadata: map[string]string{
							"key-1": "",
							"key-2": "",
							"key-3": "",
						},
					},
				}))
			})
			By("using key-values", func() {
				err := list.WithKV("key-1", "edited-value-1")(nil)
				Expect(err).To(Equal(srerr.NoOptionsProvided))

				err = list.WithKV("key-1", "edited-value-1")(opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts).To(Equal(&list.Options{
					MetadataFilters: &list.MetadataFilters{
						Metadata: map[string]string{
							"key-1": "edited-value-1",
							"key-2": "",
							"key-3": "",
						},
					},
				}))

				err = list.WithKV("new-key", "new-val")(opts)
				Expect(opts).To(Equal(&list.Options{
					MetadataFilters: &list.MetadataFilters{
						Metadata: map[string]string{
							"key-1":   "edited-value-1",
							"key-2":   "",
							"key-3":   "",
							"new-key": "new-val",
						},
					},
				}))
			})
			By("checking incompatible filters", func() {
				err := list.WithKV("new-key", "new-val")(&list.Options{
					MetadataFilters: &list.MetadataFilters{
						NoMetadata: true,
					},
				})
				Expect(err).To(Equal(srerr.IncompatibleMetadataFilters))

				err = list.WithNoMetadata()(&list.Options{
					MetadataFilters: &list.MetadataFilters{
						Metadata: map[string]string{"key-1": "val-1"},
					},
				})
				Expect(err).To(Equal(srerr.IncompatibleMetadataFilters))
			})
		})
	})
	It("applies the correct results number", func() {
		var results int32 = 22
		err := list.WithResultsNumber(-1)(nil)
		Expect(err).To(Equal(srerr.NoOptionsProvided))

		err = list.WithResultsNumber(-1)(opts)
		Expect(err).To(Equal(srerr.InvalidResultsNumber))

		err = list.WithResultsNumber(results)(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(&list.Options{
			Results: results,
		}))
	})
	Context("address filters", func() {
		It("applies the correct CIDR", func() {
			err := list.WithCIDR("")(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = list.WithCIDR("10.10.10.10/33")(opts)
			Expect(err).To(Equal(srerr.InvalidCIDRProvided))

			err = list.WithCIDR("10.10.10.0/24")(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&list.Options{
				AddressFilters: &list.AddressFilters{
					CIDR: "10.10.10.0/24",
				},
			}))

			err = list.WithCIDR("10.10.10.0/24")(&list.Options{
				AddressFilters: &list.AddressFilters{
					AddressFamily: list.IPv4AddressFamily,
				},
			})
			Expect(err).To(Equal(srerr.IncompatibleAddressFilters))
		})
		It("applies the correct address family", func() {
			err := list.WithIPv4Only()(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = list.WithIPv6Only()(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			err = list.WithIPv4Only()(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&list.Options{
				AddressFilters: &list.AddressFilters{
					AddressFamily: list.IPv4AddressFamily,
				},
			}))

			err = list.WithIPv6Only()(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).To(Equal(&list.Options{
				AddressFilters: &list.AddressFilters{
					AddressFamily: list.IPv6AddressFamily,
				},
			}))

			err = list.WithIPv4Only()(&list.Options{
				AddressFilters: &list.AddressFilters{
					CIDR: "10.10.10.0/24",
				},
			})
			Expect(err).To(Equal(srerr.IncompatibleAddressFilters))

			err = list.WithIPv6Only()(&list.Options{
				AddressFilters: &list.AddressFilters{
					CIDR: "10.10.10.0/24",
				},
			})
			Expect(err).To(Equal(srerr.IncompatibleAddressFilters))

			err = list.WithNoMetadata()(nil)
			Expect(err).To(Equal(srerr.NoOptionsProvided))

			noMetaOpts := &list.Options{}
			err = list.WithNoMetadata()(noMetaOpts)
			Expect(err).NotTo(HaveOccurred())
			Expect(noMetaOpts).To(Equal(&list.Options{
				MetadataFilters: &list.MetadataFilters{
					Metadata:   map[string]string{},
					NoMetadata: true,
				},
			}))
		})
	})

	It("applies port filters", func() {
		err := list.WithPortIn()(nil)
		Expect(err).To(Equal(srerr.NoOptionsProvided))

		err = list.WithPortIn()(opts)
		Expect(err).To(Equal(srerr.NoPortsProvided))

		err = list.WithPortIn(-1)(opts)
		Expect(err).To(MatchError(srerr.InvalidPort))

		err = list.WithPortIn(math.MaxUint16 + 1)(opts)
		Expect(err).To(MatchError(srerr.InvalidPort))

		err = list.WithPortIn(80, 8080)(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(&list.Options{
			PortFilters: &list.PortFilters{
				In: []int32{80, 8080},
			},
		}))

		err = list.WithPortIn(80, 90)(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.PortFilters.In).To(ConsistOf([]int32{80, 90, 8080}))

		err = list.WithPortRange(0, 0)(nil)
		Expect(err).To(Equal(srerr.NoOptionsProvided))

		err = list.WithPortRange(-1, 0)(&list.Options{})
		Expect(err).To(MatchError(srerr.InvalidPortRange))

		err = list.WithPortRange(2, math.MaxUint16+1)(opts)
		Expect(err).To(MatchError(srerr.InvalidPortRange))

		err = list.WithPortRange(2, 1)(opts)
		Expect(err).To(MatchError(srerr.InvalidPortRange))

		err = list.WithPortRange(80, 90)(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.PortFilters.Range).To(Equal([][2]int32{{80, 90}}))

		err = list.WithPortRange(80, 90)(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.PortFilters.Range).To(Equal([][2]int32{{80, 90}}))

		err = list.WithPortRange(8080, 8090)(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.PortFilters.Range).To(ConsistOf([][2]int32{{80, 90}, {8080, 8090}}))
	})
})

var _ = Describe("Filter", func() {
	var opts *list.Options
	BeforeEach(func() {
		opts = &list.Options{}
	})

	Context("with nil", func() {
		It("should return an error", func() {
			passed, err := opts.Filter(nil)
			Expect(passed).To(BeFalse())
			Expect(err).To(Equal(srerr.InvalidObjectToFilter))
		})
	})

	Context("with no filters", func() {
		It("should return true", func() {
			passed, err := opts.Filter(&coretypes.Namespace{})
			Expect(passed).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("name filters", func() {
		Context("in filter", func() {
			It("filters resource correctly", func() {
				opts = &list.Options{
					NameFilters: &list.NameFilters{
						In: []string{
							"payroll",
							"config",
							"lab",
						},
					},
				}

				passed, err := opts.Filter(&coretypes.Namespace{
					Name: "staging",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Service{
					Name: "payroll",
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("prefix filter", func() {
			It("filters resource correctly", func() {
				opts = &list.Options{
					NameFilters: &list.NameFilters{
						Prefix: "should-",
					},
				}

				passed, err := opts.Filter(&coretypes.Namespace{
					Name: "no-pass",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Namespace{
					Name: "should-pass",
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Testing NoMetadata filter", func() {
		Context("with metadata", func() {
			It("should return false", func() {
				opts = &list.Options{
					MetadataFilters: &list.MetadataFilters{
						NoMetadata: true,
					},
				}
				passed, err := opts.Filter(&coretypes.Namespace{
					Metadata: map[string]string{
						"key": "whatever",
					},
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Namespace{
					Metadata: map[string]string{},
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Testing Metadata filter", func() {
		Context("with not passing", func() {
			It("should return false", func() {
				opts = &list.Options{
					MetadataFilters: &list.MetadataFilters{
						Metadata: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
							"key-3": "",
						},
					},
				}

				passed, err := opts.Filter(&coretypes.Namespace{
					Metadata: map[string]string{
						"key":   "whatever",
						"key-1": "val-1",
						"key-2": "",
					},
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with key not existing", func() {
			It("should return false", func() {
				opts = &list.Options{
					MetadataFilters: &list.MetadataFilters{
						Metadata: map[string]string{
							"key-1": "",
						},
					},
				}
				passed, err := opts.Filter(&coretypes.Namespace{
					Metadata: map[string]string{
						"key": "whatever",
					},
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with passing metadata", func() {
			It("should return true", func() {
				opts = &list.Options{
					MetadataFilters: &list.MetadataFilters{
						Metadata: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
							"key-3": "",
						},
					},
				}

				obj := &coretypes.Namespace{
					Metadata: map[string]string{
						"env":   "prod",
						"v":     "1.2.1",
						"key-1": "val-1",
						"key-2": "val-2",
						"key-3": "whatever",
					},
				}
				passed, err := opts.Filter(obj)
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Testing CIDR filter", func() {
		Context("IPv4", func() {
			It("should return false", func() {
				opts = &list.Options{
					AddressFilters: &list.AddressFilters{
						CIDR: "10.10.10.0/24",
					},
				}

				passed, err := opts.Filter(&coretypes.Endpoint{
					Address: "10.10.11.1",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Endpoint{
					Address: "not-a-valid-ip",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Endpoint{
					Address: "10.10.10.1",
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("with IPv6 address not in CIDR", func() {
			It("should return false", func() {
				opts = &list.Options{
					AddressFilters: &list.AddressFilters{
						CIDR: "2001:db8::/32",
					},
				}

				passed, err := opts.Filter(&coretypes.Endpoint{
					Address: "2001:db9::",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Endpoint{
					Address: "not-a-valid-ip",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Endpoint{
					Address: "2001:db8::1",
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Testing AddressFamiliy", func() {
		Context("IPv4", func() {
			It("should return false", func() {
				opts = &list.Options{
					AddressFilters: &list.AddressFilters{
						AddressFamily: syscall.AF_INET,
					},
				}

				passed, err := opts.Filter(&coretypes.Endpoint{
					Address: "2001:db8::1",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Endpoint{
					Address: "not-a-valid-ip",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Endpoint{
					Address: "10.10.10.10",
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("IPv6", func() {
			It("should return false", func() {
				opts = &list.Options{
					AddressFilters: &list.AddressFilters{
						AddressFamily: syscall.AF_INET6,
					},
				}

				passed, err := opts.Filter(&coretypes.Endpoint{
					Address: "10.10.10.10",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Endpoint{
					Address: "not-a-valid-ip",
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())

				passed, err = opts.Filter(&coretypes.Endpoint{
					Address: "2001:db8::1",
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})

	Describe("Testing PortIn", func() {
		var (
			portFilters = &list.Options{
				PortFilters: &list.PortFilters{
					In: []int32{80, 443, 8080},
				},
			}
		)
		Context("with port out of values", func() {
			It("should return false", func() {
				passed, err := portFilters.Filter(&coretypes.Endpoint{
					Port: 22,
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with port in values", func() {
			It("should return true", func() {
				passed, err := portFilters.Filter(&coretypes.Endpoint{
					Port: 443,
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Testing PortRange", func() {
		var (
			portRange = &list.Options{
				PortFilters: &list.PortFilters{
					Range: [][2]int32{
						{80, 90},
						{8080, 8088},
					},
				},
			}
		)

		Context("with port out of ranges", func() {
			It("should return false", func() {
				passed, err := portRange.Filter(&coretypes.Endpoint{
					Port: 91,
				})
				Expect(passed).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with port in values", func() {
			It("should return true", func() {
				passed, err := portRange.Filter(&coretypes.Endpoint{
					Port: 8080,
				})
				Expect(passed).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
