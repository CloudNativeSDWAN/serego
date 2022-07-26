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

package list

import (
	"fmt"
	"math"
	"net"
	"syscall"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
)

const (
	// DefaultListResultsNumber is the number of results to retrieve when
	// getting a list of objects per page. It can be overridden in list
	// options.
	DefaultListResultsNumber int32 = 50

	minPortNumber int32 = 1
	maxPortNumber int32 = math.MaxUint16
)

// NameFilters contains filters to use for filtering an object based on its
// name.
type NameFilters struct {
	// In is a list of names of objects that you want to get.
	// It cannot be used together with the Prefix filter.
	//
	// Example:
	// 	[]string{"sales", "production", "hr"}
	In []string
	// Prefix that the object names must have.
	// It cannot be used together with the In filter.
	//
	// Examples: "prod-" or "building_",
	Prefix string
}

// MetadataFilters contains filters to use for filtering an object based on its
// metadata.
type MetadataFilters struct {
	// NoMetadata instructs the List function to only return objects that don't
	// have any metadata.
	//
	// It cannot be used together with the Metadata filter.
	NoMetadata bool
	// Metadata that the object must have.
	// It cannot be used together with the NoMetadata filter.
	//
	// Example:
	// 	map[string]string{"env": "prod", "maintainer": "team-24@company.com"}
	Metadata map[string]string
}

// AddressFilters contains filters to use for filtering an *endpoint* based on its
// address.
type AddressFilters struct {
	// CIDR instructs List to only get endpoints whose address is included in
	// the given subnet.
	// It cannot be used together with the AddressFamily filter.
	//
	// Example: "10.10.10.0/24" or "10.11.12.13/32" or
	// "2002::1234:abcd:ffff:c0a8:101/64"
	CIDR string
	// AddressFamily instructs List to only get endpoints with a certain
	// address family.
	// It cannot be used together with the CIDR filter.
	//
	// Accepted values:
	// - AnyAddressfamily (the default)
	// - IPv4AddressFamily
	// - IPv6AddressFamily
	AddressFamily AddressFamily
}

// PortFilters contains filters to use for filtering an *endpoint* based on its
// port.
type PortFilters struct {
	// In is a list of ports that the endpoint must have in order to be
	// returned.
	// It cannot be used together with the Range filter.
	//
	// Example: []int32{80, 8080, 443}
	In []int32
	// Range is exactly like the In filter except that you use this when
	// wanting to including many ports in a range. The range delimiters are
	// included.
	// Cannot be used together with the In filter.
	//
	// Example: []int32{{80, 85}, {8080, 8090}}
	Range [][2]int32
}

// AddressFamily is the protocol the address belongs to.
type AddressFamily int

const (
	// AnyAddressFamily instructs List to get address that belong to any
	// address family.
	AnyAddressFamily = 0
	// IPv4AddressFamily instructs List to only get endpoints containing IPv4
	// addresses.
	IPv4AddressFamily = syscall.AF_INET
	// IPv6AddressFamily instructs List to only get endpoints containing IPv6
	// addresses.
	IPv6AddressFamily = syscall.AF_INET6
)

type Options struct {
	// Results defines the number of results to pull per page.
	Results int32
	// NameFilters provides filters for the name of the objects.
	*NameFilters
	// MetadataFilters provides filters for the metadata of the objects.
	*MetadataFilters
	// AddressFilters provides filters for the address of the endpoint.
	*AddressFilters
	// PortFilters provides filters for the ports of the endpoint.
	*PortFilters
}

