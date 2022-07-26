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
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	sd "github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/internal/wrappers/aws/cloudmap"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
)

var _ = Describe("EndpointOperations", func() {
	const (
		nsOpID   = "ns-op-id"
		servOpID = "serv-op-id"
		endpOpID = "endp-op-id"
	)

	var (
		ctxtodo   context.Context
		f         *fakeCloudMapClient
		w         *cloudmap.AwsCloudMapWrapper
		ip        string
		port      int
		ns        types.NamespaceSummary
		serv      types.ServiceSummary
		endp      types.InstanceSummary
		endpMetas map[string]string
		endpTags  []types.Tag
	)

	BeforeEach(func() {
		ip = endpoints[2].Attributes["AWS_INSTANCE_IPV4"]
		port, _ = strconv.Atoi(endpoints[2].Attributes["AWS_INSTANCE_PORT"])
		ctxtodo = context.TODO()

		f = &fakeCloudMapClient{}
		w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime})

		f._ListNamespaces = listNamespaces
		f._ListServices = listServices
		f._ListTagsForResource = listTagsForResource
		f._ListInstances = listInstances
		ns = namespaces[0]
		serv = services[1]
		endp = endpoints[2]
		endpMetas = metas[2]
		endpTags = tags[2]
	})

	Describe("Creating an endpoint", func() {
		Context("with IPv4", func() {
			It("should call Cloud Map with correct parameters", func() {
				f._RegisterInstance = func(ctx context.Context, params *sd.RegisterInstanceInput, optFns ...func(*sd.Options)) (*sd.RegisterInstanceOutput, error) {
					Expect(params).To(Equal(&sd.RegisterInstanceInput{
						Attributes: endp.Attributes,
						InstanceId: endp.Id,
						ServiceId:  serv.Id,
					}))
					return &sd.RegisterInstanceOutput{OperationId: aws.String("op-id")}, nil
				}
				f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
					Expect(params).To(Equal(&sd.GetOperationInput{
						OperationId: aws.String("op-id"),
					}))

					return &sd.GetOperationOutput{
						Operation: &types.Operation{
							Status: types.OperationStatusSuccess,
						},
					}, nil
				}
				f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
					return &sd.GetInstanceOutput{
						Instance: &types.Instance{
							Id:         endp.Id,
							Attributes: endp.Attributes,
						},
					}, nil
				}

				e, err := w.Namespace(*ns.Name).
					Service(*serv.Name).
					Endpoint(*endp.Id).Create(ctxtodo, ip, int32(port), endpMetas)
				Expect(err).NotTo(HaveOccurred())
				Expect(e).To(Equal(&coretypes.Endpoint{
					Name:      *endp.Id,
					Namespace: *ns.Name,
					Service:   *serv.Name,
					Metadata:  endpMetas,
					Address:   ip,
					Port:      int32(port),
					OriginalObject: &types.Instance{
						Id:         endp.Id,
						Attributes: endp.Attributes,
					},
				}))

				f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
					Fail("got from cloud map instead of cache")
					return nil, nil
				}
				e, err = w.Namespace(*ns.Name).
					Service(*serv.Name).
					Endpoint(*endp.Id).Get(ctxtodo, &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(e).To(Equal(&coretypes.Endpoint{
					Name:      *endp.Id,
					Namespace: *ns.Name,
					Service:   *serv.Name,
					Metadata:  endpMetas,
					Address:   ip,
					Port:      int32(port),
					OriginalObject: &types.Instance{
						Id:         endp.Id,
						Attributes: endp.Attributes,
					},
				}))
			})
			Context("with no address or ips", func() {
				It("should call Cloud Map with correct parameters", func() {
					f._RegisterInstance = func(ctx context.Context, params *sd.RegisterInstanceInput, optFns ...func(*sd.Options)) (*sd.RegisterInstanceOutput, error) {
						Expect(params.Attributes).To(Equal(endpMetas))
						return &sd.RegisterInstanceOutput{OperationId: aws.String("op-id")}, nil
					}
					f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
						Expect(params).To(Equal(&sd.GetOperationInput{
							OperationId: aws.String("op-id"),
						}))

						return &sd.GetOperationOutput{
							Operation: &types.Operation{
								Status: types.OperationStatusSuccess,
							},
						}, nil
					}
					f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
						return &sd.GetInstanceOutput{
							Instance: &types.Instance{
								Id:         endp.Id,
								Attributes: endpMetas,
							},
						}, nil
					}

					e, err := w.Namespace(*ns.Name).
						Service(*serv.Name).
						Endpoint(*endp.Id).Create(ctxtodo, "", 0, endpMetas)
					Expect(err).NotTo(HaveOccurred())
					Expect(e).To(Equal(&coretypes.Endpoint{
						Name:      *endp.Id,
						Namespace: *ns.Name,
						Service:   *serv.Name,
						Metadata:  endpMetas,
						Address:   "",
						Port:      0,
						OriginalObject: &types.Instance{
							Id:         endp.Id,
							Attributes: endpMetas,
						},
					}))
				})
			})
		})
		Context("with IPv6", func() {
			It("should call Cloud Map with correct parameters", func() {
				ipv6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
				f._RegisterInstance = func(ctx context.Context, params *sd.RegisterInstanceInput, optFns ...func(*sd.Options)) (*sd.RegisterInstanceOutput, error) {
					Expect(params.InstanceId).To(Equal(endp.Id))
					Expect(params.ServiceId).To(Equal(serv.Id))
					Expect(params).To(Equal(&sd.RegisterInstanceInput{
						Attributes: map[string]string{
							"AWS_INSTANCE_IPV6": ipv6,
							"AWS_INSTANCE_PORT": endp.Attributes["AWS_INSTANCE_PORT"],
							*endpTags[0].Key:    *endpTags[0].Value,
							*endpTags[1].Key:    *endpTags[1].Value,
						},
						InstanceId: endp.Id,
						ServiceId:  serv.Id,
					}))
					return &sd.RegisterInstanceOutput{OperationId: aws.String("op-id")}, nil
				}
				f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
					Expect(params).To(Equal(&sd.GetOperationInput{
						OperationId: aws.String("op-id"),
					}))

					return &sd.GetOperationOutput{
						Operation: &types.Operation{
							Status: types.OperationStatusSuccess,
						},
					}, nil
				}
				f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
					return &sd.GetInstanceOutput{
						Instance: &types.Instance{
							Id: endp.Id,
							Attributes: map[string]string{
								"AWS_INSTANCE_IPV6": ipv6,
								"AWS_INSTANCE_PORT": endp.Attributes["AWS_INSTANCE_PORT"],
								*endpTags[0].Key:    *endpTags[0].Value,
								*endpTags[1].Key:    *endpTags[1].Value,
							},
						},
					}, nil
				}

				e, err := w.Namespace(*ns.Name).
					Service(*serv.Name).
					Endpoint(*endp.Id).Create(ctxtodo, ipv6, int32(port), endpMetas)
				Expect(err).NotTo(HaveOccurred())
				Expect(e).To(Equal(&coretypes.Endpoint{
					Name:      *endp.Id,
					Namespace: *ns.Name,
					Service:   *serv.Name,
					Metadata:  endpMetas,
					Address:   ipv6,
					Port:      int32(port),
					OriginalObject: &types.Instance{
						Id: endp.Id,
						Attributes: map[string]string{
							"AWS_INSTANCE_IPV6": ipv6,
							"AWS_INSTANCE_PORT": endp.Attributes["AWS_INSTANCE_PORT"],
							*endpTags[0].Key:    *endpTags[0].Value,
							*endpTags[1].Key:    *endpTags[1].Value,
						},
					},
				}))
			})
		})

		Context("in case of errors", func() {
			It("returns the same error", func() {
				By("checking if a parent exists", func() {
					s, err := w.Namespace("not-existing").
						Service(*serv.Name).
						Endpoint(*endp.Id).Create(ctxtodo, ip, int32(port), map[string]string{})
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
					Expect(s).To(BeNil())
				})
				By("checking return from CloudMap", func() {
					expErr := fmt.Errorf("whatever")
					f._RegisterInstance = func(ctx context.Context, params *sd.RegisterInstanceInput, optFns ...func(*sd.Options)) (*sd.RegisterInstanceOutput, error) {
						return nil, expErr
					}
					s, err := w.Namespace(*ns.Name).
						Service(*serv.Name).
						Endpoint(*endp.Id).Create(ctxtodo, ip, int32(port), map[string]string{})
					Expect(err).To(MatchError(expErr))
					Expect(s).To(BeNil())
				})
				By("checking operation status", func() {
					f._RegisterInstance = func(ctx context.Context, params *sd.RegisterInstanceInput, optFns ...func(*sd.Options)) (*sd.RegisterInstanceOutput, error) {
						return &sd.RegisterInstanceOutput{OperationId: aws.String("op-id")}, nil
					}
					f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
						return &sd.GetOperationOutput{
							Operation: &types.Operation{
								Status:       types.OperationStatusFail,
								ErrorCode:    aws.String("ACCESS_DENIED"),
								ErrorMessage: aws.String("whatever"),
							},
						}, nil
					}
					s, err := w.Namespace(*ns.Name).
						Service(*serv.Name).
						Endpoint(*endp.Id).Create(ctxtodo, ip, int32(port), map[string]string{})
					Expect(err).To(HaveOccurred())
					Expect(s).To(BeNil())
				})
			})
		})
	})

	Describe("Listing endpoints", func() {
		BeforeEach(func() {
			w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime})
		})
		Context("with no filters provided", func() {
			It("should return all endpoints", func() {
				resToGet := int32(2)
				timesList := 0
				f._ListInstances = func(ctx context.Context, params *sd.ListInstancesInput, optFns ...func(*sd.Options)) (*sd.ListInstancesOutput, error) {
					defer func() {
						timesList++
					}()
					Expect(timesList).To(BeNumerically("<", 3))
					Expect(params.ServiceId).To(Equal(serv.Id))
					Expect(params.MaxResults).To(Equal(&resToGet))

					switch timesList {
					case 0:
						Expect(params.NextToken).To(BeNil())
						return &sd.ListInstancesOutput{
							Instances: endpoints[:2],
							NextToken: aws.String("next-1"),
						}, nil
					case 1:
						Expect(params.NextToken).To(Equal(aws.String("next-1")))
						return &sd.ListInstancesOutput{
							Instances: endpoints[2:],
							NextToken: aws.String("next-2"),
						}, nil
					default:
						Expect(params.NextToken).To(Equal(aws.String("next-2")))
						return &sd.ListInstancesOutput{}, nil
					}
				}
				it := w.Namespace(*ns.Name).
					Service(*serv.Name).
					Endpoint("").List(&list.Options{
					Results: resToGet,
				})

				for i := 0; i < len(endpoints); i++ {
					e, eop, err := it.Next(ctxtodo)

					Expect(err).NotTo(HaveOccurred())
					Expect(e.Name).To(Equal(*endpoints[i].Id))
					Expect(e.Address).To(Equal(endpoints[i].Attributes["AWS_INSTANCE_IPV4"]))
					Expect(e.Port).To(Equal(func() int32 {
						p, _ := strconv.Atoi(endpoints[i].Attributes["AWS_INSTANCE_PORT"])
						return int32(p)
					}()))
					Expect(e.Metadata).To(Equal(metas[i]))
					f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
						Expect(params).To(Equal(&sd.GetInstanceInput{
							InstanceId: endpoints[i].Id,
							ServiceId:  serv.Id,
						}))
						metToRet := map[string]string{}
						for k, v := range metas[i] {
							metToRet[k] = v
						}
						for k, v := range endpoints[i].Attributes {
							metToRet[k] = v
						}
						return &sd.GetInstanceOutput{
							Instance: &types.Instance{
								Id:         endpoints[i].Id,
								Attributes: metToRet,
							},
						}, nil
					}
					f._ListServices = func(ctx context.Context, params *sd.ListServicesInput, optFns ...func(*sd.Options)) (*sd.ListServicesOutput, error) {
						Fail("should not call list services")
						return nil, nil
					}
					eop.Get(ctxtodo, &get.Options{})
				}

				e, eop, err := it.Next(ctxtodo)
				Expect(e).To(BeNil())
				Expect(eop).To(BeNil())
				Expect(err).To(MatchError(srerr.IteratorDone))
			})
		})

		Context("with some filters", func() {
			Context("with ipv4 address filter", func() {
				_endpoints := []types.InstanceSummary{
					{Id: aws.String("endp-name-1"), Attributes: map[string]string{
						"AWS_INSTANCE_IPV4": "10.10.10.10", "AWS_INSTANCE_PORT": "80",
						"key-1": "value-1",
					}},
					{Id: aws.String("endp-name-2"), Attributes: map[string]string{
						"AWS_INSTANCE_IPV6": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
						"AWS_INSTANCE_PORT": "8080",
						"key-1":             "value-1", "position": "2",
					}},
					{Id: aws.String("endp-name-3"), Attributes: map[string]string{
						"AWS_INSTANCE_IPV4": "10.10.10.12",
						"AWS_INSTANCE_IPV6": "2001:0db8:85a3:0000:0000:8a2e:0370:7335",
						"AWS_INSTANCE_PORT": "81",
						"key-1":             "value-1",
					}},
					{Id: aws.String("endp-name-4"), Attributes: map[string]string{
						"AWS_INSTANCE_IPV4": "10.10.10.13", "AWS_INSTANCE_PORT": "8181",
						"key-1": "value-1", "empty": "",
					}},
				}
				It("should return the correct ipv4", func() {
					f._ListInstances = func(ctx context.Context, params *sd.ListInstancesInput, optFns ...func(*sd.Options)) (*sd.ListInstancesOutput, error) {
						return &sd.ListInstancesOutput{Instances: _endpoints}, nil
					}
					it := w.Namespace(*ns.Name).
						Service(*serv.Name).
						Endpoint("").List(&list.Options{
						AddressFilters: &list.AddressFilters{
							AddressFamily: list.IPv4AddressFamily,
						},
					})

					elems := []int{0, 2, 3}
					for _, i := range elems {
						e, _, _ := it.Next(ctxtodo)
						Expect(e.Name).To(Equal(*_endpoints[i].Id))
						// No need to check the other fields because they
						// have been tested above.
					}

					_, _, err := it.Next(ctxtodo)
					Expect(err).To(MatchError(srerr.IteratorDone))
				})
				It("should return the correct ipv6", func() {
					f._ListInstances = func(ctx context.Context, params *sd.ListInstancesInput, optFns ...func(*sd.Options)) (*sd.ListInstancesOutput, error) {
						return &sd.ListInstancesOutput{Instances: _endpoints}, nil
					}
					it := w.Namespace(*ns.Name).
						Service(*serv.Name).
						Endpoint("").List(&list.Options{
						AddressFilters: &list.AddressFilters{
							AddressFamily: list.IPv6AddressFamily,
						},
					})

					elems := []int{1, 2}
					for _, i := range elems {
						e, _, _ := it.Next(ctxtodo)
						Expect(e.Name).To(Equal(*_endpoints[i].Id))
						// No need to check the other fields because they
						// have been tested above.
					}

					_, _, err := it.Next(ctxtodo)
					Expect(err).To(MatchError(srerr.IteratorDone))
				})
			})
		})

		Context("in case of errors", func() {
			It("should return the same error", func() {
				By("checking if parents name are defined", func() {
					e, eop, err := w.Namespace("").
						Service("defined").Endpoint("").List(&list.Options{}).
						Next(ctxtodo)

					Expect(e).To(BeNil())
					Expect(eop).To(BeNil())
					Expect(err).To(MatchError(srerr.MissingName))

					e, eop, err = w.Namespace("defined").
						Service("").Endpoint("").List(&list.Options{}).
						Next(ctxtodo)

					Expect(e).To(BeNil())
					Expect(eop).To(BeNil())
					Expect(err).To(MatchError(srerr.MissingName))
				})
				By("checking if parents exist", func() {
					e, eop, err := w.Namespace("not-exists").
						Service(*serv.Name).Endpoint("").List(&list.Options{}).
						Next(ctxtodo)

					Expect(e).To(BeNil())
					Expect(eop).To(BeNil())
					Expect(err).To(MatchError(srerr.NamespaceNotFound))

					e, eop, err = w.Namespace(*ns.Name).
						Service("not-exists").Endpoint("").List(&list.Options{}).
						Next(ctxtodo)

					Expect(e).To(BeNil())
					Expect(eop).To(BeNil())
					Expect(err).To(MatchError(srerr.ServiceNotFound))
				})
				By("checking that no two consecutive calls are made to cloudmap", func() {
					it := w.Namespace("doesnt-exist").
						Service(*serv.Name).
						Endpoint("").
						List(&list.Options{})
					_, _, err := it.Next(ctxtodo)
					Expect(err).NotTo(MatchError(srerr.IteratorDone))
					_, _, err = it.Next(ctxtodo)
					Expect(err).To(MatchError(srerr.IteratorDone))
				})
				By("checking the error from ListInstances", func() {
					expErr := fmt.Errorf("whatever")
					f._ListInstances = func(ctx context.Context, params *sd.ListInstancesInput, optFns ...func(*sd.Options)) (*sd.ListInstancesOutput, error) {
						return nil, expErr
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
					e, eop, err := w.Namespace(*ns.Name).
						Service(*serv.Name).
						Endpoint("").
						List(&list.Options{}).Next(ctxtodo)
					Expect(e).To(BeNil())
					Expect(eop).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
		})
	})

	Describe("Retrieving an endpoint", func() {
		It("should should get it successfully", func() {
			eop := w.Namespace(*ns.Name).Service(*serv.Name).
				Endpoint(*endp.Id)
			f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
				Expect(params).To(Equal(&sd.GetInstanceInput{
					InstanceId: endp.Id,
					ServiceId:  serv.Id,
				}))
				return &sd.GetInstanceOutput{
					Instance: &types.Instance{
						Id:         endp.Id,
						Attributes: endp.Attributes,
					},
				}, nil
			}
			e, err := eop.Get(ctxtodo, &get.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(e).To(Equal(&coretypes.Endpoint{
				Name:      *endp.Id,
				Namespace: *ns.Name,
				Service:   *serv.Name,
				Metadata:  endpMetas,
				Address:   ip,
				Port:      int32(port),
				OriginalObject: &types.Instance{
					Id:         endp.Id,
					Attributes: endp.Attributes,
				},
			}))

			f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
				Fail("should get it from cache not cloud map")
				return nil, nil
			}
			e, err = eop.Get(ctxtodo, &get.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(e).To(Equal(&coretypes.Endpoint{
				Name:      *endp.Id,
				Namespace: *ns.Name,
				Service:   *serv.Name,
				Metadata:  endpMetas,
				Address:   ip,
				Port:      int32(port),
				OriginalObject: &types.Instance{
					Id:         endp.Id,
					Attributes: endp.Attributes,
				},
			}))
		})

		Context("in case of errors", func() {
			It("should return the same error", func() {
				By("checking if the namespace exists", func() {
					e, err := w.Namespace("not-exists").Service(*serv.Name).
						Endpoint(*endp.Id).Get(ctxtodo, &get.Options{})
					Expect(e).To(BeNil())
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
				})
				By("checking if the service exists", func() {
					e, err := w.Namespace(*ns.Name).Service("not-exists").
						Endpoint(*endp.Id).Get(ctxtodo, &get.Options{})
					Expect(e).To(BeNil())
					Expect(err).To(MatchError(srerr.ServiceNotFound))
				})
				By("checking the error returned from GetInstance", func() {
					expErr := fmt.Errorf("whatever")
					f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
						return nil, expErr
					}
					e, err := w.Namespace(*ns.Name).Service(*serv.Name).
						Endpoint(*endp.Id).Get(ctxtodo, &get.Options{})
					Expect(e).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
		})
	})

	Describe("Deleting an endpoint", func() {
		It("should delete the endpoint successfully", func() {
			f._DeregisterInstance = func(ctx context.Context, params *sd.DeregisterInstanceInput, optFns ...func(*sd.Options)) (*sd.DeregisterInstanceOutput, error) {
				Expect(params).To(Equal(&sd.DeregisterInstanceInput{
					InstanceId: endp.Id,
					ServiceId:  serv.Id,
				}))
				return &sd.DeregisterInstanceOutput{
					OperationId: aws.String("op-id"),
				}, nil
			}
			f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
				return &sd.GetOperationOutput{
					Operation: &types.Operation{
						Status: types.OperationStatusSuccess,
					},
				}, nil
			}
			err := w.Namespace(*ns.Name).Service(*serv.Name).Endpoint(*endp.Id).Delete(ctxtodo)
			Expect(err).To(BeNil())

			called := false
			f._GetInstance = func(ctx context.Context, params *sd.GetInstanceInput, optFns ...func(*sd.Options)) (*sd.GetInstanceOutput, error) {
				called = true
				return nil, fmt.Errorf("whatever")
			}
			w.Namespace(*ns.Name).Service(*serv.Name).Endpoint(*endp.Id).Get(ctxtodo, &get.Options{})
			Expect(called).To(BeTrue())
		})

		Context("in case of errors", func() {
			It("should return the same error", func() {
				By("checking if the namespace has errors", func() {
					err := w.Namespace("whatever").
						Service("whatever").
						Endpoint(*endp.Id).
						Delete(ctxtodo)
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
				})
				By("checking service errors", func() {
					err := w.Namespace(*ns.Name).
						Service("whatever").
						Endpoint(*endp.Id).
						Delete(ctxtodo)
					Expect(err).To(MatchError(srerr.ServiceNotFound))
				})
				By("checking delete errors", func() {
					expErr := fmt.Errorf("whatever")
					f._DeregisterInstance = func(ctx context.Context, params *sd.DeregisterInstanceInput, optFns ...func(*sd.Options)) (*sd.DeregisterInstanceOutput, error) {
						return nil, expErr
					}
					err := w.Namespace(*ns.Name).
						Service(*serv.Name).
						Endpoint(*endp.Id).
						Delete(ctxtodo)
					Expect(err).To(MatchError(expErr))
				})
				By("checking operation result", func() {
					f._DeregisterInstance = func(ctx context.Context, params *sd.DeregisterInstanceInput, optFns ...func(*sd.Options)) (*sd.DeregisterInstanceOutput, error) {
						return &sd.DeregisterInstanceOutput{
							OperationId: aws.String("op-id"),
						}, nil
					}
					f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
						return &sd.GetOperationOutput{
							Operation: &types.Operation{
								Status:       types.OperationStatusFail,
								ErrorCode:    aws.String("error-code"),
								ErrorMessage: aws.String("error-message"),
							},
						}, nil
					}
					err := w.Namespace(*ns.Name).
						Service(*serv.Name).
						Endpoint(*endp.Id).
						Delete(ctxtodo)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
