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

	"github.com/googleapis/gax-go/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/iterator"
	pb "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/gcp/servicedirectory"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
)

var _ = Describe("ServiceOperations", func() {
	var (
		f      *fakeRegistrationClient
		w      *servicedirectory.GoogleServiceDirectoryWrapper
		parent = path.Join("projects", project,
			"locations", region,
			"namespaces", nsName,
		)
		servPathName = path.Join(parent, "services", servName)
		metadata     = map[string]string{
			"key-1": "val-1",
			"key-2": "val-2",
		}
		expectedServ = &coretypes.Service{
			Name:      servName,
			Namespace: nsName,
			Metadata:  metadata,
			OriginalObject: &pb.Service{
				Name:        servPathName,
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

	Describe("Creating a service", func() {

		Context("with valid parameters", func() {
			It("should call Service Directory with correct parameters", func() {
				f._createService = func(c context.Context, csr *pb.CreateServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					Expect(csr).To(Equal(&pb.CreateServiceRequest{
						Parent:    parent,
						ServiceId: servName,
						Service: &pb.Service{
							Name:        servPathName,
							Annotations: metadata,
						},
					}))
					return &pb.Service{
						Name:        servPathName,
						Annotations: metadata,
					}, nil
				}

				createdServ, err := w.Namespace(nsName).Service(servName).
					Create(context.Background(), metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdServ).To(Equal(expectedServ))

				f._getService = func(ctx context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					Fail("should get service from cache not service directory")
					return nil, nil
				}
				createdServ, err = w.Namespace(nsName).Service(servName).
					Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(createdServ).To(Equal(expectedServ))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				f._createService = func(c context.Context, csr *pb.CreateServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					return &pb.Service{
						Name:        servPathName,
						Annotations: metadata,
					}, nil
				}
				called := false
				f._getService = func(ctx context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					called = true
					return &pb.Service{
						Name:        servPathName,
						Annotations: metadata,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				w.Namespace(nsName).Service(servName).Create(context.Background(), metadata)
				w.Namespace(nsName).Service(servName).Get(context.TODO(), &get.Options{})
				Expect(called).To(BeTrue())
			})
		})

		Context("error returned from Service Directory", func() {
			It("should forward the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._createService = func(c context.Context, csr *pb.CreateServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).Service(servName).
					Create(context.TODO(), map[string]string{})
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Retrieving a service", func() {
		Context("with valid parameters", func() {
			It("should call Service Directory with correct parameters", func() {
				timesCalled := 0
				f._getService = func(c context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					timesCalled++
					if timesCalled > 1 {
						Fail("should get service from cache not service directory")
						return nil, nil
					}
					Expect(gsr).To(Equal(&pb.GetServiceRequest{
						Name: servPathName,
					}))
					return &pb.Service{
						Name:        servPathName,
						Annotations: metadata,
					}, nil
				}

				gotServ, err := w.Namespace(nsName).Service(servName).
					Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(gotServ).To(Equal(expectedServ))

				gotServ, err = w.Namespace(nsName).Service(servName).
					Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(gotServ).To(Equal(expectedServ))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				timesCalled := 0
				f._getService = func(c context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					timesCalled++
					return &pb.Service{
						Name:        servPathName,
						Annotations: metadata,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				w.Namespace(nsName).Service(servName).Get(context.TODO(), &get.Options{})
				w.Namespace(nsName).Service(servName).Get(context.TODO(), &get.Options{})
				Expect(timesCalled).To(Equal(2))
			})
		})

		Context("error returned from Service Directory", func() {
			It("should forward the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._getService = func(c context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).Service(servName).
					Get(context.TODO(), &get.Options{})
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Updating a service", func() {
		Context("with valid parameters", func() {
			It("should call Service Directory with correct parameters", func() {
				f._updateService = func(c context.Context, usr *pb.UpdateServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					Expect(usr).To(Equal(&pb.UpdateServiceRequest{
						Service: &pb.Service{
							Name:        servPathName,
							Annotations: metadata,
						},
						UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"annotations"}},
					}))
					return &pb.Service{
						Name:        servPathName,
						Annotations: metadata,
					}, nil
				}

				updServ, err := w.Namespace(nsName).Service(servName).Update(context.TODO(), metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(updServ).To(Equal(expectedServ))

				f._getService = func(ctx context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					Fail("should get service from cache not service directory")
					return nil, nil
				}
				updServ, err = w.Namespace(nsName).Service(servName).
					Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(updServ).To(Equal(expectedServ))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				f._updateService = func(c context.Context, usr *pb.UpdateServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					return &pb.Service{
						Name:        servPathName,
						Annotations: metadata,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				updServ, err := w.Namespace(nsName).Service(servName).Update(context.TODO(), metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(updServ).To(Equal(expectedServ))

				called := false
				f._getService = func(ctx context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					called = true
					return &pb.Service{
						Name:        servPathName,
						Annotations: metadata,
					}, nil
				}
				w.Namespace(nsName).Service(servName).Get(context.TODO(), &get.Options{})
				Expect(called).To(BeTrue())
			})
		})

		Context("error returned from Service Directory", func() {
			It("should forward the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._updateService = func(c context.Context, usr *pb.UpdateServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).Service(servName).Update(context.TODO(), metadata)
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Deleting a service", func() {
		It("should call Service Directory with correct parameters", func() {
			f._deleteService = func(c context.Context, dsr *pb.DeleteServiceRequest, co ...gax.CallOption) error {
				Expect(dsr).To(Equal(&pb.DeleteServiceRequest{
					Name: servPathName,
				}))
				return nil
			}

			err := w.Namespace(nsName).Service(servName).Delete(context.TODO())
			Expect(err).NotTo(HaveOccurred())

			called := false
			f._getService = func(ctx context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
				called = true
				return nil, fmt.Errorf("whatever")
			}
			w.Namespace(nsName).Service(servName).Get(context.TODO(), &get.Options{})
			Expect(called).To(BeTrue())
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._deleteService = func(c context.Context, dsr *pb.DeleteServiceRequest, co ...gax.CallOption) error {
					Expect(dsr).To(Equal(&pb.DeleteServiceRequest{
						Name: servPathName,
					}))
					return expErr
				}

				err := w.Namespace(nsName).Service(servName).Delete(context.TODO())
				Expect(err).To(Equal(expErr))
			})
		})
	})

	Describe("Listing services", func() {
		var fakeIt *fakeServiceIterator
		BeforeEach(func() {
			fakeIt = &fakeServiceIterator{
				_next: func() (*pb.Service, error) {
					return &pb.Service{}, nil
				},
			}
		})

		Context("with no name or filters provided", func() {
			It("should contain no filters", func() {
				it := w.Namespace(nsName).Service("").
					List(&list.Options{}).(*servicedirectory.ServiceDirectoryServiceIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListServicesRequest{
					Parent: parent,
				}))
			})
		})

		Context("with a name provided", func() {
			It("should contain a filter for the name", func() {
				it := w.Namespace(nsName).Service(servName).
					List(&list.Options{}).(*servicedirectory.ServiceDirectoryServiceIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListServicesRequest{
					Parent: parent,
					Filter: fmt.Sprintf("name=%s", servPathName),
				}))
			})
		})

		Context("with a name provided and nameIn filter", func() {
			It("should contain a filter for the name and the filter", func() {
				it := w.Namespace(nsName).Service(servName).
					List(&list.Options{
						NameFilters: &list.NameFilters{
							In: []string{"another"},
						},
					}).(*servicedirectory.ServiceDirectoryServiceIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListServicesRequest{
					Parent: parent,
					Filter: fmt.Sprintf("(name=%s OR name=%s)",
						path.Join(parent, "services", "another"),
						servPathName),
				}))
			})
		})

		Context("with multiple filters", func() {
			It("creates appropriate request data", func() {
				it := w.Namespace(nsName).Service("").
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
					}).(*servicedirectory.ServiceDirectoryServiceIterator)
				it.Iterator = &fakeServiceIterator{
					_next: func() (*pb.Service, error) {
						return nil, srerr.IteratorDone
					},
				}

				it.Next(context.Background())
				Expect(it.Request).NotTo(BeNil())
				Expect(it.Request.Parent).To(Equal(parent))
				accepted := []string{
					fmt.Sprintf("(name=%s OR name=%s) AND (annotations.key-1=val-1 AND annotations.key-2=val-2)",
						path.Join(parent, "services", "another"),
						path.Join(parent, "services", "another-2"),
					),
					fmt.Sprintf("(name=%s OR name=%s) AND (annotations.key-2=val-2 AND annotations.key-1=val-1)",
						path.Join(parent, "services", "another"),
						path.Join(parent, "services", "another-2"),
					),
				}

				if it.Request.Filter != accepted[0] && it.Request.Filter != accepted[1] {
					Fail(fmt.Sprintf("filtered names are not as expected: %s", it.Request.Filter))
				}
			})
		})

		Context("with results", func() {
			It("returns a correct data", func() {
				it := w.Namespace(nsName).Service("").
					List(&list.Options{
						NameFilters: &list.NameFilters{
							Prefix: "should",
						},
					}).(*servicedirectory.ServiceDirectoryServiceIterator)
				results := []*pb.Service{
					{
						Name: path.Join(parent, "services", "no-pass"),
						Annotations: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
					{
						Name: path.Join(parent, "services", "should-pass"),
						Annotations: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
				}
				iterating := 0
				it.Iterator = &fakeServiceIterator{
					_next: func() (*pb.Service, error) {
						if iterating < len(results) {
							toReturn := results[iterating]
							iterating++

							return toReturn, nil
						}

						return nil, iterator.Done
					},
				}

				serv, op, err := it.Next(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(serv).To(Equal(&coretypes.Service{
					Name:      "should-pass",
					Namespace: nsName,
					Metadata: map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
					},
					OriginalObject: &pb.Service{
						Name: path.Join(parent, "services", "should-pass"),
						Annotations: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
				}))
				Expect(op).To(Equal(w.Namespace(nsName).Service("should-pass")))

				f._getService = func(ctx context.Context, gsr *pb.GetServiceRequest, co ...gax.CallOption) (*pb.Service, error) {
					Fail("should get it from cache, not service directory")
					return nil, nil
				}

				serv, err = op.Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(serv).To(Equal(&coretypes.Service{
					Name:      "should-pass",
					Namespace: nsName,
					Metadata: map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
					},
					OriginalObject: &pb.Service{
						Name: path.Join(parent, "services", "should-pass"),
						Annotations: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
				}))

				serv, op, err = it.Next(context.Background())
				Expect(serv).To(BeNil())
				Expect(op).To(BeNil())
				Expect(err).To(Equal(iterator.Done))
			})
		})
	})
})
