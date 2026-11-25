package main

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// rbac prueba, desde dentro de un pod, qué le deja hacer su ServiceAccount.
func rbac(ctx context.Context, cs kubernetes.Interface) {
	probar("list pods en demo", func() error {
		_, err := cs.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		return err
	})
	probar("get deployment web", func() error {
		_, err := cs.AppsV1().Deployments(ns).Get(ctx, "web", metav1.GetOptions{})
		return err
	})
	probar("list pods en kube-system", func() error {
		_, err := cs.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
		return err
	})
}

func probar(que string, f func() error) {
	err := f()
	if err == nil {
		logf("%-26s ok", que)
		return
	}
	var codigo int32
	if s, ok := err.(apierrors.APIStatus); ok {
		codigo = s.Status().Code
	}
	logf("%-26s %d · IsForbidden=%v", que, codigo, apierrors.IsForbidden(err))
	logf("  %v", err)
}
