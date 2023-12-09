# Workloads

В данной практике предлагаю испробовать работу [Horizontal Pod Autoscaler][hpa] в связке с
[StatefulSet][sts], увеличивая потребляемые ресурсы приложения с помощью [Cronjob][].


## Metrics server

В первую очередь для работы [HPA][] в кластер необходимо установить [metrics server][metrics-server],
который позволит hpa контроллеру получать информацию о потребляемых ресурсах подами приложения.

```console
$ k apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.6.3/components.yaml
$ k patch -n kube-system deployment metrics-server --type=json -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'
```

## Application

Для проверки работы [HPA][] предлагаю использовать масштабирование на основе потребления памяти
приложением. Потребуется простое приложение, которое сможет занять память заданного размера, для этого
можно использовать вызов [mmap][], отображая файл в оперативную память. Наше приложение будет отображать
файл заданного размера в оперативную память, таким образом можно будет регулировать потребляемую память.
Вы можете написать его на своем любимом языке, в качестве примера приведу простую реализацию на C:

```c
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

int main(int argc, char *argv[]) {
  if (argc < 3) {
    printf("too few arguments\n");
    return 1;
  }

  while (1) {
    int fd = open(argv[1], O_RDONLY);
    if (fd < 0) {
      printf("error read file: %s\n", argv[1]);
      sleep(strtol(argv[2], NULL, 10));
      continue;
    }
    struct stat s;
    int status = fstat(fd, &s);
    char *data = mmap(0, s.st_size, PROT_READ, MAP_SHARED, fd, 0);
    for (int i = 0; i < s.st_size; i++) {
      char c;
      c = data[i];
    }
    printf("size: %lld\n", s.st_size);
    sleep(strtol(argv[2], NULL, 10));
    munmap(data, s.st_size);
    close(fd);
  }
}
```

Данное программа перечитывает файл заданный первым аргументом с периодичностью заданной вторым
аргументом и отображает в оперативную память.

## Image

Для запуска приложения в kubernetes нам понадобится docker образ, для приведенной выше программы
Dockerfile будет выглядеть так:

```dockerfile
FROM alpine:3.17 as build

WORKDIR /build

RUN apk add --no-cache clang musl-dev binutils gcc
COPY main.c ./
RUN clang main.c

FROM alpine:3.17

COPY --from=build /build/a.out /app

CMD ["/app", "/tmp/data", "30"]
```

Таким образом для сборки образа можно выполнить следующие команды:
```bash
mkdir testapp;cd testapp
cat << EOF > main.c
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

int main(int argc, char *argv[]) {
  if (argc < 3) {
    printf("too few arguments\n");
    return 1;
  }

  while (1) {
    int fd = open(argv[1], O_RDONLY);
    if (fd < 0) {
      printf("error read file: %s\n", argv[1]);
      sleep(strtol(argv[2], NULL, 10));
      continue;
    }
    struct stat s;
    int status = fstat(fd, &s);
    char *data = mmap(0, s.st_size, PROT_READ, MAP_SHARED, fd, 0);
    for (int i = 0; i < s.st_size; i++) {
      char c;
      c = data[i];
    }
    printf("size: %lld\n", s.st_size);
    sleep(strtol(argv[2], NULL, 10));
    munmap(data, s.st_size);
    close(fd);
  }
}
EOF
cat << EOF > Dockerfile
FROM alpine:3.17 as build

WORKDIR /build

RUN apk add --no-cache clang musl-dev binutils gcc
COPY main.c ./
RUN clang main.c

FROM alpine:3.17

COPY --from=build /build/a.out /app

CMD ["/app", "/tmp/data", "30"]
EOF
docker build -t app:test .
```

Появится образ `app:test`, который необходимо загрузить в кластер. В утилите `kind` для этого есть
команда `kind load docker-image`, которая загрузит образ на все ноды кластера.

```console
$ kind load docker-image app:test
Image: "" with ID "sha256:7966712b4b2f550489f136128f48fafd8d0224c85f6770ae4562714ce5a0c202" not yet present on node "kind-worker", loading...
Image: "" with ID "sha256:7966712b4b2f550489f136128f48fafd8d0224c85f6770ae4562714ce5a0c202" not yet present on node "kind-worker2", loading...
Image: "" with ID "sha256:7966712b4b2f550489f136128f48fafd8d0224c85f6770ae4562714ce5a0c202" not yet present on node "kind-control-plane", loading...
```

## StatefulSet

Создадим [StatefulSet][sts] с собранным образом, указав в нем:
- `resources.requests` - необходимые ресурсы для работы пода, также [hpa][] будет учитывать данный
  параметр при расчете необходимости увеличения количества реплик
