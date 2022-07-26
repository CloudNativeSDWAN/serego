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

package cloudmap

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
)

const (
	defaultPollTick     time.Duration = 2 * time.Second
	getOperationTimeout time.Duration = 3 * time.Second
)

var (
	PollTick = defaultPollTick
)

func fromSummaryToNamespace(summary *types.NamespaceSummary) *types.Namespace {
	return &types.Namespace{
		Arn:          summary.Arn,
		CreateDate:   summary.CreateDate,
		Description:  summary.Description,
		Id:           summary.Id,
		Name:         summary.Name,
		Properties:   summary.Properties,
		ServiceCount: summary.ServiceCount,
		Type:         summary.Type,
	}
}

func fromSummaryToService(summary *types.ServiceSummary) *types.Service {
	return &types.Service{
		Arn:                     summary.Arn,
		CreateDate:              summary.CreateDate,
		Description:             summary.Description,
		DnsConfig:               summary.DnsConfig,
		HealthCheckConfig:       summary.HealthCheckConfig,
		HealthCheckCustomConfig: summary.HealthCheckCustomConfig,
		Id:                      summary.Id,
		InstanceCount:           summary.InstanceCount,
		Name:                    summary.Name,
		Type:                    summary.Type,
	}
}

func fromSummaryToInstance(summary *types.InstanceSummary) *types.Instance {
	return &types.Instance{
		Id:         summary.Id,
		Attributes: summary.Attributes,
	}
}

func fromMapToTagsSlice(metadata map[string]string) []types.Tag {
	tags := []types.Tag{}

	for k, v := range metadata {
		tags = append(tags, types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return tags
}

func fromTagsSliceToMap(tags []types.Tag) map[string]string {
	metadata := map[string]string{}

	for _, t := range tags {
		metadata[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}

	return metadata
}

func pollOperationStatus(ctx context.Context, client cloudMapClientIface, operationID string) (*types.Operation, error) {
	ticker := time.NewTicker(defaultPollTick)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			opCtx, opCanc := context.WithTimeout(ctx, getOperationTimeout)
			op, err := client.GetOperation(opCtx, &servicediscovery.GetOperationInput{OperationId: aws.String(operationID)})
			if err != nil {
				opCanc()
				return nil, err
			}
			opCanc()

			switch op.Operation.Status {
			case types.OperationStatusPending, types.OperationStatusSubmitted:
				continue
			case types.OperationStatusFail:
				return nil, fmt.Errorf("operation failed: %s", aws.ToString(op.Operation.ErrorMessage))
			default:
				return op.Operation, nil
			}
		}
	}
}

func updateTags(ctx context.Context, client cloudMapClientIface, arn string, metadata map[string]string) error {
	if len(metadata) > 0 {
		// First, we add/modify all the ones that need to be inserted, so that
		// should the untag operation fail at least we have the ones we wanted.
		if _, err := client.TagResource(ctx, &servicediscovery.TagResourceInput{
			ResourceARN: &arn,
			Tags:        fromMapToTagsSlice(metadata),
		}); err != nil {
			return fmt.Errorf("error while updating tags: %w,", err)
		}
	}

	// Now we have to remove tags that are not there anymore, as TagResource is
	// always incremental.
	existing, err := client.ListTagsForResource(ctx, &servicediscovery.ListTagsForResourceInput{
		ResourceARN: &arn,
	})
	if err != nil {
		return fmt.Errorf("error while getting existing tags: %w", err)
	}

	tagsToRemove := []string{}
	for _, tag := range existing.Tags {
		if _, exists := metadata[aws.ToString(tag.Key)]; !exists {
			tagsToRemove = append(tagsToRemove, aws.ToString(tag.Key))
		}
	}

	if len(tagsToRemove) > 0 {
		if _, err := client.UntagResource(ctx, &servicediscovery.UntagResourceInput{
			ResourceARN: &arn,
			TagKeys:     tagsToRemove,
		}); err != nil {
			return fmt.Errorf("error while untagging resource: %w", err)
		}
	}

	return nil
}
