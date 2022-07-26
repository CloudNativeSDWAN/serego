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

package cloudmap_test

import (
	"context"
	"fmt"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/aws/cloudmap"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"

	// "github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/aws/aws-sdk-go-v2/aws"
	sd "github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Operations", func() {
	const (
		nsOpID   = "ns-op-id"
		servOpID = "serv-op-id"
	)

	var (
		ctxtodo   context.Context
		f         *fakeCloudMapClient
		w         *cloudmap.AwsCloudMapWrapper
		ns        types.NamespaceSummary
		serv      types.ServiceSummary
		servTags  []types.Tag
		servMetas map[string]string
	)

	BeforeEach(func() {
		ctxtodo = context.TODO()
		f = &fakeCloudMapClient{}
		w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime})
		ns = namespaces[0]
		serv = services[1]
		servTags = tags[1]
		servMetas = metas[1]
		f._ListNamespaces = listNamespaces
		f._ListServices = listServices
		f._ListTagsForResource = listTagsForResource
	})

	Describe("Creating a service", func() {
		It("should call Cloud Map with correct parameters", func() {
			f._CreateService = func(ctx context.Context, params *sd.CreateServiceInput, optFns ...func(*sd.Options)) (*sd.CreateServiceOutput, error) {
				Expect(params).NotTo(BeNil())
				Expect(ctx).NotTo(BeNil())
				Expect(params.Name).To(Equal(serv.Name))
				Expect(params.NamespaceId).To(Equal(ns.Id))
				Expect(params.Type).To(Equal(types.ServiceTypeOptionHttp))
				Expect(params.Tags).To(ConsistOf(servTags))
				return &sd.CreateServiceOutput{
					Service: &types.Service{
						Arn:         serv.Arn,
						Id:          serv.Id,
						NamespaceId: ns.Id,
						Name:        serv.Name,
					},
				}, nil
			}
			f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
				Expect(params).To(Equal(&sd.GetServiceInput{
					Id: aws.String(*serv.Id),
				}))
				return &sd.GetServiceOutput{
					Service: &types.Service{
						Arn:         serv.Arn,
						Id:          serv.Id,
						Name:        serv.Name,
						NamespaceId: ns.Id,
					},
				}, nil
			}
			f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
				Expect(params).To(Equal(&sd.GetNamespaceInput{
					Id: ns.Id,
				}))
				return &sd.GetNamespaceOutput{
					Namespace: &types.Namespace{
						Arn:  ns.Arn,
						Id:   ns.Id,
						Name: ns.Name,
					},
				}, nil
			}

			s, err := w.Namespace(*ns.Name).
				Service(*serv.Name).Create(context.Background(), servMetas)
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal(&coretypes.Service{
				Name:      *serv.Name,
				Namespace: *ns.Name,
				Metadata:  servMetas,
				OriginalObject: &types.Service{
					Arn:         serv.Arn,
					Id:          serv.Id,
					Name:        serv.Name,
					NamespaceId: ns.Id,
				},
			}))

			f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
				Fail("got from cloud map instead of cache")
				return nil, nil
			}
			s, err = w.Namespace(*ns.Name).
				Service(*serv.Name).Get(context.Background(), &get.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal(&coretypes.Service{
				Name:      *serv.Name,
				Namespace: *ns.Name,
				Metadata:  servMetas,
				OriginalObject: &types.Service{
					Arn:         serv.Arn,
					Id:          serv.Id,
					Name:        serv.Name,
					NamespaceId: ns.Id,
				},
			}))
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				By("checking if the namespace exists", func() {
					s, err := w.Namespace("not-existing").
						Service(*serv.Name).Create(ctxtodo, map[string]string{})
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
					Expect(s).To(BeNil())
				})
				By("checking the return from CloudMap", func() {
					retErr := fmt.Errorf("whatever error")
					f._CreateService = func(ctx context.Context, params *sd.CreateServiceInput, optFns ...func(*sd.Options)) (*sd.CreateServiceOutput, error) {
						return nil, retErr
					}
					s, err := w.Namespace(*ns.Name).
						Service(*serv.Name).Create(ctxtodo, map[string]string{})
					Expect(err).To(MatchError(retErr))
					Expect(s).To(BeNil())
				})
			})
		})
	})

	Describe("Listing services", func() {
		BeforeEach(func() {
			w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: 0})
		})
		Context("with no filters", func() {
			It("should return all services", func() {
				resToGet := int32(2)
				timesList := 0
				f._ListServices = func(ctx context.Context, params *sd.ListServicesInput, optFns ...func(*sd.Options)) (*sd.ListServicesOutput, error) {
					defer func() {
						timesList++
					}()
					Expect(params).NotTo(BeNil())
					Expect(params.Filters).To(
						Equal([]types.ServiceFilter{
							{
								Name:      types.ServiceFilterNameNamespaceId,
								Condition: types.FilterConditionEq,
								Values:    []string{*ns.Id},
							}},
						),
					)
					Expect(params.MaxResults).To(Equal(aws.Int32(resToGet)))

					switch timesList {
					case 0:
						Expect(params.NextToken).To(BeNil())
						return &sd.ListServicesOutput{
							Services:  services[0:2],
							NextToken: aws.String("next-1"),
						}, nil
					case 1:
						Expect(params.NextToken).To(Equal(aws.String("next-1")))
						return &sd.ListServicesOutput{
							Services:  services[2:],
							NextToken: aws.String("next-2"),
						}, nil
					default:
						Expect(params.NextToken).To(Equal(aws.String("next-2")))
						return &sd.ListServicesOutput{}, nil
					}
				}

				it := w.Namespace(*ns.Name).Service("").List(&list.Options{
					Results: int32(resToGet),
				})

				for i := 0; i < len(services); i++ {
					s, sop, err := it.Next(ctxtodo)

					Expect(err).NotTo(HaveOccurred())
					Expect(s).To(Equal(&coretypes.Service{
						Name:      *services[i].Name,
						Namespace: *ns.Name,
						Metadata:  metas[i],
						OriginalObject: &types.Service{
							Arn:         services[i].Arn,
							Name:        services[i].Name,
							Id:          services[i].Id,
							NamespaceId: ns.Id,
						},
					}))
					f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
						Expect(params).NotTo(BeNil())
						Expect(params.Id).To(Equal(services[i].Id))
						return &sd.GetServiceOutput{
							Service: &types.Service{
								Arn:         services[i].Arn,
								Id:          services[i].Id,
								Name:        services[i].Name,
								NamespaceId: ns.Id,
							},
						}, nil
					}
					sop.Get(ctxtodo, &get.Options{})
				}

				s, sop, err := it.Next(ctxtodo)
				Expect(err).To(MatchError(srerr.IteratorDone))
				Expect(s).To(BeNil())
				Expect(sop).To(BeNil())
			})
		})

		Context("in case of errors", func() {
			It("should return the same error", func() {
				By("checking if namespace is specified", func() {
					it := w.Namespace("").Service("").List(&list.Options{})
					serv, sop, err := it.Next(ctxtodo)
					Expect(serv).To(BeNil())
					Expect(sop).To(BeNil())
					Expect(err).To(MatchError(srerr.EmptyNamespaceName))
				})
				By("checking if namespace exists first", func() {
					it := w.Namespace("doesnt-exist").
						Service("").List(&list.Options{})
					serv, sop, err := it.Next(ctxtodo)
					Expect(serv).To(BeNil())
					Expect(sop).To(BeNil())
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
				})
				By("checking that no two consecutive calls are made to cloudmap", func() {
					it := w.Namespace("doesnt-exist").
						Service("").List(&list.Options{})
					_, _, err := it.Next(ctxtodo)
					Expect(err).NotTo(MatchError(srerr.IteratorDone))
					_, _, err = it.Next(ctxtodo)
					Expect(err).To(MatchError(srerr.IteratorDone))
				})
				By("checking that not found tags cause a not found error", func() {
					oldF := f._ListTagsForResource
					defer func() {
						f._ListTagsForResource = oldF
					}()

					f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
						if *params.ResourceARN != *serv.Arn {
							return oldF(ctx, params)
						}

						return nil, &types.ResourceNotFoundException{}
					}
					it := w.Namespace(*ns.Name).
						Service(*serv.Name).List(&list.Options{})
					_, _, err := it.Next(ctxtodo)
					Expect(err).To(MatchError(srerr.IteratorDone))
				})
				By("checking the error from list services", func() {
					whateverErr := fmt.Errorf("whatever error")
					f._ListServices = func(ctx context.Context, params *sd.ListServicesInput, optFns ...func(*sd.Options)) (*sd.ListServicesOutput, error) {
						return nil, whateverErr
					}
					f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
						return &sd.GetNamespaceOutput{
							Namespace: &types.Namespace{
								Arn:  ns.Arn,
								Name: ns.Name,
								Id:   ns.Id,
							},
						}, nil
					}
					it := w.Namespace(*ns.Name).
						Service(*serv.Name).List(&list.Options{})
					s, sop, err := it.Next(ctxtodo)
					Expect(s).To(BeNil())
					Expect(sop).To(BeNil())
					Expect(err).To(MatchError(whateverErr))
				})
			})
		})
	})

	Describe("Retrieving a service", func() {
		It("should get it...", func() {
			sop := w.Namespace(*ns.Name).Service(*serv.Name)
			By("... first from the list", func() {
				s, err := sop.Get(ctxtodo, &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(s.OriginalObject).To(Equal(&types.Service{
					Arn:         serv.Arn,
					Name:        serv.Name,
					Id:          serv.Id,
					NamespaceId: ns.Id,
				}))
				Expect(s.Name).To(Equal(*serv.Name))
			})
			By("... then from cache", func() {
				f._ListServices = func(ctx context.Context, params *sd.ListServicesInput, optFns ...func(*sd.Options)) (*sd.ListServicesOutput, error) {
					Fail("should not call ListServices")
					return nil, nil
				}
				f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
					Fail("should not call GetService")
					return nil, nil
				}
				f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
					Fail("should not call ListTagsForResource")
					return nil, nil
				}

				s, err := sop.Get(ctxtodo, &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(s.OriginalObject).To(Equal(&types.Service{
					Arn:         serv.Arn,
					Name:        serv.Name,
					Id:          serv.Id,
					NamespaceId: ns.Id,
				}))
				Expect(s.Name).To(Equal(*serv.Name))
			})
		})

		Context("without cache", func() {
			It("only gets the ID from cache", func() {
				w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: 0})
				w.Namespace(*ns.Name).Service(*serv.Name).Get(ctxtodo, &get.Options{})
				f._ListServices = func(ctx context.Context, params *sd.ListServicesInput, optFns ...func(*sd.Options)) (*sd.ListServicesOutput, error) {
					Fail("should not call ListServices")
					return nil, nil
				}
				f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
					Expect(params).To(Equal(&sd.GetServiceInput{
						Id: serv.Id,
					}))
					return &sd.GetServiceOutput{
						Service: &types.Service{
							Arn:         serv.Arn,
							Id:          serv.Id,
							Name:        serv.Name,
							NamespaceId: ns.Id,
						},
					}, nil
				}
				w.Namespace(*ns.Name).Service(*serv.Name).Get(ctxtodo, &get.Options{})
			})
		})

		Context("in case of errors", func() {
			It("should return the same error", func() {
				w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: 0})
				sop := w.Namespace(*ns.Name).Service(*serv.Name)

				By("checking if namespace exists", func() {
					s, err := w.Namespace("whatever").Service("whatever").Get(ctxtodo, &get.Options{})
					Expect(s).To(BeNil())
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
				})
				By("checking if service exists", func() {
					s, err := w.Namespace(*ns.Name).Service("whatever").Get(ctxtodo, &get.Options{})
					Expect(s).To(BeNil())
					Expect(err).To(MatchError(srerr.ServiceNotFound))
				})
				By("checking the error from getByID", func() {
					f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
						return &sd.GetNamespaceOutput{
							Namespace: &types.Namespace{
								Arn:  ns.Arn,
								Name: ns.Name,
								Id:   ns.Id,
							},
						}, nil
					}
					sop.Get(ctxtodo, &get.Options{})
					whateverErr := fmt.Errorf("whatever")
					f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
						return nil, whateverErr
					}
					s, err := sop.Get(ctxtodo, &get.Options{})
					Expect(s).To(BeNil())
					Expect(err).To(MatchError(whateverErr))
				})
				By("checking the error from list tags", func() {
					f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
						return &sd.GetNamespaceOutput{
							Namespace: &types.Namespace{
								Arn:  ns.Arn,
								Id:   ns.Id,
								Name: ns.Name,
							},
						}, nil
					}
					sop.Get(ctxtodo, &get.Options{})
					whateverErr := fmt.Errorf("whatever")
					f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
						return &sd.GetServiceOutput{
							Service: &types.Service{
								Arn:         serv.Arn,
								Name:        serv.Name,
								NamespaceId: ns.Id,
								Id:          serv.Id,
							},
						}, nil
					}
					f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
						return nil, whateverErr
					}
					s, err := sop.Get(ctxtodo, &get.Options{})
					Expect(s).To(BeNil())
					Expect(err).To(MatchError(whateverErr))
				})
			})
		})
	})

	Describe("Deleting a service", func() {
		It("deletes the service successfully", func() {
			f._DeleteService = func(ctx context.Context, params *sd.DeleteServiceInput, optFns ...func(*sd.Options)) (*sd.DeleteServiceOutput, error) {
				Expect(params).To(Equal(&sd.DeleteServiceInput{
					Id: serv.Id,
				}))
				return &sd.DeleteServiceOutput{}, nil
			}

			err := w.Namespace(*ns.Name).
				Service(*serv.Name).Delete(ctxtodo)
			Expect(err).NotTo(HaveOccurred())

			called := false
			f._ListServices = func(ctx context.Context, params *sd.ListServicesInput, optFns ...func(*sd.Options)) (*sd.ListServicesOutput, error) {
				called = true
				return nil, fmt.Errorf("whatever")
			}
			w.Namespace(*ns.Name).Service(*serv.Name).Get(context.Background(), &get.Options{})
			Expect(called).To(BeTrue())
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				By("checking if the namespace has errors", func() {
					err := w.Namespace("whatever").
						Service("whatever").Delete(ctxtodo)
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
				})
				By("checking service errors", func() {
					err := w.Namespace(*ns.Name).
						Service("whatever").Delete(ctxtodo)
					Expect(err).To(MatchError(srerr.ServiceNotFound))
				})
				By("checking delete errors", func() {
					expErr := fmt.Errorf("whatever")
					f._DeleteService = func(ctx context.Context, params *sd.DeleteServiceInput, optFns ...func(*sd.Options)) (*sd.DeleteServiceOutput, error) {
						return nil, expErr
					}
					err := w.Namespace(*ns.Name).
						Service(*serv.Name).Delete(ctxtodo)
					Expect(err).To(MatchError(expErr))
				})
			})
		})
	})

	Describe("Updating a service", func() {
		It("updates tags correctly", func() {
			newTags := []types.Tag{
				{Key: servTags[0].Key, Value: aws.String("value-1-edited")},
				{Key: aws.String("new-key"), Value: aws.String("new-val")},
			}
			newMetadata := map[string]string{
				*servTags[0].Key: "value-1-edited",
				"new-key":        "new-val",
			}
			f._TagResource = func(ctx context.Context, params *sd.TagResourceInput, optFns ...func(*sd.Options)) (*sd.TagResourceOutput, error) {
				// No need to test this, tested in namespace
				return &sd.TagResourceOutput{}, nil
			}
			f._UntagResource = func(ctx context.Context, params *sd.UntagResourceInput, optFns ...func(*sd.Options)) (*sd.UntagResourceOutput, error) {
				// No need to test this, tested in namespace
				return &sd.UntagResourceOutput{}, nil
			}
			oldListTags := f._ListTagsForResource
			f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
				if *params.ResourceARN != *serv.Arn {
					return oldListTags(ctx, params)
				}

				return &sd.ListTagsForResourceOutput{
					Tags: newTags,
				}, nil
			}
			f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
				return &sd.GetServiceOutput{
					Service: &types.Service{
						Arn:         serv.Arn,
						Id:          serv.Id,
						Name:        serv.Name,
						NamespaceId: ns.Id,
						Type:        types.ServiceTypeHttp,
					},
				}, nil
			}
			s, err := w.Namespace(*ns.Name).Service(*serv.Name).
				Update(ctxtodo, newMetadata)
			Expect(s).To(Equal(&coretypes.Service{
				Name:      *serv.Name,
				Namespace: *ns.Name,
				Metadata:  newMetadata,
				OriginalObject: &types.Service{
					Arn:         serv.Arn,
					Name:        serv.Name,
					Id:          serv.Id,
					NamespaceId: ns.Id,
					Type:        types.ServiceTypeHttp,
				},
			}))
			Expect(err).NotTo(HaveOccurred())

			f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
				Fail("should get it from cache not cloud map")
				return nil, nil
			}
			f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
				Fail("should get it from cache not cloud map")
				return nil, nil
			}
			w.Namespace(*ns.Name).Service(*serv.Name).Get(context.Background(), &get.Options{})
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				By("checking that namespace exists", func() {
					s, err := w.Namespace("whatever").
						Service("whatever").Update(ctxtodo, map[string]string{})
					Expect(s).To(BeNil())
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
				})
				By("checking that service exists", func() {
					s, err := w.Namespace(*ns.Name).
						Service("whatever").Update(ctxtodo, map[string]string{})
					Expect(s).To(BeNil())
					Expect(err).To(MatchError(srerr.ServiceNotFound))
				})
				By("checking tags operation", func() {
					// No need to check both functions, they are tested in namespace
					expErr1 := fmt.Errorf("whatever")
					expErr2 := fmt.Errorf("whatever-2")
					f._TagResource = func(ctx context.Context, params *sd.TagResourceInput, optFns ...func(*sd.Options)) (*sd.TagResourceOutput, error) {
						return nil, expErr1
					}
					f._UntagResource = func(ctx context.Context, params *sd.UntagResourceInput, optFns ...func(*sd.Options)) (*sd.UntagResourceOutput, error) {
						return nil, expErr2
					}
					s, err := w.Namespace(*ns.Name).
						Service(*serv.Name).Update(ctxtodo, map[string]string{})
					Expect(s).To(BeNil())
					Expect(err).To(MatchError(expErr2))
				})
				By("checking the final get", func() {
					expErr := fmt.Errorf("whatever")
					f._TagResource = func(ctx context.Context, params *sd.TagResourceInput, optFns ...func(*sd.Options)) (*sd.TagResourceOutput, error) {
						return &sd.TagResourceOutput{}, nil
					}
					f._UntagResource = func(ctx context.Context, params *sd.UntagResourceInput, optFns ...func(*sd.Options)) (*sd.UntagResourceOutput, error) {
						return &sd.UntagResourceOutput{}, nil
					}
					f._GetService = func(ctx context.Context, params *sd.GetServiceInput, optFns ...func(*sd.Options)) (*sd.GetServiceOutput, error) {
						return nil, expErr
					}
					ns, err := w.Namespace(*ns.Name).Service(*serv.Name).Update(ctxtodo, map[string]string{
						"key": "val",
					})
					Expect(ns).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
		})
	})

})
