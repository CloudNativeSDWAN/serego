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

package etcd_test

import (
	"context"
	"fmt"
	"time"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/etcd"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	etcdns "go.etcd.io/etcd/client/v3/namespace"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Namespace Operations", func() {
	var (
		cli *clientv3.Client
		ns  *coretypes.Namespace
		e   *etcd.EtcdWrapper
	)

	BeforeEach(func() {
		cli = &clientv3.Client{}
		e, _ = etcd.NewEtcdWrapper(cli, &wrapper.Options{CacheExpirationTime: time.Minute})
		etcd.NewKV = etcdns.NewKV
		ns = namespaces[0]
	})

	Describe("Creating a namespace", func() {
		It("calls etcd with the correct parameters", func() {
			timesGet := 0
			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				Expect(prefix).To(Equal("/namespaces"))
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Put: func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
						Expect(key).To(Equal("/" + ns.Name))
						Expect(opts).To(BeEmpty())
						Expect(val).To(Equal(string(kvsNamespaces[0].Value)))
						return &clientv3.PutResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						timesGet++
						if timesGet > 1 {
							Fail("should get it from cache")
							return nil, nil
						}
						return &clientv3.GetResponse{
							Header: &etcdserverpb.ResponseHeader{
								Revision: 0,
							},
							Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
						}, nil
					},
				}
			}
			createdNs, err := e.Namespace(ns.Name).Create(ctx, ns.Metadata)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdNs).To(Equal(ns))

			createdNs, err = e.Namespace(ns.Name).Get(ctx, &get.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(createdNs).To(Equal(ns))
		})

		Context("without cache", func() {
			It("always calls etcd", func() {
				timesGet := 0
				etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
					Expect(prefix).To(Equal("/namespaces"))
					return &fakeKV{
						_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
							// Just to make this work
							return clientv3.OpResponse{}, nil
						},
						_Put: func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
							return &clientv3.PutResponse{}, nil
						},
						_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
							timesGet++
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						},
					}
				}
				e, _ = etcd.NewEtcdWrapper(cli, &wrapper.Options{})
				e.Namespace(ns.Name).Create(ctx, ns.Metadata)
				e.Namespace(ns.Name).Get(ctx, &get.Options{})
			})
		})

		Context("when errors happen", func() {
			It("returns the same error", func() {
				nsName := "whatever"
				expErr := fmt.Errorf("whatever")
				etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
					return &fakeKV{
						_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
							// Just to make this work
							return clientv3.OpResponse{}, nil
						},
						_Put: func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
							var n coretypes.Namespace
							yaml.Unmarshal([]byte(val), &n)
							Expect(n).To(Equal(coretypes.Namespace{
								Name:     nsName,
								Metadata: map[string]string{},
							}))
							return nil, expErr
						},
					}
				}
				createdNs, err := e.Namespace(nsName).Create(ctx, nil)
				Expect(err).To(Equal(expErr))
				Expect(createdNs).To(BeNil())
			})
		})
	})

	Describe("Retrieving a namespace", func() {
		Context("in case of errors", func() {
			It("returns the same error", func() {
				By("checking the get operation", func() {
					expErr := fmt.Errorf("whatever")
					etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
						return &fakeKV{
							_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
								// Just to make this work
								return clientv3.OpResponse{}, nil
							},
							_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
								return nil, expErr
							},
						}
					}
					n, err := e.Namespace(ns.Name).Get(ctx, &get.Options{})
					Expect(err).To(Equal(expErr))
					Expect(n).To(BeNil())
				})
				By("checking the unmarshal operation", func() {
					etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
						return &fakeKV{
							_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
								// Just to make this work
								return clientv3.OpResponse{}, nil
							},
							_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
								return &clientv3.GetResponse{
									Header: &etcdserverpb.ResponseHeader{
										Revision: 0,
									},
									Kvs: []*mvccpb.KeyValue{
										{
											Key:   []byte(ns.Name),
											Value: []byte("<"),
										},
									},
								}, nil
							},
						}
					}
					_, err := e.Namespace(ns.Name).Get(ctx, &get.Options{})
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("Deleting a namespace", func() {
		It("deletes the namespace and its children", func() {
			getCalled := false
			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Delete: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
						Expect(key).To(Equal("/" + ns.Name))
						Expect(opts).To(HaveLen(1))
						expOp, provOp := clientv3.Op{}, clientv3.Op{}
						clientv3.WithPrefix()(&expOp)
						opts[0](&provOp)
						Expect(provOp).To(Equal(expOp))
						return &clientv3.DeleteResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						getCalled = true
						return nil, fmt.Errorf("whatever")
					},
				}
			}
			err := e.Namespace(ns.Name).Delete(ctx)
			Expect(err).NotTo(HaveOccurred())

			e.Namespace(ns.Name).Get(ctx, &get.Options{})
			Expect(getCalled).To(BeTrue())
		})
		Context("in case of errors", func() {
			It("returns the same error in ", func() {
				expErr := fmt.Errorf("whatever")
				etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
					return &fakeKV{
						_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
							// Just to make this work
							return clientv3.OpResponse{}, nil
						},
						_Delete: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
							return nil, expErr
						},
					}
				}
				err := e.Namespace(ns.Name).Delete(ctx)
				Expect(err).To(Equal(expErr))
			})
		})
	})

	Describe("Listing namespaces", func() {
		It("returns all the namespaces", func() {
			timesCalled := 0
			limit := 3
			expOpts := clientv3.Op{}
			clientv3.WithFromKey()(&expOpts)
			clientv3.WithLimit(int64(limit))(&expOpts)

			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						defer func() {
							timesCalled++
						}()
						Expect(opts).To(HaveLen(2))
						provOpts := clientv3.Op{}
						for _, opt := range opts {
							opt(&provOpts)
						}
						Expect(provOpts).To(Equal(expOpts))

						switch timesCalled {
						case 0:
							Expect(key).To(Equal("/"))
							return &clientv3.GetResponse{
								Kvs: kvsNamespaces[0:3],
							}, nil
						case 1:
							Expect(key).To(Equal("/" + namespaces[2].Name))
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{
									kvsNamespaces[2],
									{
										Key: []byte("/ns-3/services/my-service"),
										Value: func() []byte {
											v, _ := yaml.Marshal(&coretypes.Service{
												Name:      "my-service",
												Namespace: "ns-3",
												Metadata: map[string]string{
													"whatever": "whatever",
												},
											})
											return v
										}(),
									},
									kvsNamespaces[3],
								},
							}, nil
						case 2:
							Expect(key).To(Equal("/" + namespaces[3].Name))
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{
									kvsNamespaces[3],
									{
										Key:   []byte("/invalid-ns"),
										Value: []byte("<invalid"),
									},
								},
							}, nil
						default:
							Fail("list called more than 3 times.")
							return nil, nil
						}
					},
				}
			}
			nit := e.Namespace("").List(&list.Options{
				Results: int32(limit),
			})

			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						Fail("should get it from cache")
						return nil, nil
					},
				}
			}

			for i := 0; i < len(namespaces); i++ {
				val, nop, err := nit.Next(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(namespaces[i]))
				Expect(nop).NotTo(BeNil())
				n, err := nop.Get(ctx, &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(namespaces[i]))
			}

			val, nop, err := nit.Next(ctx)
			Expect(err).To(MatchError(srerr.IteratorDone))
			Expect(val).To(BeNil())
			Expect(nop).To(BeNil())
		})

		Context("with name filters", func() {
			It("should only return requested namespaces", func() {
				timesCalled := 0
				expOpts := clientv3.Op{}
				clientv3.WithFromKey()(&expOpts)
				clientv3.WithLimit(int64(list.DefaultListResultsNumber))(&expOpts)

				etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
					return &fakeKV{
						_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
							// Just to make this work
							return clientv3.OpResponse{}, nil
						},
						_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
							defer func() {
								timesCalled++
							}()
							if timesCalled > 1 {
								Fail("list called more than 1 time.")
								return nil, nil
							}
							Expect(opts).To(HaveLen(2))
							provOpts := clientv3.Op{}
							for _, opt := range opts {
								opt(&provOpts)
							}
							Expect(provOpts).To(Equal(expOpts))

							return &clientv3.GetResponse{
								Kvs: kvsNamespaces,
							}, nil
						},
					}
				}
				nit := e.Namespace("ns-1").List(&list.Options{
					NameFilters: &list.NameFilters{
						In: []string{"ns-3", "ns-5"},
					},
				})

				expLoops := 2
				expResults := []*coretypes.Namespace{
					namespaces[0], namespaces[2],
				}
				for i := 0; i < expLoops; i++ {
					val, nop, err := nit.Next(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(expResults[i]))
					Expect(nop).NotTo(BeNil())
				}

				val, nop, err := nit.Next(ctx)
				Expect(val).To(BeNil())
				Expect(nop).To(BeNil())
				Expect(err).To(MatchError(srerr.IteratorDone))
			})
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
					return &fakeKV{
						_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
							// Just to make this work
							return clientv3.OpResponse{}, nil
						},
						_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
							return nil, expErr
						},
					}
				}
				nit := e.Namespace("ns-1").List(&list.Options{
					NameFilters: &list.NameFilters{
						In: []string{"ns-3", "ns-5"},
					},
				})

				val, nop, err := nit.Next(ctx)
				Expect(err).To(MatchError(expErr))
				Expect(val).To(BeNil())
				Expect(nop).To(BeNil())

				val, nop, err = nit.Next(ctx)
				Expect(val).To(BeNil())
				Expect(nop).To(BeNil())
				Expect(err).To(MatchError(srerr.IteratorDone))
			})
		})
	})
})