// Filter returns true if the object provided as argument passes all the
// filters in the options. The object must be one of those provided in the
// api/core/types package: Namespace, Service or Endpoint. Otherwise, this
// will return false and an error.
func (o *Options) Filter(object interface{}) (bool, error) {
	var (
		name     string
		metadata map[string]string
	)

	switch obj := object.(type) {
	case *coretypes.Namespace:
		name = obj.Name
		metadata = obj.Metadata
	case *coretypes.Service:
		name = obj.Name
		metadata = obj.Metadata
	case *coretypes.Endpoint:
		name = obj.Name
		metadata = obj.Metadata
	default:
		return false, srerr.InvalidObjectToFilter
	}

	if o.NameFilters != nil {
		if len(o.NameFilters.In) > 0 && !nameInFilter(name, o.NameFilters.In...) {
			return false, nil
		}

		if o.NameFilters.Prefix != "" && !namePrefixFilter(name, o.NameFilters.Prefix) {
			return false, nil
		}
	}

	if o.MetadataFilters != nil {
		if o.MetadataFilters.NoMetadata && len(metadata) > 0 {
			return false, nil
		}

		if len(o.MetadataFilters.Metadata) > 0 && !metadataFilter(metadata, o.MetadataFilters.Metadata) {
			return false, nil
		}
	}

	if endp, ok := object.(*coretypes.Endpoint); ok {
		if o.AddressFilters != nil {
			if o.AddressFilters.CIDR != "" && !isInsideCIDR(o.AddressFilters.CIDR, endp.Address) {
				return false, nil
			}

			if o.AddressFilters.AddressFamily == syscall.AF_INET && !isIPv4(endp.Address) {
				return false, nil
			}

			if o.AddressFilters.AddressFamily == syscall.AF_INET6 && !isIPv6(endp.Address) {
				return false, nil
			}
		}

		if o.PortFilters != nil {
			if len(o.PortFilters.In) > 0 && endp.Port != 0 && !portIsIn(endp.Port, o.PortFilters.In...) {
				return false, nil
			}

			if len(o.PortFilters.Range) > 0 && !portIsInRange(endp.Port, o.PortFilters.Range) {
				return false, nil
			}
		}
	}

	return true, nil
}

type Option func(*Options) error

// WithNamePrefix instructs List to only get endpoints that have a certain
// prefix, ignore all others. It cannot be used together with WithNameIn.
//
// Each call to this function replaces values provided by any precedent call.
//
// Example:
// 	sr.Namespace(core.Any).List(list.WithNamePrefix("prod-"))
func WithNamePrefix(prefix string) Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if prefix == "" {
			return srerr.InvalidNamePrefixFilter
		}

		if lo.NameFilters == nil {
			lo.NameFilters = &NameFilters{}
		}

		if len(lo.NameFilters.In) > 0 {
			return srerr.IncompatibleNameFilters
		}

		lo.NameFilters.Prefix = prefix
		return nil
	}
}

// WithNameIn instructs List to only retreive the given endpoints, ignore
// all others. It cannot be used together with WithNamePrefix.
//
// Each call to this function appends its values to the ones provided by any
// precedent call, skipping duplicate names.
//
// Example:
// 	sr.Namespace(core.Any).List(list.WithNameIn("sales", "it", "hr"))
func WithNameIn(names ...string) Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if len(names) == 0 {
			return srerr.EmptyNameInFilter
		}

		if lo.NameFilters == nil {
			lo.NameFilters = &NameFilters{}
		}

		if lo.NameFilters.Prefix != "" {
			return srerr.IncompatibleNameFilters
		}

		// Note: we don't do any dedup here.
		// TODO: add this on future versions?

		lo.NameFilters.In = append(lo.NameFilters.In, names...)
		return nil
	}
}

// WithMetadataKeys instructs List to only retreive endpoints whose metadata
// include the given keys - even if they have no values, ignore all others.
// It cannot be used together with WithNoMetadata.
//
// Each call to this function appends its values to the ones provided by any
// precedent call, skipping duplicate keys.
//
// This is a shortcut for
// 	WithMetadata(map[string]string{"key-1": "", "key-2": ""})
//
// Example:
// 	sr.Namespace(core.Any).List(list.WithMetadataKeys("team", "contact", "building"))
func WithMetadataKeys(keys ...string) Option {
	return func(lo *Options) error {
		if len(keys) == 0 {
			return srerr.EmptyMetadataKeysFilter
		}

		metadata := map[string]string{}
		for _, key := range keys {
			metadata[key] = ""
		}

		return WithMetadata(metadata)(lo)
	}
}

