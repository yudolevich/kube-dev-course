## Operators

```{image} img/operatorframework.svg
:width: 200px
```

### Operator

* Custom Resource
* Control Loop

### Operator

Examples:
```{revealjs-fragments}
* Etcd Operator
* Prometheus Operator
* MySQL Operator
```

### Custom Resources

```{revealjs-fragments}
* API Aggregation
* CRD
```

### API Aggregation

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|4|6-7|8-9|10-12|13
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: <name of the registration object>
spec:
  group: <API group name this extension apiserver hosts>
  version: <API version this extension apiserver hosts>
  groupPriorityMinimum: <priority this APIService for this group, see API documentation>
  versionPriority: <prioritizes ordering of this version within a group, see API documentation>
  service:
    namespace: <namespace of the extension apiserver service>
    name: <name of the extension apiserver service>
  caBundle: <pem encoded ca cert that signs the server cert used by the webhook>
```

### API Aggregation

```{image} img/api-aggregation.svg
:width: 800px
```

### CRD

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|4-5|7-8|9-10|11-20|12-13|14-15|16-17|18-20|23-40|23-27|28-40
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: crontabs.stable.example.com
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: stable.example.com
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: crontabs
    # singular name to be used as an alias on the CLI and for display
    singular: crontab
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CronTab
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - ct
  # list of versions supported by this CustomResourceDefinition
  versions:
    - name: v1
      # Each version can be enabled/disabled by Served flag.
      served: true
      # One and only one version must be marked as the storage version.
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                cronSpec:
                  type: string
                image:
                  type: string
                replicas:
                  type: integer
```

### Operator Pattern

```{image} img/operator_pattern.png
:width: 1000px
```

### Operator Components

```{image} img/operator_components.png
:width: 1000px
```

### Kubernetes Clients

Officially-supported
| | | |
| -------- | ------ | ----------- |
| C	       | Go     | CSharp      |
| Haskell	 | Java	  | JavaScript	|
| Perl	   | Python | Ruby	      |
| | | |

### Kubernetes Clients

Community-maintained
| | | |
| -------- | ------ | ----------- |
| Clojure	 | Elixir | Lisp        |
| Node.js	 | PHP	  | Rust      	|
| Scala	   | Swift  | DotNet      |
| | | |

### Operator Frameworks

* KUDO (Kubernetes Universal Declarative Operator)
* Metacontroller (along with WebHooks that you implement yourself)
* Shell-operator

### Operator Frameworks

* Java Operator SDK
* Kopf (Kubernetes Operator Pythonic Framework)
* KubeOps (.NET operator SDK)

### Operator Frameworks

* kubebuilder
* Operator Framework

### Operator Framework

```{revealjs-fragments}
* BUILD: operator-sdk
* MANAGE: olm
* DISCOVER: operatorhub.io
```

### Operator SDK

```{revealjs-fragments}
* GO
* Ansible
* Helm
```

### Operator SDK

```{revealjs-code-block} console
---
data-line-numbers: 1-3|4-6|7-19
---
$ operator-sdk init \
  --domain example.com \
  --repo github.com/example/memcached-operator
$ operator-sdk create api \
  --group cache --version v1alpha1 \
  --kind Memcached --resource --controller
$ ls -1
Dockerfile
Makefile
PROJECT
README.md
api/
bin/
config/
controllers/
go.mod
go.sum
hack/
main.go
```

### Operator SDK

`api/v1alpha1/memcached_types.go`
```{revealjs-code-block} go
---
data-line-numbers: 1-2|3-4|5-9|11-13|15-17|20-21|22-29|31-33|36-44
---
// MemcachedSpec defines the desired state of Memcached
type MemcachedSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=5
	// +kubebuilder:validation:ExclusiveMaximum=false

	// Size defines the number of Memcached instances
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Size int32 `json:"size,omitempty"`

	// Port defines the port that will be used to init the container with the image
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ContainerPort int32 `json:"containerPort,omitempty"`
}

// MemcachedStatus defines the observed state of Memcached
type MemcachedStatus struct {
	// Represents the observations of a Memcached's current state.
	// Memcached.status.conditions.type are: "Available", "Progressing", and "Degraded"
	// Memcached.status.conditions.status are one of True, False, Unknown.
	// Memcached.status.conditions.reason the value should be a CamelCase string and producers of specific
	// condition types may define expected values and meanings for this field, and whether the values
	// are considered a guaranteed API.
	// Memcached.status.conditions.Message is a human readable message indicating details about the transition.
	// For further information see: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// Conditions store the status conditions of the Memcached instances
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// Memcached is the Schema for the memcacheds API
//+kubebuilder:subresource:status
type Memcached struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MemcachedSpec   `json:"spec,omitempty"`
	Status MemcachedStatus `json:"status,omitempty"`
}
```

### Operator SDK

```{revealjs-code-block} console
---
data-line-numbers: 1-2|3-4
---
$ make generate
# api/v1alpha1/zz_generated.deepcopy.go
$ make manifests
# config/crd/bases/cache.example.com_memcacheds.yaml
```

### Operator SDK

`controllers/memcached_controller.go`
```{revealjs-code-block} go
---
data-line-numbers: 1-6|3|4|8-15
---
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Memcached{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(
  ctx context.Context, req ctrl.Request,
) (ctrl.Result, error) {
  // Lookup the Memcached instance for this reconcile request
  memcached := &cachev1alpha1.Memcached{}
  err := r.Get(ctx, req.NamespacedName, memcached)
  ...
}
```

### Operator SDK

```{revealjs-code-block} console
$ make manifests
$ IMG=example.com/memcached-operator:v0.0.1 make docker-build
$ IMG=example.com/memcached-operator:v0.0.1 make deploy
```
