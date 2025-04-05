package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GrafanaAlertContentEnv struct {
	Name string `json:"name"`
	// Inline env value
	// +optional
	Value string `json:"value,omitempty"`
	// Reference on value source, might be the reference on a secret or config map
	// +optional
	ValueFrom GrafanaAlertContentEnvFromSource `json:"valueFrom,omitempty"`
}

type GrafanaAlertContentEnvFromSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *v1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	// +optional
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type GrafanaAlertContentSpec struct {
	// model from configmap
	// +optional
	ConfigMapRef *v1.ConfigMapKeySelector `json:"configMapRef,omitempty"`

	// maps required data sources to existing ones
	// +optional
	Datasources []GrafanaContentDatasource `json:"datasources,omitempty"`

	// environments variables as a map
	// +optional
	Envs []GrafanaAlertContentEnv `json:"envs,omitempty"`

	// environments variables from secrets or config maps
	// +optional
	EnvsFrom []GrafanaAlertContentEnvFromSource `json:"envFrom,omitempty"`
}

type GrafanaAlertContentStatus struct {
	ContentCache     []byte      `json:"contentCache,omitempty"`
	ContentTimestamp metav1.Time `json:"contentTimestamp,omitempty"`
	ContentUrl       string      `json:"contentUrl,omitempty"`
	Hash             string      `json:"hash,omitempty"`
	UID              string      `json:"uid,omitempty"`
}

// GrafanaAlertContentResource
// Common interface for any resource that embeds or references Grafana-native model content.
// +kubebuilder:object:generate=false
type GrafanaAlertContentResource interface {
	client.Object
	GrafanaAlertContentSpec() *GrafanaAlertContentSpec
	GrafanaAlertContentStatus() *GrafanaAlertContentStatus
}