- `resources.limits` - максимально возможное потребление для пода
- `image: app:test` - собранный образ
- `imagePullPolicy: Never` - для того, чтобы kubelet не пытался скачать собранный локально образ
  из внешнего источника
- `volumeClaimTemplates` с необходимыми параметрами дискового хранилища, в котором будет лежать файл,
  загружаемый в память
- `volumeMounts` с указанием пути для монтирования хранилища

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: app
spec:
  selector:
    matchLabels:
      app: app
  podManagementPolicy: OrderedReady
  replicas: 1
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: app:test
        imagePullPolicy: Never
        resources:
          limits:
            memory: 300Mi
          requests:
            memory: 100Mi
        volumeMounts:
        - name: data
          mountPath: /tmp
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 300Mi
```

Для создания можно воспользоваться командой `kubectl apply`:
```bash
k apply -f - << EOF
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: app
spec:
  selector:
    matchLabels:
      app: app
  podManagementPolicy: OrderedReady
  replicas: 1
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: app:test
        imagePullPolicy: Never
        resources:
          limits:
            memory: 300Mi
          requests:
            memory: 100Mi
        volumeMounts:
        - name: data
          mountPath: /tmp
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 300Mi
EOF
```

После создания пода при установленном [metrics server][metrics-server] потребление ресурсов подами
можно посмотреть командой `kubectl top pod`:
```console
$ k top pod
NAME    CPU(cores)   MEMORY(bytes)
app-0   1m           1Mi
```

## Horizontal Pod Autoscaler

Для конфигурации автоматического масштабирования необходимо создать ресурс `HorizontalPodAutoscaler`,
где необходимо указать:
- `maxReplicas` - максимальное количество реплик, до которого можно масштабироваться
- `minReplicas` - минимальное количество реплик
- `averageUtilization` - средняя утилизация между всеми репликами в процентах от параметра
  `resource.requests`

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app
spec:
  maxReplicas: 5
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: app
  metrics:
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 60
```

```{note}
`HorizontalPodAutoscaler` имеет несколько версий своей спецификации, важно указывать правильную версию
в поле `apiVersion`. Управление на основе ресурса памяти есть в стабильной версии `autoscaling/v2`.
Информацию по параметрам конкретной версии можно посмотреть командой `kubectl explain` с указанием
опции `--api-version`, например `k explain --api-version='autoscaling/v2' hpa`.
```

Для создания можно воспользоваться командой `kubectl apply`:
```bash
k apply -f - << EOF
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app
spec:
  maxReplicas: 5
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: app
  metrics:
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 60
EOF
```

После создания можно посмотреть текущее состояние:
```console
$ k get hpa
NAME   REFERENCE         TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
app    StatefulSet/app   1%/60%    1         5         1          4s
```

Как видно из вывода - текущее среднее потребление памяти составляет 1% от `resource.requests`.

## Cronjob

Теперь необходимо сделать механизм для постепенного увеличения потребления, для этого можно сделать
небольшой скрипт, который будет создавать файл заданного размера внутри каждого пода. Сохранять состояние
между запусками скрипта можно в ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app
data:
  max: "500"
  size: "50"
```

Где size - текущий размер, который необходимо распределить между подами, max - максимальный размер.

Сам скрипт будет выглядеть так:
```bash
#!/bin/bash
set -e

NAME="app"
ADD="10"

data="$(kubectl get configmap "$NAME" -o json)"
max="$(echo "$data" | jq -rc '.data.max')"
size="$(echo "$data" | jq -rc '.data.size')"

size="$((size + ADD))"
if [[ "$size" -gt "$max" ]];then
  size="$max"
fi

pods=($(kubectl get pods -l app="$NAME" -o name))
count="${#pods[@]}"
pod_size="$((size/count))"

for pod in "${pods[@]}";do
  echo "$pod - $pod_size"
  kubectl exec "$pod" -- /bin/sh -c \
    "dd if=/dev/zero of=/tmp/data.tmp bs=1M count=$pod_size;mv /tmp/data.tmp /tmp/data"
done

kubectl patch configmap "$NAME" -p "{\"data\":{\"size\":\"$size\"}}"
```

Сам скрипт можно также сохранить в ConfigMap.

Для работы данного скрипта внутри кластера потребуется дать права сервисному аккаунту default, для этого
необходимо создать Role и RoleBinding такого вида:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: app
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/exec
  - configmaps
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: app
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: app
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
```

Применить данные ресурсы в кластер можно также командой `kubectl apply`:
```bash
k apply -f - << EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: app
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/exec
  - configmaps
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: app
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: app
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app
data:
  max: "500"
  size: "50"
EOF
```

