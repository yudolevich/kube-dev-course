# Основы работы с kubernetes

## kind
Для взаимодействия с кластером kubernetes нам в первую очередь необходим сам кластер.
Для локальной работы и обучения предлагаю воспользоваться открытым проектом
[kind(Kubernetes in Docker)][kind], который позволяет запустить локальный кластер в Docker контейнере.

### Установка
Варианты установки описаны на сайте [kind.sigs.k8s.io][kind-install], можно воспользоваться
пакетным менеджером или достаточно просто скачать и положить в каталог из переменной `PATH`
исполняемый файл для своей ОС и архитектуры со страницы [releases][kind-releases].

Вариант для Linux:
```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.17.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### Создание кластера

Для создания кластера достаточно запустить команду:
```bash
kind create cluster
```

Посмотреть созданные кластера можно командой:
```bash
kind get clusters
```

## kubectl
Для взаимодействия с кластером нам понадобится утилита [kubectl][kubectl-overview] - это инструмент
командной строки для управления кластерами Kubernetes.

### Установка
Варианты установки описаны на сайте [kubernetes][kubectl-install] для каждой ОС. Здесь также достаточно
скачать исполняемый файл и положить в каталог из переменной `PATH`.

Вариант для Linux:
```bash
curl -Lo ./kubectl \
  "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```

### Конфигурация
По умолчанию конфигурация находится в домашней директории по пути `~/.kube/config`, переопределить
путь конфигурационного файла можно либо через переменную среды `KUBECONFIG`, либо явно указывая
опцию `--kubeconfig`. Сам файл конфигурации можно посмотреть командой `kubectl config view`.
```bash
kubectl config view --kubeconfig=~/.kube/config1
export KUBECONFIG=~/.kube/config2
kubectl config view
```

```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: DATA+OMITTED
    server: https://127.0.0.1:6443
  name: kind-kind
contexts:
- context:
    cluster: kind-kind
    namespace: default
    user: kind-kind
  name: kind-kind
current-context: kind-kind
kind: Config
preferences: {}
users:
- name: kind-kind
  user:
    client-certificate-data: DATA+OMITTED
    client-key-data: DATA+OMITTED
```

В самой конфигурации можно выделить основные параметры:
* clusters - список кластеров, где указаны параметры подключения
* users - список учетных данных для подключения
* contexts - список именованных объединений конкретного кластера с учетными данными для подключения
* current-context - текущий выбранный контекст с которым и будет работать утилита kubectl

### Автодополнение
Для удобного использования утилиты `kubectl` можно воспользоваться функцией автоматического дополнения
вводимой команды. На текущий момент это работает для командных оболочек `bash`, `zsh`, `fish` и
`powershell`. Для подробной информации можно выполнить команду `kubectl completion -h`.

Вариант для bash:
```bash
# для текущей сессии bash, должен быть установлен пакет bash-completion
source <(kubectl completion bash)

# для запуска при инициализации bash
echo "source <(kubectl completion bash)" >> ~/.bashrc
```

Также для удобного использования можно создать псевдоним для `kubectl` в виде команды `k`:
```bash
# для текущей сессии bash
alias k=kubectl
complete -F __start_kubectl k

# для запуска при инициализации bash
echo -e "alias k=kubectl\ncomplete -F __start_kubectl k" >> ~/.bashrc
```

### Информация о ресурсах
Получить список всех типов ресурсов в кластере можно командой `kubectl api-resources`:
```console
$ k api-resources | head
NAME                   SHORTNAMES   APIVERSION   NAMESPACED   KIND
bindings                            v1           true         Binding
componentstatuses      cs           v1           false        ComponentStatus
configmaps             cm           v1           true         ConfigMap
endpoints              ep           v1           true         Endpoints
events                 ev           v1           true         Event
limitranges            limits       v1           true         LimitRange
namespaces             ns           v1           false        Namespace
nodes                  no           v1           false        Node
persistentvolumeclaims pvc          v1           true         PersistentVolumeClaim
```

Здесь можно увидеть следующие столбцы:
* NAME - имя ресурса
* SHORTNAMES - короткие имена, которые можно использовать при работе с этим ресурсом
* APIVERSION - поле apiVersion, которое используется в объектах данного типа ресурса
* NAMESPACED - является ли данный ресурс namespaced или cluster-wide
* KIND - поле kind, которое используется в объектах данного типа ресурса

