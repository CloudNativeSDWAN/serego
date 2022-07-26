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

// Package errors contains functions to help you check for some of the
// errors returned by the package without you having to refer to the
// documentation of the service registry you chose.
//
// The project either returns errors from this package whenever it can, or
// the "original" error from the service registry's API, as
// different service registries throw different errors because of different
// reasons. The original errors are either "untouched" or wrapped, which you
// can still check with either one of these functions, or with golang's errors
// package, or the service registry's API errors.
//
// The Variables section of this package contains the errors that are "native"
// to this package, and their usage or reasons for being returned should be
// self explanatory.
//
// So, in order to facilitate error checking, you can use some of these
// functions. If have any needs/suggestions, we will be happy to hear it
// and add other functions to this package.
//
// Take a look at the examples to learn more.
package errors
