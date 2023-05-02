## Pod
```{image} img/pod-unl.svg
:width: 200px
```

###

Pod - это минимальная единица развертывания

```{revealjs-fragments}
* один или несколько контейнеров на одной ноде
* общая сеть
* общее хранилище
* общий набор параметров
```

###

```{revealjs-code-block} yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    ports:
    - containerPort: 80
```
```{revealjs-code-block} console
$ kubectl apply -f pod.yaml
```

### Несколько контейнеров

```{image} img/pod-multicontainer.svg
:width: 300px
:align: center
```

###

```{revealjs-code-block} yaml
---
data-line-numbers: 7-8|9-11|12-13|14-16|17-20
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    volumeMounts:
    - mountPath: /opt/nginx/conf
      name: conf
  - name: reloader
    image: config-reloader:1.2.3
    volumeMounts:
    - mountPath: /opt/nginx/conf
      name: conf
  volumes:
  - name: conf
    configMap:
      name: conf
```

### Жизненный цикл

```{image} img/lifecycle.svg
```

### Pod Phases

```{revealjs-fragments}
* **Pending** - под создан, но не назначен на ноду
* **Running** - под назначен и все контейнеры запущены
* **Succeeded** - все контейнеры успешно завершены
* **Failed** - все контейнеры завершены, но хотя бы один неуспешно
* **Unknown** - не удалось определить статус пода
```

### Pod Conditions

```{revealjs-fragments}
 * **PodScheduled** - под назначен на ноду
* **ContainersReady** - все контейнеры пода подготовлены
* **Initialized** - все инит контейнеры в поде выполнились успешно
* **Ready** - под готов принимать запросы
```

### Container states

```{revealjs-fragments}
* **Waiting** - подготовительные операции для старта
* **Running** - контейнер запустился и работает
* **Terminated** - контейнер завершился
```

### Container restart policy

```{revealjs-fragments}
* **Always** - при любом завершении контейнера
* **OnFailure** - при некорректном завершении
* **Never** - никогда не перезапускать
```

### Container hooks

```{revealjs-fragments}
* **PostStart** - запускается сразу после создания контейнера
* **PreStop** - запускается непосредственно перед завершением контейнера
```

### Container probes

```{revealjs-fragments}
* **exec** - выполняет команду внутри контейнера
* **httpGet** - выполняет HTTP GET запрос по IP пода
* **tcpSocket** - выполняет TCP проверку по IP и порту
* **grpc** - выполняет вызов удаленной процедуры по протоколу gRPC
```

### Container probes

```{revealjs-fragments}
* **livenessProbe** - определяет, что контейнер работает
* **readinessProbe** - определяет, что контейнер готов принимать запросы
* **startupProbe** - определяет, что приложение в контейнере запустилось
```

### Pod status

```{revealjs-code-block} yaml
---
data-line-numbers: 2|3-6|7-10|11-14|15-18|19|20-27|28-30|31-37
---
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2023-03-03T14:38:28Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2023-03-03T14:39:31Z"
    status: "True"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2023-03-03T14:39:31Z"
    status: "True"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2023-03-03T14:38:28Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: cri-o://83ae1cecc32ff20bc915438c265bae638b8741c371cd4b3006cbbfbf1aeb255c
    image: docker.io/library/nginx:latest
    imageID: docker.io/library/nginx@sha256:3f13b4376446cf92b0cb9a5c46ba75d57c41f627c4edb8b635fa47386ea29e20
    lastState: {}
    name: nginx
    ready: true
    restartCount: 0
    started: true
    state:
      running:
        startedAt: "2023-03-03T14:39:31Z"
  hostIP: 172.18.0.3
  phase: Running
  podIP: 10.244.2.5
  podIPs:
  - ip: 10.244.2.5
  qosClass: Burstable
  startTime: "2023-03-03T14:38:28Z"
```

### Pod status

```{revealjs-code-block} console
$ k get po
NAME    READY   STATUS    RESTARTS   AGE
nginx   1/1     Running   0          3d17h
```

### Pod termination
```{image} img/terminate.svg
```

### Конфигурация

###

```{revealjs-code-block} console
$ # kubectl explain pod.<field>
$ kubectl explain pod.spec.containers.image
KIND:     Pod
VERSION:  v1

FIELD:    image <string>

DESCRIPTION:
     Docker image name. More info:
     https://kubernetes.io/docs/concepts/containers/images This field is
     optional to allow higher level config management to default or override
     container images in workload controllers like Deployments and StatefulSets.
```

### Scheduling

```{revealjs-code-block} yaml
---
data-line-numbers: 6
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  nodeName: kind-worker
  containers:
  - name: nginx
    image: nginx
```

### Scheduling

```{revealjs-code-block} yaml
---
data-line-numbers: 6-7
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  nodeSelector:
    kubernetes.io/hostname: kind-worker
  containers:
  - name: nginx
    image: nginx
```

### Scheduling

