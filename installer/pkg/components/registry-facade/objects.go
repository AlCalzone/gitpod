package registryfacade

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	clusterrole,
	configmap,
	daemonset,
	networkpolicy,
	podsecuritypolicy,
	rolebinding,
	common.GenerateService(Component, map[string]common.ServicePort{
		"registry": {
			ContainerPort: ContainerPort,
			ServicePort:   ServicePort,
		},
	}, nil),
	common.DefaultServiceAccount(Component),
)
