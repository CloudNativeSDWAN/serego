# SeReGo

**Se**rvice **re**gistry. The universal and common way. Written in **go**.

## About the project

*SeReGo* is a library software written in go that allows you to connect to and
perform operations on different service registries using the same APIs and
semantic regardless of the underlying provider you choose.

Use a service registry, i.e. *Google Service Directory* or *AWS Cloud Map*,
with the same set of operations without learning each one's documentation and
switching seamlessly between the two.

You can also use a key-value database like *etcd* as the provider and let
*SeReGo* treat it as a service registry and thus have a sort of in-house
service registry.

*SeReGo* also implements some features that may not be supported or are only
partially supported by the registry allowing you to just focus on the business
logic of your application, e.g. filtering services based on metadata or other
data.

We currently support *Google Service Directory*, *AWS Cloud Map* and
*etcd*. We're open to support more registries or databases and
if you have suggestions please feel free to post a feature request via *Issues*
or to discuss it by opening a discussion in the *Discussions* section.

To learn more about service registries and the objects that *Serego* will work
with, please read our
[documentation section about service registries](./docs/service_registry.md).

## Quickstart

In this quick example, create a client from *Google Service Directory* and pass
it to *SeReGo* to perform operations from then on.

In this case we will list all services that are inside a namespace called
`sales` and that have at least the key-value pair
`maintainer: alice.smith@company.com`.

```go
import (
    "context"
    "fmt"

    servicedirectory "cloud.google.com/go/servicedirectory/apiv1"
    "github.com/CloudNativeSDWAN/serego/api/core"
    "github.com/CloudNativeSDWAN/serego/api/errors"
    "github.com/CloudNativeSDWAN/serego/api/options/list"
    "github.com/CloudNativeSDWAN/serego/api/options/register"
    "github.com/CloudNativeSDWAN/serego/api/options/wrapper"
)

func main() {

    // -----------------------------------
    // Get the client from the service registry
    // -----------------------------------

    // (there are many ways to get a client from Service Directory, this is just
    // an example)
    cl, err := servicedirectory.NewRegistrationClient(
        context.Background(),
        option.WithCredentialsFile("path/to/the/service-account.json"),
    )
    if err != nil {
        fmt.Println("could not get service directory client:", err, ". Exiting...")
        return
    }
    defer cl.Close()

    // -----------------------------------
    // Pass it to SeReGo
    // -----------------------------------

    // Initialize the client and the settings...
    sd, err := core.NewServiceRegistryFromServiceDirectory(
        client, wrapper.WithRegion("us-east1"), wrapper.WithProjectID("my-project-id"))
    if err != nil {
        fmt.Println("error while instantiating serego wrapper:", err)
        return
    }

    // -----------------------------------
    // Perform operations
    // -----------------------------------

    // List all services inside a namespace called "sales" and that are
    // maintained by *Alice Smith*:
    servIterator := sd.Namespace("sales").
        Service(core.Any).
        List(list.WithMetadataKeyValue("maintainer", "alice.smith@company.com"))

    // Loop through all found results
    for {
        service, servOp, err := servIterator.Next(context.TODO())
        if err != nil {
            if errors.IsIteratorDone(err) {
                // Finished iterating.
                fmt.Println("finished.")
            } else {
                fmt.Println("could not get next service:", err)
            }

            break
        }

        fmt.Printf("found service with name %s and metadata %+v\n", service.Name, service.Metadata)

        // Update its metadata by setting "status: beta"
        err := servOp.Register(context.TODO(), register.WithMetadataKeyValue("status", "beta"))
        if err != nil {
            fmt.Println("could not update service's metadata:", err)
            return
        }
    }

    // Read the API documentation to learn more about operations and options.
}
```

You need to do the same for multiple service registries? Just put whatever is
after the `Perform operations` section into a separate function, and create
multiple Serego wrappers to call that function with.

For example, you can also create an AWS Cloud Map wrapper through the
`core.NewServiceRegistryFromCloudMap` function, the syntax will just stay
the same. Once again, please look at the documentation for more examples.

## Resources

The objects that will be abstracted by *Serego* are *Namespaces*, *Services*,
and *Endpoints*. Please read our
[documentation section about service registries](./docs/service_registry.md) to
learn more about their use and purpose.

## Operations

*Serego* allows you to perform the following operations on a service registry:

- `Get` to retrieve a resource
- `Register` to create the object or update it if it already exists
- `Deregister` to remove the object
- `List` to list all objects based on filters

Please refer to our SDK documentation to learn more about these operations and
their options.

You should first define the object you wish to define the operation for,
examples:

```go
// Get a service called "payroll" inside namespace "hr"
srv, err := Namespace("hr").Service("payroll").Get(context.Background())

// Register an endpoint for service "profile" inside namespace "users"
err := Namespace("users").
    Service("profile").
    Endpoint("internal").
    Register(context.Background(), register.WithAddress("10.10.10.22"), register.WithPort(8080))
```

Once again please refer to our SDK documentation for more thorough
descriptions and examples.

## Future developments

- Watching for changes in real-time
- Server that acts as a full-fledged service registry that you can interact
    with via `gRPC` or `REST` and in any language and supporting `RBAC`.
- `CLI` application
- Experiment with go `1.18` generics
