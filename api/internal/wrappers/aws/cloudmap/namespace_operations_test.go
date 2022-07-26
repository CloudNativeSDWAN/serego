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

var _ = Describe("Namespace Operations", func() {
	const (
		nsOpID = "ns-op-id"
	)

	var (
		f       *fakeCloudMapClient
		w       *cloudmap.AwsCloudMapWrapper
		ns      types.NamespaceSummary
		nsTags  []types.Tag
		nsMetas map[string]string
	)

	BeforeEach(func() {
		f = &fakeCloudMapClient{}
		w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: wrapper.DefaultCacheExpirationTime})
		ns = namespaces[0]
		nsTags = tags[0]
		nsMetas = metas[0]
		f._ListNamespaces = listNamespaces
		f._ListTagsForResource = listTagsForResource
	})

	Describe("Creating a namespace", func() {
		It("should call Cloud Map with the correct parameters", func() {
			f._CreateHttpNamespace = func(ctx context.Context, params *sd.CreateHttpNamespaceInput, optFns ...func(*sd.Options)) (*sd.CreateHttpNamespaceOutput, error) {
				Expect(params.Name).To(Equal(ns.Name))
				Expect(params.Tags).To(ConsistOf(nsTags))
				return &sd.CreateHttpNamespaceOutput{
					OperationId: aws.String(nsOpID),
				}, nil
			}
			f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
				Expect(params).To(Equal(&sd.GetOperationInput{
					OperationId: aws.String(nsOpID),
				}))
				return &sd.GetOperationOutput{
					Operation: &types.Operation{
						Status:  types.OperationStatusSuccess,
						Targets: map[string]string{"NAMESPACE": *ns.Id},
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
			f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
				Expect(params).To(Equal(&sd.ListTagsForResourceInput{
					ResourceARN: ns.Arn,
				}))
				return &sd.ListTagsForResourceOutput{
					Tags: nsTags,
				}, nil
			}

			createdNs, err := w.Namespace(*ns.Name).Create(context.Background(), nsMetas)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdNs).To(Equal(&coretypes.Namespace{
				Name:     *ns.Name,
				Metadata: nsMetas,
				OriginalObject: &types.Namespace{
					Arn:  ns.Arn,
					Id:   ns.Id,
					Name: ns.Name,
				},
			}))

			f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
				Fail("got from cloud map instead of cache")
				return nil, nil
			}

			createdNs, err = w.Namespace(*ns.Name).Get(context.Background(), &get.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(createdNs).To(Equal(&coretypes.Namespace{
				Name:     *ns.Name,
				Metadata: nsMetas,
				OriginalObject: &types.Namespace{
					Arn:  ns.Arn,
					Id:   ns.Id,
					Name: ns.Name,
				},
			}))
		})

		Context("when errors happen", func() {
			It("should return the error", func() {
				By("checking the creation of the namespace", func() {
					f._CreateHttpNamespace = func(ctx context.Context, params *sd.CreateHttpNamespaceInput, optFns ...func(*sd.Options)) (*sd.CreateHttpNamespaceOutput, error) {
						return nil, fmt.Errorf("whatever")
					}

					ns, err := w.Namespace(*ns.Name).
						Create(context.Background(), nsMetas)
					Expect(err).To(HaveOccurred())
					Expect(ns).To(BeNil())
				})
				By("checking error in get operation", func() {
					expErr := fmt.Errorf("whatever-get-operation")
					f._CreateHttpNamespace = func(ctx context.Context, params *sd.CreateHttpNamespaceInput, optFns ...func(*sd.Options)) (*sd.CreateHttpNamespaceOutput, error) {
						return &sd.CreateHttpNamespaceOutput{
							OperationId: aws.String(nsOpID),
						}, nil
					}
					f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
						return nil, expErr
					}
					ns, err := w.Namespace(*ns.Name).
						Create(context.Background(), nsMetas)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(expErr))
					Expect(ns).To(BeNil())
				})
				By("checking operation status", func() {
					f._CreateHttpNamespace = func(ctx context.Context, params *sd.CreateHttpNamespaceInput, optFns ...func(*sd.Options)) (*sd.CreateHttpNamespaceOutput, error) {
						return &sd.CreateHttpNamespaceOutput{
							OperationId: aws.String(nsOpID),
						}, nil
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
					ns, err := w.Namespace(*ns.Name).
						Create(context.Background(), nsMetas)
					Expect(err).To(HaveOccurred())
					Expect(ns).To(BeNil())
				})
			})
		})

	})

	Describe("Listing namespaces", func() {
		BeforeEach(func() {
			w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: 0})
		})
		Context("with no filters provided", func() {
			It("should return all namespaces", func() {
				resToGet := int32(2)
				reqs := []*sd.ListNamespacesInput{}
				it := w.Namespace("").List(&list.Options{
					Results: resToGet,
				})
				n := 0
				f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
					reqs = append(reqs, params)
					defer func() {
						n++
					}()
					switch n {
					case 0:
						return &sd.ListNamespacesOutput{
							Namespaces: namespaces[:2],
							NextToken:  aws.String("next-1"),
						}, nil
					case 1:
						return &sd.ListNamespacesOutput{
							Namespaces: namespaces[2:],
							NextToken:  aws.String("next-2"),
						}, nil
					default:
						return &sd.ListNamespacesOutput{}, nil
					}
				}

				for i := 0; i < len(namespaces); i++ {
					val, valOp, err := it.Next(context.Background())

					Expect(err).NotTo(HaveOccurred())
					Expect(val.Name).To(Equal(*namespaces[i].Name))
					Expect(val.Metadata).To(Equal(metas[i]))
					Expect(val.OriginalObject).To(Equal(&types.Namespace{
						Arn:  namespaces[i].Arn,
						Name: namespaces[i].Name,
						Id:   namespaces[i].Id,
					}))
					f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
						// Here to check that it has ID
						Expect(params).To(Equal(&sd.GetNamespaceInput{
							Id: namespaces[i].Id,
						}))
						return &sd.GetNamespaceOutput{
							Namespace: &types.Namespace{
								Arn:  namespaces[i].Arn,
								Name: namespaces[i].Name,
								Id:   namespaces[i].Id,
							},
						}, nil
					}
					valOp.Get(context.Background(), &get.Options{})
				}

				val, valOp, err := it.Next(context.Background())
				Expect(err).To(Equal(srerr.IteratorDone))
				Expect(val).To(BeNil())
				Expect(valOp).To(BeNil())

				Expect(reqs).To(HaveLen(3))
				Expect(reqs[0]).To(Equal(&sd.ListNamespacesInput{
					MaxResults: aws.Int32(resToGet),
				}))
				Expect(reqs[1]).To(Equal(&sd.ListNamespacesInput{
					MaxResults: aws.Int32(resToGet),
					NextToken:  aws.String("next-1"),
				}))
				Expect(reqs[2]).To(Equal(&sd.ListNamespacesInput{
					MaxResults: aws.Int32(resToGet),
					NextToken:  aws.String("next-2"),
				}))
			})
		})

		Context("with name filters", func() {
			It("should only return requested namespaces", func() {
				it := w.Namespace("ns-name-1").List(&list.Options{
					NameFilters: &list.NameFilters{
						In: []string{"ns-name-2", "ns-name-5"},
					},
				})
				timesCalled := 0
				f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
					defer func() {
						timesCalled++
					}()
					if timesCalled > 0 {
						Fail("called more than once")
					}
					Expect(params).To(Equal(&sd.ListNamespacesInput{
						MaxResults: aws.Int32(list.DefaultListResultsNumber),
					}))
					return &sd.ListNamespacesOutput{
						Namespaces: namespaces,
					}, nil
				}

				expectedLoops := 2
				for i := 0; i < expectedLoops; i++ {
					val, valOp, err := it.Next(context.Background())

					Expect(err).NotTo(HaveOccurred())
					Expect(val.Name).To(Equal(*namespaces[i].Name))
					Expect(val.Metadata).To(Equal(metas[i]))
					Expect(val.OriginalObject).To(Equal(&types.Namespace{
						Arn:  namespaces[i].Arn,
						Name: namespaces[i].Name,
						Id:   namespaces[i].Id,
					}))
					f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
						// Here to check that it has ID
						Expect(params).To(Equal(&sd.GetNamespaceInput{
							Id: namespaces[i].Id,
						}))
						return &sd.GetNamespaceOutput{
							Namespace: &types.Namespace{
								Arn:  namespaces[i].Arn,
								Name: namespaces[i].Name,
								Id:   namespaces[i].Id,
							},
						}, nil
					}
					valOp.Get(context.Background(), &get.Options{})
				}

				val, valOp, err := it.Next(context.Background())
				Expect(err).To(Equal(srerr.IteratorDone))
				Expect(val).To(BeNil())
				Expect(valOp).To(BeNil())
			})
		})

		Context("in case of errors", func() {
			Context("during namespace check", func() {
				It("ignores non-existing namespaces", func() {
					it := w.Namespace("ns-name-1").List(&list.Options{})
					f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
						return &sd.ListNamespacesOutput{
							Namespaces: namespaces,
						}, nil
					}
					f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
						switch aws.ToString(params.ResourceARN) {
						case "ns-arn-1":
							return nil, &types.ResourceNotFoundException{}
						default:
							return nil, &types.InvalidInput{}
						}
					}

					val, valOp, err := it.Next(context.Background())
					Expect(err).To(Equal(srerr.IteratorDone))
					Expect(val).To(BeNil())
					Expect(valOp).To(BeNil())
				})
			})

			Context("during listing", func() {
				It("stops iteration", func() {
					e := &types.InvalidInput{}
					it := w.Namespace("").List(&list.Options{})
					f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
						return nil, e
					}

					val, valOp, err := it.Next(context.Background())
					Expect(err).To(MatchError(e))
					Expect(val).To(BeNil())
					Expect(valOp).To(BeNil())
				})
			})
		})
	})

	Describe("Retrieving a namespace", func() {
		It("should get it...", func() {
			nsop := w.Namespace(*ns.Name)
			timesLookedForTags := 0
			f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
				timesLookedForTags++
				Expect(timesLookedForTags).To(Equal(1))
				Expect(params).NotTo(BeNil())
				Expect(params).To(Equal(&sd.ListTagsForResourceInput{
					ResourceARN: ns.Arn,
				}))
				return &sd.ListTagsForResourceOutput{
					Tags: nsTags,
				}, nil
			}

			By("... first from the list...", func() {
				timesListed := 0
				timesLookedForTags = 0
				f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
					timesListed++
					Expect(timesListed).To(Equal(1))
					Expect(params).NotTo(BeNil())
					Expect(params).To(Equal(&sd.ListNamespacesInput{
						MaxResults: aws.Int32(list.DefaultListResultsNumber),
					}))
					return &sd.ListNamespacesOutput{
						Namespaces: namespaces,
					}, nil
				}

				op, err := nsop.Get(context.Background(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(op).To(Equal(&coretypes.Namespace{
					Name:     *ns.Name,
					Metadata: nsMetas,
					OriginalObject: &types.Namespace{
						Arn:  ns.Arn,
						Id:   ns.Id,
						Name: ns.Name,
					},
				}))
			})

			By("... then from cache", func() {
				f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
					Fail("should not call ListNamespaces")
					return nil, nil
				}
				f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
					Fail("should not call GetNamespace")
					return nil, nil
				}
				f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
					Fail("should not call ListTagsForResource")
					return nil, nil
				}

				op, err := nsop.Get(context.Background(), &get.Options{})
				Expect(err).NotTo(HaveOccurred())
				Expect(op).To(Equal(&coretypes.Namespace{
					Name:     *ns.Name,
					Metadata: nsMetas,
					OriginalObject: &types.Namespace{
						Arn:  ns.Arn,
						Id:   ns.Id,
						Name: ns.Name,
					},
				}))
			})
		})

		Context("without cache", func() {
			It("only gets the ID from cache", func() {
				w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: 0})
				nsop := w.Namespace(*ns.Name)
				f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
					return &sd.ListTagsForResourceOutput{
						Tags: nsTags,
					}, nil
				}

				f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
					return &sd.ListNamespacesOutput{
						Namespaces: namespaces,
					}, nil
				}

				nsop.Get(context.Background(), &get.Options{})

				f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
					Fail("should not call ListNamespaces")
					return nil, nil
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
				nsop.Get(context.Background(), &get.Options{})
			})
		})

		Context("when errors happen", func() {
			BeforeEach(func() {
				f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
					switch aws.ToString(params.ResourceARN) {
					case "ns-arn-1":
						return &sd.ListTagsForResourceOutput{
							Tags: tags[0],
						}, nil
					case "ns-arn-2":
						return &sd.ListTagsForResourceOutput{
							Tags: tags[1],
						}, nil
					case "ns-arn-3":
						return &sd.ListTagsForResourceOutput{
							Tags: tags[2],
						}, nil
					case "ns-arn-4":
						return &sd.ListTagsForResourceOutput{
							Tags: tags[3],
						}, nil
					default:
						return nil, &types.ResourceNotFoundException{}
					}
				}
				w, _ = cloudmap.NewCloudMapWrapper(f, &wrapper.Options{CacheExpirationTime: 0})
			})

			Context("during searching", func() {
				It("returns the error", func() {
					nsop := w.Namespace(*ns.Name)
					fakeError := &types.InvalidInput{}
					f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
						return nil, fakeError
					}

					gotNs, err := nsop.Get(context.Background(), &get.Options{})
					Expect(gotNs).To(BeNil())
					Expect(err).To(MatchError(fakeError))
				})
				It("returns namespace not found", func() {
					nsop := w.Namespace(*ns.Name)
					f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
						return nil, srerr.IteratorDone
					}

					gotNs, err := nsop.Get(context.Background(), &get.Options{})
					Expect(gotNs).To(BeNil())
					Expect(err).To(MatchError(srerr.NamespaceNotFound))
				})
			})

			Context("during getting by ID", func() {
				It("returns the error", func() {
					nsop := w.Namespace(*ns.Name)
					f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
						return &sd.ListNamespacesOutput{
							Namespaces: namespaces,
						}, nil
					}
					nsop.Get(context.Background(), &get.Options{})
					notFound := &types.NamespaceNotFound{}
					By("... checking the list tags operation", func() {
						f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
							return &sd.GetNamespaceOutput{
								Namespace: &types.Namespace{
									Arn:  ns.Arn,
									Id:   ns.Id,
									Name: ns.Name,
								},
							}, nil
						}
						expErr := &types.ResourceNotFoundException{}
						f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
							return nil, expErr
						}
						ns, err := nsop.Get(context.Background(), &get.Options{})
						Expect(ns).To(BeNil())
						Expect(err).To(MatchError(expErr))
					})
					By("... checking the get operation", func() {
						f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
							return nil, notFound
						}
						ns, err := nsop.Get(context.Background(), &get.Options{})
						Expect(ns).To(BeNil())
						Expect(err).To(MatchError(notFound))
					})
				})
			})
		})
	})

	Describe("Deleting a namespace", func() {
		fakeOpID := "abc123"
		It("deletes the namespace successfully", func() {
			f._DeleteNamespace = func(ctx context.Context, params *sd.DeleteNamespaceInput, optFns ...func(*sd.Options)) (*sd.DeleteNamespaceOutput, error) {
				Expect(params).To(Equal(&sd.DeleteNamespaceInput{
					Id: ns.Id,
				}))
				return &sd.DeleteNamespaceOutput{
					OperationId: &fakeOpID,
				}, nil
			}
			f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
				Expect(params).To(Equal(&sd.GetOperationInput{
					OperationId: &fakeOpID,
				}))
				return &sd.GetOperationOutput{
					Operation: &types.Operation{
						Status: types.OperationStatusSuccess,
					},
				}, nil
			}
			err := w.Namespace(*ns.Name).Delete(context.Background())
			Expect(err).NotTo(HaveOccurred())
			called := false
			f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
				called = true
				return nil, fmt.Errorf("whatever")
			}
			w.Namespace(*ns.Name).Get(context.Background(), &get.Options{})
			Expect(called).To(BeTrue())
		})

		Context("in case of errors", func() {
			Context("from get", func() {
				It("forwards the same error", func() {
					expErr := &types.InvalidInput{}
					f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
						return nil, expErr
					}
					err := w.Namespace("whatever-name").Delete(context.Background())
					Expect(err).To(MatchError(expErr))
				})
			})
			Context("from delete", func() {
				It("forwards the same error", func() {
					expErr := &types.InvalidInput{}
					f._DeleteNamespace = func(ctx context.Context, params *sd.DeleteNamespaceInput, optFns ...func(*sd.Options)) (*sd.DeleteNamespaceOutput, error) {
						return nil, expErr
					}
					err := w.Namespace(*ns.Name).Delete(context.Background())
					Expect(err).To(MatchError(expErr))
				})
			})
			Context("from delete", func() {
				It("forwards the same error", func() {
					f._DeleteNamespace = func(ctx context.Context, params *sd.DeleteNamespaceInput, optFns ...func(*sd.Options)) (*sd.DeleteNamespaceOutput, error) {
						Expect(params).To(Equal(&sd.DeleteNamespaceInput{
							Id: ns.Id,
						}))
						return &sd.DeleteNamespaceOutput{
							OperationId: &fakeOpID,
						}, nil
					}
					f._GetOperation = func(ctx context.Context, params *sd.GetOperationInput, optFns ...func(*sd.Options)) (*sd.GetOperationOutput, error) {
						Expect(params).To(Equal(&sd.GetOperationInput{
							OperationId: &fakeOpID,
						}))
						return &sd.GetOperationOutput{
							Operation: &types.Operation{
								Status:       types.OperationStatusFail,
								ErrorCode:    aws.String("ACCESS_DENIED"),
								ErrorMessage: aws.String("whatever"),
							},
						}, nil
					}
					err := w.Namespace(*ns.Name).Delete(context.Background())
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("Updating a namespace", func() {
		It("updates tags correctly", func() {
			newTags := []types.Tag{
				{Key: nsTags[0].Key, Value: aws.String("value-1-edited")},
				{Key: aws.String("new-key"), Value: aws.String("new-val")},
			}
			f._TagResource = func(ctx context.Context, params *sd.TagResourceInput, optFns ...func(*sd.Options)) (*sd.TagResourceOutput, error) {
				Expect(params.ResourceARN).To(Equal(ns.Arn))
				Expect(params.Tags).To(ConsistOf(newTags))
				return &sd.TagResourceOutput{}, nil
			}
			f._UntagResource = func(ctx context.Context, params *sd.UntagResourceInput, optFns ...func(*sd.Options)) (*sd.UntagResourceOutput, error) {
				Expect(params).To(Equal(&sd.UntagResourceInput{
					ResourceARN: ns.Arn,
					TagKeys:     []string{*nsTags[1].Key},
				}))
				return &sd.UntagResourceOutput{}, nil
			}
			expNs := &types.Namespace{
				Arn:  ns.Arn,
				Id:   ns.Id,
				Name: ns.Name,
			}
			f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
				return &sd.GetNamespaceOutput{
					Namespace: expNs,
				}, nil
			}
			oldTagsRes := f.ListTagsForResource
			defer func() {
				f._ListTagsForResource = oldTagsRes
			}()
			f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
				return &sd.ListTagsForResourceOutput{
					Tags: newTags,
				}, nil
			}

			newMaps := map[string]string{"key-1": "value-1-edited", "new-key": "new-val"}
			updNs, err := w.Namespace(*ns.Name).Update(context.Background(), newMaps)
			Expect(err).To(BeNil())
			Expect(updNs).To(Equal(&coretypes.Namespace{
				Name:           *ns.Name,
				Metadata:       newMaps,
				OriginalObject: expNs,
			}))

			f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
				Fail("should get it from cache not cloud map")
				return nil, nil
			}
			f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
				Fail("should get it from cache not cloud map")
				return nil, nil
			}
			w.Namespace(*ns.Name).Get(context.Background(), &get.Options{})
		})
		Context("in case of errors", func() {
			expErr := &types.InvalidInput{}
			Context("during get", func() {
				It("returns the error", func() {
					f._ListNamespaces = func(ctx context.Context, params *sd.ListNamespacesInput, optFns ...func(*sd.Options)) (*sd.ListNamespacesOutput, error) {
						return nil, expErr
					}
					ns, err := w.Namespace("whatever").Update(context.Background(), map[string]string{})
					Expect(ns).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
			Context("during tag insertion", func() {
				It("returns the error", func() {
					f._TagResource = func(ctx context.Context, params *sd.TagResourceInput, optFns ...func(*sd.Options)) (*sd.TagResourceOutput, error) {
						return nil, expErr
					}
					ns, err := w.Namespace(*namespaces[0].Name).Update(context.Background(), map[string]string{
						"key": "val",
					})
					Expect(ns).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
			Context("during tag list", func() {
				It("returns the error", func() {
					f._TagResource = func(ctx context.Context, params *sd.TagResourceInput, optFns ...func(*sd.Options)) (*sd.TagResourceOutput, error) {
						return &sd.TagResourceOutput{}, nil
					}
					f._ListTagsForResource = func(ctx context.Context, params *sd.ListTagsForResourceInput, optFns ...func(*sd.Options)) (*sd.ListTagsForResourceOutput, error) {
						return nil, expErr
					}
					ns, err := w.Namespace(*namespaces[0].Name).Update(context.Background(), map[string]string{
						"key": "val",
					})
					Expect(ns).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
			Context("during untag", func() {
				It("returns the error", func() {
					f._TagResource = func(ctx context.Context, params *sd.TagResourceInput, optFns ...func(*sd.Options)) (*sd.TagResourceOutput, error) {
						return &sd.TagResourceOutput{}, nil
					}
					f._UntagResource = func(ctx context.Context, params *sd.UntagResourceInput, optFns ...func(*sd.Options)) (*sd.UntagResourceOutput, error) {
						return nil, expErr
					}
					ns, err := w.Namespace(*namespaces[0].Name).Update(context.Background(), map[string]string{
						"key": "val",
					})
					Expect(ns).To(BeNil())
					Expect(err).To(MatchError(expErr))
				})
			})
			Context("during get final state", func() {
				It("returns the error", func() {
					f._TagResource = func(ctx context.Context, params *sd.TagResourceInput, optFns ...func(*sd.Options)) (*sd.TagResourceOutput, error) {
						return &sd.TagResourceOutput{}, nil
					}
					f._UntagResource = func(ctx context.Context, params *sd.UntagResourceInput, optFns ...func(*sd.Options)) (*sd.UntagResourceOutput, error) {
						return &sd.UntagResourceOutput{}, nil
					}
					nsExpErr := &types.NamespaceNotFound{}
					f._GetNamespace = func(ctx context.Context, params *sd.GetNamespaceInput, optFns ...func(*sd.Options)) (*sd.GetNamespaceOutput, error) {
						return nil, nsExpErr
					}
					ns, err := w.Namespace(*namespaces[0].Name).Update(context.Background(), map[string]string{
						"key": "val",
					})
					Expect(ns).To(BeNil())
					Expect(err).To(MatchError(nsExpErr))
				})
			})
		})
	})
})
