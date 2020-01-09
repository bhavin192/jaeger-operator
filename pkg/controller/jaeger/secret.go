package jaeger

import (
	"context"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/jaegertracing/jaeger-operator/pkg/apis/jaegertracing/v1"
	"github.com/jaegertracing/jaeger-operator/pkg/inventory"
	"github.com/jaegertracing/jaeger-operator/pkg/util"
)

func (r *ReconcileJaeger) applySecrets(jaeger v1.Jaeger, desired []corev1.Secret) error {
	opts := []client.ListOption{
		client.InNamespace(jaeger.Namespace),
		client.MatchingLabels(util.ProcessLabels(
			map[string]string{
				"app.kubernetes.io/instance":   jaeger.Name,
				"app.kubernetes.io/managed-by": "jaeger-operator",
			})),
	}
	list := &corev1.SecretList{}
	if err := r.client.List(context.Background(), list, opts...); err != nil {
		return err
	}

	inv := inventory.ForSecrets(list.Items, desired)
	for _, d := range inv.Create {
		jaeger.Logger().WithFields(log.Fields{
			"secret":    d.Name,
			"namespace": d.Namespace,
		}).Debug("creating secrets")
		if err := r.client.Create(context.Background(), &d); err != nil {
			return err
		}
	}

	for _, d := range inv.Update {
		jaeger.Logger().WithFields(log.Fields{
			"secret":    d.Name,
			"namespace": d.Namespace,
		}).Debug("updating secrets")
		if err := r.client.Update(context.Background(), &d); err != nil {
			return err
		}
	}

	for _, d := range inv.Delete {
		jaeger.Logger().WithFields(log.Fields{
			"secret":    d.Name,
			"namespace": d.Namespace,
		}).Debug("deleting secrets")
		if err := r.client.Delete(context.Background(), &d); err != nil {
			return err
		}
	}

	return nil
}
