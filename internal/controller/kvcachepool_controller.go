package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	llmv1alpha1 "github.com/vishalsanfran/llama-shepherd/api/v1alpha1"
)

// KVCachePoolReconciler reconciles a KVCachePool object
type KVCachePoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=llm.example.com,resources=kvcachepools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=llm.example.com,resources=kvcachepools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=llm.example.com,resources=kvcachepools/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

func (r *KVCachePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// 1. Load KVCachePool
	var pool llmv1alpha1.KVCachePool
	if err := r.Get(ctx, req.NamespacedName, &pool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Derive desired replicas
	replicas := int32(1)
	if pool.Spec.Replicas != nil {
		replicas = *pool.Spec.Replicas
	}

	deployName := pool.Name + "-cache"

	// 2. Ensure Deployment exists
	var deploy appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Name: deployName, Namespace: pool.Namespace}, &deploy)
	if err != nil && apierrors.IsNotFound(err) {
		// Create a new Deployment of "cache nodes".
		deploy = appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployName,
				Namespace: pool.Namespace,
				Labels: map[string]string{
					"app": deployName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To(replicas),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": deployName,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": deployName,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "cache",
								// Placeholder KV node. Later, replace with your own Go-based KV service.
								Image: "redis:7-alpine",
								Ports: []corev1.ContainerPort{
									{
										Name:          "redis",
										ContainerPort: 6379,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "KVCACHE_STRATEGY",
										Value: pool.Spec.Strategy,
									},
									{
										Name:  "KVCACHE_TOTAL_MEMORY_GB",
										Value: fmt.Sprintf("%d", pool.Spec.TotalMemoryGB),
									},
								},
							},
						},
					},
				},
			},
		}

		// 3. Ensure headless Service exists
		svcName := pool.Name + "-cache"

		var svc corev1.Service
		err = r.Get(ctx, client.ObjectKey{Name: svcName, Namespace: pool.Namespace}, &svc)
		if err != nil && apierrors.IsNotFound(err) {
			// Need to create headless Service
			svc = corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      svcName,
					Namespace: pool.Namespace,
					Labels: map[string]string{
						"app": deployName,
					},
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "None", // headless service
					Selector: map[string]string{
						"app": deployName,
					},
					Ports: []corev1.ServicePort{
						{
							Name:       "redis",
							Port:       6379,
							TargetPort: intstr.FromInt(6379),
						},
					},
				},
			}

			if err := ctrl.SetControllerReference(&pool, &svc, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}

			if err := r.Create(ctx, &svc); err != nil {
				log.Error(err, "failed to create headless KV Service", "service", svcName)
				return ctrl.Result{}, err
			}

			log.Info("created headless KV Service", "service", svcName)

		} else if err != nil {
			// Some other error
			return ctrl.Result{}, err
		}

		if err := ctrl.SetControllerReference(&pool, &deploy, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, &deploy); err != nil {
			log.Error(err, "failed to create KV cache Deployment", "deployment", deployName)
			return ctrl.Result{}, err
		}

		log.Info("created KV cache Deployment", "deployment", deployName)
	} else if err == nil {
		// Keep replicas in sync
		if deploy.Spec.Replicas == nil || *deploy.Spec.Replicas != replicas {
			deploy.Spec.Replicas = ptr.To(replicas)
			if err := r.Update(ctx, &deploy); err != nil {
				log.Error(err, "failed to update KV cache Deployment replicas", "deployment", deployName)
				return ctrl.Result{}, err
			}
			log.Info("updated KV cache Deployment replicas", "deployment", deployName, "replicas", replicas)
		}
	} else {
		return ctrl.Result{}, err
	}

	// 3. Update status.ReadyReplicas based on Deployment status
	ready := deploy.Status.ReadyReplicas
	if pool.Status.ReadyReplicas != ready {
		pool.Status.ReadyReplicas = ready
		if err := r.Status().Update(ctx, &pool); err != nil {
			log.Error(err, "failed to update KVCachePool status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KVCachePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&llmv1alpha1.KVCachePool{}).
		Owns(&appsv1.Deployment{}).
		Named("kvcachepool").
		Complete(r)
}
