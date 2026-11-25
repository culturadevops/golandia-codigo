// El controlador del operador (vídeo 517 de golandia). Reemplaza al que genera `kubebuilder create api`.
// Bucle de reconciliación: mira la PaginaWeb y deja el clúster como dice: un ConfigMap con el HTML, un Deployment de nginx
// y un Service. Todo con el dueño puesto (ownerReference): si se borra la PaginaWeb, Kubernetes borra lo demás.
package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	webv1 "paginas/api/v1"
)

// PaginaWebReconciler deja el clúster como dice cada PaginaWeb.
type PaginaWebReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=web.golandia.dev,resources=paginawebs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.golandia.dev,resources=paginawebs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services;configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *PaginaWebReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var p webv1.PaginaWeb
	if err := r.Get(ctx, req.NamespacedName, &p); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // borrada: sus hijos los borra Kubernetes
	}
	etiquetas := map[string]string{"app": p.Name}

	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Data = map[string]string{"index.html": "<h1>" + p.Spec.Mensaje + "</h1>\n"}
		return controllerutil.SetControllerReference(&p, cm, r.Scheme)
	}); err != nil {
		return ctrl.Result{}, err
	}

	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace}}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, dep, func() error {
		dep.Spec.Replicas = &p.Spec.Replicas
		dep.Spec.Selector = &metav1.LabelSelector{MatchLabels: etiquetas}
		dep.Spec.Template.Labels = etiquetas
		dep.Spec.Template.Spec.Containers = []corev1.Container{{
			Name:         "nginx",
			Image:        "nginx:1.29-alpine",
			Ports:        []corev1.ContainerPort{{ContainerPort: 80}},
			VolumeMounts: []corev1.VolumeMount{{Name: "html", MountPath: "/usr/share/nginx/html"}},
		}}
		dep.Spec.Template.Spec.Volumes = []corev1.Volume{{Name: "html", VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: p.Name}}}}}
		return controllerutil.SetControllerReference(&p, dep, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		log.Info("Deployment", "pagina", p.Name, "operacion", op, "replicas", p.Spec.Replicas)
	}

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec.Selector = etiquetas
		svc.Spec.Ports = []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt32(80)}}
		return controllerutil.SetControllerReference(&p, svc, r.Scheme)
	}); err != nil {
		return ctrl.Result{}, err
	}

	if p.Status.ReplicasListas != dep.Status.ReadyReplicas {
		p.Status.ReplicasListas = dep.Status.ReadyReplicas
		if err := r.Status().Update(ctx, &p); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager: además de las PaginaWeb, vigila lo que crea (Owns). Si alguien borra el Deployment, vuelve a entrar aquí.
func (r *PaginaWebReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1.PaginaWeb{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Named("paginaweb").
		Complete(r)
}
