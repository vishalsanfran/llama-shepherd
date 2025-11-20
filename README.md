## 🦙 llama-shepherd

llama-shepherd is a modular Kubernetes Operator (written in Go using Kubebuilder) for exploring LLM runtime control planes, including:
	•	LLMInferenceJob — batch-style inference jobs
	•	InferenceService — scalable router-based online inference endpoints
	•	KVCachePool — distributed KV cache pool for advanced scheduling & vLLM-style research

This project is designed as a foundation for experiments in LLM scheduling, distributed KV caching, and cluster-level inference orchestration.

Full feature details live in the /docs directory:
	•	docs/InferenceJob.md￼ — Batch-style inference (LLMInferenceJob)
	•	docs/InferenceService.md￼ — Router-based online inference
	•	docs/KVCachePool.md￼ — Distributed KV cache pool (headless service)

These documents describe the CRDs, controllers, and architecture in depth.

## 📦 Getting Started

Prerequisites
	•	Go 1.24+
	•	Docker 17.03+
	•	kubectl 1.19+ (any modern version works)
	•	A running Kubernetes cluster (kind or Docker Desktop are fine)

### ⚙️ Install CRDs

Kubebuilder generates code + YAML based on what’s inside api/ and controllers/.
```
make manifests
make install
```

### Run the Operator Locally
`make run`

 Try an Example (LLMInferenceJob)

```
kubectl apply -f config/samples/llm_v1alpha1_llminferencejob.yaml
kubectl get jobs
kubectl logs job/<your-job-name>
```

Additional examples for InferenceService and KVCachePool are provided in
config/samples/.

📤 Deploy to a Cluster

Build and push an image:
`make docker-build docker-push IMG=<registry>/llama-shepherd:<tag>`

Deploy the operator:
`make deploy IMG=<registry>/llama-shepherd:<tag>`

Cleanup:
```
make undeploy
make uninstall
```