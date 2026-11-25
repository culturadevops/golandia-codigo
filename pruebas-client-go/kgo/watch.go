package main

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// El API server corta los watch por su cuenta (min-request-timeout, 30-60 min por defecto).
// En la demo lo forzamos a los 5 s con TimeoutSeconds para verlo pasar.
var cinco = int64(5)

// escalarDespues cambia las réplicas de web a 3 y luego a 1, mientras el watch mira.
func escalarDespues(ctx context.Context, cs kubernetes.Interface) {
	for _, paso := range []struct {
		en time.Duration
		n  int32
	}{{1 * time.Second, 3}, {8 * time.Second, 1}} {
		time.Sleep(paso.en - time.Since(t0))
		sc, err := cs.AppsV1().Deployments(ns).GetScale(ctx, "web", metav1.GetOptions{})
		if err != nil {
			logf("escalar: %v", err)
			return
		}
		sc.Spec.Replicas = paso.n
		_, err = cs.AppsV1().Deployments(ns).UpdateScale(ctx, "web", sc, metav1.UpdateOptions{})
		logf("[escalador] replicas=%d %v", paso.n, errOK(err))
	}
}

func errOK(err error) string {
	if err != nil {
		return err.Error()
	}
	return "ok"
}

func imprimir(ev watch.Event, modificados *int) {
	p, ok := ev.Object.(*corev1.Pod)
	if !ok {
		logf("%s %T", ev.Type, ev.Object)
		return
	}
	if ev.Type == watch.Modified {
		*modificados++
		return
	}
	logf("%-8s %s", ev.Type, p.Name)
}

func watchIngenuo(ctx context.Context, cs kubernetes.Interface) error {
	go escalarDespues(ctx, cs)
	w, err := cs.CoreV1().Pods(ns).Watch(ctx, metav1.ListOptions{
		LabelSelector: "app=web", TimeoutSeconds: &cinco})
	if err != nil {
		return err
	}
	mod := 0
	for ev := range w.ResultChan() {
		imprimir(ev, &mod)
	}
	logf("el canal se cerró (%d MODIFIED) · el bucle terminó, el programa ya no ve nada", mod)
	time.Sleep(14 * time.Second)
	pods, err := cs.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: "app=web"})
	if err != nil {
		return err
	}
	logf("realidad al final: %d pods de web", len(pods.Items))
	return nil
}

func watchReintento(ctx context.Context, cs kubernetes.Interface) error {
	pods := cs.CoreV1().Pods(ns)
	lista, err := pods.List(ctx, metav1.ListOptions{LabelSelector: "app=web"})
	if err != nil {
		return err
	}
	rv := lista.ResourceVersion
	logf("lista inicial: %d pods · resourceVersion %s", len(lista.Items), rv)
	go escalarDespues(ctx, cs)
	mod, aperturas := 0, 0
	for time.Since(t0) < 22*time.Second {
		w, err := pods.Watch(ctx, metav1.ListOptions{
			LabelSelector: "app=web", ResourceVersion: rv, TimeoutSeconds: &cinco})
		if err != nil {
			return err
		}
		aperturas++
		for ev := range w.ResultChan() {
			if ev.Type == watch.Error {
				return apierrors.FromObject(ev.Object)
			}
			if m, err := meta.Accessor(ev.Object); err == nil {
				rv = m.GetResourceVersion()
			}
			imprimir(ev, &mod)
		}
		logf("canal cerrado · reabro desde resourceVersion %s", rv)
	}
	logf("%d watch abiertos · %d MODIFIED · peticiones:%s", aperturas, mod, resumenPeticiones())

	// ¿Y si reabres con un resourceVersion demasiado viejo?
	w, err := pods.Watch(ctx, metav1.ListOptions{LabelSelector: "app=web", ResourceVersion: "1", TimeoutSeconds: &cinco})
	if err != nil {
		logf("watch desde resourceVersion 1: %v", err)
		return nil
	}
	viejos := 0
	for ev := range w.ResultChan() {
		if ev.Type == watch.Error {
			e := apierrors.FromObject(ev.Object)
			logf("watch desde resourceVersion 1: %s · código %d · IsResourceExpired=%v",
				e.Error(), e.(apierrors.APIStatus).Status().Code, apierrors.IsResourceExpired(e))
			w.Stop()
			return nil
		}
		viejos++
	}
	logf("watch desde resourceVersion 1: sin error · %d eventos viejos repetidos", viejos)
	return nil
}