// Shortcut for WithMetadataKeys().
func WithKs(keys ...string) Option {
	return WithMetadataKeys(keys...)
}

// WithMetadataKeyValue instructs List to only retreive endpoints whose
// metadata include the given keys with the given values, ignore all others.
// It cannot be used together with WithNoMetadata.
//
// Each call to this function appends its values to the ones provided by any
// precedent call. In case of duplicates, only the last one will be saved.
//
// This is a shortcut for
// 	WithMetadata(map[string]string{"key": "value"})
//
// Example:
// 	sr.Namespace(core.Any).List(
// 		list.WithMetadataKeyValue("maintainer", "alice.smith"),
//		list.WithMetadataKeyValue("location", "building-24"),
// 	)
func WithMetadataKeyValue(key, value string) Option {
	return WithMetadata(map[string]string{key: value})
}

// See WithMetadataKeyValue().
func WithKV(key, value string) Option {
	return WithMetadataKeyValue(key, value)
}

// WithMetadata instructs List to only retreive endpoints with the provided
// metadata key-value pairs. Add a key with an empty value if you don't care
// about its value. All other endpoints will be ignored.
// It cannot be used together with WithNoMetadata.
//
// Each call to this function appends its values to the ones provided by any
// precedent call. In case of duplicate pairs, only the last ones will be
// saved.
//
// Example:
// 	sd.Namespace(core.Any).List(list.WithMetadata(map[string]string{
// 		"version": "v1.2.0",
// 		"stage":   "alpha",
// 	}))
func WithMetadata(kv map[string]string) Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if lo.MetadataFilters == nil {
			lo.MetadataFilters = &MetadataFilters{
				Metadata: map[string]string{},
			}
		}

		if lo.MetadataFilters.NoMetadata {
			return srerr.IncompatibleMetadataFilters
		}

		// First check for empty keys
		for k := range kv {
			if k == "" {
				return srerr.EmptyMetadataKey
			}

			// TODO: check for long values, as well.
		}

		// Then push the kvs
		for k, v := range kv {
			lo.Metadata[k] = v
		}

		return nil
	}
}

// WithNoMetadata instructs List to only retreive namespaces with no metadata
// key-value pairs, or in other words with empty metadata.
//
// It cannot be used together with any of the other metadata options.
//
// Example:
// 	sd.Namespace(core.Any).List(list.WithNoMetadata())
func WithNoMetadata() Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if lo.MetadataFilters == nil {
			lo.MetadataFilters = &MetadataFilters{
				Metadata: map[string]string{},
			}
		}

		if len(lo.MetadataFilters.Metadata) > 0 {
			return srerr.IncompatibleMetadataFilters
		}

		lo.MetadataFilters.NoMetadata = true
		return nil
	}
}

// WithResultsNumber provides a custom value for the number of objects to get
// per page.
//
// If you don't have any special needs, you can skip this function entirely,
// letting the project use the default value, which is 50 results per page.
//
// Example:
// 	sd.Namespace(core.Any).List(list.WithResultsNumber(100))
func WithResultsNumber(number int32) Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if number <= 0 {
			return srerr.InvalidResultsNumber
		}

		lo.Results = number
		return nil
	}
}

// WithCIDR instructs List to only get endpoints with an address inside the
// given network.
//
// Note that version of the CIDR is important, as providing an IPv4 network
// will only return endpoints with IPv4 addresses. Therefore, this cannot be
// used together with WithIPv4Only nor WithIPv6Only().
//
// This option is ignored if used on namespaces or services.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Endpoints(core.Any).
// 		List(list.WithCIDR("10.10.10.0/24"))
func WithCIDR(cidr string) Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return srerr.InvalidCIDRProvided
		}

		if lo.AddressFilters == nil {
			lo.AddressFilters = &AddressFilters{}
		}

		if lo.AddressFilters.AddressFamily != AnyAddressFamily {
			return srerr.IncompatibleAddressFilters
		}

		lo.AddressFilters.CIDR = cidr
		return nil
	}
}

