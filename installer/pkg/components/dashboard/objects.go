package dashboard

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	deployment,
	networkpolicy,
	rolebinding,
	common.GenerateService(Component, map[string]common.ServicePort{
		"http": {
			ContainerPort: ContainerPort,
			ServicePort:   ServicePort,
		},
	}, nil),
	common.DefaultServiceAccount(Component),
)
