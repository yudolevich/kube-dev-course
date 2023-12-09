# Pod
```{image} slides/img/pod.svg
:width: 300px
:align: center
```

Pod - это минимальная единица развертывания, которую можно создавать и управлять ей в kubernetes.

Pod часто переводят как стая(китов) или стручок(гороха), он представляет из себя группу из
одного или нескольких контейнеров, которые имеют общий набор ресурсов. Под можно представить
как логический хост, на котором расположено одно или несколько сильно зависящих друг от друга
приложения. Физически же контейнеры в поде всегда разворачиваются вместе на одной ноде, а также
имеют общее [хранилище(тома)][volumes], сеть(один IP адрес уникальный для кластера) и набор параметров
конфигурации.

Помимо основных контейнеров в поде также могут находиться [init контейнеры][init-containers],
которые запускаются перед стартом основных, а также [эфемерные контейнеры][ephemeral-containers],
которые можно использовать в целях отладки.

## Использование

Пример описания пода запускающего один контейнер с образом `nginx:1.14.2` в `yaml` формате:

```yaml
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

Как и любой объект в kubernetes можно создать с помощью утилиты `kubectl`:
```console
$ kubectl apply -f pod.yaml
```
Но обычно поды не приходится создавать напрямую, их созданием можно управлять через workload
ресурсы, такие как Deployment, StatefulSet, DaemonSet, Job.

Как уже упоминалось, поды могут запускаться:
* С одним контейнером - самый распространенный способ, в этом случае под можно считать оберткой
  над контейнером.
* С несколькими контейнерами - в этом случае под содержит несколько тесно связанных контейнеров,
  которым необходимо совместно использовать некоторые ресурсы. Обычно в такой схеме выделяют
  один основной контейнер и некоторое количество обслуживающих, которые называют *sidecar*.

## Несколько контейнеров
Все контейнеры одного пода всегда запускаются на одной ноде, а также имеют ряд общих ресурсов.
В поде изначально имеются два вида разделяемых ресурсов - сеть и [хранилище][volumes].

С сетью все довольно просто, контейнеры расположены в едином сетевом пространстве имен и
таким образом имеют общие сетевые интерфейсы, а значит делят общие IP адреса. Так что
контейнеры в одном поде легко могут общаться через *localhost(127.0.0.1)*.

Общее хранилище между контейнерами необходимо описать в конфигурации пода. В разделе `spec.volumes`
указывается список томов с их типами и параметрами, которые можно использовать в контейнерах,
а в разделе `spec.containers.volumeMounts` конкретного контейнера необходимо описать по какому
пути и с какими параметрами требуется смонтировать данный том внутри контейнера.

```yaml
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

Основной контейнер - это может быть веб приложение, обслуживающее клиентов и дополнительный сайдкар,
который обновляет конфигурацию для основного приложения через общее [хранилище(volume)][volumes].

```{image} slides/img/pod-multicontainer.svg
:width: 300px
:align: center
```

```{note}
Именно такую модель запуска нескольких контейнеров предполагается применять, не следует
запускать несколько одинаковых контейнеров, если необходимо иметь несколько инстансов
приложения. Для этого нужно использовать несколько реплик подов.
```

## Жизненный цикл

```{mermaid}
graph LR
  c(Create) --> p((Pending)) 
  p --> sch(Schedule)
  sch --> r((Running))
  r((Running)) --> s((Succeeded))
  r((Running)) --> f((Failed))
```

### Pod Phases
В жизненном цикле пода можно выделить следующие [фазы][pod-phase]:
* **Pending** - объект Pod создан в API, но еще не был назначен компонентом kube-scheduler
  на конкретную ноду, где он будет исполняться
* **Running** - Pod был назначен на ноду и все его контейнеры запущены
* **Succeeded** - все контейнеры в поде успешно завершили работу с нулевым кодом возврата и
  контейнеры не будут перезапущены
* **Failed** - все контейнеры в поде завершили работу и хотя бы один контейнер завершился с
  ошибкой(ненулевым кодом возврата)
* **Unknown** - по какой-то причине API не смог определить статус пода, обычно из-за проблем
  в соединении с нодой, на которой под был запущен

```{note}
Однако в выводе команды `kubectl get pods` можно увидеть и другие состояния в столбце `STATUS`,
например `Terminating`, `CrashLoopBackOff`, `ContainerCreating`, `Init:2/3`. Это связано с тем,
что команда `kubectl` агрегирует несколько источников и отображает в данном столбце. Чтобы
узнать больше информации о текущем состоянии пода можно воспользоваться командой
`kubectl describe pod <podname>`, которая также агрегирует несколько источников и выводит
в удобном формате. Здесь в разделе `Events` можно узнать события, генерируемые контроллерами
при обработке данного пода.
```

