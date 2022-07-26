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
	"net"
	"strings"
)

func nameInFilter(name string, needle ...string) bool {
	for _, n := range needle {
		if name == n {
			return true
		}
	}

	return false
}

func namePrefixFilter(name, prefix string) bool {
	return strings.HasPrefix(name, prefix)
}

func metadataFilter(metadata, needle map[string]string) bool {
	for k, v := range needle {
		val, exists := metadata[k]
		if !exists {
			return false
		}

		if v != "" && val != v {
			return false
		}
	}

	return true
}

func isInsideCIDR(cidr, addr string) bool {
	// We do not check error because it is done before
	_, ipnet, _ := net.ParseCIDR(cidr)

	ip := net.ParseIP(addr)
	if ip == nil {
		return false
	}

	return ipnet.Contains(ip)
}

func isIPv4(addr string) bool {
	ip := net.ParseIP(addr)
	return ip != nil && ip.To4() != nil
}

func isIPv6(addr string) bool {
	ip := net.ParseIP(addr)
	// To4 returns nil if it is not a valid ipv4 address.
	// Since service directory and cloud map do their own
	// validation and we do the same for etcd, then if it is nil
	// we're sure it is an ipv6.
	return ip != nil && ip.To4() == nil
}

func portIsIn(port int32, ports ...int32) bool {
	for _, p := range ports {
		if port == p {
			return true
		}
	}

	return false
}

func portIsInRange(port int32, ranges [][2]int32) bool {
	for _, r := range ranges {
		if port >= r[0] && port <= r[1] {
			return true
		}
	}

	return false
}
