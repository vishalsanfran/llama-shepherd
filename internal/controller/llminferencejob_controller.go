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

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	llmv1alpha1 "github.com/vishalsanfran/llama-shepherd/api/v1alpha1"
)

// LLMInferenceJobReconciler reconciles a LLMInferenceJob object
type LLMInferenceJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=llm.example.com,resources=llminferencejobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=llm.example.com,resources=llminferencejobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=llm.example.com,resources=llminferencejobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LLMInferenceJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *LLMInferenceJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	var cr llmv1alpha1.LLMInferenceJob
	err := r.Get(ctx, req.NamespacedName, &cr)
	if err != nil {
		return ctrl.Result{}, nil
	}

	if cr.Status.Completed {
		return ctrl.Result{}, nil
	}

	jobName := cr.Name + "-runner"
	var k8sJob batchv1.Job
	err = r.Get(ctx, client.ObjectKey{Name: jobName, Namespace: cr.Namespace}, &k8sJob)
	if err == nil {
		if k8sJob.Status.Succeeded > 0 {
			cr.Status.Completed = true
			cr.Status.Output = "finished (dummy output)"
			r.Status().Update(ctx, &cr)
		}
	}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: cr.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "runner",
							Image: "busybox",
							Command: []string{
								"sh", "-c", fmt.Sprintf("echo '%s'; sleep 2", cr.Spec.Prompt),
							},
						},
					},
				},
			},
		},
	}

	err = ctrl.SetControllerReference(&cr, &job, r.Scheme)
	if err != nil {
		return ctrl.Result{}, nil
	}
	err = r.Create(ctx, &job)
	if err != nil {
		return ctrl.Result{}, nil
	}

	log.Info("created job", "job", jobName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LLMInferenceJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&llmv1alpha1.LLMInferenceJob{}).
		Named("llminferencejob").
		Complete(r)
}