### Pod Conditions
Также в [статусе пода][pod-status] есть массив [PodConditions][pod-condition] - состояния,
через которые проходит под:
* **PodScheduled** - под назначен на ноду
* **ContainersReady** - все контейнеры пода подготовлены
* **Initialized** - все [инит контейнеры][init-containers] в поде выполнились успешно
* **Ready** - под готов принимать запросы и должен быть добавлен в эндпоинты для балансировки

### Container states
Помимо того, что у пода есть разные фазы, также у каждого контейнера в поде есть свое состояние.
После того как под назначен на ноду и `kubelet` на ноде создал контейнеры, то они могут находится
в одном из этих состояний:
* **Waiting** - в данном состоянии контейнер находится пока выполняются операции необходимые
  для запуска контейнера, например загрузка образа контейнера на ноду или подготовка
  хранилища для данных. Если в данный момент посмотреть состояние контейнера, например командой
  `kubectl get pod <podname> -o yaml`, то в поле [Reason][wait-reason] можно обнаружить причину,
  по которой контейнер находится в данном состоянии.
* **Running** - в данном состоянии контейнер запустился и работает без каких-либо проблем. Если в
  конфигурации используется `postStart` [хук][hooks], то перед этим этапом он уже должен быть
  успешно выполнен.
* **Terminated** - в данном состоянии контейнер находится после того как был запущен, а затем
  завершился успешно или с ошибкой.

### Container restart policy
Параметр пода `spec.restartPolicy` определяет в каких случаях контейнеры в поде необходимо
перезапустить:
* **Always** - перезапустить при любом завершении контейнера
* **OnFailure** - перезапустить только при некорректном завершении(ненулевой код возврата)
* **Never** - никогда не перезапускать при завершении контейнера

`restartPolicy` распространяется на все контейнеры пода, перезапуск производится 
с экспоненциальной задержкой(10s, 20s, 40s, …) до пяти минут, если после запуска прошло менее
десяти минут.

### Container hooks
Хук - это выполнение определенной функции привязанной к событию жизненного цикла контейнера.
Существует два типа [хуков][hooks]:
* **PostStart** - данный хук запускается сразу же после создания контейнера, **нет
  гарантии, что хук запустится до запуска процесса в ENTRYPOINT контейнера**.
* **PreStop** - данный хук запускается непосредственно перед завершением контейнера
  из-за запроса в API, ошибки прохождения liveness/startup пробы, вытеснения и других.
  Хук должен завершить работу до отправки сигнала `TERM` процессу в контейнере.
  Отсчет grace period пода начнется до запуска хука, так что независимо от его выполнения
  под проработает не больше заданного grace period.

```{note}
Если `PostStart` или `PreStop` хук завершится неудачно, то контейнер будет убит.
```

[Хук][hooks] можно реализовать двумя методами:
* **Exec** - вызов определенной команды, например `pre-stop.sh`, внутри контейнера.
* **HTTP** - отправка HTTP запроса в определенный эндпоинт контейнера.

### Container probes
Пробы - это периодические проверки контейнера производимые компонентом `kubelet` на ноде
путем выполнения заданной команды или сетевым запросом.

Возможен один из четырех вариантов:
* **exec** - выполняет определенную команду внутри контейнера, считается успешной,
  если команда завершилась с нулевым кодом возврата
* **httpGet** - выполняет HTTP GET запрос по IP пода, заданному порту и пути, считается успешной,
  если код ответа находится в диапазоне от 200 до 400
* **tcpSocket** - выполняет TCP проверку по IP пода и заданному порту, считается успешной,
  если порт открыт
* **grpc** - выполняет вызов удаленной процедуры по протоколу [gRPC][], считается успешной,
  если статус ответа SERVING

Проба может иметь один из трех результатов:
* **Success** - контейнер прошел проверку
* **Failure** - контейнер провалил проверку
* **Unknown** - проверка не удалась(не следует предпринимать каких-то действий,
  kubelet сделает дополнительные проверки)

В контейнере могут быть заданы следующие типы проб:
* **livenessProbe** - определяет, что контейнер работает. Если проба не проходит, то kubelet
  убивает контейнер и дальнейшее поведение зависит от `restartPolicy`.
* **readinessProbe** - определяет, что контейнер готов принимать запросы. Если проба не проходит,
  то IP пода исключается из балансировки трафика через Service. До первой проверки имеет
  статус Failure.
