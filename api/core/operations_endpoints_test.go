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

package core_test

import (
	"context"
	"fmt"
	"math"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/CloudNativeSDWAN/serego/api/core"
	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/fake"
	"github.com/CloudNativeSDWAN/serego/api/options/deregister"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
)

var _ = Describe("Endpoint operations", func() {
	const (
		nsName   = "ns"
		servName = "serv"
		endpName = "endp"
	)

	var (
		sr     *core.ServiceRegistry
		nsop   = &fake.NamespaceOperation{}
		servop = &fake.ServiceOperation{}
		fop    *fake.EndpointOperation
		ctx    = context.TODO()
		wrp, _ = fake.NewFakeWrapper()
		eop    *core.EndpointOperation
	)

	BeforeEach(func() {
		// fop is the fake internal endpoint operation, used to mock a real
		// service registry operation.
		fop = &fake.EndpointOperation{}
		servop.Endpoint_ = func(name string) ops.EndpointOperation {
			fop.Name_ = name
			return fop
		}
		nsop.Service_ = func(s string) ops.ServiceOperation {
			return servop
		}
		wrp.Namespace_ = func(name string) ops.NamespaceOperation {
			return nsop
		}
		sr, _ = core.NewServiceRegistryFromWrapper(wrp)
		eop = sr.Namespace(nsName).Service(servName).Endpoint(endpName)
	})

	Describe("Getting an endpoint", func() {
		Context("in case of user errors", func() {
			It("should return an error", func() {
				By("checking if its parent namespace name is provided")
				res, err := sr.Namespace("").
					Service(servName).
					Endpoint(endpName).
					Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(srerr.EmptyNamespaceName))

				By("checking if the service name is provided")
				res, err = sr.Namespace(nsName).
					Service("").
					Endpoint(endpName).
					Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(srerr.EmptyServiceName))

				By("checking if the endpoint name is provided")
				res, err = sr.Namespace(nsName).
					Service(servName).
					Endpoint("").
					Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(srerr.EmptyEndpointName))
			})
		})

		Context("in case of errors from the service registry", func() {
			It("should return exactly the same error", func() {
				expectedError := fmt.Errorf("expected error")
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
					return nil, expectedError
				}

				res, err := eop.Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(Equal(expectedError))
			})
		})

		It("should return the correct endpoint", func() {
			metadata := map[string]string{"key-1": "val-1", "key-2": "val2"}
			fop.Get_ = func(_ context.Context, g *get.Options) (*coretypes.Endpoint, error) {
				Expect(fop.Name_).To(Equal(endpName))
				Expect(g).NotTo(BeNil())
				return &coretypes.Endpoint{
					Name:      endpName,
					Namespace: nsName,
					Service:   servName,
					Metadata:  metadata,
				}, nil
			}

			By("querying the service registry")
			res, err := sr.Namespace(nsName).
				Service(servName).
				Endpoint(endpName).
				Get(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(&coretypes.Endpoint{
				Name:      endpName,
				Namespace: nsName,
				Service:   servName,
				Metadata:  metadata,
			}))

			By("... and ignoring the cache if user so desires")
			fop.Get_ = func(_ context.Context, g *get.Options) (*coretypes.Endpoint, error) {
				Expect(g).To(Equal(&get.Options{ForceRefresh: true}))
				return &coretypes.Endpoint{
					Name:      endpName,
					Namespace: nsName,
					Service:   servName,
					Metadata:  metadata,
				}, nil
			}
			res, err = sr.Namespace(nsName).
				Service(servName).
				Endpoint(endpName).
				Get(ctx, get.WithForceRefresh())
			Expect(err).NotTo(HaveOccurred())

			Expect(res).To(Equal(&coretypes.Endpoint{
				Name:      endpName,
				Namespace: nsName,
				Service:   servName,
				Metadata:  metadata,
			}))
		})
	})

	Describe("Registering an endpoint", func() {
		Context("in case of user errors", func() {
			It("should return an error", func() {
				By("checking if namespace name is provided")
				Expect(
					sr.Namespace("").
						Service(servName).
						Endpoint(endpName).
						Register(ctx),
				).To(Equal(srerr.EmptyNamespaceName))

				By("checking if service name is provided")
				Expect(
					sr.Namespace(nsName).
						Service("").
						Endpoint(endpName).
						Register(ctx),
				).To(Equal(srerr.EmptyServiceName))

				By("checking if endpoint name is provided and generateName is false")
				Expect(
					sr.Namespace(nsName).
						Service(servName).
						Endpoint("").
						Register(ctx),
				).To(Equal(srerr.EmptyEndpointName))

				By("checking the provided address is valid")
				Expect(
					sr.Namespace(nsName).
						Service(servName).
						Endpoint("").
						Register(ctx,
							register.WithGenerateName(),
							register.WithAddress("10.10"),
						),
				).To(Equal(srerr.InvalidAddress))

				By("checking the provided port is valid")
				Expect(eop.Register(ctx, register.WithPort(-1))).
					To(Equal(srerr.InvalidPort))

				Expect(eop.Register(ctx, register.WithPort(math.MaxUint16+1))).
					To(Equal(srerr.InvalidPort))

				// We don't check MatchErrors because that is tested on
				// options_test.
				By("checking if other options are correct")
				Expect(eop.Register(ctx, register.WithKV("", ""))).
					To(And(HaveOccurred(), Not(MatchError(srerr.EmptyEndpointName))))
			})
		})

		Context("in case of errors from the service registry", func() {
			Context("and it happens while we check if it exists", func() {
				It("wraps the error", func() {
					permD := fmt.Errorf("get permissions denied")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
						return nil, permD
					}
					Expect(eop.Register(ctx)).To(MatchError(permD))
				})
			})

			It("returns exactly the same error", func() {
				By("reforwarding from Create", func() {
					permD := fmt.Errorf("create permissions denied")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
						return nil, srerr.EndpointNotFound
					}
					fop.Create_ = func(_ context.Context, address string, port int32, m map[string]string) (*coretypes.Endpoint, error) {
						return nil, permD
					}
					Expect(eop.Register(ctx)).To(Equal(permD))
				})

				By("... or reforwarding from Update", func() {
					permD := fmt.Errorf("update permissions denied")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
						return &coretypes.Endpoint{}, nil
					}
					fop.Update_ = func(_ context.Context, _ string, _ int32, _ map[string]string) (*coretypes.Endpoint, error) {
						return nil, permD
					}
					Expect(eop.Register(ctx)).To(Equal(permD))
				})
			})
		})

		Context("when creating an endpoint", func() {
			const (
				expAddress string = "10.10.10.10"
				expPort    int32  = 80
			)
			It("registers the endpoint", func() {
				var (
					providedMap     map[string]string
					providedAddress string
					providedPort    int32
				)
				fop.Get_ = func(c context.Context, g *get.Options) (*coretypes.Endpoint, error) {
					return nil, srerr.EndpointNotFound
				}
				fop.Create_ = func(_ context.Context, address string, port int32, m map[string]string) (*coretypes.Endpoint, error) {
					providedMap = m
					providedAddress = address
					providedPort = port
					return nil, nil
				}

				By("just registering its name", func() {
					Expect(eop.Register(ctx)).NotTo(HaveOccurred())
					Expect(providedMap).To(And(BeEmpty(), Not(BeNil())))
					Expect(providedAddress).To(BeEmpty())
					Expect(providedPort).To(BeZero())
				})

				By("just registering a generated name", func() {
					Expect(sr.Namespace(nsName).Service(servName).
						Endpoint("").Register(ctx, register.WithGenerateName())).
						NotTo(HaveOccurred())
					Expect(fop.Name_).NotTo(BeEmpty())
					if !strings.HasPrefix(fop.Name_, servName) {
						Fail(
							fmt.Sprintf(
								"generated endpoint name does not start with its service name: %s",
								fop.Name_,
							),
						)
					}
					Expect(providedMap).To(And(BeEmpty(), Not(BeNil())))
					Expect(providedAddress).To(BeEmpty())
					Expect(providedPort).To(BeZero())
				})

				By("providing address and ports from options", func() {
					addr := expAddress
					port := expPort
					Expect(eop.Register(ctx,
						register.WithAddress(addr),
						register.WithPort(port),
					),
					).NotTo(HaveOccurred())
					Expect(providedMap).To(And(BeEmpty(), Not(BeNil())))
					Expect(providedAddress).To(Equal(expAddress))
					Expect(providedPort).To(Equal(expPort))
				})

				By("providing address and ports from helpers", func() {
					Expect(
						eop.Register(
							ctx,
							register.WithAddress(expAddress),
							register.WithPort(expPort),
						),
					).NotTo(HaveOccurred())
					Expect(providedMap).To(And(BeEmpty(), Not(BeNil())))
					Expect(providedAddress).To(Equal(expAddress))
					Expect(providedPort).To(Equal(expPort))
				})

				By("providing metadata from options", func() {
					metadata := map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
						"key-3": "",
					}
					Expect(
						eop.Register(
							ctx,
							register.WithMetadata(metadata),
						),
					).NotTo(HaveOccurred())
					Expect(providedMap).To(Equal(metadata))
				})

				By("... or providing metadata from helpers", func() {
					expMetadata := map[string]string{
						"key-1": "val-1",
						"key-2": "val-2-overridden",
						"key-3": "val-3",
						"key-4": "val-4-overridden",
						"key-5": "",
						"key-6": "",
					}
					Expect(eop.Register(ctx,
						register.WithKV("key-1", "val-1"),
						register.WithKV("key-2", "val-2"),
						register.WithMetadata(map[string]string{
							"key-2": "val-2-overridden",
							"key-3": "val-3",
							"key-4": "val-4",
							"key-5": "",
						}),
						register.WithMetadataKeyValue("key-4", "val-4-overridden"),
						register.WithMetadataKeyValue("key-6", ""),
					)).NotTo(HaveOccurred())
					Expect(providedMap).To(Equal(expMetadata))
				})
			})

			Context("and the service already exists", func() {
				It("stops the operation", func() {
					By("checking the registration mode and the error")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
						return &coretypes.Endpoint{}, nil
					}
					Expect(eop.Register(ctx, register.WithCreateMode())).
						To(Equal(srerr.EndpointAlreadyExists))
				})
			})
		})

		Context("when updating an endpoint", func() {
			Context("but there is no name and GenerateName is provided", func() {
				It("should return an error", func() {
					Expect(
						sr.Namespace(nsName).
							Service(servName).
							Endpoint("").
							Register(
								ctx,
								register.WithGenerateName(),
								register.WithUpdateMode(),
							),
					).To(Equal(srerr.EmptyEndpointName))
				})
			})

			It("does so successfully", func() {
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
					return &coretypes.Endpoint{
						Name:      nsName,
						Namespace: nsName,
						Service:   servName,
						Address:   "10.10.10.10",
						Port:      8080,
						Metadata: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
							"key-3": "",
						},
					}, nil
				}
				metadata := map[string]string{}
				address := ""
				port := int32(0)
				fop.Update_ = func(_ context.Context, a string, p int32, m map[string]string) (*coretypes.Endpoint, error) {
					metadata = m
					address = a
					port = p
					return nil, nil
				}

				By("adding new metadata, address and port", func() {
					err := eop.Register(ctx,
						register.WithKV("key-4", "val-4"),
						register.WithAddress("10.10.10.11"),
						register.WithPort(8081),
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(metadata).To(Equal(map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
						"key-3": "",
						"key-4": "val-4",
					}))
					Expect(address).To(Equal("10.10.10.11"))
					Expect(port).To(Equal(int32(8081)))
				})

				By("updating a metadata key-value", func() {
					err := eop.Register(
						ctx,
						register.WithKV("key-4", "val-4-overridden"),
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(metadata).To(Equal(map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
						"key-3": "",
						"key-4": "val-4-overridden",
					}))
				})

				By("... or by replacing all metadata", func() {
					err := eop.Register(
						ctx,
						register.WithKV("key-0", "val-0"),
						register.WithReplaceMetadata(),
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(metadata).To(Equal(map[string]string{
						"key-0": "val-0",
					}))
				})
			})

			It("resets everything", func() {
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
					return &coretypes.Endpoint{
						Name:      nsName,
						Address:   "10.10.10.10",
						Port:      8080,
						Namespace: nsName,
						Service:   servName,
						Metadata: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
							"key-3": "",
						},
					}, nil
				}

				metadata := map[string]string{"key": "to-be-reset"}
				address := "to-be-reset"
				port := int32(9000)
				fop.Update_ = func(_ context.Context, a string, p int32, m map[string]string) (*coretypes.Endpoint, error) {
					metadata = m
					address = a
					port = p
					return nil, nil
				}

				By("providing options", func() {
					err := eop.Register(ctx,
						register.WithMetadata(map[string]string{}),
						register.WithAddress(""),
						register.WithPort(0),
						register.WithReplaceMetadata(),
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(metadata).To(And(BeEmpty(), Not(BeNil())))
					Expect(address).To(BeEmpty())
					Expect(port).To(BeZero())
				})
				By("providing helper options", func() {
					metadata = map[string]string{"key": "to-be-reset"}
					address = "to-be-reset"
					port = int32(9000)
					err := eop.Register(ctx,
						register.WithAddress(""),
						register.WithPort(0),
						register.WithMetadata(map[string]string{}),
						register.WithReplaceMetadata(),
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(metadata).To(And(BeEmpty(), Not(BeNil())))
					Expect(address).To(BeEmpty())
					Expect(port).To(BeZero())
				})
			})

			Context("but the endpoint does not exist", func() {
				It("returns an error", func() {
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
						return nil, srerr.EndpointNotFound
					}
					timesCalled := 0
					fop.Update_ = func(_ context.Context, _ string, _ int32, _ map[string]string) (*coretypes.Endpoint, error) {
						timesCalled++
						return nil, fmt.Errorf("another error")
					}

					Expect(eop.Register(ctx, register.WithUpdateMode())).
						To(Equal(srerr.EndpointNotFound))
					Expect(timesCalled).To(BeZero())
				})
			})

			Context("but the new data is the same as the current one", func() {
				It("does nothing silently", func() {
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Endpoint, error) {
						return &coretypes.Endpoint{
							Name:      endpName,
							Namespace: nsName,
							Service:   servName,
							Address:   "10.10.10.10",
							Port:      8080,
							Metadata: map[string]string{
								"key-1": "val-1",
								"key-2": "val=-2",
							},
						}, nil
					}
					timesCalled := 0
					fop.Update_ = func(_ context.Context, _ string, _ int32, _ map[string]string) (*coretypes.Endpoint, error) {
						timesCalled++
						return nil, fmt.Errorf("another error")
					}
					eop.Register(ctx)
					Expect(timesCalled).To(BeZero())
				})
			})
		})
	})

	Describe("Deregistering an endpoint", func() {
		Context("in case of user errors", func() {
			It("returns an error", func() {
				By("checking the name of the namespace")
				Expect(sr.Namespace("").Service(servName).
					Endpoint(endpName).Deregister(ctx)).
					To(MatchError(srerr.EmptyNamespaceName))

				By("checking the name of the service")
				Expect(sr.Namespace(nsName).Service("").
					Endpoint(endpName).Deregister(ctx)).
					To(MatchError(srerr.EmptyServiceName))

				By("checking the name of the endpoint")
				Expect(sr.Namespace(nsName).Service(servName).
					Endpoint("").Deregister(ctx)).
					To(MatchError(srerr.EmptyEndpointName))
			})
		})

		Context("in case of service registry errors", func() {
			Context("and the error is != not found", func() {
				It("returns exactly the same error", func() {
					expErr := fmt.Errorf("error")
					fop.Delete_ = func(_ context.Context) error {
						return expErr
					}

					err := eop.Deregister(ctx)
					Expect(err).To(Equal(expErr))
				})
			})
			Context("when the error is not found", func() {
				Context("and the default behavior is in place", func() {
					It("does not return any errors", func() {
						fop.Delete_ = func(_ context.Context) error {
							return srerr.EndpointNotFound
						}
						Expect(eop.Deregister(ctx)).NotTo(HaveOccurred())
					})
				})
				Context("but user still wants to know about that", func() {
					It("does return the error", func() {
						fop.Delete_ = func(_ context.Context) error {
							return srerr.EndpointNotFound
						}
						err := eop.Deregister(ctx, deregister.WithFailIfNotExists())
						Expect(srerr.IsNotFound(err)).
							To(BeTrue())
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
		Context("when no errors are there", func() {
			It("is successful", func() {
				called := false
				fop.Delete_ = func(_ context.Context) error {
					called = true
					return nil
				}
				eop.Deregister(ctx)
				Expect(called).To(BeTrue())
			})
		})
	})

	Describe("Listing endpoints", func() {
		const (
			expAddress string = "10.10.10.10"
			expPort    int32  = 80
		)
		Context("in case of user errors", func() {
			It("returns the same error", func() {
				By("checking the name of the namespace")
				lister := sr.Namespace("").
					Service(servName).
					Endpoint(endpName).
					List()

				for i := 0; i < 2; i++ {
					endp, next, err := lister.Next(ctx)
					Expect(next).To(BeNil())
					Expect(endp).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyNamespaceName))
				}

				By("checking the name of the namespace")
				lister = sr.Namespace("").
					Service("").
					Endpoint(endpName).
					List()

				for i := 0; i < 2; i++ {
					endp, next, err := lister.Next(ctx)
					Expect(next).To(BeNil())
					Expect(endp).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyServiceName))
				}

				By("checking the provided options")
				lister = sr.Namespace(nsName).
					Service(servName).
					Endpoint("").
					List(list.WithNameIn())

				for i := 0; i < 2; i++ {
					endp, next, err := lister.Next(ctx)
					Expect(next).To(BeNil())
					Expect(endp).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyNameInFilter))
				}
			})
		})

		Context("in case of service registry errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				fsiter := &fake.FakeEndpointIterator{
					Next_: func(_ context.Context) (*coretypes.Endpoint, ops.EndpointOperation, error) {
						return nil, nil, expErr
					},
				}
				fop.List_ = func(_ *list.Options) ops.EndpointLister {
					return fsiter
				}
				lister := eop.List()
				endp, resOp, err := lister.Next(ctx)
				Expect(resOp).To(BeNil())
				Expect(endp).To(BeNil())
				Expect(err).To(Equal(expErr))
			})
		})

		It("returns the next element", func() {
			elems := []*coretypes.Endpoint{}
			elemOps := []*fake.EndpointOperation{}
			for i := 0; i < 2; i++ {
				elems = append(elems, &coretypes.Endpoint{
					Name:      fmt.Sprintf("ns-%d", i),
					Namespace: nsName,
					Service:   servName,
					Address:   expAddress,
					Port:      expPort,
					Metadata: map[string]string{
						fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
					},
				})
				elemOps = append(elemOps, &fake.EndpointOperation{
					Name_: elems[i].Name,
				})
			}

			i := 0
			fsiter := &fake.FakeEndpointIterator{
				Next_: func(_ context.Context) (*coretypes.Endpoint, ops.EndpointOperation, error) {
					if i < 2 {
						return elems[i], elemOps[i], nil
					}

					return nil, nil, srerr.IteratorDone
				},
			}
			fop.List_ = func(_ *list.Options) ops.EndpointLister {
				return fsiter
			}
			l := eop.List()

			for i < 2 {
				endp, resOp, err := l.Next(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(resOp).NotTo(BeNil())
				Expect(endp).To(Equal(elems[i]))

				var called bool
				elemOps[i].Delete_ = func(_ context.Context) error {
					called = true
					return nil
				}
				resOp.Deregister(ctx)

				Expect(called).To(BeTrue())
				i++
			}

			res, resOp, err := l.Next(ctx)
			Expect(resOp).To(BeNil())
			Expect(res).To(BeNil())
			Expect(srerr.IsIteratorDone(err)).To(BeTrue())
		})
	})

	Context("starting a direct operation", func() {
		It("is not allowed", func() {
			eop := &core.EndpointOperation{}

			By("checking Get operations", func() {
				res, err := eop.Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.EmptyEndpointName))
			})

			By("checking Register operations", func() {
				Expect(eop.Register(ctx)).To(Equal(srerr.UninitializedOperation))
			})

			By("checking Register operations", func() {
				Expect(eop.Deregister(ctx)).To(Equal(srerr.EmptyEndpointName))
			})

			By("checking List operations", func() {
				_, res, err := eop.List().Next(context.TODO())
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.UninitializedOperation))
			})

			By("initializing an iterator", func() {
				_, res, err := (&core.EndpointsIterator{}).Next(context.TODO())
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.UninitializedOperation))
			})
		})
	})
})
