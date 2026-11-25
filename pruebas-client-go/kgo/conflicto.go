package main

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func ptr[T any](v T) *T { return &v }

func conflicto(ctx context.Context, cs kubernetes.Interface) error {
	deps := cs.AppsV1().Deployments(ns)

	// Dos copias del mismo objeto: como dos programas que lo leyeron a la vez.
	a, err := deps.Get(ctx, "web", metav1.GetOptions{})
	if err != nil {
		return err
	}
	b, _ := deps.Get(ctx, "web", metav1.GetOptions{})
	logf("a y b leídos · replicas=%d · resourceVersion %s", *a.Spec.Replicas, a.ResourceVersion)

	a.Spec.Replicas = ptr(int32(2))
	a2, err := deps.Update(ctx, a, metav1.UpdateOptions{})
	logf("Update(a) replicas=2: %s · resourceVersion ahora %s", errOK(err), a2.ResourceVersion)

	b.Labels["equipo"] = "pagos"
	_, err = deps.Update(ctx, b, metav1.UpdateOptions{})
	logf("Update(b) con la copia vieja: IsConflict=%v", apierrors.IsConflict(err))
	logf("  %v", err)

	// Lo que evita el resourceVersion: sin él, el update de b pisa las réplicas de a.
	b.ResourceVersion = ""
	b3, err := deps.Update(ctx, b, metav1.UpdateOptions{})
	if err != nil {
		logf("Update(b) sin resourceVersion: %v", err)
	} else {
		logf("Update(b) sin resourceVersion: ok · replicas=%d (a había puesto 2)", *b3.Spec.Replicas)
	}

	// Un patch manda solo lo que cambia: no lleva resourceVersion y no pisa nada.
	_, _ = deps.Patch(ctx, "web", types.MergePatchType, []byte(`{"spec":{"replicas":2}}`), metav1.PatchOptions{})
	p, err := deps.Patch(ctx, "web", types.MergePatchType,
		[]byte(`{"metadata":{"labels":{"equipo":"datos"}}}`), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	logf("Patch(labels): ok · replicas=%d · equipo=%s", *p.Spec.Replicas, p.Labels["equipo"])

	// Si de verdad necesitas Update: vuelve a leer y reintenta.
	intentos := 0
	viejo, _ := deps.Get(ctx, "web", metav1.GetOptions{})
	_, _ = deps.Patch(ctx, "web", types.MergePatchType, []byte(`{"metadata":{"labels":{"otro":"si"}}}`), metav1.PatchOptions{})
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		intentos++
		if intentos > 1 {
			var e error
			if viejo, e = deps.Get(ctx, "web", metav1.GetOptions{}); e != nil {
				return e
			}
		}
		viejo.Spec.Replicas = ptr(int32(1))
		_, e := deps.Update(ctx, viejo, metav1.UpdateOptions{})
		return e
	})
	logf("RetryOnConflict: %s en %d intentos", errOK(err), intentos)
	return nil
}
