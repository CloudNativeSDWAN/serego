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

	"github.com/CloudNativeSDWAN/serego/api/core"
	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/fake"
	"github.com/CloudNativeSDWAN/serego/api/options/deregister"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespace Operations", func() {
	const (
		nsName = "ns"
	)

	var (
		sr     *core.ServiceRegistry
		fop    *fake.NamespaceOperation
		ctx    = context.TODO()
		wrp, _ = fake.NewFakeWrapper()
		nsop   *core.NamespaceOperation
	)

	BeforeEach(func() {
		// fop is the fake namespace operation, the one we created to mock an
		// actual internal operation.
		fop = &fake.NamespaceOperation{}
		wrp.Namespace_ = func(name string) ops.NamespaceOperation {
			fop.Name_ = name
			return fop
		}
		sr, _ = core.NewServiceRegistryFromWrapper(wrp)
		nsop = sr.Namespace(nsName)
	})

	Describe("Getting a namespace", func() {
		Context("in case of user errors", func() {
			It("should return an error", func() {
				By("checking if the name is provided")
				res, err := sr.Namespace("").Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(srerr.EmptyNamespaceName))
			})
		})

		Context("in case of errors from the service registry", func() {
			It("should return exactly the same error", func() {
				expectedError := fmt.Errorf("expected error")
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Namespace, error) {
					return nil, expectedError
				}

				res, err := nsop.Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(Equal(expectedError))
			})
		})

		It("returns the correct namespace", func() {
			metadata := map[string]string{"key-1": "val-1", "key-2": "val2"}
			fop.Get_ = func(_ context.Context, g *get.Options) (*coretypes.Namespace, error) {
				Expect(fop.Name_).To(Equal(nsName))
				Expect(g).NotTo(BeNil())
				return &coretypes.Namespace{
					Name:     nsName,
					Metadata: metadata,
				}, nil
			}

			By("querying the service registry")
			res, err := nsop.Get(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(&coretypes.Namespace{
				Name:     nsName,
				Metadata: metadata,
			}))

			By("... and ignoring the cache if user so desires")
			fop.Get_ = func(_ context.Context, g *get.Options) (*coretypes.Namespace, error) {
				Expect(g).To(Equal(&get.Options{ForceRefresh: true}))
				return &coretypes.Namespace{
					Name:     nsName,
					Metadata: metadata,
				}, nil
			}
			res, err = nsop.Get(ctx, get.WithForceRefresh())
		})
	})

	Describe("Registering a namespace", func() {
		Context("in case of user errors", func() {
			It("should return an error", func() {
				By("checking if namespace name is provided")
				Expect(sr.Namespace("").Register(ctx)).
					To(MatchError(srerr.EmptyNamespaceName))

				By("checking if options are correct", func() {
					err := nsop.
						Register(ctx, register.WithKV("", "invalid"))
					Expect(err).To(MatchError(srerr.EmptyMetadataKey))
				})
			})
		})

		Context("in case of errors from the service registry", func() {
			Context("and it happens while we check if it exists", func() {
				It("wraps the error", func() {
					permD := fmt.Errorf("get permissions denied")
					fop.Get_ = func(c context.Context, g *get.Options) (*coretypes.Namespace, error) {
						return nil, permD
					}
					Expect(nsop.Register(ctx)).To(MatchError(permD))
				})
			})

			It("returns exactly the same error", func() {
				By("reforwarding from Create", func() {
					permD := fmt.Errorf("create permissions denied")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Namespace, error) {
						return nil, srerr.NamespaceNotFound
					}
					fop.Create_ = func(_ context.Context, _ map[string]string) (*coretypes.Namespace, error) {
						return nil, permD
					}
					Expect(nsop.Register(ctx)).To(Equal(permD))
				})

				By("... or reforwarding from Update", func() {
					permD := fmt.Errorf("update permissions denied")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Namespace, error) {
						return &coretypes.Namespace{}, nil
					}
					fop.Update_ = func(_ context.Context, _ map[string]string) (*coretypes.Namespace, error) {
						return nil, permD
					}
					Expect(nsop.Register(ctx)).To(Equal(permD))
				})
			})
		})

		Context("when creating a namespace", func() {
			It("registers the namespace", func() {
				var providedMap map[string]string
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Namespace, error) {
					return nil, srerr.NamespaceNotFound
				}
				fop.Create_ = func(_ context.Context, m map[string]string) (*coretypes.Namespace, error) {
					providedMap = m
					return nil, nil
				}

				By("just registering its name", func() {
					Expect(nsop.Register(ctx)).NotTo(HaveOccurred())
					Expect(providedMap).To(And(BeEmpty(), Not(BeNil())))
				})

				By("providing metadata from options", func() {
					metadata := map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
						"key-3": "",
					}
					Expect(nsop.
						Register(ctx, register.WithMetadata(metadata)),
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
					Expect(nsop.
						Register(ctx,
							register.WithKV("key-1", "val-1"),
							register.WithKV("key-2", "val-2"),
							register.WithMetadata(map[string]string{
								"key-2": "val-2-overridden",
								"key-3": "val-3",
								"key-4": "val-4",
								"key-5": "",
							}),
							register.WithMetadataKeyValue(
								"key-4", "val-4-overridden",
							),
							register.WithMetadataKeyValue("key-6", ""),
						)).
						NotTo(HaveOccurred())
					Expect(providedMap).To(Equal(expMetadata))
				})
			})

			Context("and the namespace already exists", func() {
				It("stops the operation", func() {
					By("checking the registration mode and the error")
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Namespace, error) {
						return &coretypes.Namespace{}, nil
					}
					Expect(nsop.Register(ctx, register.WithCreateMode())).
						To(Equal(srerr.NamespaceAlreadyExists))
				})
			})
		})

		Context("when updating a namespace", func() {
			It("does so successfully", func() {
				fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Namespace, error) {
					return &coretypes.Namespace{
						Name: nsName,
						Metadata: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
							"key-3": "",
						},
					}, nil
				}
				metadata := map[string]string{}
				fop.Update_ = func(_ context.Context, m map[string]string) (*coretypes.Namespace, error) {
					metadata = m
					return nil, nil
				}

				By("adding new metadata", func() {
					err := nsop.Register(
						ctx,
						register.WithKV("key-4", "val-4"),
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(metadata).To(Equal(map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
						"key-3": "",
						"key-4": "val-4",
					}))
				})

				By("updating a metadata key-value", func() {
					err := nsop.Register(
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
					err := nsop.Register(
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

			Context("but the namespace does not exist", func() {
				It("returns an error", func() {
					fop.Get_ = func(_ context.Context, _ *get.Options) (*coretypes.Namespace, error) {
						return nil, srerr.NamespaceNotFound
					}
					timesCalled := 0
					fop.Update_ = func(_ context.Context, _ map[string]string) (*coretypes.Namespace, error) {
						timesCalled++
						return nil, fmt.Errorf("another error")
					}

					Expect(nsop.Register(ctx, register.WithUpdateMode())).
						To(Equal(srerr.NamespaceNotFound))
					Expect(timesCalled).To(BeZero())
				})
			})

			Context("but the new data is the same as the current one", func() {
				It("does nothing silently", func() {
					fop.Get_ = func(c context.Context, _ *get.Options) (*coretypes.Namespace, error) {
						return &coretypes.Namespace{
							Name: nsName,
							Metadata: map[string]string{
								"key-1": "val-1",
								"key-2": "val=-2",
							},
						}, nil
					}
					timesCalled := 0
					fop.Update_ = func(_ context.Context, _ map[string]string) (*coretypes.Namespace, error) {
						timesCalled++
						return nil, fmt.Errorf("another error")
					}
					nsop.Register(ctx)
					Expect(timesCalled).To(BeZero())
				})
			})
		})
	})

	Describe("Deregistering a namespace", func() {
		Context("in case of user errors", func() {
			It("returns an error", func() {
				By("checking the name of the namespace")
				Expect(sr.Namespace("").Deregister(ctx)).
					To(MatchError(srerr.EmptyNamespaceName))
			})
		})

		Context("in case of service registry errors", func() {
			Context("and the error is != not found", func() {
				It("returns exactly the same error", func() {
					expErr := fmt.Errorf("error")
					fop.Delete_ = func(_ context.Context) error {
						return expErr
					}

					err := nsop.Deregister(ctx)
					Expect(err).To(Equal(expErr))
				})
			})
			Context("when the error is not found", func() {
				Context("and the default behavior is in place", func() {
					It("does not return any errors", func() {
						fop.Delete_ = func(_ context.Context) error {
							return srerr.NamespaceNotFound
						}
						Expect(nsop.Deregister(ctx)).NotTo(HaveOccurred())
					})
				})
				Context("but user still wants to know about that", func() {
					It("does return the error", func() {
						fop.Delete_ = func(_ context.Context) error {
							return srerr.NamespaceNotFound
						}
						err := nsop.Deregister(ctx, deregister.WithFailIfNotExists())
						Expect(srerr.IsNotFound(err)).To(BeTrue())
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
				nsop.Deregister(ctx)
				Expect(called).To(BeTrue())
			})
		})
	})

	Describe("Listing namespaces", func() {
		Context("in case of user errors", func() {
			It("returns the same error", func() {
				lister := sr.Namespace("").List(list.WithNameIn())

				for i := 0; i < 2; i++ {
					ns, next, err := lister.Next(ctx)
					Expect(ns).To(BeNil())
					Expect(next).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyNameInFilter))
				}
			})
		})

		Context("in case of service registry errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				fniter := &fake.FakeNamespaceIterator{

					Next_: func(_ context.Context) (*coretypes.Namespace, ops.NamespaceOperation, error) {
						return nil, nil, expErr
					},
				}
				fop.List_ = func(_ *list.Options) ops.NamespaceLister {
					return fniter
				}
				lister := sr.Namespace("").List()
				ns, op, err := lister.Next(ctx)
				Expect(ns).To(BeNil())
				Expect(op).To(BeNil())
				Expect(err).To(Equal(expErr))
			})
		})

		It("returns the next element", func() {
			elems := []*coretypes.Namespace{}
			elemOps := []*fake.NamespaceOperation{}
			for i := 0; i < 2; i++ {
				elems = append(elems, &coretypes.Namespace{
					Name: fmt.Sprintf("ns-%d", i),
					Metadata: map[string]string{
						fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
					},
				})
				elemOps = append(elemOps, &fake.NamespaceOperation{
					Name_: elems[i].Name,
				})
			}

			i := 0
			fniter := &fake.FakeNamespaceIterator{
				Next_: func(_ context.Context) (*coretypes.Namespace, ops.NamespaceOperation, error) {
					if i < 2 {
						return elems[i], elemOps[i], nil
					}

					return nil, nil, srerr.IteratorDone
				},
			}
			fop.List_ = func(_ *list.Options) ops.NamespaceLister {
				return fniter
			}
			l := sr.Namespace("").List()

			for i < 2 {
				ns, resOp, err := l.Next(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(resOp).NotTo(BeNil())
				Expect(ns).To(Equal(elems[i]))

				var called bool
				elemOps[i].Delete_ = func(_ context.Context) error {
					called = true
					return nil
				}
				resOp.Deregister(ctx)

				Expect(called).To(BeTrue())
				i++
			}

			ns, resOp, err := l.Next(ctx)
			Expect(resOp).To(BeNil())
			Expect(ns).To(BeNil())
			Expect(srerr.IsIteratorDone(err)).To(BeTrue())
		})
	})

	Context("starting a direct operation", func() {
		It("is not allowed", func() {
			nsop := &core.NamespaceOperation{}

			By("checking Get operations", func() {
				res, err := nsop.Get(ctx)
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.EmptyNamespaceName))
			})

			By("checking Register operations", func() {
				Expect(nsop.Register(ctx)).To(Equal(srerr.EmptyNamespaceName))
			})

			By("checking Register operations", func() {
				Expect(nsop.Deregister(ctx)).To(Equal(srerr.EmptyNamespaceName))
			})

			By("checking List operations", func() {
				_, res, err := nsop.List().Next(context.TODO())
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.UninitializedOperation))
			})

			By("initializing an iterator", func() {
				_, res, err := (&core.NamespacesIterator{}).Next(context.TODO())
				Expect(res).To(BeNil())
				Expect(err).To(Equal(srerr.UninitializedOperation))
			})
		})
	})
})
