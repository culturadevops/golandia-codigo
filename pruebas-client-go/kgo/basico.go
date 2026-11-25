package main

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func listarPods(ctx context.Context, cs kubernetes.Interface) error {
	pods, err := cs.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, p := range pods.Items {
		logf("%-20s %-45s %s", p.Namespace, p.Name, p.Status.Phase)
	}
	logf("%d pods · resourceVersion de la lista: %s · peticiones:%s", len(pods.Items), pods.ResourceVersion, resumenPeticiones())
	return nil
}

func deploymentWeb() *appsv1.Deployment {
	uno := int32(1)
	et := map[string]string{"app": "web"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Labels: et},
		Spec: appsv1.DeploymentSpec{
			Replicas: &uno,
			Selector: &metav1.LabelSelector{MatchLabels: et},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: et},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name: "nginx", Image: "nginx:1.30.5-alpine", ImagePullPolicy: corev1.PullIfNotPresent,
				}}},
			},
		},
	}
}

func desplegar(ctx context.Context, cs kubernetes.Interface) error {
	deps := cs.AppsV1().Deployments(ns)
	d, err := deps.Create(ctx, deploymentWeb(), metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		logf("ya existe: %v", err)
		d, err = deps.Get(ctx, "web", metav1.GetOptions{})
	}
	if err != nil {
		return err
	}
	logf("creado %s/%s · resourceVersion %s", d.Namespace, d.Name, d.ResourceVersion)
	if err := esperarListas(ctx, cs, 1); err != nil {
		return err
	}

	// Escalar por la subresource /scale: solo toca spec.replicas.
	sc, err := deps.GetScale(ctx, "web", metav1.GetOptions{})
	if err != nil {
		return err
	}
	sc.Spec.Replicas = 3
	if _, err := deps.UpdateScale(ctx, "web", sc, metav1.UpdateOptions{}); err != nil {
		return err
	}
	logf("escalado a 3")
	if err := esperarListas(ctx, cs, 3); err != nil {
		return err
	}
	logf("peticiones:%s", resumenPeticiones())
	return nil
}

func esperarListas(ctx context.Context, cs kubernetes.Interface, n int32) error {
	inicio := time.Now()
	for {
		d, err := cs.AppsV1().Deployments(ns).Get(ctx, "web", metav1.GetOptions{})
		if err != nil {
			return err
		}
		if d.Status.ReadyReplicas == n && d.Status.UpdatedReplicas == n {
			logf("%d/%d listas en %.1fs", d.Status.ReadyReplicas, n, time.Since(inicio).Seconds())
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
}
