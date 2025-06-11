# anareta
Automated Namespace-based Assertion &amp; Resiliency Evaluation and Telemetry Agent.

anareta doesn’t spin up new clusters—it simply creates a dedicated Kubernetes namespace (e.g. `anareta-<devenv-name>`) for each feature branch. Within that namespace your application’s Helm release, ConfigMaps, Secrets and other resources live in complete isolation from other workloads. The operator attaches a TTL to each DevEnv so after your specified duration it automatically deletes the namespace and everything inside, giving you fast, per-branch sandboxes on a single shared cluster without the overhead of managing separate control planes.

The eventual intent is to provide a way to run the same tests/chaos engineering and telemetry in an isolated environment as you would in production, but without the need to create a separate cluster for each feature branch. This allows for faster development cycles and easier testing of new features.
