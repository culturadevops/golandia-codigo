package main

import (
	"context"
	"sync/atomic"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func kb(n int64) float64 { return float64(n) / 1024 }

func compararLecturas(ctx context.Context, cs kubernetes.Interface) error {
	// A · listar en un bucle apretado, con el rate limiter por defecto de client-go
	reiniciarContador()
	inicio := time.Now()
	for i := 0; i < 100; i++ {
		if _, err := cs.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{}); err != nil {
			return err
		}
	}
	logf("A · 100 List seguidos: %.1fs · %d peticiones · %.1f KB", time.Since(inicio).Seconds(), peticiones.Load(), kb(bytesLeidos.Load()))

	// B · sondeo: un List por segundo durante 20 s
	reiniciarContador()
	inicio = time.Now()
	for time.Since(inicio) < 20*time.Second {
		if _, err := cs.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	logf("B · sondeo 1/s durante 20 s: %d peticiones · %.1f KB", peticiones.Load(), kb(bytesLeidos.Load()))

	// C · informer: un List y un Watch; después todo se lee de la caché
	reiniciarContador()
	fab := informers.NewSharedInformerFactoryWithOptions(cs, 0, informers.WithNamespace(ns))
	inf := fab.Core().V1().Pods()
	lister := inf.Lister()
	var eventos atomic.Int64
	inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(any) { eventos.Add(1) },
		UpdateFunc: func(_, _ any) { eventos.Add(1) },
		DeleteFunc: func(any) { eventos.Add(1) },
	})
	stop := make(chan struct{})
	defer close(stop)
	inicio = time.Now()
	fab.Start(stop)
	fab.WaitForCacheSync(stop)
	logf("C · caché sincronizada en %s · peticiones:%s", time.Since(inicio).Round(time.Millisecond), resumenPeticiones())

	antes := peticiones.Load()
	inicio = time.Now()
	var pods []*corev1.Pod
	for i := 0; i < 100; i++ {
		var err error
		if pods, err = lister.Pods(ns).List(labels.Everything()); err != nil {
			return err
		}
	}
	logf("C · 100 List del lister: %s · %d pods · peticiones nuevas: %d", time.Since(inicio), len(pods), peticiones.Load()-antes)

	go escalarInformer(ctx, cs)
	time.Sleep(20 * time.Second)
	n, _ := lister.Pods(ns).List(labels.Everything())
	logf("C · 20 s más, con escalado 1→3→1 de por medio: %d eventos recibidos · lister ve %d pods · peticiones totales:%s · %.1f KB",
		eventos.Load(), len(n), resumenPeticiones(), kb(bytesLeidos.Load()))
	return nil
}

// escalarInformer usa otro cliente para no contar sus peticiones en el informer.
func escalarInformer(ctx context.Context, _ kubernetes.Interface) {
	cfg, _, err := cargarConfig()
	if err != nil {
		return
	}
	otro := kubernetes.NewForConfigOrDie(cfg)
	for _, n := range []int32{3, 1} {
		time.Sleep(4 * time.Second)
		sc, err := otro.AppsV1().Deployments(ns).GetScale(ctx, "web", metav1.GetOptions{})
		if err != nil {
			return
		}
		sc.Spec.Replicas = n
		otro.AppsV1().Deployments(ns).UpdateScale(ctx, "web", sc, metav1.UpdateOptions{})
	}
}
