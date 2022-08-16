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
	"context"
	"fmt"
	"path"
	"time"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/gcp/servicedirectory"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	"github.com/googleapis/gax-go/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/iterator"
	pb "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

var _ = Describe("EndpointOperations", func() {
	var (
		f      *fakeRegistrationClient
		w      *servicedirectory.GoogleServiceDirectoryWrapper
		parent = path.Join("projects", project,
			"locations", region,
			"namespaces", nsName,
			"services", servName,
		)
		addr       = "10.10.10.10"
		port       = int32(80)
		epPathName = path.Join(parent, "endpoints", epName)
		metadata   = map[string]string{
			"key-1": "val-1",
			"key-2": "val-2",
		}
		expectedEndp = &coretypes.Endpoint{
			Name:      epName,
			Service:   servName,
			Namespace: nsName,
			Address:   addr,
			Port:      port,
			Metadata:  metadata,
			OriginalObject: &pb.Endpoint{
				Name:        epPathName,
				Address:     addr,
				Port:        port,
				Annotations: metadata,
			},
		}
	)

	BeforeEach(func() {
		f = &fakeRegistrationClient{}
		w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
			ProjectID:           project,
			Region:              region,
			CacheExpirationTime: time.Minute,
		})
	})

	Describe("Creating an endpoint", func() {
		Context("with valid parameters", func() {
			It("should call Service Directory with correct parameters", func() {
				f._createEndpoint = func(c context.Context, cer *pb.CreateEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					Expect(cer).To(Equal(&pb.CreateEndpointRequest{
						Parent:     parent,
						EndpointId: epName,
						Endpoint: &pb.Endpoint{
							Name:        epPathName,
							Address:     addr,
							Port:        port,
							Annotations: metadata,
						},
					}))
					return &pb.Endpoint{
						Name:        epPathName,
						Address:     addr,
						Port:        port,
						Annotations: metadata,
					}, nil
				}

				createdEp, err := w.Namespace(nsName).
					Service(servName).
					Endpoint(epName).
					Create(context.TODO(), addr, port, metadata)

				Expect(err).NotTo(HaveOccurred())
				Expect(createdEp).To(Equal(expectedEndp))

				f._getEndpoint = func(ctx context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					Fail("should get endpoint from cache not service directory")
					return nil, nil
				}
				createdEp, err = w.Namespace(nsName).
					Service(servName).
					Endpoint(epName).
					Get(context.TODO(), &get.Options{})

				Expect(err).NotTo(HaveOccurred())
				Expect(createdEp).To(Equal(expectedEndp))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				f._createEndpoint = func(c context.Context, cer *pb.CreateEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					return &pb.Endpoint{
						Name:        epPathName,
						Address:     addr,
						Port:        port,
						Annotations: metadata,
					}, nil
				}

				createdEp, err := w.Namespace(nsName).
					Service(servName).
					Endpoint(epName).
					Create(context.TODO(), addr, port, metadata)

				Expect(err).NotTo(HaveOccurred())
				Expect(createdEp).To(Equal(expectedEndp))

				called := false
				f._getEndpoint = func(ctx context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					called = true
					return &pb.Endpoint{
						Name:        epPathName,
						Address:     addr,
						Port:        port,
						Annotations: metadata,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				w.Namespace(nsName).
					Service(servName).
					Endpoint(epName).
					Get(context.TODO(), &get.Options{})
				Expect(called).To(BeTrue())
			})
		})

		Context("error returned from Service Directory", func() {
			It("should forward the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._createEndpoint = func(c context.Context, cer *pb.CreateEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).
					Service(servName).
					Endpoint(epName).Create(context.TODO(), addr, port, map[string]string{})
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Retrieving an endpoint", func() {
		Context("with valid parameters", func() {
			It("should call Service Directory with correct parameters", func() {
				timesCalled := 0
				f._getEndpoint = func(c context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					timesCalled++
					if timesCalled > 1 {
						Fail("should get endpoint from cache not service directory")
						return nil, nil
					}
					Expect(ger).To(Equal(&pb.GetEndpointRequest{
						Name: epPathName,
					}))
					return &pb.Endpoint{
						Name:        epPathName,
						Annotations: metadata,
						Address:     addr,
						Port:        port,
					}, nil
				}

				gotEndp, err := w.Namespace(nsName).Service(servName).
					Endpoint(epName).Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(gotEndp).To(Equal(expectedEndp))

				gotEndp, err = w.Namespace(nsName).Service(servName).
					Endpoint(epName).Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(gotEndp).To(Equal(expectedEndp))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				timesCalled := 0
				f._getEndpoint = func(c context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					timesCalled++
					return &pb.Endpoint{
						Name:        epPathName,
						Annotations: metadata,
						Address:     addr,
						Port:        port,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				w.Namespace(nsName).Service(servName).
					Endpoint(epName).Get(context.TODO(), &get.Options{})
				w.Namespace(nsName).Service(servName).
					Endpoint(epName).Get(context.TODO(), &get.Options{})

				Expect(timesCalled).To(Equal(2))
			})
		})

		Context("error returned from Service Directory", func() {
			It("should forward the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._getEndpoint = func(c context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).Service(servName).
					Endpoint(epName).Get(context.TODO(), &get.Options{})
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Updating an endpoint", func() {
		Context("with valid parameters", func() {
			It("should call Service Directory with correct parameters", func() {
				f._updateEndpoint = func(c context.Context, uer *pb.UpdateEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					Expect(uer).To(Equal(&pb.UpdateEndpointRequest{
						Endpoint: &pb.Endpoint{
							Name:        epPathName,
							Address:     addr,
							Port:        port,
							Annotations: metadata,
						},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"annotations", "address", "port"},
						},
					}))
					return &pb.Endpoint{
						Name:        epPathName,
						Address:     addr,
						Port:        port,
						Annotations: metadata,
					}, nil
				}

				updEndp, err := w.Namespace(nsName).Service(servName).
					Endpoint(epName).Update(context.TODO(), addr, port, metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(updEndp).To(Equal(expectedEndp))

				f._getEndpoint = func(ctx context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					Fail("should get endpoint from cache not service directory")
					return nil, nil
				}
				updEndp, err = w.Namespace(nsName).Service(servName).
					Endpoint(epName).Update(context.TODO(), addr, port, metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(updEndp).To(Equal(expectedEndp))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				f._updateEndpoint = func(c context.Context, uer *pb.UpdateEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					return &pb.Endpoint{
						Name:        epPathName,
						Address:     addr,
						Port:        port,
						Annotations: metadata,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				w.Namespace(nsName).Service(servName).
					Endpoint(epName).Update(context.TODO(), addr, port, metadata)

				called := false
				f._getEndpoint = func(ctx context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					called = true
					return &pb.Endpoint{
						Name:        epPathName,
						Address:     addr,
						Port:        port,
						Annotations: metadata,
					}, nil
				}
				w.Namespace(nsName).Service(servName).
					Endpoint(epName).Get(context.TODO(), &get.Options{})
				Expect(called).To(BeTrue())
			})
		})

		Context("error returned from Service Directory", func() {
			It("should forward the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._updateEndpoint = func(c context.Context, uer *pb.UpdateEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).Service(servName).
					Endpoint(epName).Update(context.TODO(), addr, port, map[string]string{})
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Deleting an endpoint", func() {
		It("should call Service Directory with correct parameters", func() {
			f._deleteEndpoint = func(c context.Context, der *pb.DeleteEndpointRequest, co ...gax.CallOption) error {
				Expect(der).To(Equal(&pb.DeleteEndpointRequest{
					Name: epPathName,
				}))
				return nil
			}

			err := w.Namespace(nsName).Service(servName).
				Endpoint(epName).Delete(context.TODO())
			Expect(err).NotTo(HaveOccurred())

			called := false
			f._getEndpoint = func(ctx context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
				called = true
				return nil, fmt.Errorf("whatever")
			}
			w.Namespace(nsName).Service(servName).
				Endpoint(epName).Get(context.TODO(), &get.Options{})
			Expect(called).To(BeTrue())
		})
		Context("in case of errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._deleteEndpoint = func(c context.Context, der *pb.DeleteEndpointRequest, co ...gax.CallOption) error {
					return expErr
				}

				err := w.Namespace(nsName).Service(servName).
					Endpoint(epName).Delete(context.TODO())
				Expect(err).To(Equal(expErr))
			})
		})
	})

	Describe("Listing endpoints", func() {
		var fakeIt *fakeEndpointIterator
		BeforeEach(func() {
			fakeIt = &fakeEndpointIterator{
				_next: func() (*pb.Endpoint, error) {
					return &pb.Endpoint{}, nil
				},
			}
		})

		Context("with no name or filters provided", func() {
			It("should contain no filters", func() {
				it := w.Namespace(nsName).Service(servName).Endpoint("").
					List(&list.Options{}).(*servicedirectory.ServiceDirectoryEndpointIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListEndpointsRequest{
					Parent: parent,
					Filter: "",
				}))
			})
		})

		Context("with a name provided", func() {
			It("should contain a filter for the name", func() {
				it := w.Namespace(nsName).Service(servName).Endpoint(epName).
					List(&list.Options{}).(*servicedirectory.ServiceDirectoryEndpointIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListEndpointsRequest{
					Parent: parent,
					Filter: fmt.Sprintf("name=%s", epPathName),
				}))
			})
		})

		Context("with a name provided and nameIn filter", func() {
			It("should contain a filter for the name and the filter", func() {
				it := w.Namespace(nsName).Service(servName).Endpoint(epName).
					List(&list.Options{
						NameFilters: &list.NameFilters{
							In: []string{"another"},
						},
					}).(*servicedirectory.ServiceDirectoryEndpointIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListEndpointsRequest{
					Parent: parent,
					Filter: fmt.Sprintf("(name=%s OR name=%s)",
						path.Join(parent, "endpoints", "another"),
						epPathName),
				}))
			})
		})

		Context("with multiple filters", func() {
			It("creates appropriate request data", func() {
				it := w.Namespace(nsName).Service(servName).Endpoint("").
					List(&list.Options{
						NameFilters: &list.NameFilters{
							In: []string{"another", "another-2"},
						},
						MetadataFilters: &list.MetadataFilters{
							Metadata: map[string]string{
								"key-1": "val-1",
								"key-2": "val-2",
								"key-3": "",
							},
						},
					}).(*servicedirectory.ServiceDirectoryEndpointIterator)
				it.Iterator = &fakeEndpointIterator{
					_next: func() (*pb.Endpoint, error) {
						return nil, srerr.IteratorDone
					},
				}

				it.Next(context.Background())
				Expect(it.Request).NotTo(BeNil())
				Expect(it.Request.Parent).To(Equal(parent))
				accepted := []string{
					fmt.Sprintf("(name=%s OR name=%s) AND (annotations.key-1=val-1 AND annotations.key-2=val-2)",
						path.Join(parent, "endpoints", "another"),
						path.Join(parent, "endpoints", "another-2"),
					),
					fmt.Sprintf("(name=%s OR name=%s) AND (annotations.key-2=val-2 AND annotations.key-1=val-1)",
						path.Join(parent, "endpoints", "another"),
						path.Join(parent, "endpoints", "another-2"),
					),
				}

				if it.Request.Filter != accepted[0] && it.Request.Filter != accepted[1] {
					Fail("filtered names are not as expected")
				}
			})
		})

		Context("with results", func() {
			It("returns a correct data", func() {
				it := w.Namespace(nsName).Service(servName).Endpoint("").
					List(&list.Options{
						NameFilters: &list.NameFilters{
							Prefix: "should",
						},
					}).(*servicedirectory.ServiceDirectoryEndpointIterator)
				results := []*pb.Endpoint{
					{
						Name:    path.Join(parent, "endpoints", "no-pass"),
						Address: "10.10.10.1",
						Port:    80,
						Annotations: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
					{
						Name:    path.Join(parent, "endpoints", "should-pass"),
						Address: "10.10.10.2",
						Port:    81,
						Annotations: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
				}
				iterating := 0
				it.Iterator = &fakeEndpointIterator{
					_next: func() (*pb.Endpoint, error) {
						if iterating < len(results) {
							toReturn := results[iterating]
							iterating++

							return toReturn, nil
						}

						return nil, iterator.Done
					},
				}

				ns, op, err := it.Next(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(ns).To(Equal(&coretypes.Endpoint{
					Name:      "should-pass",
					Namespace: nsName,
					Service:   servName,
					Metadata: map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
					},
					Address: "10.10.10.2",
					Port:    81,
					OriginalObject: &pb.Endpoint{
						Name:    path.Join(parent, "endpoints", "should-pass"),
						Address: "10.10.10.2",
						Port:    81,
						Annotations: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
				}))
				Expect(op).To(Equal(w.Namespace(nsName).Service(servName).Endpoint("should-pass")))

				f._getEndpoint = func(ctx context.Context, ger *pb.GetEndpointRequest, co ...gax.CallOption) (*pb.Endpoint, error) {
					Fail("should get endpoint from cache not service directory")
					return nil, nil
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(ns).To(Equal(&coretypes.Endpoint{
					Name:      "should-pass",
					Namespace: nsName,
					Service:   servName,
					Metadata: map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
					},
					Address: "10.10.10.2",
					Port:    81,
					OriginalObject: &pb.Endpoint{
						Name:    path.Join(parent, "endpoints", "should-pass"),
						Address: "10.10.10.2",
						Port:    81,
						Annotations: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
				}))

				ns, op, err = it.Next(context.Background())
				Expect(ns).To(BeNil())
				Expect(op).To(BeNil())
				Expect(err).To(Equal(iterator.Done))
			})
		})
	})
})
