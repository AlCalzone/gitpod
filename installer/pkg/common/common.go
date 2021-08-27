package common

import (
	"strings"

	config "github.com/gitpod-io/gitpod/installer/pkg/config/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
)

const (
	AffinityLabelMeta              = "gitpod.io/workload_meta"
	AffinityLabelWorkspaceServices = "gitpod.io/workload_workspace_services"
	AffinityLabelWorkspaces        = "gitpod.io/workload_workspaces"
	AffinityLabelHeadless          = "gitpod.io/workload_headless"
)

func DefaultLabels(component string) map[string]string {
	return map[string]string{
		"component": component,
	}
}

func MergeEnv(envs ...[]corev1.EnvVar) (res []corev1.EnvVar) {
	for _, e := range envs {
		res = append(res, e...)
	}
	return
}

func DefaultEnv(cfg *config.Config) []corev1.EnvVar {
	logLevel := "debug"
	if cfg.Observability.LogLevel != nil {
		logLevel = string(*cfg.Observability.LogLevel)
	}

	return []corev1.EnvVar{
		{Name: "GITPOD_DOMAIN", Value: cfg.Domain},
		{Name: "LOG_LEVEL", Value: strings.ToLower(logLevel)},
	}
}

func TracingEnv(cfg *config.Config) (res []corev1.EnvVar) {
	if cfg.Observability.Tracing == nil {
		return
	}

	if cfg.Observability.Tracing.Endpoint != nil {
		res = append(res, corev1.EnvVar{Name: "JAEGER_ENDPOINT", Value: *cfg.Observability.Tracing.Endpoint})
	} else if cfg.Observability.Tracing.AgentHost != nil {
		res = append(res, corev1.EnvVar{Name: "JAEGER_AGENT_HOST", Value: *cfg.Observability.Tracing.Endpoint})
	} else {
		// TODO(cw): think about proper error handling here.
		//			 Returning an error would be the appropriate thing to do,
		//			 but would make env var composition more cumbersome.
	}

	res = append(res,
		corev1.EnvVar{Name: "JAEGER_SAMPLER_TYPE", Value: "const"},
		corev1.EnvVar{Name: "JAEGER_SAMPLER_PARAM", Value: "1"},
	)

	return
}

func KubeRBACProxyContainer() *corev1.Container {
	return &corev1.Container{
		Name:  "kube-rbac-proxy",
		Image: "quay.io/brancz/kube-rbac-proxy:v0.9.0",
		Args: []string{
			"--v=10",
			"--logtostderr",
			"--insecure-listen-address=[$(IP)]:9500",
			"--upstream=http://127.0.0.1:9500/",
		},
		Ports: []corev1.ContainerPort{
			{Name: "metrics", ContainerPort: 9500},
		},
		Env: []corev1.EnvVar{
			{
				Name: "IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		},
		Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
			corev1.ResourceName("cpu"):    resource.MustParse("1m"),
			corev1.ResourceName("memory"): resource.MustParse("30Mi"),
		}},
		TerminationMessagePolicy: corev1.TerminationMessagePolicy("FallbackToLogsOnError"),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:    pointer.Int64(65532),
			RunAsGroup:   pointer.Int64(65532),
			RunAsNonRoot: pointer.Bool(true),
		},
	}
}

func Affinity(orLabels ...string) *corev1.Affinity {
	var terms []corev1.NodeSelectorTerm
	for _, lbl := range orLabels {
		terms = append(terms, corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      lbl,
					Operator: corev1.NodeSelectorOperator("Exists"),
				},
			},
		})
	}

	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: terms,
			},
		},
	}
}
