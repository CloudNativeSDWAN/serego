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

package errors

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	IteratorDone                = errors.New("iterator done")
	NotFound                    = errors.New("not found")
	NamespaceNotFound           = errors.New("namespace not found")
	NamespaceAlreadyExists      = errors.New("namespace already exists")
	ServiceNotFound             = errors.New("service not found")
	ServiceAlreadyExists        = errors.New("service already exists")
	EndpointNotFound            = errors.New("endpoint not found")
	EndpointAlreadyExists       = errors.New("endpoint already exists")
	EmptyNamespaceName          = errors.New("no namespace name provided")
	EmptyServiceName            = errors.New("no service name provided")
	EmptyEndpointName           = errors.New("no endpoint name provided")
	MissingName                 = errors.New("name not found")
	UnknownRegisterMode         = errors.New("unknown register mode")
	NamespaceNotEmpty           = errors.New("namespace is not empty")
	InvalidPort                 = errors.New("invalid port")
	NoPortsProvided             = errors.New("no ports provided")
	InvalidAddress              = errors.New("invalid address provided")
	InvalidNamePrefixFilter     = errors.New("invalid name prefix filter provided")
	IncompatibleNameFilters     = errors.New(`"NameIn" and "NamePrefix" filters cannot be used together`)
	EmptyName                   = errors.New("empty name provided")
	EmptyNameInFilter           = errors.New("empty nameIn filter provided")
	EmptyMetadataKeysFilter     = errors.New("empty metadata keys filter provided")
	EmptyMetadataKeyValueFilter = errors.New("empty metadata key value filter provided")
	EmptyMetadataKey            = errors.New("metadata contains an empty key")
	EmptyMetadataFilter         = errors.New("empty metadata filter provided")
	InvalidResultsNumber        = errors.New("invalid results number provided")
	IncompatibleMetadataFilters = errors.New("'noMetadata' and 'metadata' filters cannot be used together'")
	UnsupportedAddressFamily    = errors.New("invalid address family provided")
	InvalidCIDRProvided         = errors.New("invalid CIDR provided")
	IncompatiblePortFilters     = errors.New("'portIn' and 'portRange' cannot be used together")
	EmptyPortInFilter           = errors.New("empty 'portIn' filter provided")
	InvalidPortRange            = errors.New("invalid port range provided")
	NoOptionsProvided           = errors.New("no options provided")
	NoOperationSet              = errors.New("no operation set")
	InvalidCacheExpirationTime  = errors.New("invalid cache expiration time provided")
	NameTooLong                 = errors.New("name is longer than 63 characters")
	NameIsNotRFC1035            = errors.New("name is not complaint to RFC1035")
	IncompatibleAddressFilters  = errors.New(`"AddressFamily" and "CIDR" cannot be used together`)
	NoWrapperProvided           = errors.New("no wrapper provided")
	NoClientProvided            = errors.New("no client provided")
	NoSettingsProvided          = errors.New("no settings provided")
	NoProjectIDSet              = errors.New("no project ID set")
	NoLocationSet               = errors.New("no default location set")
	UninitializedOperation      = errors.New("operation not initialized")
	InvalidObjectToFilter       = errors.New("object to filter is invalid")
	InvalidRegionProvided       = errors.New("empty or invalid region provided")
	InvalidProjectProvided      = errors.New("empty or invalid project provided")
)

// IsIteratorDone returns true if the error provided as argument is
// because the iterator has iterated through all elements already.
//
// If false, then some other error was thrown by the service registry
// which you should either check with another function from this package
// or by referring to the documentation of the underlying service registry
// that you are using with the core package.
func IsIteratorDone(err error) bool {
	if err == nil {
		// Finished unwrapping
		return false
	}

	if err == IteratorDone || err == iterator.Done {
		return true
	}

	return IsIteratorDone(errors.Unwrap(err))
}

// IsNotFound returns true if the error provided as argument was thrown by the
// service registry because the error does not exist there.
//
// If false, then some other error was thrown by the service registry
// which you should either check with another function from this package
// or by referring to the documentation of the underlying service registry
// that you are using with the core package.
func IsNotFound(err error) bool {
	if err == nil {
		// Finished unwrapping
		return false
	}

	// Our internal errors
	if errors.Is(err, NotFound) ||
		errors.Is(err, NamespaceNotFound) ||
		errors.Is(err, ServiceNotFound) ||
		errors.Is(err, EndpointNotFound) {
		return true
	}

	// Service Directory gRPC errors.
	if status.Code(err) == codes.NotFound {
		return true
	}

	// Cloud Map Errors
	{
		var (
			rnf *types.ResourceNotFoundException
			nnf *types.NamespaceNotFound
			snf *types.ServiceNotFound
			inf *types.InstanceNotFound
		)

		if errors.As(err, &rnf) ||
			errors.As(err, &nnf) ||
			errors.As(err, &snf) ||
			errors.As(err, &inf) {
			return true
		}
	}

	// Etcd errors
	if errors.Is(err, rpctypes.ErrGRPCKeyNotFound) {
		return true
	}

	return IsNotFound(errors.Unwrap(err))
}

// IsPermissionsError returns true if the error provided as argument was
// thrown by the service registry because of insufficient errors.
func IsPermissionsError(err error) bool {
	if err == nil {
		// Finished unwrapping
		return false
	}

	if status.Code(err) == codes.PermissionDenied {
		return true
	}

	// TODO: other situations

	return IsPermissionsError(errors.Unwrap(err))
}

// IsAlreadyExists returns true if the error provided as argument was
// thrown by the service registry because the object already exists.
func IsAlreadyExists(err error) bool {
	if err == nil {
		// Finished unwrapping
		return false
	}

	if errors.Is(err, NamespaceAlreadyExists) ||
		errors.Is(err, ServiceAlreadyExists) ||
		errors.Is(err, EndpointAlreadyExists) {
		return true
	}

	// Cloud Map Errors
	{
		var (
			nae *types.NamespaceAlreadyExists
			sae *types.ServiceAlreadyExists
		)

		if errors.As(err, &nae) ||
			errors.As(err, &sae) {
			return true
		}
	}

	// Service Directory gRPC errors.
	if status.Code(err) == codes.AlreadyExists {
		return true
	}

	return IsAlreadyExists(errors.Unwrap(err))
}