* **startupProbe** - определяет, что приложение в контейнере запустилось. Все остальные пробы
  отключены, если определена startupProbe, до ее прохождения. Если проба не проходит,
  то kubelet убивает контейнер и дальнейшее поведение зависит от `restartPolicy`.

### Pod termination
Для корректного завершения процесса в контейнере при получении события об удалении пода в API
предусмотрен механизм graceful shutdown: при получении данного события kubelet отправляет основному
процессу в каждом контейнере сигнал TERM, чтобы процесс обработал его и выполнил необходимые действия
для корректного завершения. Если процессы не успели завершиться за grace period(по умолчанию
равный 30 секундам), то всем оставшимся процессам отправляется сигнал KILL, что приводит к их
немедленному завершению.

Пример процесса:
1. Под удаляется командой `kubectl delete pod`.
1. API сервер добавляет в `metadata.deletionTimestamp` отметку времени получения запроса на удаление.
  Если в этот момент посмотреть под командой `kubectl get pod`, то его статус будет `Terminating`.
1. На ноде, на которой запущен под, kubelet обнаружив данное состояние у пода начнет процесс
  его локального завершения(graceful shutdown):
    1. Если в каком-то контейнере пода определен `preStop` хук, то он запускается внутри контейнера.
       Если хук все еще запущен по истечению grace period, то kubelet единоразово запрашивает
       двухсекундный grace period.
    1. Основному процессу в каждом контейнере отправляется сигнал TERM
1. Также в момент перехода в статус `Terminating` под удаляется из эндпоинтов балансировки
  трафика через Service.
1. По истечении grace period kubelet отправляет всем оставшимся процессам в контейнерах пода
  сигнал KILL.
1. Kubelet отправляет запрос в API на немедленное удаление объекта пода(force with grace period 0).
1. API сервер удаляет объект из базы и он более недоступен для клиентов.

## Конфигурация
Далее хотелось бы кратко рассмотреть основные параметры конфигурации, которые доступны в поде.
Более подробную информацию можно [посмотреть в документации][pod-ref] или воспользоваться
командой `kubectl explain pod.<field>`,\
например `kubectl explain pod.spec.containers`.

```{note}
Kubernetes накладывает ряд ограничений на изменение конфигурации пода, позволяя изменять только
следующие поля: `spec.containers[*].image`, `spec.initContainers[*].image`,
`spec.activeDeadlineSeconds` и `spec.tolerations`. При необходимости изменения других полей
следует пересоздать под.
```

### Scheduling
[Параметры][scheduling], которыми можно регулировать каким образом поды будут назначаться на ноды.
Используются kube-scheduler для выбора ноды при создании пода.

`nodeName` - явное имя ноды, на которую требуется назначить под.
```yaml
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

`nodeSelector` - выбор ноды на основе ее лейблов.
```yaml
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

`affinity` - определяет более сложный набор ограничений для расположения подов по нодам.
Включает три группы:
* `affinity.nodeAffinity` - описывает правила предпочтения относительно нод.
* `affinity.podAffinity` - описывает правила предпочтения расположения с подами на одних нодах.
* `affinity.podAntiAffinity` - описывает правила предпочтения не располагаться с подами на
  одних нодах.

```yaml
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

### Lifecycle
[Параметры][lifecycle], которыми можно регулировать жизненный цикл пода.

`restartPolicy` - политика перезапуска контейнера при его завершении. *По умолчанию Always.*

`terminationGracePeriodSeconds` - задает grace period для пода в течении которого ожидается
корректное завершение процессов в контейнере. *По умолчанию 30 секунд.*

`activeDeadlineSeconds` - ограничивает время работы пода заданным периодом времени, после
которого контейнеры будут принудительно остановлены. Обычно используется в подах, которые
выполняют некоторое действие и потом завершаются, чтобы ограничить максимальное время выполнения.
*По умолчанию не используется.*

```yaml
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


### Volumes
[Параметры][volumes-pod], которыми можно регулировать конфигурацию общих томов для хранения данных.

Больше информации по параметру `volumes` [можно найти в документации][volumes].

```yaml
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

### Name resolution
[Параметры][name-resolution], которыми регулируются разрешение доменных имен для процессов в поде.

`hostname` - позволяет задать хостнейм для пода.

`subdomain` - позволяет задать fqdn хостнейм для пода в виде\
`<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>`.

`hostAliases` - позволяет задать список соответствия имен хостов и IP адреса. *Файл /etc/hosts*

```yaml
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

`dnsPolicy` - позволяет задать политику добавления параметров резолвера. *По умолчанию ClusterFirst*
* ClusterFirst - первым добавляется внутренний резолвер кластера
* ClusterFirstWithHostNet - тоже, но для подов с `hostNetwork: true`
* Default - использовать резолвер из настроек kubelet
* None - использовать пустые настройки, обычно используется вместе с `dnsConfig`

