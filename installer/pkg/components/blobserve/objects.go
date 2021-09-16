package blobserve

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	configmap,
	deployment,
	networkpolicy,
	rolebinding,
	common.GenerateService(Component, map[string]common.ServicePort{
		"service": {
			ContainerPort: ContainerPort,
			ServicePort:   ServicePort,
		},
	}, nil),
	common.DefaultServiceAccount(Component),
)
