## KVCachePool

This represents a distributed KV cache pool backing inference workloads. This resource will manage a Deployment of cache nodes (initially Redis), track readiness, and expose endpoints for use by InferenceService router pods. This is the first foundational step toward a distributed KV-cache-aware inference system

Spec Fields
• totalMemoryGB — total memory across KV nodes
• replicas — number of cache nodes
• strategy — eviction/placement strategy

Controller Behavior

The controller should:
1. Fetch/create/update a Deployment named -cache
2. Sync .spec.replicas with the Deployment replica count
3. Pass totalMemoryGB and strategy as environment variables
4. Create a headless Service (ClusterIP: None)
• Enables router pods to discover individual KV nodes
5. Update .status.readyReplicas from Deployment status

This creates a proper “cache pool” control-plane object.

```
+------------------------+
                         |     InferenceService   |
                         |       (control plane)  |
                         +-----------+------------+
                                     |
                                     | selects KVCachePool
                                     v
                      +------------------------------+
                      |        KVCachePool CRD       |
                      |   (desired replicas, strategy)|
                      +------------------------------+
                                     |
                                     | reconciles
                                     v
                     +---------------------------------+
                     |   Headless Service (ClusterIP: None)
                     |   Name: <pool-name>-cache
                     +------------------+--------------+
                                        |
             +--------------------------+---------------------------+
             |                          |                           |
             v                          v                           v
   +----------------+        +----------------+        +----------------+
   |  KV Node Pod 1 |        |  KV Node Pod 2 |        |  KV Node Pod 3 |
   |   (cache-0)    |        |   (cache-1)    |        |   (cache-2)    |
   +--------+-------+        +--------+-------+        +--------+-------+
            ^                         ^                         ^
            |                         |                         |
            +-------------+-----------+-------------+-----------+
                          |                         |
                          v                         v
                 +----------------------+   +----------------------+
                 |      Router Pod 1    |   |     Router Pod 2     |
                 | (part of Deployment) |   | (part of Deployment) |
                 +-----------+----------+   +-----------+----------+
                             |                          |
                             |  Discovers ALL KV nodes  |
                             |    via headless Service  |
                             +------------+-------------+
                                          |
                                          v
                               Direct Pod-by-Pod Access
```

This diagram shows how an InferenceService connects to a distributed KV cache through a KVCachePool managed by the operator. The KVCachePool controller creates a headless Service (ClusterIP: None), which exposes the individual pod IPs of all KV cache nodes instead of load-balancing them behind a single virtual IP.

Router pods created by the InferenceService use this headless Service to directly discover each KV node, enabling advanced distributed-systems behaviors such as:
• sharded KV placement
• cache locality awareness
• custom eviction policies
• pod-level load balancing

This structure mirrors how systems like Cassandra, Redis Cluster, and Kafka expose nodes for topology-aware clients.