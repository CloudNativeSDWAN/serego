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

package core

import (
	"fmt"
	"math/rand"
	"reflect"

	"github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
)

func prepareRegisterOperation(opts interface{}, getObj interface{}, getErr error) (register.RegisterMode, map[string]string, error) {
	var (
		errAlreadyExists error
		errNotFound      error
		metadata         = map[string]string{}
		regOpts          = reflect.ValueOf(opts)
		registerMode     = register.RegisterMode(
			regOpts.Elem().FieldByName("RegisterMode").Uint())
	)

	switch getObj.(type) {
	case *types.Namespace:
		errAlreadyExists = srerr.NamespaceAlreadyExists
		errNotFound = srerr.NamespaceNotFound
	case *types.Service:
		errAlreadyExists = srerr.ServiceAlreadyExists
		errNotFound = srerr.ServiceNotFound
	case *types.Endpoint:
		errAlreadyExists = srerr.EndpointAlreadyExists
		errNotFound = srerr.EndpointNotFound
	}

	if getObj != nil {
		val := reflect.ValueOf(getObj)
		if !val.IsNil() {
			elemMetadata := val.Elem().
				FieldByName("Metadata").
				Interface().(map[string]string)
			metadata = deepCopyMap(elemMetadata)
		}
	}

	switch {
	case getErr == nil:
		if registerMode == register.CreateMode {
			return register.CreateOrUpdateMode, nil, errAlreadyExists
		}
		registerMode = register.UpdateMode
	case srerr.IsNotFound(getErr):
		if registerMode == register.UpdateMode {
			return register.CreateOrUpdateMode, nil, errNotFound
		}
		registerMode = register.CreateMode
	default:
		return register.CreateOrUpdateMode, nil, fmt.Errorf("error while checking if object exists: %w", getErr)
	}

	if regOpts.Elem().FieldByName("ReplaceMetadata").Bool() {
		metadata = map[string]string{}
	}

	metaIterator := regOpts.Elem().FieldByName("Metadata").MapRange()

	for metaIterator.Next() {
		metadata[metaIterator.Key().String()] = metaIterator.Value().String()
	}

	return registerMode, metadata, nil
}

func deepCopyMap(src map[string]string) (dest map[string]string) {
	dest = map[string]string{}
	for k, v := range src {
		dest[k] = v
	}
	return
}

func generateRandomName(serviceName string) string {
	// We want to be RFC1035-complaint, so we will only accept lower case
	// and numbers.
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return fmt.Sprintf("%s-%s", serviceName, string(b))

}