```bash
k apply -f - << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-job-sh
data:
  job.sh: |
    #!/bin/bash
    set -e

    NAME="app"
    ADD="10"

    data="$(kubectl get configmap "$NAME" -o json)"
    max="$(echo "$data" | jq -rc '.data.max')"
    size="$(echo "$data" | jq -rc '.data.size')"

    size="$((size + ADD))"
    if [[ "$size" -gt "$max" ]];then
      size="$max"
    fi

    pods=($(kubectl get pods -l app="$NAME" -o name))
    count="${#pods[@]}"
    pod_size="$((size/count))"

    for pod in "${pods[@]}";do
      echo "$pod - $pod_size"
      kubectl exec "$pod" -- /bin/sh -c \
        "dd if=/dev/zero of=/tmp/data.tmp bs=1M count=$pod_size;mv /tmp/data.tmp /tmp/data"
    done

    kubectl patch configmap "$NAME" -p "{\"data\":{\"size\":\"$size\"}}"
EOF
```

Для запуска скрипта с заданной периодичностью можно воспользоваться ресурсом [Cronjob][]:
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  labels:
    app: app
  name: app
spec:
  jobTemplate:
    metadata:
      name: app
    spec:
      template:
        metadata:
          labels:
            app: app-job
        spec:
          containers:
          - image: alpine/k8s:1.25.8
            name: app
            command:
            - /bin/bash
            - /script/job.sh
            resources: {}
            volumeMounts:
            - name: script
              mountPath: /script
          restartPolicy: OnFailure
          volumes:
          - name: script
            configMap:
              name: app-job-sh
  schedule: '*/5 * * * *'
```

Данный Cronjob будет запускать созданный скрипт каждые 5 минут, создадим также через `kubectl apply`:
```bash
k apply -f - << EOF
apiVersion: batch/v1
kind: CronJob
metadata:
  labels:
    app: app
  name: app
spec:
  jobTemplate:
    metadata:
      name: app
    spec:
      template:
        metadata:
          labels:
            app: app-job
        spec:
          containers:
          - image: alpine/k8s:1.25.8
            name: app
            command:
            - /bin/bash
            - /script/job.sh
            resources: {}
            volumeMounts:
            - name: script
              mountPath: /script
          restartPolicy: OnFailure
          volumes:
          - name: script
            configMap:
              name: app-job-sh
  schedule: '*/5 * * * *'
EOF
```

## Scaling

После запуска скрипта из [CronJob][] с некоторой задержкой будет отображаться новое потребление у пода:
```console
$ k top po
NAME    CPU(cores)   MEMORY(bytes)
app-0   10m          64Mi
$ k get hpa
NAME   REFERENCE         TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
app    StatefulSet/app   64%/60%   1         5         1          52m
```

Последующие запуски будут увеличивать потребление и разделять между подами, в свою очередь [hpa][]
при прохождении порога в 60% средней нагрузки между подами и нахождении за этим порогом достаточное
время будет увеличивать количество реплик.

```console
$ k top po
NAME    CPU(cores)   MEMORY(bytes)
app-0   19m          44Mi
app-1   7m           43Mi
$ k get hpa
NAME   REFERENCE         TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
app    StatefulSet/app   44%/60%   1         5         2          61m
```

По итогу постепенно развернутся все поды до максимального количества, заданного в hpa:
```console
$ k top po
NAME    CPU(cores)   MEMORY(bytes)
app-0   0m           58Mi
app-1   0m           59Mi
app-2   14m          58Mi
app-3   17m          57Mi
app-4   0m           57Mi
$ k get hpa
NAME   REFERENCE         TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
app    StatefulSet/app   58%/60%   1         5         5          159m
```

```{note}
Чтобы процесс шел быстрее можно увеличить частоту запуска Job в CronJob через параметр `schedule`.
```

После можно вернуть количество потребляемых ресурсов в ConfigMap в 0:
```console
$ k patch cm app -p '{"data":{"size": "0"}}'
configmap/app patched
```

Что через некоторое время приведет к уменьшению количества реплик:
```console
$ k get po -l app=app
NAME                 READY   STATUS        RESTARTS   AGE
app-0                1/1     Running       0          3h48m
app-1                1/1     Running       0          117m
app-2                0/1     Terminating   0          87m
$ k get po -l app=app
NAME    READY   STATUS    RESTARTS   AGE
app-0   1/1     Running   0          3h49m
$ k get hpa
NAME   REFERENCE         TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
app    StatefulSet/app   7%/60%    1         5         1          172m
```

Все примеры, используемые здесь, можно также посмотреть на [github][example].

[hpa]:https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[sts]:https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/
[cronjob]:https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/
[metrics-server]:https://github.com/kubernetes-sigs/metrics-server
[mmap]:https://en.wikipedia.org/wiki/Mmap
[example]:https://github.com/yudolevich/kube-dev-course/tree/main/examples/workloads/