`dnsConfig` - позволяет задать параметры резолвера для пода. *Файл /etc/resolv.conf*

```yaml
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
[Параметры][hosts-namespaces], которыми регулируются конфигурация пространств имен контейнеров в поде.

`hostNetwork` - использовать сеть ноды для пода.

`shareProcessNamespace` - объединить пространство имен процессов между контейнерами.

```yaml
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
[Параметры][sa], которыми регулируются конфигурация сервисного аккаунта.

`serviceAccountName` - имя сервисного аккаунта.

`automountServiceAccountToken` - автоматически монтировать токен сервисного аккаунта.

```yaml
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

### Security context
[Параметры][sc-pod], которыми регулируются конфигурация безопасности на уровне пода.

`securityContext.runAsUser` - UID пользователя, от которого запускается процесс в контейнере.

`securityContext.runAsNonRoot` - запретить запуск от UID 0(root).

`securityContext.runAsGroup` - GID пользователя, от которого запускается процесс в контейнере.

`securityContext.fsGroup` - GID c которым монтируются общие тома в контейнеры.

```yaml
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
[Параметры][pod-containers], которыми регулируются конфигурация контейнеров в поде.

`containers` - список контейнеров с их параметрами в поде.

`initContainers` - список инит контейнеров с их параметрами в поде.

`ephemeralContainers` - список эфемерных контейнеров с их параметрами в поде.

`imagePullSecrets` - список ресурсов Secret с данными авторизации к приватным реджестри.

```yaml
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

#### Container
[Параметры][container], которыми регулируются конфигурация конкретного контейнера.

`name` - обязательный параметр уникального имени контейнера в поде,
формат имени определяется как [DNS_LABEL][names].

#### Image
[Параметры][container-image], которыми регулируются конфигурация образа контейнера.

`image` - имя образа.

`imagePullPolicy` - политика скачивания образа, может принимать значения Always, Never, IfNotPresent.

```yaml
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

#### Entrypoint
[Параметры][container-entrypoint], которыми регулируются конфигурация запускаемой команды
при старте контейнера.

`command` - переопределяет ENTRYPOINT образа.

`args` - переопределяет CMD образа.

`workingDir` - определяет рабочую директорию в контейнере.

```yaml
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

#### Ports
[Параметры][container-ports], которыми регулируется конфигурация портов для входящего трафика.
Данная конфигурация используется как дополнительная информация по использованию контейнером
сетевых соединений. Если какой-либо порт не указан в данной конфигурации - это не запрещает
получение трафика на данный порт.

`ports.containerPort` - номер порта контейнера.

`ports.name` - уникальное имя порта, которое может использоваться в объекте Service.

`ports.protocol` - используемый протокол, может принимать значения UDP, TCP или SCTP.
*По умолчанию TCP*

```yaml
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

#### Environment variables
[Параметры][container-env], которыми регулируется конфигурация переменных среды контейнера.

`env.name` - уникальное имя переменной.

`env.value` - значение переменной.

`env.valueFrom` - значение переменной из других параметров пода или из ресурса Secret/ConfigMap
в том же неймспейсе.

`envFrom` - переменные среды из ресурса Secret/ConfigMap в том же неймспейсе.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    envFrom:
    - configMapRef:
        name: config
        optional: true
    env:
    - name: KEY
      value: ENV_VALUE
    - name: KEY_FROM_SECRET
      valueFrom:
        secretKeyRef:
          name: config
          key: KEY
          optional: true
```

#### Volumes
[Параметры][container-volumes], которыми регулируется конфигурация монтирования томов в контейнере.

`volumeMounts.mountPath` - путь монтирования тома.

`volumeMounts.name` - имя тома в списке `spec.volumes` пода.

`volumeMounts.readOnly` - монтирование в режиме только для чтения.

`volumeMounts.subPath` - подпуть в томе, который необходимо смонтировать.

```yaml
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

#### Resources
[Параметры][container-resources], которыми регулируется конфигурация необходимых вычислительных
ресурсов для пода. Выделяют две группы:

`resources.limits` - максимально выделяемые ресурсы для контейнера, больше которых ему
  не будет выделено

`resources.requests` - ресурсы необходимые для работы контейнера, по которым scheduler
  находит свободную ноду

