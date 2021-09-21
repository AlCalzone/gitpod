package server

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	configmap,
	deployment,
	ideconfigmap,
	networkpolicy,
	role,
	rolebinding,
	common.GenerateService(Component, map[string]common.ServicePort{
		"http": {
			ContainerPort: ContainerPort,
			ServicePort:   ServicePort,
		},
		"metrics": {
			ContainerPort: PrometheusPort,
			ServicePort:   PrometheusPort,
		},
	}, nil),
	common.DefaultServiceAccount(Component),
)
