package content

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

// Unchanged checks if the stored content hash on the status matches the input
func Unchanged(cr v1beta1.GrafanaContentResource, hash string) bool {
	status := cr.GrafanaContentStatus()
	// This indicates an implementation error
	if status == nil {
		return true
	}

	return status.Hash == hash
}

type AlertContentResolver struct {
	Client   client.Client
	resource v1beta1.GrafanaAlertContentResource
}

type Option func(r *AlertContentResolver)

func NewContentResolver(cr v1beta1.GrafanaAlertContentResource, client client.Client, opts ...Option) *AlertContentResolver {
	resolver := &AlertContentResolver{
		Client:   client,
		resource: cr,
	}

	for _, opt := range opts {
		opt(resolver)
	}

	return resolver
}

// map data sources that are required in the content model to data sources that exist in the instance
func (h *AlertContentResolver) resolveDatasources(contentJson []byte) ([]byte, error) {
	spec := h.resource.GrafanaAlertContentSpec()
	if len(spec.Datasources) == 0 {
		return contentJson, nil
	}

	for _, input := range spec.Datasources {
		if input.DatasourceName == "" || input.InputName == "" {
			return nil, fmt.Errorf("invalid datasource input rule in content resource %v/%v, input or datasource empty", h.resource.GetNamespace(), h.resource.GetName())
		}

		searchValue := fmt.Sprintf("${%s}", input.InputName)
		contentJson = bytes.ReplaceAll(contentJson, []byte(searchValue), []byte(input.DatasourceName))
	}

	return contentJson, nil
}

func (h *AlertContentResolver) getAlertContentEnvs(ctx context.Context) (map[string]string, error) {
	spec := h.resource.GrafanaAlertContentSpec()

	envs := make(map[string]string)
	if spec.EnvsFrom != nil {
		for _, ref := range spec.EnvsFrom {
			key, val, err := h.getReferencedValue(ctx, h.resource, ref)
			if err != nil {
				return nil, fmt.Errorf("something went wrong processing envs, error: %w", err)
			}
			envs[key] = val
		}
	}
	if spec.Envs != nil {
		for _, ref := range spec.Envs {
			if ref.Value != "" {
				envs[ref.Name] = ref.Value
			} else {
				_, val, err := h.getReferencedValue(ctx, h.resource, ref.ValueFrom)
				if err != nil {
					return nil, fmt.Errorf("something went wrong processing referenced env %s, error: %w", ref.Name, err)
				}
				envs[ref.Name] = val
			}
		}
	}
	return envs, nil
}

func (h *AlertContentResolver) getReferencedValue(ctx context.Context, cr v1beta1.GrafanaAlertContentResource, source v1beta1.GrafanaAlertContentEnvFromSource) (string, string, error) {
	if source.SecretKeyRef != nil {
		s := &v1.Secret{}
		err := h.Client.Get(ctx, client.ObjectKey{Namespace: cr.GetNamespace(), Name: source.SecretKeyRef.Name}, s)
		if err != nil {
			return "", "", err
		}
		if val, ok := s.Data[source.SecretKeyRef.Key]; ok {
			return source.SecretKeyRef.Key, string(val), nil
		} else {
			return "", "", fmt.Errorf("missing key %s in secret %s", source.SecretKeyRef.Key, source.SecretKeyRef.Name)
		}
	}
	if source.ConfigMapKeyRef != nil {
		s := &v1.ConfigMap{}
		err := h.Client.Get(ctx, client.ObjectKey{Namespace: cr.GetNamespace(), Name: source.ConfigMapKeyRef.Name}, s)
		if err != nil {
			return "", "", err
		}
		if val, ok := s.Data[source.ConfigMapKeyRef.Key]; ok {
			return source.ConfigMapKeyRef.Key, val, nil
		} else {
			return "", "", fmt.Errorf("missing key %s in configmap %s", source.ConfigMapKeyRef.Key, source.ConfigMapKeyRef.Name)
		}
	}
	return "", "", fmt.Errorf("source couldn't be parsed source: %s", source)
}

// getContentModel resolves datasources, updates uid (if needed) and converts raw json to type grafana client accepts
func (h *AlertContentResolver) getAlertContentModel(contentJson []byte) (map[string]interface{}, string, error) {
	contentJson, err := h.resolveDatasources(contentJson)
	if err != nil {
		return map[string]interface{}{}, "", err
	}

	hash := sha256.New()
	hash.Write(contentJson)

	var contentModel map[string]interface{}
	err = json.Unmarshal(contentJson, &contentModel)
	if err != nil {
		return map[string]interface{}{}, "", err
	}

	// NOTE: id should never be hardcoded in a model, otherwise grafana will try to update a model by id instead of uid.
	//       And, in case the id is non-existent, grafana will respond with 404. https://github.com/grafana/grafana-operator/issues/1108
	contentModel["id"] = nil

	// uid, _ := contentModel["uid"].(string) //nolint:errcheck
	// contentModel["uid"] = CustomUIDOrUID(h.resource, uid)

	return contentModel, fmt.Sprintf("%x", hash.Sum(nil)), nil
}
