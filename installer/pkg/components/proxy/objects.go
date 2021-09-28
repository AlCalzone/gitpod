package proxy

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	configmap,
	deployment,
	networkpolicy,
	rolebinding,
	common.GenerateService(Component, map[string]common.ServicePort{
		"http": {
			ContainerPort: ContainerHTTPPort,
			ServicePort:   ContainerHTTPPort,
		},
		"https": {
			ContainerPort: ContainerHTTPSPort,
			ServicePort:   ContainerHTTPSPort,
		},
		"metrics": {
			ContainerPort: PrometheusPort,
			ServicePort:   PrometheusPort,
		},
	}, nil),
	common.DefaultServiceAccount(Component),
)