Для получения информации о ресурсе можно воспользоваться командой `kubectl explain`:
```console
$ k explain pod
KIND:     Pod
VERSION:  v1

DESCRIPTION:
     Pod is a collection of containers that can run on a host. This resource is
     created by clients and scheduled onto hosts.

FIELDS:
   apiVersion   <string>
     APIVersion defines the versioned schema of this representation of an
     object. Servers should convert recognized schemas to the latest internal
     value, and may reject unrecognized values. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources

   kind <string>
     Kind is a string value representing the REST resource this object
     represents. Servers may infer this from the endpoint the client submits
     requests to. Cannot be updated. In CamelCase. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds

   metadata     <Object>
     Standard object's metadata. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata

   spec <Object>
     Specification of the desired behavior of the pod. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

   status       <Object>
     Most recently observed status of the pod. This data may not be up to date.
     Populated by the system. Read-only. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
```
Тут представлено описание `DESCRIPTION`, а также в `FIELDS` все известные api поля и их тип.
Можно обратиться к конкретному полю за дополнительной вложенной информацией:
```console
$ k explain pod.spec | head -20
KIND:     Pod
VERSION:  v1

RESOURCE: spec <Object>

DESCRIPTION:
     Specification of the desired behavior of the pod. More info:
     https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

     PodSpec is a description of a pod.

FIELDS:
   activeDeadlineSeconds        <integer>
     Optional duration in seconds the pod may be active on the node relative to
     StartTime before the system will actively try to mark it failed and kill
     associated containers. Value must be a positive integer.

   affinity     <Object>
     If specified, the pod's scheduling constraints

$ k explain pod.spec.containers.command
KIND:     Pod
VERSION:  v1

FIELD:    command <[]string>

DESCRIPTION:
     Entrypoint array. Not executed within a shell. The container image's
     ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME)
     are expanded using the container's environment. If a variable cannot be
     resolved, the reference in the input string will be unchanged. Double $$
     are reduced to a single $, which allows for escaping the $(VAR_NAME)
     syntax: i.e. "$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
     Escaped references will never be expanded, regardless of whether the
     variable exists or not. Cannot be updated. More info:
     https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
```

### Работа с объектами
Утилита `kubectl` позволяет производить [CRUD][] операции над объектами в kubernetes. Для этого
можно использовать команды `create`, `get`, `replace`, `patch`, `delete` - по их названию можно
понять какие операции они производят. Отдельно хочется выделить команды `apply`, `diff`,
`describe` и `edit`:
* **apply** - создает или обновляет существующий объект, сохраняя примененную конфигурацию в
  аннотацию `kubectl.kubernetes.io/last-applied-configuration` для расчета необходимых изменений
  при последующих обновлениях
* **diff** - показывает изменения, которые произойдут в кластере при применении новой конфигурации объекта
* **describe** - выдает состояние объекта в удобочитаемом формате
* **edit** - открывает ресурс для редактирования в текстовом редакторе, определенном в переменной среды
  `KUBE_EDITOR` или `EDITOR`

Также есть ряд специализированных команд, которые позволяют выполнять конкретные действия в кластере:
`run`, `logs`, `set`, `label`, `exec`, `debug`, `scale` и многие другие. Их мы рассмотрим позже.

#### Полезные функции
Прежде чем начать работать с объектами хочется затронуть некоторые полезные функции в утилите `kubectl`.

Через флаг `-v<N>` можно задать уровень логирования утилиты, где `N` - необходимый уровень. Это полезно
для отладки или для обучения, так как позволит понять какой именно REST запрос отправляется на сервер.
Отмечу три полезных уровня:
* `-v6` - отображается HTTP запрос отправляемый в api
* `-v8` - помимо запроса также отображается и тело ответа от api
* `-v10` - максимальный уровень, отображает также и пример данного запроса с использованием утилиты
  `curl` для возможности сформировать запрос в ручном режиме

Через флаг `-o` или `--output` можно управлять форматом вывода.
* `-o name` выводит только имена объектов
* `-o wide` расширенный вывод в табличном виде
* `-o yaml` выводит содержимое в `yaml` формате
* `-o json` выводит содержимое в `json` формате
* `-o jsonpath=` выводит поля заданные выражением [jsonpath][]
* `-o custom-columns=` выводит содержимое в табличном виде с заданными столбцами

