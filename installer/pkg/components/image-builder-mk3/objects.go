package image_builder_mk3

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	clusterrole,
	configmap,
	deployment,
	networkpolicy,
	rolebinding,
	secret,
	common.GenerateService(Component, map[string]common.ServicePort{
		"service": {
			ContainerPort: RPCPort,
			ServicePort:   RPCPort,
		},
	}, nil),
	common.DefaultServiceAccount(Component),
)
