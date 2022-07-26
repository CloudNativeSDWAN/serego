# Service registry

A service registry is basically a database of services: each registered service
provides information about its instances, addresses, ports and other data.
Often times, *metadata* can be registered as well, which provide additional
data about a resource.

Therefore, a client can log in to the service registry and *discover*
registered services to know how to connect to them and get data about them.
Sounds familiar? That's because it is very similar to how DNS works, but the
service registry pattern is a
[key concept of microservices](https://auth0.com/blog/an-introduction-to-microservices-part-3-the-service-registry/).

**Note** that this document defines how a service registry is implement and
abstracted by *Serego* and, therefore, may also be used as a *guideline* for
designing one on your backend; but it does not mean that *all* existing service
registries are implemented as you see here or abide to thise guidelines.

## Metadata

Some service registries allow for *metadata* to be published along with the
resource that you are publishing. These are things that may not really have a
meaning for the object itself but more to you, your team, your application or
your use case in general.

The following is just list of `key: value` pairs and we're confident that the
examples in this document will clarify the concept more:

* `hash-commit: asd043qv`, `branch: master` to define repository data
* `protocol: RTP`, `authentication: enabled` to define more information on how
  to connect to the application
* `team: best-team-in-company`, `contact: bu-manager@company.com` to define
  information about the team and how to get help about the application

Please keep in mind that *some* service registries may provide *partial*
support for metadata, i.e. a maximum number of values or allow them only
for some objects and not all.

## Which one to choose?

Which one you choose depends on different factors, including:

* restrictions on metadata if you plan to have many
* your budget
* visibility

and so on.

We currently support *Google Service Directory*, *AWS Cloud Map* and
*etcd*. We're open to support more registries or databases and
if you have suggestions please feel free to post a feature request via *Issues*
or to discuss it by opening a discussion in the *Discussions* section.

## Objects

Let's now cover the objects that *Serego* will work with and/or abstract for
you. Keep in mind that some products, i.e. *Google Service Directory*, may
work with objects that have similar names but different formats. Nonetheless,
*Serego* will provide you with data that always look like the ones included in
this guide. Even more, some service registries may not even support or include
some of these objects: as we said *Serego* will always try to abstract them
while still using the service registry's own format to interact with it.

### Namespace

A namespace is like a virtual cluster or group where you contain
applications/services that belong to the same project, or team, environment,
purpose, and so on.

Here are some example use cases for namespaces:

* an environment, i.e. `production` or `dev`
* a business unit or team, i.e. `hr` or `my-software-team`

A namespace will look like this in a *YAML* format:

```yaml
name: team
metadata:
    env: production
    manager: John Smith
```

This is a more formal description:

| Field       | Type        | Description
| ----------- | ----------- | -----------
| name        | string      | the name of the namespace
| metadata    | map (dictionary) | A list of key -> value pairs that provide more information about this namespace. Look at the example. Keys and values are both strings.

If you are using the SDK, you will also have another field called
`OriginalObject` which contains the object as it exists in the abstracted
service registry, in case you need to use some special features and fields
that are available only to that service registry. In this case you will have to
cast appropriately: please refer to each package documentation to learn how to
do that.

### Service

As the name suggests, a service is something that provides users/softwares a
resource or that performs some operations with a final result.
For simplicity, think of a service as an *application*.

It cannot exist by itself but only when it is part of a namespace. For example,
the same `payroll` service may exist in both namespace `prod` and
`dev` namespaces, but they may differ slightly - e.g. different container
images - or even be completely different applications/programs.

Here are some example services:

* `payroll` or `user-profile`
* `mysql` or `redis`

A service looks like this in a *YAML* format:

```yaml
name: payroll
namespace: production
metadata:
    traffic-profile: standard
    version: v1.2.1
    maintainers: software-team
    contact: software-team@company.com
```

This is a more formal description:

| Field       | Type        | Description
| ----------- | ----------- | -----------
| name        | string      | the name of the service
| namespace   | string      | the name of the namespace this service belongs to
| metadata    | map (dictionary) | A list of key -> value pairs that provide more information about this service. Look at the example. Keys and values are both strings.

As for namespaces, services have the `OriginalObject` field as well, please
read the documentation to understand how to cast the object appropriately.

### Endpoint

This is the actual "place" where you can reach a service/application.

It cannot exist by itself, as it obviously only has a meaning within a service.

Here are some example endpoints:

* `payroll-tcp` or `payroll-8080`
* `user-profile-internal` or `user-profile-vpn`

An endpoint looks like this in a *YAML* format:

```yaml
name: payroll-internal
service: payroll
namespace: production
address: 10.11.12.13
port: 9876
metadata:
    protocol: UDP
    weight: 0.25
```

This is a more formal description:

| Field       | Type        | Description
| ----------- | ----------- | -----------
| name        | string      | the name of the endpoint
| service     | string      | the name of the service that will be reached with this endpoint
| namespace   | string      | the name of the namespace that contains the parent service (and therefore the endpoint as well)
| address     | string      | the IP address of the endpoint
| port        | 32 bit integer | the port of the endpoint
| metadata    | map (dictionary) | A list of key -> value pairs that provide more information about this endpoint. Look at the example. Keys and values are both strings.

Finally, endpoints do have an `OriginalObject` field, too, that contains the
original object from the service registry and must be cast appropriately.
