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

var _ = Describe("Service Operations", func() {
	const (
		nsName   = "ns"
		servName = "serv"
	)

	var (
		sr     *core.ServiceRegistry
		nsop   = &fake.NamespaceOperation{}
		fop    *fake.ServiceOperation
		ctx    = context.TODO()
		wrp, _ = fake.NewFakeWrapper()
		sop    *core.ServiceOperation
	)

	BeforeEach(func() {
		// fop is the fake internal service operation use to mock a real
		// service registry operation.
		fop = &fake.ServiceOperation{}
		nsop.Service_ = func(s string) ops.ServiceOperation {
			fop.Name_ = s
			return fop
		}
		wrp.Namespace_ = func(name string) ops.NamespaceOperation {
			nsop.Name_ = name
			return nsop
		}
		sr, _ = core.NewServiceRegistryFromWrapper(wrp)
		sop = sr.Namespace(nsName).Service(servName)
	})

	Describe("Getting a service", func() {
		Context("in case of user errors", func() {
			It("should return an error", func() {
				By("checking if its parent namespace name is provided")
				res, err := sr.Namespace("").Service(servName).Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(srerr.EmptyNamespaceName))

				By("checking if the service name is provided")
				res, err = sr.Namespace(nsName).Service("").Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(srerr.EmptyServiceName))
			})
		})

		Context("in case of errors from the service registry", func() {
			It("should return exactly the same error", func() {
				expectedError := fmt.Errorf("expected error")
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
					return nil, expectedError
				}

				res, err := sop.Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(Equal(expectedError))
			})
		})

		It("should return the correct service", func() {
			metadata := map[string]string{"key-1": "val-1", "key-2": "val2"}
			fop.Get_ = func(_ context.Context, g *get.Options) (*coretypes.Service, error) {
				Expect(fop.Name_).To(Equal(servName))
				Expect(g).NotTo(BeNil())
				return &coretypes.Service{
					Name:      servName,
					Namespace: nsName,
					Metadata:  metadata,
				}, nil
			}

			By("querying the service registry")
			res, err := sop.Get(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(&coretypes.Service{
				Name:      servName,
				Namespace: nsName,
				Metadata:  metadata,
			}))

			By("... and ignoring the cache if user so desires")
			fop.Get_ = func(_ context.Context, g *get.Options) (*coretypes.Service, error) {
				Expect(g).To(Equal(&get.Options{ForceRefresh: true}))
				return &coretypes.Service{
					Name:      servName,
					Namespace: nsName,
					Metadata:  metadata,
				}, nil
			}
			res, err = sop.Get(ctx, get.WithForceRefresh())
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(&coretypes.Service{
				Name:      servName,
				Namespace: nsName,
				Metadata:  metadata,
			}))
		})
	})

	Describe("Registering a service", func() {
		Context("in case of user errors", func() {
			It("should return an error", func() {
				By("checking if namespace name is provided")
				Expect(sr.Namespace("").Service(servName).Register(ctx)).
					To(MatchError(srerr.EmptyNamespaceName))

				By("checking if service name is provided")
				Expect(sr.Namespace(nsName).Service("").Register(ctx)).
					To(MatchError(srerr.EmptyServiceName))

				By("checking if options are correct", func() {
					err := sr.Namespace(nsName).Service(servName).
						Register(ctx, register.WithKV("", "invalid"))
					Expect(err).To(MatchError(srerr.EmptyMetadataKey))
				})
			})
		})

		Context("in case of errors from the service registry", func() {
			Context("and it happens while we check if it exists", func() {
				It("wraps the error", func() {
					permD := fmt.Errorf("get permissions denied")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
						return nil, permD
					}
					Expect(sop.Register(ctx)).To(MatchError(permD))
				})
			})

			It("returns exactly the same error", func() {
				By("reforwarding from Create", func() {
					permD := fmt.Errorf("create permissions denied")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
						return nil, srerr.ServiceNotFound
					}
					fop.Create_ = func(_ context.Context, _ map[string]string) (*coretypes.Service, error) {
						return nil, permD
					}
					Expect(sop.Register(ctx)).To(Equal(permD))
				})

				By("... or reforwarding from Update", func() {
					permD := fmt.Errorf("update permissions denied")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
						return &coretypes.Service{}, nil
					}
					fop.Update_ = func(_ context.Context, _ map[string]string) (*coretypes.Service, error) {
						return nil, permD
					}
					Expect(sop.Register(ctx)).To(Equal(permD))
				})
			})
		})

		Context("when creating a service", func() {
			It("registers the service", func() {
				var providedMap map[string]string
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
					return nil, srerr.ServiceNotFound
				}
				fop.Create_ = func(_ context.Context, m map[string]string) (*coretypes.Service, error) {
					providedMap = m
					return nil, nil
				}

				By("just registering its name", func() {
					Expect(sop.Register(ctx)).NotTo(HaveOccurred())
					Expect(providedMap).To(And(BeEmpty(), Not(BeNil())))
				})

				By("providing metadata from options", func() {
					metadata := map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
						"key-3": "",
					}
					Expect(
						sop.Register(
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
					Expect(sop.Register(ctx,
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
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
						return &coretypes.Service{}, nil
					}
					Expect(sop.Register(ctx, register.WithCreateMode())).
						To(Equal(srerr.ServiceAlreadyExists))
				})
			})
		})

		Context("when updating a service", func() {
			It("does so successfully", func() {
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
					return &coretypes.Service{
						Name:      servName,
						Namespace: nsName,
						Metadata: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
							"key-3": "",
						},
					}, nil
				}
				metadata := map[string]string{}
				fop.Update_ = func(_ context.Context, m map[string]string) (*coretypes.Service, error) {
					metadata = m
					return nil, nil
				}

				By("adding new metadata", func() {
					err := sop.Register(ctx, register.WithKV("key-4", "val-4"))
					Expect(err).NotTo(HaveOccurred())
					Expect(metadata).To(Equal(map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
						"key-3": "",
						"key-4": "val-4",
					}))
				})

				By("updating a metadata key-value", func() {
					err := sop.Register(
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
					err := sop.Register(
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

			Context("but the service does not exist", func() {
				It("returns an error", func() {
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
						return nil, srerr.ServiceNotFound
					}
					timesCalled := 0
					fop.Update_ = func(_ context.Context, _ map[string]string) (*coretypes.Service, error) {
						timesCalled++
						return nil, fmt.Errorf("another error")
					}

					Expect(sop.Register(ctx, register.WithUpdateMode())).
						To(Equal(srerr.ServiceNotFound))
					Expect(timesCalled).To(BeZero())
				})
			})

			Context("but the new data is the same as the current one", func() {
				It("does nothing silently", func() {
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Service, error) {
						return &coretypes.Service{
							Name:      servName,
							Namespace: nsName,
							Metadata: map[string]string{
								"key-1": "val-1",
								"key-2": "val=-2",
							},
						}, nil
					}
					timesCalled := 0
					fop.Update_ = func(_ context.Context, _ map[string]string) (*coretypes.Service, error) {
						timesCalled++
						return nil, fmt.Errorf("another error")
					}
					sop.Register(ctx)
					Expect(timesCalled).To(BeZero())
				})
			})
		})
	})

	Describe("Deregistering a service", func() {
		Context("in case of user errors", func() {
			It("returns an error", func() {
				By("checking the name of the namespace")
				Expect(sr.Namespace("").Service(servName).Deregister(ctx)).
					To(MatchError(srerr.EmptyNamespaceName))

				By("checking the name of the service")
				Expect(sr.Namespace(nsName).Service("").Deregister(ctx)).
					To(MatchError(srerr.EmptyServiceName))
			})
		})

		Context("in case of service registry errors", func() {
			Context("and the error is != not found", func() {
				It("returns exactly the same error", func() {
					expErr := fmt.Errorf("error")
					fop.Delete_ = func(_ context.Context) error {
						return expErr
					}

					err := sop.Deregister(ctx)
					Expect(err).To(Equal(expErr))
				})
			})
			Context("when the error is not found", func() {
				Context("and the default behavior is in place", func() {
					It("does not return any errors", func() {
						fop.Delete_ = func(_ context.Context) error {
							return srerr.ServiceNotFound
						}
						Expect(sop.Deregister(ctx)).NotTo(HaveOccurred())
					})
				})
				Context("but user still wants to know about that", func() {
					It("does return the error", func() {
						fop.Delete_ = func(_ context.Context) error {
							return srerr.ServiceNotFound
						}
						err := sop.
							Deregister(ctx, deregister.WithFailIfNotExists())
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
				sop.Deregister(ctx)
				Expect(called).To(BeTrue())
			})
		})
	})

	Describe("Listing services", func() {
		Context("in case of user errors", func() {
			It("returns the same error", func() {
				By("checking the name of the namespace")
				lister := sr.Namespace("").Service("").
					List()

				for i := 0; i < 2; i++ {
					serv, next, err := lister.Next(ctx)
					Expect(next).To(BeNil())
					Expect(serv).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyNamespaceName))
				}

				By("checking the provided options")
				lister = sop.List(list.WithNameIn())

				for i := 0; i < 2; i++ {
					serv, next, err := lister.Next(ctx)
					Expect(next).To(BeNil())
					Expect(serv).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyNameInFilter))
				}
			})
		})

		Context("in case of service registry errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				fsiter := &fake.FakeServiceIterator{

					Next_: func(_ context.Context) (*coretypes.Service, ops.ServiceOperation, error) {
						return nil, nil, expErr
					},
				}
				fop.List_ = func(_ *list.Options) ops.ServiceLister {
					return fsiter
				}
				lister := sr.Namespace(nsName).Service("").List()
				serv, resOp, err := lister.Next(ctx)
				Expect(resOp).To(BeNil())
				Expect(serv).To(BeNil())
				Expect(err).To(Equal(expErr))
			})
		})

		It("returns the next element", func() {
			elems := []*coretypes.Service{}
			elemOps := []*fake.ServiceOperation{}
			for i := 0; i < 2; i++ {
				elems = append(elems, &coretypes.Service{
					Name:      fmt.Sprintf("ns-%d", i),
					Namespace: nsName,
					Metadata: map[string]string{
						fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
					},
				})
				elemOps = append(elemOps, &fake.ServiceOperation{
					Name_: elems[i].Name,
				})
			}

			i := 0
			fsiter := &fake.FakeServiceIterator{
				Next_: func(_ context.Context) (*coretypes.Service, ops.ServiceOperation, error) {
					if i < 2 {
						return elems[i], elemOps[i], nil
					}

					return nil, nil, srerr.IteratorDone
				},
			}
			fop.List_ = func(_ *list.Options) ops.ServiceLister {
				return fsiter
			}
			l := sr.Namespace(servName).Service("").List()

			for i < 2 {
				serv, resOp, err := l.Next(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(resOp).NotTo(BeNil())
				Expect(serv).To(Equal(elems[i]))

				var called bool
				elemOps[i].Delete_ = func(_ context.Context) error {
					called = true
					return nil
				}
				resOp.Deregister(ctx)

				Expect(called).To(BeTrue())
				i++
			}

			serv, resOp, err := l.Next(ctx)
			Expect(resOp).To(BeNil())
			Expect(serv).To(BeNil())
			Expect(srerr.IsIteratorDone(err)).To(BeTrue())
		})
	})

	Context("starting a direct operation", func() {
		It("is not allowed", func() {
			servop := &core.ServiceOperation{}

			By("checking Get operations", func() {
				res, err := servop.Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.EmptyServiceName))
			})

			By("checking Register operations", func() {
				Expect(servop.Register(ctx)).To(Equal(srerr.EmptyServiceName))
			})

			By("checking Register operations", func() {
				Expect(servop.Deregister(ctx)).To(Equal(srerr.EmptyServiceName))
			})

			By("checking List operations", func() {
				_, res, err := servop.List().Next(context.TODO())
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.UninitializedOperation))
			})

			By("initializing an iterator", func() {
				_, res, err := (&core.ServicesIterator{}).Next(context.TODO())
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.UninitializedOperation))
			})
		})
	})
})
