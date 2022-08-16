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
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ = Describe("Namespace Operations", func() {
	var (
		f              *fakeRegistrationClient
		w              *servicedirectory.GoogleServiceDirectoryWrapper
		parent         = "projects/project/locations/region"
		expectedNsPath = path.Join(parent, "namespaces", nsName)
		metadata       = map[string]string{
			"key-1": "val-1",
			"key-2": "val-2",
		}
		expectedNs = &coretypes.Namespace{
			Name:     nsName,
			Metadata: metadata,
			OriginalObject: &pb.Namespace{
				Name:   expectedNsPath,
				Labels: metadata,
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

	Describe("Creating a namespace", func() {
		Context("with valid parameters", func() {
			It("calls Service Directory with correct parameters", func() {
				f._createNamespace = func(c context.Context, cnr *pb.CreateNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					Expect(cnr).To(Equal(&pb.CreateNamespaceRequest{
						NamespaceId: nsName,
						Parent:      parent,
						Namespace: &pb.Namespace{
							Name:   expectedNsPath,
							Labels: metadata,
						},
					}))
					return &pb.Namespace{
						Name:   path.Join("projects", project, "locations", region, "namespaces", nsName),
						Labels: metadata,
					}, nil
				}

				createdNs, err := w.Namespace(nsName).Create(context.TODO(), metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdNs).To(Equal(expectedNs))

				f._getNamespace = func(ctx context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					Fail("should get namespace from cache not service directory")
					return nil, nil
				}
				createdNs, err = w.Namespace(nsName).Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(createdNs).To(Equal(expectedNs))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				f._createNamespace = func(c context.Context, cnr *pb.CreateNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					return &pb.Namespace{
						Name:   path.Join("projects", project, "locations", region, "namespaces", nsName),
						Labels: metadata,
					}, nil
				}
				called := false
				f._getNamespace = func(ctx context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					called = true
					return &pb.Namespace{
						Name:   expectedNsPath,
						Labels: metadata,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				w.Namespace(nsName).Create(context.TODO(), metadata)
				w.Namespace(nsName).Get(context.TODO(), &get.Options{})
				Expect(called).To(BeTrue())
			})
		})

		Context("error from Service Directory", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._createNamespace = func(c context.Context, cnr *pb.CreateNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).Create(context.TODO(), map[string]string{})
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Retrieving a namespace", func() {
		Context("with valid parameters", func() {
			It("should call Service Directory with correct parameters", func() {
				times := 0
				f._getNamespace = func(c context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					times++
					if times > 1 {
						Fail("should get namespace from cache not service directory")
						return nil, nil
					}

					Expect(gnr).To(Equal(&pb.GetNamespaceRequest{
						Name: expectedNsPath,
					}))
					return &pb.Namespace{
						Name:   expectedNsPath,
						Labels: metadata,
					}, nil
				}

				gotNs, err := w.Namespace(nsName).Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(gotNs).To(Equal(expectedNs))

				gotNs, err = w.Namespace(nsName).Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(gotNs).To(Equal(expectedNs))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				timesCalled := 0
				f._getNamespace = func(ctx context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					timesCalled++
					return &pb.Namespace{
						Name:   expectedNsPath,
						Labels: metadata,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				w.Namespace(nsName).Get(context.TODO(), &get.Options{})
				w.Namespace(nsName).Get(context.TODO(), &get.Options{})
				Expect(timesCalled).To(Equal(2))
			})
		})

		Context("error returned from Service Directory", func() {
			It("should forward the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._getNamespace = func(c context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).Get(context.TODO(), &get.Options{})
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Updating a namespace", func() {
		Context("with valid parameters", func() {
			It("should call Service Directory with correct parameters", func() {
				f._updateNamespace = func(c context.Context, unr *pb.UpdateNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					Expect(unr).To(Equal(&pb.UpdateNamespaceRequest{
						Namespace: &pb.Namespace{
							Name:   expectedNsPath,
							Labels: metadata,
						},
						UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"labels"}},
					}))
					return &pb.Namespace{
						Name:   expectedNsPath,
						Labels: metadata,
					}, nil
				}

				updNs, err := w.Namespace(nsName).Update(context.TODO(), metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(updNs).To(Equal(expectedNs))

				f._getNamespace = func(ctx context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					Fail("should get namespace from cache not service directory")
					return nil, nil
				}
				updNs, err = w.Namespace(nsName).Update(context.TODO(), metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(updNs).To(Equal(expectedNs))
			})
		})

		Context("without cache", func() {
			It("always calls service directory", func() {
				f._updateNamespace = func(c context.Context, unr *pb.UpdateNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					return &pb.Namespace{
						Name:   expectedNsPath,
						Labels: metadata,
					}, nil
				}
				called := false
				f._getNamespace = func(ctx context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					called = true
					return &pb.Namespace{
						Name:   expectedNsPath,
						Labels: metadata,
					}, nil
				}

				w, _ = servicedirectory.NewServiceDirectoryWrapper(f, &wrapper.Options{
					ProjectID: project,
					Region:    region,
				})
				w.Namespace(nsName).Update(context.TODO(), metadata)
				w.Namespace(nsName).Get(context.TODO(), &get.Options{})
				Expect(called).To(BeTrue())
			})
		})

		Context("error returned from Service Directory", func() {
			It("should forward the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._updateNamespace = func(c context.Context, unr *pb.UpdateNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					return nil, expErr
				}

				res, err := w.Namespace(nsName).Update(context.TODO(), map[string]string{})
				Expect(res).To(BeNil())
				Expect(err).To(MatchError(expErr))
			})
		})
	})

	Describe("Deleting a namespace", func() {
		It("should call Service Directory with correct parameters", func() {
			f._deleteNamespace = func(c context.Context, dnr *pb.DeleteNamespaceRequest, co ...gax.CallOption) error {
				Expect(dnr).To(Equal(&pb.DeleteNamespaceRequest{
					Name: expectedNsPath,
				}))
				return nil
			}

			err := w.Namespace(nsName).Delete(context.TODO())
			Expect(err).NotTo(HaveOccurred())

			called := false
			f._getNamespace = func(ctx context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
				called = true
				return nil, fmt.Errorf("whatever")
			}
			w.Namespace(nsName).Get(context.TODO(), &get.Options{})
			Expect(called).To(BeTrue())
		})
		Context("in case of errors", func() {
			It("returns the same error", func() {
				expErr := fmt.Errorf("whatever")
				f._deleteNamespace = func(c context.Context, dnr *pb.DeleteNamespaceRequest, co ...gax.CallOption) error {
					return expErr
				}

				err := w.Namespace(nsName).Delete(context.TODO())
				Expect(err).To(Equal(expErr))
			})
		})
	})

	Describe("Listing namespaces", func() {
		var fakeIt *fakeNamespaceIterator
		BeforeEach(func() {
			fakeIt = &fakeNamespaceIterator{
				_next: func() (*pb.Namespace, error) {
					return &pb.Namespace{}, nil
				},
			}
		})

		Context("with no name or filters provided", func() {
			It("should contain no filters", func() {
				it := w.Namespace("").
					List(&list.Options{}).(*servicedirectory.ServiceDirectoryNamespaceIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListNamespacesRequest{
					Parent: parent,
				}))
			})
		})

		Context("with a name provided", func() {
			It("should contain a filter for the name", func() {
				it := w.Namespace(nsName).
					List(&list.Options{}).(*servicedirectory.ServiceDirectoryNamespaceIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListNamespacesRequest{
					Parent: parent,
					Filter: fmt.Sprintf("name=%s", expectedNsPath),
				}))
			})
		})

		Context("with a name provided and nameIn filter", func() {
			It("should contain a filter for the name and the filter", func() {
				it := w.Namespace(nsName).
					List(&list.Options{
						NameFilters: &list.NameFilters{
							In: []string{"another"},
						},
					}).(*servicedirectory.ServiceDirectoryNamespaceIterator)
				it.Iterator = fakeIt

				it.Next(context.Background())
				Expect(it.Request).To(Equal(&pb.ListNamespacesRequest{
					Parent: parent,
					Filter: fmt.Sprintf("(name=%s OR name=%s)",
						path.Join(parent, "namespaces", "another"),
						expectedNsPath),
				}))
			})
		})

		Context("with multiple filters", func() {
			It("creates appropriate request data", func() {
				it := w.Namespace("").
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
					}).(*servicedirectory.ServiceDirectoryNamespaceIterator)
				it.Iterator = &fakeNamespaceIterator{
					_next: func() (*pb.Namespace, error) {
						return nil, srerr.IteratorDone
					},
				}

				it.Next(context.Background())
				Expect(it.Request).NotTo(BeNil())
				Expect(it.Request.Parent).To(Equal(parent))
				accepted := []string{
					fmt.Sprintf("(name=%s OR name=%s) AND (labels.key-1=val-1 AND labels.key-2=val-2)",
						path.Join(parent, "namespaces", "another"),
						path.Join(parent, "namespaces", "another-2"),
					),
					fmt.Sprintf("(name=%s OR name=%s) AND (labels.key-2=val-2 AND labels.key-1=val-1)",
						path.Join(parent, "namespaces", "another"),
						path.Join(parent, "namespaces", "another-2"),
					),
				}

				if it.Request.Filter != accepted[0] && it.Request.Filter != accepted[1] {
					Fail("filtered names are not as expected")
				}
			})

			It("does not include parenthesis when only one metadata value is there", func() {
				it := w.Namespace("").
					List(&list.Options{
						NameFilters: &list.NameFilters{
							In: []string{"another", "another-2"},
						},
						MetadataFilters: &list.MetadataFilters{
							Metadata: map[string]string{
								"key-1": "val-1",
							},
						},
					}).(*servicedirectory.ServiceDirectoryNamespaceIterator)
				it.Iterator = &fakeNamespaceIterator{
					_next: func() (*pb.Namespace, error) {
						return nil, srerr.IteratorDone
					},
				}

				it.Next(context.Background())
				Expect(it.Request.Filter).To(Equal(fmt.Sprintf("(name=%s OR name=%s) AND labels.key-1=val-1",
					path.Join(parent, "namespaces", "another"),
					path.Join(parent, "namespaces", "another-2"),
				)))
			})
		})

		Context("with results", func() {
			It("returns a correct data", func() {
				it := w.Namespace("").
					List(&list.Options{
						NameFilters: &list.NameFilters{
							Prefix: "should",
						},
					}).(*servicedirectory.ServiceDirectoryNamespaceIterator)
				results := []*pb.Namespace{
					{
						Name: path.Join(parent, "namespaces", "no-pass"),
						Labels: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
					{
						Name: path.Join(parent, "namespaces", "should-pass"),
						Labels: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
				}
				iterating := 0
				it.Iterator = &fakeNamespaceIterator{
					_next: func() (*pb.Namespace, error) {
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
				Expect(ns).To(Equal(&coretypes.Namespace{
					Name: "should-pass",
					Metadata: map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
					},
					OriginalObject: &pb.Namespace{
						Name: path.Join(parent, "namespaces", "should-pass"),
						Labels: map[string]string{
							"key-1": "val-1",
							"key-2": "val-2",
						},
					},
				}))
				Expect(op).To(Equal(w.Namespace("should-pass")))
				f._getNamespace = func(ctx context.Context, gnr *pb.GetNamespaceRequest, co ...gax.CallOption) (*pb.Namespace, error) {
					Fail("should get it from cache, not service directory")
					return nil, nil
				}
				ns, err = op.Get(context.TODO(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(ns).To(Equal(&coretypes.Namespace{
					Name: "should-pass",
					Metadata: map[string]string{
						"key-1": "val-1",
						"key-2": "val-2",
					},
					OriginalObject: &pb.Namespace{
						Name: path.Join(parent, "namespaces", "should-pass"),
						Labels: map[string]string{
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
