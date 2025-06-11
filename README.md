# anareta
Automated Namespace-based Assertion &amp; Resiliency Evaluation and Telemetry Agent.

anareta doesn’t spin up new clusters—it simply creates a dedicated Kubernetes namespace (e.g. `anareta-<devenv-name>`) for each feature branch. Within that namespace your application’s Helm release, ConfigMaps, Secrets and other resources live in complete isolation from other workloads. The operator attaches a TTL to each DevEnv so after your specified duration it automatically deletes the namespace and everything inside, giving you fast, per-branch sandboxes on a single shared cluster without the overhead of managing separate control planes.
