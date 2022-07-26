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

package servicedirectory

import (
	"fmt"
	"path"
	"strings"

	"github.com/CloudNativeSDWAN/serego/api/options/list"
)

func getRequestFilters(basePath string, lo *list.Options) string {
	reqFilter := []string{}

	func() {
		if lo.NameFilters == nil {
			return
		}

		nf := lo.NameFilters
		namesFilter := []string{}
		for _, name := range nf.In {
			namesFilter = append(namesFilter, fmt.Sprintf("name=%s", path.Join(basePath, name)))
		}

		if len(namesFilter) > 0 {
			if len(namesFilter) == 1 {
				reqFilter = append(reqFilter, namesFilter[0])
			} else {
				reqFilter = append(reqFilter, fmt.Sprintf("(%s)",
					strings.Join(namesFilter, " OR ")))
			}
		}

		// Reset this, so we don't have to use filters there
		lo.NameFilters.In = []string{}
	}()

	func() {
		if lo.MetadataFilters == nil {
			return
		}

		mf := lo.MetadataFilters
		emptyMetadata := map[string]string{}
		metadataVals := []string{}
		for k, v := range mf.Metadata {
			if v == "" {
				emptyMetadata[k] = v
				continue
			}

			metadataName := ""
			switch path.Base(basePath) {
			case "namespaces":
				metadataName = "labels"
			case "services", "endpoints":
				metadataName = "annotations"
			}

			if metadataName != "" {
				metadataVals = append(metadataVals, fmt.Sprintf("%s.%s=%s", metadataName, k, v))
			}
		}
		if len(metadataVals) > 0 {
			if len(metadataVals) == 1 {
				reqFilter = append(reqFilter, metadataVals[0])
			} else {
				reqFilter = append(reqFilter, fmt.Sprintf("(%s)",
					strings.Join(metadataVals, " AND ")))
			}
		}

		// Reset this, so we don't have to use filters there
		lo.MetadataFilters.Metadata = emptyMetadata
	}()

	return strings.Join(reqFilter, " AND ")
}