Через флаг `--dry-run` можно производить пробный прогон без реальных изменений в кластере.
Есть два варианта:
* `--dry-run=client` - клиент сам формирует конфигурацию объекта без отправки его в api
* `--dry-run=server` - клиент формирует конфигурацию объекта с отправкой в api, а api производит
  операции валидации и мутации объекта, но не сохраняет его в базе

#### Работа c объектами на примере kind: Pod
Для создания объекта пода есть специальная команда `run` -\
`kubectl run --image=<docker_image> <name>`:

```console
$ k run --image nginx nginx --dry-run=client -o yaml # для начала посмотрим содержимое без применения
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: null
  labels:
    run: nginx
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    resources: {}
  dnsPolicy: ClusterFirst
  restartPolicy: Always
status: {}
$ k run --image nginx nginx
pod/nginx created
```

Для получения информации о созданном поде можно воспользоваться командами `get` и `describe`:
```console
$ k get po
NAME    READY   STATUS    RESTARTS   AGE
nginx   1/1     Running   0          80s
$ k describe po nginx
Name:             nginx
Namespace:        default
Priority:         0
Service Account:  default
Node:             kind-control-plane/172.18.0.2
Start Time:       Sun, 19 Feb 2023 16:54:24 +0300
Labels:           run=nginx
Annotations:      <none>
Status:           Running
IP:               10.244.0.6
IPs:
  IP:  10.244.0.6
Containers:
  nginx:
    Container ID:   containerd://5b640dd7561025918ada550c1c513cf60392acc443a5c27ee64504efa0024cba
    Image:          nginx
    Image ID:       docker.io/library/nginx@sha256:6650513efd1d27c1f8a5351cbd33edf85cc7e0d9d0fcb4ffb23d8fa89b601ba8
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Sun, 19 Feb 2023 16:54:37 +0300
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-dzptg (ro)
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True
Volumes:
  kube-api-access-dzptg:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type    Reason     Age    From               Message
  ----    ------     ----   ----               -------
  Normal  Scheduled  2m18s  default-scheduler  Successfully assigned default/nginx to kind-control-plane
  Normal  Pulling    2m17s  kubelet            Pulling image "nginx"
  Normal  Pulled     2m5s   kubelet            Successfully pulled image "nginx" in 11.713549374s
  Normal  Created    2m5s   kubelet            Created container nginx
  Normal  Started    2m5s   kubelet            Started container nginx
```

Выполнить команду внутри контейнера можно командой `kubectl exec <pod> -- <command>`:
```console
$ kubectl exec nginx -- cat /proc/1/cmdline
nginx: master process nginx -g daemon off;
```

Посмотреть лог можно командой `kubectl logs <pod>`, если в поде несколько контейнеров, то указать
конкретный можно через ключ `-c` - `kubectl logs <pod> -c <container>`:
```console
$ k logs nginx
/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
/docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
10-listen-on-ipv6-by-default.sh: info: Getting the checksum of /etc/nginx/conf.d/default.conf
10-listen-on-ipv6-by-default.sh: info: Enabled listen on IPv6 in /etc/nginx/conf.d/default.conf
/docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
/docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
/docker-entrypoint.sh: Configuration complete; ready for start up
2023/02/19 13:54:37 [notice] 1#1: using the "epoll" event method
2023/02/19 13:54:37 [notice] 1#1: nginx/1.23.3
2023/02/19 13:54:37 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6)
2023/02/19 13:54:37 [notice] 1#1: OS: Linux 4.19.130-boot2docker
2023/02/19 13:54:37 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
2023/02/19 13:54:37 [notice] 1#1: start worker processes
2023/02/19 13:54:37 [notice] 1#1: start worker process 36
2023/02/19 13:54:37 [notice] 1#1: start worker process 37
2023/02/19 13:54:37 [notice] 1#1: start worker process 38
2023/02/19 13:54:37 [notice] 1#1: start worker process 39
```

