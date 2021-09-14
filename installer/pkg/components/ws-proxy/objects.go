package wsproxy

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	configmap,
	deployment,
	networkpolicy,
	rolebinding,
	common.DefaultServiceAccount(Component),
	common.GenerateService(Component, map[string]common.ServicePort{
		"httpProxy": {
			ContainerPort: HTTPProxyPort,
			ServicePort:   HTTPProxyPort,
		},
		"httpsProxy": {
			ContainerPort: HTTPSProxyPort,
			ServicePort:   HTTPSProxyPort,
		},
		"metrics": {
			ContainerPort: MetricsPort,
			ServicePort:   MetricsPort,
		},
	}, nil),
)