```{revealjs-code-block} yaml
---
data-line-numbers: 6-14
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
            - kind-worker
  containers:
  - name: nginx
    image: nginx
```

### Pod Lifecycle

```{revealjs-code-block} yaml
---
data-line-numbers: 6-8
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  restartPolicy: Never
  terminationGracePeriodSeconds: 5
  activeDeadlineSeconds: 10
  containers:
  - image: nginx
    name: nginx
```

### Pod Volumes

```{revealjs-code-block} yaml
---
data-line-numbers: 12-14
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    volumeMounts:
    - name: data
      mountPath: /data/
  volumes:
  - name: data
    emptyDir: {}
```

### Name Resolution

```{revealjs-fragments}
* *hostname* - позволяет задать хостнейм для пода.

* *subdomain* - позволяет задать fqdn хостнейм для пода в виде\
  *hostname.subdomain.namespace.svc.cluster.local*.
```

### Name Resolution

```{revealjs-code-block} yaml
---
data-line-numbers: 6-14
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  hostAliases:
  - ip: "127.0.0.1"
    hostnames:
    - "foo.local"
    - "bar.local"
  - ip: "10.1.2.3"
    hostnames:
    - "foo.remote"
    - "bar.remote"
  containers:
  - image: nginx
    name: nginx
```

### Name Resolution

```{revealjs-code-block} yaml
---
data-line-numbers: 6|7-10
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  dnsPolicy: "None"
  dnsConfig:
    nameservers:
    - 8.8.8.8
    - 8.8.4.4
  containers:
  - image: nginx
    name: nginx
```

### Hosts namespaces

```{revealjs-code-block} yaml
---
data-line-numbers: 6-7
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  hostNetwork: true
  shareProcessNamespace: true
  containers:
  - image: nginx
    name: nginx
```

### Service account

```{revealjs-code-block} yaml
---
data-line-numbers: 6-7
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  serviceAccountName: default
  automountServiceAccountToken: false
  containers:
  - image: nginx
    name: nginx
```

### Pod Security Context

```{revealjs-code-block} yaml
---
data-line-numbers: 6-10
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    runAsGroup: 3000
    fsGroup: 2000
  containers:
  - image: nginx
    name: nginx
```

### Containers

```{revealjs-code-block} yaml
---
data-line-numbers: 6-7|8-11|12-14
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  imagePullSecrets:
  - name: registry-cred
  initContainers:
  - name: init
    image: busybox:1.28
    command: ['sh', '-c', 'echo hello']
  containers:
  - image: nginx
    name: nginx
```

### Конфигурация контейнера

### Image

```{revealjs-code-block} yaml
---
data-line-numbers: 6|8
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  imagePullPolicy: Always
  containers:
  - image: nginx
    name: nginx
```

### Entrypoint

```{revealjs-code-block} yaml
---
data-line-numbers: 9-11
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: echo
    image: busybox:1.28
    command: ['sh', '-c']
    args: ['echo $PWD']
    workingDir: /
```

### Ports

```{revealjs-code-block} yaml
---
data-line-numbers: 9-14
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    ports:
    - name: http
      containerPort: 80
      protocol: TCP
    - name: https
      containerPort: 443
```

### Environment

```{revealjs-code-block} yaml
---
data-line-numbers: 9-11|12-17|18-21
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    env:
    - name: KEY
      value: ENV_VALUE
    - name: KEY_FROM_SECRET
      valueFrom:
        secretKeyRef:
          name: config
          key: KEY
          optional: true
    envFrom:
    - configMapRef:
        name: config
        optional: true
```

### Container Volumes

```{revealjs-code-block} yaml
---
data-line-numbers: 9-13
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - mountPath: /opt/nginx/conf
      readOnly: true
      subPath: conf
      name: conf
  volumes:
  - name: conf
    configMap:
      name: conf
```

### Resources

```{revealjs-code-block} yaml
---
data-line-numbers: 9-15|10-12|13-15
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    resources:
      requests:
        cpu: "500m"
        memory: "100Mi"
      limits:
        cpu: "1"
        memory: "200Mi"
```

### Container Lifecycle

```{revealjs-code-block} yaml
---
data-line-numbers: 9-15
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    lifecycle:
      postStart:
        exec:
          command: ["/bin/sh", "-c", "echo > /var/index.html"]
      preStop:
        exec:
          command: ["/bin/sh","-c","nginx -s quit"]
```

### Container Lifecycle

```{revealjs-code-block} yaml
---
data-line-numbers: 9-15|16-20|21-29
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    startupProbe:
      exec:
        command:
        - cat
        - /tmp/healthy
      initialDelaySeconds: 5
      periodSeconds: 5
    readinessProbe:
      tcpSocket:
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 10
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080
        httpHeaders:
        - name: Custom-Header
          value: Awesome
      initialDelaySeconds: 3
      periodSeconds: 3
```

### Container Security Context

```{revealjs-code-block} yaml
---
data-line-numbers: 9-12
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    securityContext:
      readOnlyRootFilesystem: true
      privileged: true
```