Если процесс внутри контейнера завершился и перезапустился, то лог предыдущей работы можно посмотреть
с помощью флага `--previous`:
```console
$ k exec nginx -- /bin/sh -c 'kill 1'
$ k get po nginx
NAME    READY   STATUS    RESTARTS      AGE
nginx   1/1     Running   2 (22s ago)   83m
$ k logs nginx --previous
/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
/docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
10-listen-on-ipv6-by-default.sh: info: Getting the checksum of /etc/nginx/conf.d/default.conf
10-listen-on-ipv6-by-default.sh: info: Enabled listen on IPv6 in /etc/nginx/conf.d/default.conf
/docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
/docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
/docker-entrypoint.sh: Configuration complete; ready for start up
2023/02/19 15:11:03 [notice] 1#1: using the "epoll" event method
2023/02/19 15:11:03 [notice] 1#1: nginx/1.23.3
2023/02/19 15:11:03 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6)
2023/02/19 15:11:03 [notice] 1#1: OS: Linux 4.19.130-boot2docker
2023/02/19 15:11:03 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
2023/02/19 15:11:03 [notice] 1#1: start worker processes
2023/02/19 15:11:03 [notice] 1#1: start worker process 35
2023/02/19 15:11:03 [notice] 1#1: start worker process 36
2023/02/19 15:11:03 [notice] 1#1: start worker process 37
2023/02/19 15:11:03 [notice] 1#1: start worker process 38
2023/02/19 15:17:19 [notice] 1#1: signal 15 (SIGTERM) received from 82, exiting
2023/02/19 15:17:19 [notice] 35#35: exiting
2023/02/19 15:17:19 [notice] 35#35: exit
2023/02/19 15:17:19 [notice] 36#36: exiting
2023/02/19 15:17:19 [notice] 36#36: exit
2023/02/19 15:17:19 [notice] 37#37: exiting
2023/02/19 15:17:19 [notice] 37#37: exit
2023/02/19 15:17:19 [notice] 38#38: exiting
2023/02/19 15:17:19 [notice] 38#38: exit
2023/02/19 15:17:19 [notice] 1#1: signal 17 (SIGCHLD) received from 36
2023/02/19 15:17:19 [notice] 1#1: worker process 35 exited with code 0
2023/02/19 15:17:19 [notice] 1#1: worker process 36 exited with code 0
2023/02/19 15:17:19 [notice] 1#1: signal 29 (SIGIO) received
2023/02/19 15:17:19 [notice] 1#1: signal 17 (SIGCHLD) received from 37
2023/02/19 15:17:19 [notice] 1#1: worker process 37 exited with code 0
2023/02/19 15:17:19 [notice] 1#1: worker process 38 exited with code 0
2023/02/19 15:17:19 [notice] 1#1: exit
```

На объекты можно навешивать метки, чтобы потом было удобно фильтровать объекты по определенным меткам.
Для этого можно либо изменить непосредственно в самом объекте секцию `metadata.labels`,
либо воспользоваться командой `kubectl label`:
```console
$ k label pod nginx app=old
pod/nginx labeled
```

Давайте добавим еще пару подов и пролейблим их:
```console
$ k run --image nginx nginx1
pod/nginx1 created
$ k run --image nginx nginx2
pod/nginx2 created
pod/nginx labeled
$ k label pod nginx1 app=new
pod/nginx1 labeled
$ k label pod nginx2 app=new
pod/nginx2 labeled
```

Посмотреть все лейблы на объектах можно добавив к команде `kubectl get` опцию `--show-labels`:
```console
$ k get po --show-labels
NAME     READY   STATUS    RESTARTS       AGE     LABELS
nginx    1/1     Running   2              4h13m   app=old,run=nginx
nginx1   1/1     Running   0              3h20m   app=new,run=nginx1
nginx2   1/1     Running   0              3h20m   app=new,run=nginx2
```

Теперь при получении объектов мы можем использовать селектор с помощью ключа `-l`:
```console
$ k get po -l app=old
NAME    READY   STATUS    RESTARTS   AGE
nginx   1/1     Running   0          59m
$ k get po -l app=new
NAME     READY   STATUS    RESTARTS   AGE
nginx1   1/1     Running   0          6m44s
nginx2   1/1     Running   0          6m42s
```

Через запятую можно указывать несколько лейблов `-l key1=value1,key2=value2`, объект должен
удовлетворять сразу всем условиям:
```console
$ k get po -l run=nginx1,app=new
NAME     READY   STATUS    RESTARTS   AGE
nginx1   1/1     Running   0          11m
```

