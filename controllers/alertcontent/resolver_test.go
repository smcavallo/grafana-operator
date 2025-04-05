package content

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAlertRuleEnvs(t *testing.T) {
	alertRule := v1beta1.GrafanaAlertRuleGroup{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-alert-rule-group",
			Namespace: "grafana-operator-system",
		},
		Spec: v1beta1.GrafanaAlertRuleGroupSpec{
			GrafanaAlertContentSpec: v1beta1.GrafanaAlertContentSpec{
				Envs: []v1beta1.GrafanaAlertContentEnv{
					{
						Name:  "TEST_ENV",
						Value: "test-env-value",
					},
				},
			},
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	defer stop()

	var contentResource v1beta1.GrafanaAlertContentResource = &alertRule
	assert.NotNil(t, contentResource.GrafanaAlertContentSpec(), "resource does not properly implement content spec or status fields; this indicates a bug in implementation")
	assert.NotNil(t, contentResource.GrafanaAlertContentStatus(), "resource does not properly implement content spec or status fields; this indicates a bug in implementation")

	resolver := NewContentResolver(&alertRule, k8sClient)

	envs, err := resolver.getAlertContentEnvs(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, envs)
	assert.True(t, len(envs) == 1, "Expected 1 env, got %d", len(envs))
}
