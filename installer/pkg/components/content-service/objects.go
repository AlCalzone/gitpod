package content_service

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	configmap,
	deployment,
	networkpolicy,
	rolebinding,
	common.GenerateService(Component, map[string]common.ServicePort{
		"rpc": {
			ContainerPort: RPCPort,
			ServicePort:   RPCPort,
		},
		"metrics": {
			ContainerPort: PrometheusPort,
			ServicePort:   PrometheusPort,
		},
	}, nil),
	common.DefaultServiceAccount(Component),
)
