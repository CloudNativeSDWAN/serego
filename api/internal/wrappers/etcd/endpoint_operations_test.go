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
	"path"
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
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	etcdns "go.etcd.io/etcd/client/v3/namespace"
	"gopkg.in/yaml.v3"
)

var _ = Describe("EndpointOperations", func() {
	var (
		cli  *clientv3.Client
		endp *coretypes.Endpoint
		e    *etcd.EtcdWrapper
	)

	BeforeEach(func() {
		cli = &clientv3.Client{}
		e, _ = etcd.NewEtcdWrapper(cli, &wrapper.Options{CacheExpirationTime: time.Minute})
		etcd.NewKV = etcdns.NewKV
		endp = endpoints[0]
	})

	Describe("Creating an endpoint", func() {
		It("calls etcd with the correct parameters", func() {
			timesOps := 0
			timesGet := 0
			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				defer func() {
					timesOps++
				}()
				switch timesOps {
				case 0:
					Expect(prefix).To(Equal("/namespaces"))
				case 1:
					Expect(prefix).To(Equal(path.Join("/", endp.Namespace, "services")))
				case 2:
					Expect(prefix).To(Equal(path.Join("/", endp.Service, "endpoints")))
				default:
					Fail("operations called more than 3 times.")
				}
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Put: func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
						Expect(key).To(Equal("/" + endp.Name))
						Expect(opts).To(BeEmpty())
						Expect(val).To(Equal(string(kvsEndpoints[0].Value)))
						return &clientv3.PutResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						switch key {
						case "/" + endp.Namespace:
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						case "/" + endp.Service:
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsServices[0]},
							}, nil
						}
						if key == "/"+endp.Name {
							timesGet++
							if timesGet > 1 {
								Fail("should get it from cache")
								return nil, nil
							}
						}

						return &clientv3.GetResponse{
							Header: &etcdserverpb.ResponseHeader{
								Revision: 0,
							},
							Kvs: []*mvccpb.KeyValue{kvsEndpoints[0]},
						}, nil
					},
				}
			}
			createdEndp, err := e.Namespace(endp.Namespace).
				Service(endp.Service).
				Endpoint(endp.Name).Create(ctx, endp.Address, endp.Port, endp.Metadata)

			Expect(err).NotTo(HaveOccurred())
			Expect(createdEndp).To(Equal(endp))

			timesOps = 0
			createdEndp, err = e.Namespace(endp.Namespace).
				Service(endp.Service).
				Endpoint(endp.Name).Get(ctx, &get.Options{})

			Expect(err).NotTo(HaveOccurred())
			Expect(createdEndp).To(Equal(endp))
		})

		Context("without cache", func() {
			It("calls etcd with the correct parameters", func() {
				called := false
				etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
					return &fakeKV{
						_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
							// Just to make this work
							return clientv3.OpResponse{}, nil
						},
						_Put: func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
							return &clientv3.PutResponse{}, nil
						},
						_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
							if key == "/"+endp.Name {
								called = true
							}

							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsEndpoints[0]},
							}, nil
						},
					}
				}

				e, _ = etcd.NewEtcdWrapper(cli, &wrapper.Options{})
				e.Namespace(endp.Namespace).Service(endp.Service).
					Endpoint(endp.Name).Create(ctx, endp.Address, endp.Port, endp.Metadata)
				e.Namespace(endp.Namespace).Service(endp.Service).
					Endpoint(endp.Name).Get(ctx, &get.Options{})
				Expect(called).To(BeTrue())
			})
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				f := &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						return nil, nil
					},
				}
				etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
					return f
				}
				By("checking if parents can be pulled", func() {
					f._Get = func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						if key == "/"+endp.Namespace {
							return nil, rpctypes.ErrGRPCKeyNotFound
						}

						return nil, fmt.Errorf("whatever")
					}
					f._Delete = func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
						Expect(key).To(Equal("/" + endp.Service))
						return &clientv3.DeleteResponse{}, nil
					}
					ep, err := e.Namespace(endp.Namespace).Service(endp.Service).
						Endpoint(endp.Name).Create(ctx, endp.Address, endp.Port, endp.Metadata)
					Expect(ep).To(BeNil())
					Expect(err).To(MatchError(srerr.NamespaceNotFound))

					f._Get = func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						if key == "/"+endp.Namespace {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						}

						if key == "/"+endp.Service {
							return nil, rpctypes.ErrGRPCKeyNotFound
						}

						return nil, fmt.Errorf("whatever")
					}
					ep, err = e.Namespace(endp.Namespace).Service(endp.Service).
						Endpoint(endp.Name).Create(ctx, endp.Address, endp.Port, endp.Metadata)
					Expect(ep).To(BeNil())
					Expect(err).To(MatchError(rpctypes.ErrGRPCKeyNotFound))
				})
				By("checking if etcd throws an error", func() {
					expErr := fmt.Errorf("whatever")
					etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
						return &fakeKV{
							_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
								// Just to make this work
								return clientv3.OpResponse{}, nil
							},
							_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
								return &clientv3.GetResponse{
									Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
								}, nil
							},
							_Put: func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
								return nil, expErr
							},
						}
					}
					ep, err := e.Namespace(endp.Namespace).Service(endp.Service).
						Endpoint(endp.Name).Create(ctx, endp.Address, endp.Port, nil)
					Expect(ep).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
		})
	})

	Describe("Retrieving an endpoint", func() {
		It("calls etcd with the right parameters", func() {
			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						if key == "/"+endp.Namespace {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						}
						if key == "/"+endp.Service {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsServices[0]},
							}, nil
						}
						Expect(key).To(Equal("/" + endp.Name))
						return &clientv3.GetResponse{
							Kvs: []*mvccpb.KeyValue{kvsEndpoints[0]},
						}, nil
					},
				}
			}

			ep, err := e.Namespace(endp.Namespace).
				Service(endp.Service).Endpoint(endp.Name).Get(ctx, &get.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(ep).To(Equal(endp))
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				By("checking that parents exist", func() {
					etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
						return &fakeKV{
							_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
								// Just to make this work
								return clientv3.OpResponse{}, nil
							},
							_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
								if key == "/"+endp.Namespace {
									return nil, rpctypes.ErrGRPCKeyNotFound
								}

								return nil, nil
							},
							_Delete: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
								return &clientv3.DeleteResponse{}, nil
							},
						}
					}
					ep, err := e.Namespace(endp.Namespace).
						Service(endp.Service).Endpoint(endp.Name).Get(ctx, &get.Options{})
					Expect(err).To(Equal(srerr.ServiceNotFound))
					Expect(ep).To(BeNil())
				})
				By("checking the get operation", func() {
					etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
						return &fakeKV{
							_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
								// Just to make this work
								return clientv3.OpResponse{}, nil
							},
							_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
								if key == "/"+endp.Namespace {
									return &clientv3.GetResponse{
										Header: &etcdserverpb.ResponseHeader{
											Revision: 0,
										},
										Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
									}, nil
								}
								if key == "/"+endp.Service {
									return &clientv3.GetResponse{
										Header: &etcdserverpb.ResponseHeader{
											Revision: 0,
										},
										Kvs: []*mvccpb.KeyValue{kvsServices[0]},
									}, nil
								}
								return nil, expErr
							},
						}
					}
					ep, err := e.Namespace(endp.Namespace).
						Service(endp.Service).Endpoint(endp.Name).Get(ctx, &get.Options{})
					Expect(err).To(Equal(expErr))
					Expect(ep).To(BeNil())
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
											Key:   []byte(endp.Name),
											Value: []byte("<"),
										},
									},
								}, nil
							},
						}
					}
					_, err := e.Namespace(endp.Namespace).
						Service(endp.Service).Endpoint(endp.Name).Get(ctx, &get.Options{})
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("Deleting a endpoint", func() {
		It("deletes the endpoint", func() {
			getCalled := false
			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Delete: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
						Expect(key).To(Equal("/" + endp.Name))
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
			err := e.Namespace(endp.Namespace).
				Service(endp.Service).Endpoint(endp.Name).Delete(ctx)
			Expect(err).NotTo(HaveOccurred())

			e.Namespace(endp.Namespace).
				Service(endp.Service).Endpoint(endp.Name).Get(ctx, &get.Options{})
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
				err := e.Namespace(endp.Namespace).
					Service(endp.Service).Endpoint(endp.Name).Delete(ctx)
				Expect(err).To(Equal(expErr))
			})
		})
	})

	Describe("Listing endpoints", func() {
		var (
			listKvEndpoints []*mvccpb.KeyValue
			listEndpoints   []*coretypes.Endpoint
		)

		BeforeEach(func() {
			listKvEndpoints = []*mvccpb.KeyValue{}
			listEndpoints = []*coretypes.Endpoint{}

			for i := 1; i < 5; i++ {
				s := &coretypes.Endpoint{
					Name:      fmt.Sprintf("endp-%d", i),
					Namespace: endp.Namespace,
					Service:   endp.Service,
					Metadata: map[string]string{
						fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
						"another-key":            "another-value",
					},
					Address: fmt.Sprintf("%d0.%d1.%d2.%d3", i, i, i, i),
					Port:    80 + int32(i),
				}
				listEndpoints = append(listEndpoints, s)
				listKvEndpoints = append(listKvEndpoints, &mvccpb.KeyValue{
					Key: []byte("/" + s.Name),
					Value: func() []byte {
						v, _ := yaml.Marshal(s)
						return v
					}(),
				})
				listEndpoints[i-1].OriginalObject = listKvEndpoints[i-1]
			}
		})

		It("returns all the endpoints", func() {
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
								Kvs: listKvEndpoints[0:3],
							}, nil
						case 1:
							Expect(key).To(Equal("/" + listEndpoints[2].Name))
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{
									listKvEndpoints[2],
									{
										Key:   []byte("/my-endpoints/whatever/else"),
										Value: []byte("whatever"),
									},
									listKvEndpoints[3],
								},
							}, nil
						case 2:
							Expect(key).To(Equal("/" + listEndpoints[3].Name))
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{
									listKvEndpoints[3],
									{
										Key:   []byte("/invalid-service"),
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
			eit := e.Namespace(endp.Namespace).Service(endp.Service).
				Endpoint("").List(&list.Options{
				Results: int32(limit),
			})

			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						if key == "/"+endp.Namespace {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						}
						if key == "/"+endp.Service {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsServices[0]},
							}, nil
						}
						switch key {
						case "/" + listEndpoints[0].Name:
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{listKvEndpoints[0]},
							}, nil
						case "/" + listEndpoints[1].Name:
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{listKvEndpoints[1]},
							}, nil
						case "/" + listEndpoints[2].Name:
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{listKvEndpoints[2]},
							}, nil
						case "/" + listEndpoints[3].Name:
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{listKvEndpoints[3]},
							}, nil
						default:
							Fail("call undefined endpoint")
							return nil, nil
						}
					},
				}
			}

			for i := 0; i < len(listEndpoints); i++ {
				val, nop, err := eit.Next(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(listEndpoints[i]))
				Expect(nop).NotTo(BeNil())
				n, err := nop.Get(ctx, &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(listEndpoints[i]))
			}

			val, nop, err := eit.Next(ctx)
			Expect(err).To(MatchError(srerr.IteratorDone))
			Expect(val).To(BeNil())
			Expect(nop).To(BeNil())
		})

		Context("with name filters", func() {
			It("should only return requested services", func() {
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
								Kvs: listKvEndpoints,
							}, nil
						},
					}
				}
				eit := e.Namespace(endp.Namespace).Service(endp.Service).
					Endpoint(endp.Name).List(&list.Options{
					NameFilters: &list.NameFilters{
						In: []string{"endp-3", "endp-5"},
					},
				})

				expLoops := 2
				expResults := []*coretypes.Endpoint{listEndpoints[0], listEndpoints[2]}
				for i := 0; i < expLoops; i++ {
					val, sop, err := eit.Next(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(expResults[i]))
					Expect(sop).NotTo(BeNil())
				}

				val, sop, err := eit.Next(ctx)
				Expect(val).To(BeNil())
				Expect(sop).To(BeNil())
				Expect(err).To(MatchError(srerr.IteratorDone))
			})
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				By("checking that a namespace name is provided", func() {
					eit := e.Namespace("").Service("").Endpoint("").
						List(&list.Options{})
					s, sop, err := eit.Next(ctx)
					Expect(s).To(BeNil())
					Expect(sop).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyNamespaceName))
				})
				By("checking that a namespace name is provided", func() {
					eit := e.Namespace(endp.Namespace).Service("").Endpoint("").
						List(&list.Options{})
					s, sop, err := eit.Next(ctx)
					Expect(s).To(BeNil())
					Expect(sop).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyServiceName))
				})
				By("checking the error from etcd", func() {
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
					eit := e.Namespace(endp.Namespace).Service(endp.Service).
						Endpoint(endp.Name).List(&list.Options{
						NameFilters: &list.NameFilters{
							In: []string{"serv-3", "serv-5"},
						},
					})

					val, sop, err := eit.Next(ctx)
					Expect(err).To(MatchError(expErr))
					Expect(val).To(BeNil())
					Expect(sop).To(BeNil())

					val, sop, err = eit.Next(ctx)
					Expect(val).To(BeNil())
					Expect(sop).To(BeNil())
					Expect(err).To(MatchError(srerr.IteratorDone))
				})
			})
		})
	})
})
