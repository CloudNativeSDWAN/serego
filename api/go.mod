module github.com/CloudNativeSDWAN/serego/api

go 1.18

require (
	cloud.google.com/go/servicedirectory v1.3.0
	github.com/aws/aws-sdk-go-v2 v1.16.13
	github.com/aws/aws-sdk-go-v2/config v1.17.4
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.17.15
	github.com/googleapis/gax-go/v2 v2.5.1
	github.com/onsi/ginkgo/v2 v2.1.6
	github.com/onsi/gomega v1.20.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	go.etcd.io/etcd/api/v3 v3.5.5
	go.etcd.io/etcd/client/v3 v3.5.5
	google.golang.org/api v0.94.0
	google.golang.org/genproto v0.0.0-20220725144611-272f38e5d71b
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/apimachinery v0.25.0
)

require (
	cloud.google.com/go/compute v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.17 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.20 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.16 // indirect
	github.com/aws/smithy-go v1.13.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.5 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/oauth2 v0.0.0-20220822191816-0ebed06d0094 // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
)

replace github.com/CloudNativeSDWAN/serego => ../
