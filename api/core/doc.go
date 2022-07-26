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

// Package core contains code and data that connects to a service registry
// and operate on its objects and resources.
//
// It acts as a common denominator for multiple service registries, as it
// abstracts resources in a common way and common semantics, and allows you
// to perform operations on them with the same API regardless of the one you
// choose to utilize.
//
// The package - and thus the full project - has three essential object types
// that map to the actual ones defined by the provider, abstracting and
// adapting them when necessary: namespaces, services, and endpoints.
//
// Finally the whole API reflects the hierarchical nature of a service
// registry, as the operations follow the namespace -> service -> endpoint
// hierarchy. Read further or look at the examples to know more.
//
// For a better description of the objects, please refer to the objects
// documentation at github.com/CloudNativeSDWAN/serego/docs/service_registry.md.
//
// Wrappers
//
// To start using the package you will need to wrap a ServiceRegistry structure
// around one of the supported service registries. To do so, refer to
// pkg.go.dev/github.com/CloudNativeSDWAN/serego/api/pkg/core.
//
// Operations
//
// You first define the object you want to perform the operation on,
// and then the actual operation:
// 	// Start an operation from wrapper myWrapper on Namespace "sales", and
// 	// then perform a Get operation.
// 	ns, err := myWrapper.Namespace("sales").Get(context.Background())
//
// 	// Start an operation on a service called "support" on namespace "sales".
//	// Then update its metadata.
// 	servOp := myWrapper.Namespace("sales").Service("support")
// 	servOp.Register(context.Background(), register.WithMetadata(map[string]string{
//				"version": "1.2.2",
// 				"maintainer": "team-22",
// 			}))
//
// The package performs some actions to improve performance and cache
// operations internally, so you should reuse operations as much as you can to
// take full advantage of this, for example:
//
// 	nsOp := wrap.Namespace("sales")
//
// 	support, err := nsOp.Service("support").Get(context.Background())
// 	err = nsOp.Service("contact").Register(context.Background())
//
// To pass options, you will have to use functions provided in the options
// package: pkg.go.dev/github.com/CloudNativeSDWAN/serego/api/pkg/options.
//
// Finally, all operations will return an error not only for errors from
// the service registry but also in case of user errors, e.g. when invalid
// options are provided or an empty name is provided.
//
// Errors
//
// All operations will return an error either from the project's
// errors package (pkg.go.dev/github.com/CloudNativeSDWAN/serego/api/pkg/errors)
// or from the service registry's own API
// package, so that you can have more clues and information about the error.
//
// To facilitate error checking without casting or multiple checks, please take
// a look at the errors package on the link provided above, as some functions
// are already defined there to help you with error checking.
//
// Get Operations
//
// Get retrieves the object by checking the internal cache and, if
// not found there, by performing a call to the service registry.
//
// Check out the the examples provided in each Get function and read their
// description to learn more about their behavior and options they accept.
//
// Register Operations
//
// All register operations create the object if it does not exist or update it
// otherwise. You can override this through the option RegisterMode, e.g. you
// can set it to CreateMode to specify that it must be created and return an
// error if it already exists.
//
// Note that register operations are incremental: by default, whatever you
// pass to it will be added to the object and will not replace/destroy all
// existing data. For example if a service has metadata:
// 	"version": "1.2.1"
// 	"maintainer": "alice.smith@company.com"
// and you call register with a metadata
// 	"commit-hash": "abc123",
// then this last pair will be *added* to the existing ones, unless you provide
// ReplaceMetadata to its options, which will still add the pair but destroy
// all the existing ones.
//
// Check out the the examples provided in each Register function and read their
// description to learn more about their behavior and options they accept.
//
// Deregister Operations
//
// Deregister acts exactly as the opposite of Register, as it attempts to
// delete the object from the service registry.
//
// If the object does not exist, Deregister *does not* return any
// error, since the intention was that of removing the object anyways.
// If you still want to receive an error though, you need to pass
// FailIfNotExists as option. Note that this will not change anything for any
// other errors, which will always still be returned.
//
// Note that as of now it does not check if the object is empty, e.g. it
// still tries to remove a namespace even if it has services inside it.
// This is because some registries do not provide this capability and others
// do, though this may change and the API could do this for you in future.
// As of now, to make sure an object is empty, you will have to do a List()
// and check if you do at least one iteration before exhausting the iterator.
//
// Check out the the examples provided in each Deregister function and read
// their description to learn more about their behavior and options they
// accept.
//
// List operations
//
// List operations create an iterator that you can use to loop through all
// found objects. You can provide filters for the list operation, e.g. the
// name, metadata, address or ports, and for a full list of options and
// filters check out each List function to see what they accept.
//
// You can then iterate through the objects with the Next function of each
// iterator, checking if there are no more elements with the provided errors
// package (look at the examples).
//
// If an error occurs, e.g. if an invalid option is passed, it will
// not be returned by the List function but by *any* subsequent call to Next
// function, preventing it from successfully getting the list.
//
// List functions return both the object and an operation: this is
// done to allow you to perform an operation on the object immediately, without
// instantiating one for it or repeating yourself.
//
// Next
//
// Next functions can be called on the results of a List function and they
// automatically return the next element from the pulled results from the
// registry.
//
// It returns the operations *only* if the object passes *all* filters - or if
// no filters are passed, and it automatically tries the next element if it
// does not.
//
// Quickstart example
//
// Take a look at this example:
package core