// WithIPv4Only instructs List to only get endpoints with an IPv4 address,
// ignoring all other endpoints.
//
// Only the last one between WithIPv4Only and WithIPv6Only will be considered,
// and cannot be used together with WithCIDR.
//
// This option is ignored if used on namespaces or services.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Endpoints(core.Any).
// 		List(list.WithIPv4Only())
func WithIPv4Only() Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if lo.AddressFilters == nil {
			lo.AddressFilters = &AddressFilters{}
		}

		if lo.AddressFilters.CIDR != "" {
			return srerr.IncompatibleAddressFilters
		}

		lo.AddressFilters.AddressFamily = IPv4AddressFamily
		return nil
	}
}

// WithIPv6Only instructs List to only get endpoints with an IPv6 address,
// ignoring all other endpoints.
//
// Only the last one between WithIPv4Only and WithIPv6Only will be considered,
// and cannot be used together with WithCIDR.
//
// This option is ignored if used on namespaces or services.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Endpoints(core.Any).
// 		List(list.WithIPv6Only())
func WithIPv6Only() Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if lo.AddressFilters == nil {
			lo.AddressFilters = &AddressFilters{}
		}

		if lo.AddressFilters.CIDR != "" {
			return srerr.IncompatibleAddressFilters
		}

		lo.AddressFilters.AddressFamily = IPv6AddressFamily
		return nil
	}
}

// WithPortIn instructs List to only get endpoints with a port included in the
// provided list, ignoring all other endpoints.
//
// Each call to this function appends its values to the ones provided
// previously.
//
// This option is ignored if used on namespaces or services.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Endpoints(core.Any).
// 		List(list.WithPortIn(443, 6443))
func WithPortIn(ports ...int32) Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if lo.PortFilters == nil {
			lo.PortFilters = &PortFilters{}
		}

		if lo.PortFilters.In == nil {
			lo.PortFilters.In = []int32{}
		}

		if len(ports) == 0 {
			return srerr.NoPortsProvided
		}

		// First check the values
		toAdd := []int32{}
		for _, ins := range ports {
			if ins < 0 || ins > math.MaxUint16 {
				return fmt.Errorf(`invalid port (%d) provided: %w`, ins, srerr.InvalidPort)
			}

			found := false
			for _, p := range lo.PortFilters.In {
				if p == ins {
					found = true
					break
				}
			}
			if !found {
				toAdd = append(toAdd, ins)
			}
		}

		lo.PortFilters.In = append(lo.PortFilters.In, toAdd...)
		return nil
	}
}

// WithPortRange instructs List to only get endpoints within a certain range,
// ignoring all other endpoints. The range delimiters are *included*.
//
// Each call to this function appends its values to the ones provided
// previously.
//
// This option is ignored if used on namespaces or services.
//
// Example:
// 	sd.Namespace("hr").Service("payroll").Endpoints(core.Any).
// 		List(list.EndpointListWithPortRange(8080, 8090))
func WithPortRange(start, end int32) Option {
	return func(lo *Options) error {
		if lo == nil {
			return srerr.NoOptionsProvided
		}

		if lo.PortFilters == nil {
			lo.PortFilters = &PortFilters{}
		}

		if lo.PortFilters.Range == nil {
			lo.PortFilters.Range = [][2]int32{}
		}

		if start > end || start < 0 || end > math.MaxUint16 {
			return fmt.Errorf("invalid range (%d-%d) provided: %w", start, end, srerr.InvalidPortRange)
		}

		for _, r := range lo.PortFilters.Range {
			if r[0] == start && r[1] == end {
				return nil
			}
		}

		// We just add the range without checking if this is overlapping, nor do
		// we check if ports in the `In` filter already included in one of these
		// ranges. We do this for simplicity.
		// TODO: change this in future?
		lo.PortFilters.Range = append(lo.PortFilters.Range, [2]int32{start, end})
		return nil
	}
}
