/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	llmv1alpha1 "github.com/vishalsanfran/llama-shepherd/api/v1alpha1"
)

// InferenceServiceReconciler reconciles a InferenceService object
type InferenceServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// RBAC markers for kubebuilder.
// +kubebuilder:rbac:groups=llm.example.com,resources=inferenceservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=llm.example.com,resources=inferenceservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=llm.example.com,resources=inferenceservices/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *InferenceServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	var isvc llmv1alpha1.InferenceService
	if err := r.Get(ctx, req.NamespacedName, &isvc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	replicas := int32(1)
	if isvc.Spec.Replicas != nil {
		replicas = *isvc.Spec.Replicas
	}

	deployName := isvc.Name + "-router"
	svcName := isvc.Name

	var kvpool llmv1alpha1.KVCachePool
	if isvc.Spec.CachePoolRef != "" {
		if err := r.Get(ctx, client.ObjectKey{
			Name:      isvc.Spec.CachePoolRef,
			Namespace: isvc.Namespace,
		}, &kvpool); err != nil {
			log.Error(err, "failed to find referenced KVCachePool")
			// Do NOT return error — instead, requeue later
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
	}

	cacheEndpoints := []string{}
	if isvc.Spec.CachePoolRef != "" {
		// routers will talk to pods directly (headless service)
		headlessSvcName := kvpool.Name + "-cache"
		cacheEndpoints = append(cacheEndpoints,
			fmt.Sprintf("%s.%s.svc.cluster.local:6379",
				headlessSvcName,
				kvpool.Namespace,
			),
		)
	}

	var deploy appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Name: deployName, Namespace: isvc.Namespace}, &deploy)
	if err != nil && errors.IsNotFound(err) {
		// Create a new router Deployment
		deploy = appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployName,
				Namespace: isvc.Namespace,
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
								Name: "router",
								// For now, use a simple HTTP echo server.
								// Later, replace with a real Go router image.
								Image: "ghcr.io/vishalsanfran/llama-shepherd-router:latest",
								Args: []string{
									"-text=llama-shepherd router: " + isvc.Spec.ModelRef,
								},
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										ContainerPort: 5678,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "MODEL_REF",
										Value: isvc.Spec.ModelRef,
									},
									{
										Name:  "MAX_CONCURRENCY",
										Value: strconv.Itoa(int(isvc.Spec.MaxConcurrency)),
									},
									{
										Name:  "KV_ENDPOINTS",
										Value: strings.Join(cacheEndpoints, ","),
									},
								},
							},
						},
					},
				},
			},
		}

		if err := ctrl.SetControllerReference(&isvc, &deploy, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, &deploy); err != nil {
			log.Error(err, "failed to create router Deployment", "deployment", deployName)
			return ctrl.Result{}, err
		}
		log.Info("created router Deployment", "deployment", deployName)
	} else if err == nil {
		// Update replicas if changed
		if deploy.Spec.Replicas == nil || *deploy.Spec.Replicas != replicas {
			deploy.Spec.Replicas = ptr.To(replicas)
			if err := r.Update(ctx, &deploy); err != nil {
				log.Error(err, "failed to update router Deployment replicas", "deployment", deployName)
				return ctrl.Result{}, err
			}
			log.Info("updated router Deployment replicas", "deployment", deployName, "replicas", replicas)
		}
	} else {
		return ctrl.Result{}, err
	}

	// Ensure Service exists
	var svc corev1.Service
	err = r.Get(ctx, client.ObjectKey{Name: svcName, Namespace: isvc.Namespace}, &svc)
	if err != nil && errors.IsNotFound(err) {
		svc = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: isvc.Namespace,
				Labels: map[string]string{
					"app": deployName,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": deployName,
				},
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(5678),
					},
				},
			},
		}

		if err := ctrl.SetControllerReference(&isvc, &svc, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, &svc); err != nil {
			log.Error(err, "failed to create router Service", "service", svcName)
			return ctrl.Result{}, err
		}
		log.Info("created router Service", "service", svcName)
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Update status with available replicas
	available := int32(0)
	if deploy.Status.AvailableReplicas > 0 {
		available = deploy.Status.AvailableReplicas
	}

	if isvc.Status.AvailableReplicas != available {
		isvc.Status.AvailableReplicas = available
		if err := r.Status().Update(ctx, &isvc); err != nil {
			log.Error(err, "failed to update InferenceService status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InferenceServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&llmv1alpha1.InferenceService{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("inferenceservice").
		Complete(r)
}
