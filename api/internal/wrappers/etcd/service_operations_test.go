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

var _ = Describe("Service Operations", func() {
	var (
		cli  *clientv3.Client
		serv *coretypes.Service
		e    *etcd.EtcdWrapper
		ctx  context.Context
	)

	BeforeEach(func() {
		cli = &clientv3.Client{}
		e, _ = etcd.NewEtcdWrapper(cli, &wrapper.Options{CacheExpirationTime: time.Minute})
		etcd.NewKV = etcdns.NewKV
		serv = services[0]
	})

	Describe("Creating a service", func() {
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
					Expect(prefix).To(Equal(path.Join("/", serv.Namespace, "services")))
				default:
					Fail("operations called more than 2 times.")
				}

				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Put: func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
						Expect(key).To(Equal("/" + serv.Name))
						Expect(opts).To(BeEmpty())
						Expect(val).To(Equal(string(kvsServices[0].Value)))
						return &clientv3.PutResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						if key == "/"+serv.Name {
							timesGet++
							if timesGet > 1 {
								Fail("should get it from cache")
								return nil, nil
							}
						}
						if key == "/"+serv.Namespace {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						}
						return &clientv3.GetResponse{
							Header: &etcdserverpb.ResponseHeader{
								Revision: 0,
							},
							Kvs: []*mvccpb.KeyValue{kvsServices[0]},
						}, nil
					},
				}
			}
			createdServ, err := e.Namespace(serv.Namespace).
				Service(serv.Name).Create(ctx, serv.Metadata)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdServ).To(Equal(serv))

			timesOps = 0
			createdServ, err = e.Namespace(serv.Namespace).
				Service(serv.Name).Get(ctx, &get.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(createdServ).To(Equal(serv))
		})

		Context("without cache", func() {
			It("always calls etcd", func() {
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
							if key == "/"+serv.Name {
								called = true
							}
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsServices[0]},
							}, nil
						},
					}
				}

				e, _ = etcd.NewEtcdWrapper(cli, &wrapper.Options{})
				e.Namespace(serv.Namespace).
					Service(serv.Name).Create(ctx, serv.Metadata)
				e.Namespace(serv.Namespace).
					Service(serv.Name).Get(ctx, &get.Options{})
				Expect(called).To(BeTrue())
			})
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				By("checking if parent namespace can be pulled", func() {
					expErr := rpctypes.ErrGRPCKeyNotFound
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
					serv, err := e.Namespace(serv.Namespace).Service(serv.Name).Create(ctx, nil)
					Expect(serv).To(BeNil())
					Expect(err).To(MatchError(expErr))
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
					serv, err := e.Namespace(serv.Namespace).Service(serv.Name).Create(ctx, nil)
					Expect(serv).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
		})
	})

	Describe("Retrieving a service", func() {
		It("calls etcd with the right parameters", func() {
			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						if key == "/"+serv.Namespace {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						}
						Expect(key).To(Equal("/" + serv.Name))
						return &clientv3.GetResponse{
							Kvs: []*mvccpb.KeyValue{kvsServices[0]},
						}, nil
					},
				}
			}

			serv, err := e.Namespace(serv.Namespace).Service(serv.Name).Get(ctx, &get.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(serv).To(Equal(serv))
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				By("checking if the parent exists", func() {
					etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
						return &fakeKV{
							_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
								// Just to make this work
								return clientv3.OpResponse{}, nil
							},
							_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
								if key == "/"+serv.Namespace {
									return nil, rpctypes.ErrGRPCKeyNotFound
								}
								return nil, nil
							},
							_Delete: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
								Expect(key).To(Equal("/" + serv.Name))
								return &clientv3.DeleteResponse{}, nil
							},
						}
					}
					s, err := e.Namespace(serv.Namespace).Service(serv.Name).Get(ctx, &get.Options{})
					Expect(err).To(Equal(srerr.NamespaceNotFound))
					Expect(s).To(BeNil())
				})
				By("checking the get operation", func() {
					etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
						return &fakeKV{
							_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
								// Just to make this work
								return clientv3.OpResponse{}, nil
							},
							_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
								if key == "/"+serv.Namespace {
									return &clientv3.GetResponse{
										Header: &etcdserverpb.ResponseHeader{
											Revision: 0,
										},
										Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
									}, nil
								}
								return nil, expErr
							},
						}
					}
					s, err := e.Namespace(serv.Namespace).Service(serv.Name).Get(ctx, &get.Options{})
					Expect(err).To(Equal(expErr))
					Expect(s).To(BeNil())
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
											Key:   []byte(serv.Name),
											Value: []byte("<"),
										},
									},
								}, nil
							},
						}
					}
					_, err := e.Namespace(serv.Namespace).Service(serv.Name).Get(ctx, &get.Options{})
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("Deleting a service", func() {
		It("deletes the service and its children", func() {
			getCalled := false
			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Delete: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
						Expect(key).To(Equal("/" + serv.Name))
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
			err := e.Namespace(serv.Namespace).Service(serv.Name).Delete(ctx)
			Expect(err).NotTo(HaveOccurred())

			e.Namespace(serv.Namespace).Service(serv.Name).Get(ctx, &get.Options{})
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
				err := e.Namespace(serv.Namespace).Service(serv.Name).Delete(ctx)
				Expect(err).To(Equal(expErr))
			})
		})
	})

	Describe("Listing services", func() {
		var (
			listKvServices []*mvccpb.KeyValue
			listServices   []*coretypes.Service
		)

		BeforeEach(func() {
			listKvServices = []*mvccpb.KeyValue{}
			listServices = []*coretypes.Service{}

			for i := 1; i < 5; i++ {
				s := &coretypes.Service{
					Name:      fmt.Sprintf("serv-%d", i),
					Namespace: serv.Namespace,
					Metadata: map[string]string{
						fmt.Sprintf("key-%d", i): fmt.Sprintf("val-%d", i),
						"another-key":            "another-value",
					},
				}
				listServices = append(listServices, s)
				listKvServices = append(listKvServices, &mvccpb.KeyValue{
					Key: []byte("/" + s.Name),
					Value: func() []byte {
						v, _ := yaml.Marshal(s)
						return v
					}(),
				})
				listServices[i-1].OriginalObject = listKvServices[i-1]
			}
		})

		It("returns all the services", func() {
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
						if key == "/"+serv.Namespace {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						}
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
								Kvs: listKvServices[0:3],
							}, nil
						case 1:
							Expect(key).To(Equal("/" + listServices[2].Name))
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{
									listKvServices[2],
									{
										Key: []byte("/my-service/endpoints/my-endpoint"),
										Value: func() []byte {
											v, _ := yaml.Marshal(&coretypes.Endpoint{
												Name:      "my-endpoint",
												Namespace: serv.Namespace,
												Service:   "my-service",
												Metadata: map[string]string{
													"whatever": "whatever",
												},
												Address: "10.10.10.10",
												Port:    8080,
											})
											return v
										}(),
									},
									listKvServices[3],
								},
							}, nil
						case 2:
							Expect(key).To(Equal("/" + listServices[3].Name))
							return &clientv3.GetResponse{
								Kvs: []*mvccpb.KeyValue{
									listKvServices[3],
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
			sit := e.Namespace(serv.Namespace).Service("").List(&list.Options{
				Results: int32(limit),
			})

			etcd.NewKV = func(kv clientv3.KV, prefix string) clientv3.KV {
				return &fakeKV{
					_Do: func(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
						// Just to make this work
						return clientv3.OpResponse{}, nil
					},
					_Get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						if key == "/"+serv.Namespace {
							return &clientv3.GetResponse{
								Header: &etcdserverpb.ResponseHeader{
									Revision: 0,
								},
								Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
							}, nil
						}
						Fail("should get it from cache")
						return nil, nil
					},
				}
			}

			for i := 0; i < len(listServices); i++ {
				val, nop, err := sit.Next(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(listServices[i]))
				Expect(nop).NotTo(BeNil())
				n, err := nop.Get(ctx, &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(listServices[i]))
			}

			val, nop, err := sit.Next(ctx)
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
							if key == "/"+serv.Namespace {
								return &clientv3.GetResponse{
									Header: &etcdserverpb.ResponseHeader{
										Revision: 0,
									},
									Kvs: []*mvccpb.KeyValue{kvsNamespaces[0]},
								}, nil
							}

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
								Kvs: listKvServices,
							}, nil
						},
					}
				}
				sit := e.Namespace("ns-1").Service("serv-1").List(&list.Options{
					NameFilters: &list.NameFilters{
						In: []string{"serv-3", "serv-5"},
					},
				})

				expLoops := 2
				expResults := []*coretypes.Service{listServices[0], listServices[2]}
				for i := 0; i < expLoops; i++ {
					val, sop, err := sit.Next(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(expResults[i]))
					Expect(sop).NotTo(BeNil())
				}

				val, sop, err := sit.Next(ctx)
				Expect(val).To(BeNil())
				Expect(sop).To(BeNil())
				Expect(err).To(MatchError(srerr.IteratorDone))
			})
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				By("checking that a namespace name is provided", func() {
					sit := e.Namespace("").Service("").List(&list.Options{})
					s, sop, err := sit.Next(ctx)
					Expect(s).To(BeNil())
					Expect(sop).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyNamespaceName))
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
					nit := e.Namespace("ns-1").Service("serv-1").List(&list.Options{
						NameFilters: &list.NameFilters{
							In: []string{"serv-3", "serv-5"},
						},
					})

					val, sop, err := nit.Next(ctx)
					Expect(err).To(MatchError(expErr))
					Expect(val).To(BeNil())
					Expect(sop).To(BeNil())

					val, sop, err = nit.Next(ctx)
					Expect(val).To(BeNil())
					Expect(sop).To(BeNil())
					Expect(err).To(MatchError(srerr.IteratorDone))
				})
			})
		})
	})
})
