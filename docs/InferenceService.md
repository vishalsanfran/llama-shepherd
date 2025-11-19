
## InferenceService

InferenceService defines a logical inference endpoint.

This endpoint:
	•	abstracts away pods, deployments, replicas, services
	•	exposes a consistent model entrypoint
	•	standardizes how inference workloads are handled
	•	becomes a unit of scaling, routing, scheduling, and optimization

This is similar to:

Kubeflow → InferenceService

KServe → ISVC

vLLM → serving endpoint configs

OpenAI → model deployment docs

HuggingFace TGI → model server



What are “router pods”?

A router pod is a Deployment-managed pod that:

🔹 Accepts inference requests (HTTP or gRPC)

🔹 Queues them

🔹 Batches them (future step)

🔹 Sends them to worker pods (future LLMModel CRD)

🔹 Optionally performs scheduling / concurrency / QoS logic

The operator is the control plane, and router pods are the data plane.