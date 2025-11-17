🦙 llama-shepherd

A minimal Kubernetes Operator (written in Go with Kubebuilder) that manages LLM inference jobs.
It introduces a custom resource, LLMInferenceJob, and automatically creates a Kubernetes Job to execute simple text-based inference workloads.

📦 Getting Started

Prerequisites
	•	Go 1.24+
	•	Docker 17.03+
	•	kubectl 1.19+ (any modern version works)
	•	A running Kubernetes cluster (kind or Docker Desktop are fine)

Generate CRDs and RBAC manifests

Kubebuilder generates code + YAML based on what’s inside api/ and controllers/.
`make manifests`
This regenerates:
	•	CRDs → `config/crd/bases

Install CRDs
`make install`

Run the Operator Locally
`make run`

Create a Sample Inference Job

`kubectl apply -f config/samples/llm_v1alpha1_llminferencejob.yaml`

Check the Job:
```
kubectl get jobs
kubectl logs job/<your-job-name>
```

📤 Deploy to a Cluster

Build and push an image:
`make docker-build docker-push IMG=<registry>/llama-shepherd:<tag>`

Deploy the operator:
`make deploy IMG=<registry>/llama-shepherd:<tag>`

Uninstall:
```
make undeploy
make uninstall
```