Операция просмотра логов также может выполняться с селектором, для добавления информации из какого пода
конкретная строка лога можно с помощью опции `--prefix`:
```console
$ k logs -l app=new --prefix
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: using the "epoll" event method
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: nginx/1.23.3
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6)
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: OS: Linux 4.19.130-boot2docker
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: start worker processes
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: start worker process 36
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: start worker process 37
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: start worker process 38
[pod/nginx1/nginx1] 2023/02/19 14:47:46 [notice] 1#1: start worker process 39
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: using the "epoll" event method
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: nginx/1.23.3
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6)
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: OS: Linux 4.19.130-boot2docker
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: start worker processes
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: start worker process 35
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: start worker process 36
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: start worker process 37
[pod/nginx2/nginx2] 2023/02/19 14:47:49 [notice] 1#1: start worker process 38
```

Аннотации можно добавлять с помощью команды `kubectl annotate` также как и лейблы:
```console
$ k annotate pod nginx key=value
pod/nginx annotated
```

Для изменения пода можно воспользоваться командой `kubectl edit`, которая откроет его в текстовом
редакторе определенном в переменной среды `KUBE_EDITOR`:
```console
$ export KUBE_EDITOR=nano
$ k edit po nginx
  UW PICO 5.09  File: /var/folders/62/_p4ngvx105qdskglcfwb2j8hr9_rkn/T/kubectl-edit-4203624632.yaml   Modified

# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
apiVersion: v1
kind: Pod
metadata:
  annotations:
    key: value1
  creationTimestamp: "2023-02-19T13:54:24Z"

^G Get Help       ^O WriteOut       ^R Read File      ^Y Prev Pg        ^K Cut Text       ^C Cur Pos
^X Exit           ^J Justify        ^W Where is       ^V Next Pg        ^U UnCut Text     ^T To Spell
```
Изменим значение аннотации и сохраним `Ctrl-X`:
```console
pod/nginx edited
```

Другой способ редактирования объектов возможен через сохранение текущего состояния в файл,
редактирование этого файла и последующее применение в кластере:

* Сперва мы получаем под командой `kubectl get` и сохраняем в текущей директории
  ```console
  $ k get po nginx -o yaml > nginx.yaml
  ```
* Редактируем данный файл удобным средствами
  ```console
  $ sed -i 's/value1/value2/' nginx.yaml
  ```
* Можем посмотреть разницу значений в файле с значениями в кластере командой `kubectl diff`
  ```console
  $ k diff -f nginx.yaml
  diff -u -N /var/folders/62/_p4ngvx105qdskglcfwb2j8hr9_rkn/T/LIVE-3347262426/v1.Pod.default.nginx /var/folders/62/_p4ngvx105qdskglcfwb2j8hr9_rkn/T/MERGED-1669273324/v1.Pod.default.nginx
  --- /var/folders/62/_p4ngvx105qdskglcfwb2j8hr9_rkn/T/LIVE-3347262426/v1.Pod.default.nginx       2023-02-19 20:25:10.000000000 +0300
  +++ /var/folders/62/_p4ngvx105qdskglcfwb2j8hr9_rkn/T/MERGED-1669273324/v1.Pod.default.nginx     2023-02-19 20:25:10.000000000 +0300
  @@ -2,7 +2,7 @@
   kind: Pod
   metadata:
     annotations:
  -    key: value1
  +    key: value2
     creationTimestamp: "2023-02-19T13:54:24Z"
     labels:
       app: old
  ```
* Применяем значения из файла в кластере командой `kubectl apply`
  ```console
  $ k apply -f nginx.yaml
  Warning: resource pods/nginx is missing the kubectl.kubernetes.io/last-applied-configuration annotation which is required by kubectl apply. kubectl apply should only be used on resources created declaratively by either kubectl create --save-config or kubectl apply. The missing annotation will be patched automatically.
  pod/nginx configured
  ```


[kind]:https://kind.sigs.k8s.io/
[kind-install]:https://kind.sigs.k8s.io/docs/user/quick-start/#installation
[kind-releases]:https://github.com/kubernetes-sigs/kind/releases
[kubectl-overview]:https://kubernetes.io/ru/docs/reference/kubectl/overview/
[kubectl-install]:https://kubernetes.io/docs/tasks/tools/#kubectl
[crud]:https://ru.wikipedia.org/wiki/CRUD
[jsonpath]:https://kubernetes.io/ru/docs/reference/kubectl/jsonpath/