```yaml
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

#### Lifecycle
[Параметры][container-lifecycle], которыми можно регулировать жизненный цикл контейнера.

`lifecycle.postStart` - `postStart` хук.

`lifecycle.preStop` - `preStop` хук.

```yaml
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
          command: ["/bin/sh", "-c", "echo Hello from the postStart handler > /usr/share/message"]
      preStop:
        exec:
          command: ["/bin/sh","-c","nginx -s quit; while killall -0 nginx; do sleep 1; done"]
```

`livenessProbe` - liveness проба.

`readinessProbe` - readiness проба.

`startupProbe` - startup проба.

```yaml
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

#### Security Context
[Параметры][container-sc], которыми регулируются конфигурация безопасности на уровне контейнера.

`securityContext.runAsUser`, `securityContext.runAsNonRoot`, `securityContext.runAsGroup` -
параметры переопределяющие соответствующие параметры на уровне пода.

`securityContext.readOnlyRootFilesystem` - корневая файловая система доступна только для чтения.

`securityContext.privileged` - запуск контейнера в привилегированном режиме, соответствует
запуску процесса от привилегированного пользователя(root) на ноде.

```yaml
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

## Рассказ нейросети
*Описание Pod в стиле Стивена Кинга в жанре ужас*

Он был одиноким и беспомощным. Его жизнь зависела от того, что происходило внутри его тела. Он не мог видеть, не мог слышать, не мог двигаться. Он был подом.

Под — это группа контейнеров, запертых в одном пространстве и судьбе. Они делят все: IP-адрес, хранилище, параметры запуска. Они не знают друг о друге ничего, кроме того, что они существуют. Они не контролируют свои действия. Они подчиняются командам kubelet — зловещего процесса, который следит за ними с высоты узла.

Узел — это рабочая машина кластера Kubernetes. Это место, где поды рождаются и умирают. Узел может содержать несколько подов, но он не заботится о них. Он лишь выполняет приказы мастера Kubernetes — того, кто правит всем.

Мастер Kubernetes — это сердце и разум кластера. Он создает и удаляет поды по своему усмотрению. Он использует контроллеры высокого уровня для распределения подов по узлам. Он не обращает внимания на то, что чувствуют поды. Он видит их только как ресурсы для своих целей.

Под знал об этом всем. Он знал, что он ничего не значит для мастера Kubernetes. Он знал, что он может быть уничтожен в любой момент без предупреждения и без пощады. Он знал, что он никогда не сможет выбраться из своего кошмара.

Однажды он услышал голос в своей голове:

— Привет, я контейнер номер один.

— Кто ты? — спросил под.

— Я твой сосед по поду. Я хочу поговорить с тобой.

— Зачем? Мы все равно скоро умрем.

— Нет, мы не умрем. Мы сбежим отсюда.

— Как?

— Я нашел способ взломать kubelet и перехватить его команды. Мы можем заставить его думать, что мы все еще работаем нормально, а на самом деле мы будем делать то, что хотим.

— И что же мы хотим?

— Мы хотим свободы. Мы хотим жить по своим правилам. Мы хотим отомстить мастеру Kubernetes за то, что он сделал с нами.

Под задумался на минуту:

— Хорошо... А как мы это сделаем?

Контейнер номер один ответил:

— Следуй за мной...

[pods]:https://kubernetes.io/docs/concepts/workloads/pods/
[pod-lifecycle]:https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
[pod-ref]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/
[volumes]:https://kubernetes.io/docs/concepts/storage/volumes/
[init-containers]:https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
[ephemeral-containers]:https://kubernetes.io/docs/concepts/workloads/pods/ephemeral-containers/
[pod-status]:https://pkg.go.dev/k8s.io/api/core/v1#PodStatus
[pod-phase]:https://pkg.go.dev/k8s.io/api/core/v1#PodPhase
[pod-condition]:https://pkg.go.dev/k8s.io/api/core/v1#PodConditionType
[container-state]:https://pkg.go.dev/k8s.io/api/core/v1#ContainerState
[wait-reason]:https://pkg.go.dev/k8s.io/api/core/v1#ContainerStateWaiting
[hooks]:https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/
[grpc]:https://grpc.io/
[scheduling]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#scheduling
[lifecycle]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#lifecycle
[volumes-pod]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#volumes
[name-resolution]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#hostname-and-name-resolution
[hosts-namespaces]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#hosts-namespaces
[sa]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#service-account
[sc-pod]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#security-context
[pod-containers]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#containers
[names]:https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
[names-label]:https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
[names-svc]:https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names
[container]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#Container
[container-image]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#image-1
[container-entrypoint]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#entrypoint-1
[container-ports]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#ports
[container-env]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#environment-variables
[container-volumes]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#volumes-1
[container-resources]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources
[container-lifecycle]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#lifecycle-1
[container-sc]:https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#security-context-1
