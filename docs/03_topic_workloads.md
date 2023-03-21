# Workloads

```{mermaid}
graph TB
    u((User)) --> lb{{LoadBalancer}}
    subgraph n1 [node]
      a1(app)
    end
    subgraph n2 [node]
      a2(app)
    end
    subgraph n3 [node]
      a3(app)
    end
    lb --> a1 & a2 & a3
```

Kubernetes позволяет запускать рабочую нагрузку в виде подов, однако, когда речь заходит о промышленной
эксплуатации, запуска в единственном экземпляре оказывается недостаточно, чтобы обеспечить высокую
доступность приложения при обработке большого количества запросов клиентов. Для удобного управления
набором подов в зависимости от типа рабочей нагрузки в kubernetes существует несколько типов ресурсов.

* **Deployment** и **ReplicaSet** - для рабочих нагрузок, которые не хранят состояние(stateless), где
  поды взаимозаменяемы и могут создаваться/удаляться при необходимости.
* **StatefulSet** - для нагрузок, которые требуют сохранение некоторого состояния.
* **DaemonSet** - для приложений, которые необходимо выполнять на каждой ноде.
* **Job** и **CronJob** - для задач, которые выполняются единоразово или периодически по расписанию.

```{note}
В общем случае не предполагается управление напрямую подами вручную, используйте преведенные типы
ресурсов для управления рабочими нагрузками в kubernetes.
```

## Deployment/ReplicaSet

```{mermaid}
graph TB
    d([Deployment]) --> rs1([ReplicaSet])
    d([Deployment]) --> rs2([ReplicaSet])
    rs1 --> p11((Pod))
    rs1 --> p12((Pod))
    rs2 --> p21((Pod))
    rs2 --> p22((Pod))
```

Если экземпляр приложения не требует хранения состояния внутри себя и не важно какой именно экземпляр
будет обрабатывать пришедший запрос, то для управления этими экземплярами отлично подойдет ресурс
**Deployment**. Он позволяет запускать нескольких реплик приложения, а также управлять процессом
обновления.

Для управления количеством копий пода используется дополнительный ресурс **ReplicaSet**:
```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: nginx-rs
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
```
Он описывает необходимое состояние в виде шаблона пода - `spec.template`, селектора подов по лейблам -
`spec.selector`, за которыми нужно следить, а также количество реплик пода - `spec.replicas`.
Таким образом ReplicaSet контроллер отслеживает по селектору текущее состояние подов и пытается привести
их количество к ожидаемому через создание/удаление дополнительных экземпляров.

Обычно использовать **ReplicaSet** напрямую не приходится, для этого стоит использовать **Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
  revisionHistoryLimit: 5
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
```
Он добавляет функционал управления процессом обновления и отката, через создание новых ReplicaSet и
управление количеством реплик в них. В конфигурации появляются такие поля как `spec.strategy` - стратегия
обновления, `spec.revisionHistoryLimit` - количество старых ReplicaSet для возможности отката на них,
`spec.paused` - для возможности выставления паузы в процессе обновления и некоторые другие для
управления процессом обновления.

### Селектор
ReplicaSet определяет через поле `spec.selector` лейблы, по которым находятся соответствующие поды
управляемые ReplicaSet контроллером. Таким образом лейблы в этом поле должны  совпадать с лейблами в
`spec.template.metadata.labels`. Когда ReplicaSet управляется через Deployment, то в селектор
также добавляется лейбл `pod-template-hash`, который содержит хэш от шаблона пода, таким образом
каждый объект ReplicaSet будет иметь уникальное значение лейбла `pod-template-hash`.
```{note}
Поле `spec.selector` неизменяемо после создания.
```

### Обновление
При изменении шаблона пода в `spec.template` в Deployment срабатывает триггер на обновление. Для
Deployment есть два типа стратегии обновления - **Recreate** и **RollingUpdate**.

#### Recreate
Стратегия Recreate позволяет быть уверенным, что старая версия приложения полностью завершила работу
перед запуском новой версии. Процесс состоит из следующих шагов:
* Контроллер выставляет количество реплик в текущем ReplicaSet в 0
* Ожидает пока поды данного ReplicaSet завершат свою работу и удалятся
* Создает новый ReplicaSet с новым шаблоном пода и необходимым количеством реплик

Можно проследить как контроллер управляет ресурсами с момента изменения `spec.template`, например,
с помощью команды `kubectl set` изменив значение переменной среды.
```console
$ k get deploy
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           47h
$ k get rs
NAME                         DESIRED   CURRENT   READY   AGE
nginx-deployment-c5f8dc5d6   3         3         3       19s
$ k set env deploy/nginx-deployment ENV="$(date)"
deployment.apps/nginx-deployment env updated
$ k get rs
NAME                         DESIRED   CURRENT   READY   AGE
nginx-deployment-c5f8dc5d6   0         0         0       36s
$ k get po
NAME                               READY   STATUS        RESTARTS   AGE
nginx-deployment-c5f8dc5d6-bqmcl   1/1     Terminating   0          38s
nginx-deployment-c5f8dc5d6-v7btw   1/1     Terminating   0          38s
nginx-deployment-c5f8dc5d6-wkx9w   1/1     Terminating   0          38s
$ k get rs
NAME                          DESIRED   CURRENT   READY   AGE
nginx-deployment-6c86565b44   3         3         0       4s
nginx-deployment-c5f8dc5d6    0         0         0       45s
$ k get po
NAME                                READY   STATUS              RESTARTS   AGE
nginx-deployment-6c86565b44-4cx9r   0/1     ContainerCreating   0          8s
nginx-deployment-6c86565b44-dlvcz   0/1     ContainerCreating   0          8s
nginx-deployment-6c86565b44-dnnht   0/1     ContainerCreating   0          8s
$ k get po
NAME                                READY   STATUS    RESTARTS   AGE
nginx-deployment-6c86565b44-4cx9r   1/1     Running   0          16s
nginx-deployment-6c86565b44-dlvcz   1/1     Running   0          16s
nginx-deployment-6c86565b44-dnnht   1/1     Running   0          16s
```

#### RollingUpdate
Стратегия RollingUpdate позволяет плавно переключить нагрузку со старой версии приложения на новую.
Процесс в общем случае состоит из шагов:
* Контроллер создает новый ReplicaSet с новым шаблоном пода и количеством реплик 0
* Поочередно уменьшает количество реплик в старом ReplicaSet и увеличивает в новом

Он также включает дополнительные параметры `spec.strategy.rollingUpdate.maxSurge` и
`spec.strategy.rollingUpdate.maxUnavailable`:
* `maxUnavailable` - максимальное количество подов, которые могут быть недоступны в процессе обновления
* `maxSurge` - максимальное количество подов, которое может быть создано сверх желаемого количества
  `spec.replicas` в процессе обновления
Данные параметры могу задаваться в как абсолютных значениях так и в процентах. По умолчанию равны 25%.

С помощью опции `-w` или `--watch` утилиты `kubectl` можно отследить как происходит переключение реплик
между двумя ReplicaSet, которыми управляет Deployment.
```console
$ k set env deploy/nginx-deployment ENV="$(date)"
deployment.apps/nginx-deployment env updated
$ k get rs -w
NAME                          DESIRED   CURRENT   READY   AGE
nginx-deployment-9ccdcc857    3         3         3       28s
nginx-deployment-7c7dcd74b5   1         0         0       0s
nginx-deployment-7c7dcd74b5   1         0         0       0s
nginx-deployment-7c7dcd74b5   1         1         0       0s
nginx-deployment-7c7dcd74b5   1         1         1       5s
nginx-deployment-9ccdcc857    2         3         3       40s
nginx-deployment-9ccdcc857    2         3         3       41s
nginx-deployment-7c7dcd74b5   2         1         1       6s
nginx-deployment-9ccdcc857    2         2         2       41s
nginx-deployment-7c7dcd74b5   2         1         1       6s
nginx-deployment-7c7dcd74b5   2         2         1       7s
nginx-deployment-7c7dcd74b5   2         2         2       9s
nginx-deployment-9ccdcc857    1         2         2       44s
nginx-deployment-7c7dcd74b5   3         2         2       9s
nginx-deployment-9ccdcc857    1         2         2       44s
nginx-deployment-9ccdcc857    1         1         1       44s
nginx-deployment-7c7dcd74b5   3         2         2       9s
nginx-deployment-7c7dcd74b5   3         3         2       9s
nginx-deployment-7c7dcd74b5   3         3         3       15s
nginx-deployment-9ccdcc857    0         1         1       50s
nginx-deployment-9ccdcc857    0         1         1       50s
nginx-deployment-9ccdcc857    0         0         0       50s
```

### Состояние

Как и с другими ресурсами о Deployment подробную информацию можно получить командой `kubectl describe`:

```console
$ k describe deploy nginx-deployment
Name:                   nginx-deployment
Namespace:              default
CreationTimestamp:      Sat, 18 Mar 2023 23:42:34 +0300
Labels:                 app=nginx
Annotations:            deployment.kubernetes.io/revision: 3
Selector:               app=nginx
Replicas:               3 desired | 3 updated | 3 total | 3 available | 0 unavailable
StrategyType:           RollingUpdate
MinReadySeconds:        0
RollingUpdateStrategy:  25% max unavailable, 25% max surge
Pod Template:
  Labels:  app=nginx
  Containers:
   nginx:
    Image:        nginx
    Port:         <none>
    Host Port:    <none>
    Environment:  <none>
    Mounts:       <none>
  Volumes:        <none>
Conditions:
  Type           Status  Reason
  ----           ------  ------
  Available      True    MinimumReplicasAvailable
  Progressing    True    NewReplicaSetAvailable
OldReplicaSets:  <none>
NewReplicaSet:   nginx-deployment-76d6c9b8c (3/3 replicas created)
Events:
  Type    Reason             Age                    From                   Message
  ----    ------             ----                   ----                   -------
  Normal  ScalingReplicaSet  5m26s                  deployment-controller  Scaled up replica set nginx-deployment-76d6c9b8c to 3
  Normal  ScalingReplicaSet  4m44s                  deployment-controller  Scaled up replica set nginx-deployment-bc8fc6c46 to 1
  Normal  ScalingReplicaSet  4m35s                  deployment-controller  Scaled down replica set nginx-deployment-76d6c9b8c to 2 from 3
  Normal  ScalingReplicaSet  4m35s                  deployment-controller  Scaled up replica set nginx-deployment-bc8fc6c46 to 2 from 1
  Normal  ScalingReplicaSet  4m24s                  deployment-controller  Scaled down replica set nginx-deployment-76d6c9b8c to 1 from 2
  Normal  ScalingReplicaSet  4m24s                  deployment-controller  Scaled up replica set nginx-deployment-bc8fc6c46 to 3 from 2
  Normal  ScalingReplicaSet  4m17s                  deployment-controller  Scaled down replica set nginx-deployment-76d6c9b8c to 0 from 1
  Normal  ScalingReplicaSet  3m                     deployment-controller  Scaled up replica set nginx-deployment-76d6c9b8c to 1 from 0
  Normal  ScalingReplicaSet  2m57s                  deployment-controller  Scaled down replica set nginx-deployment-bc8fc6c46 to 2 from 3
  Normal  ScalingReplicaSet  2m39s (x4 over 2m57s)  deployment-controller  (combined from similar events): Scaled down replica set nginx-deployment-bc8fc6c46 to 0 from 1
```

Текущее состояние реплик отражено в поле Replicas, в процессе обновления здесь будут отражены:
* **desired** - ожидаемое количество реплик
* **updated** - количество реплик с новой версией
* **total** - общее количество существующих реплик(старой и новой версии)
* **available** - количество реплик в состоянии Ready
* **unavailable** - количество реплик в состоянии NotReady
```
Replicas: 3 desired | 2 updated | 4 total | 3 available | 1 unavailable
```

В момент обновления поля `OldReplicaSets` и `NewReplicaSet` будут содержать информацию о старом и
новом ReplicaSet, а также текущее количество реплик в них.
```
OldReplicaSets:  nginx-deployment-76d6c9b8c (2/2 replicas created)
NewReplicaSet:   nginx-deployment-bc8fc6c46 (2/2 replicas created)
```

Поле [Conditions][] может отражать три типа состояний, в которых находится Deployment:
* **Available** - указывает на то, что количество реплик в состоянии Ready не меньше указанного в
  `spec.replicas`
* **Progressing** - указывает на то, что процесс обновления производится или успешно завершен
* **ReplicaFailure** - указывает на то, что в процессе обновления не удается создать новые реплики

```{note}
С другими параметрами ресурса [Deployment можно также ознакомиться в документации][deploy] или командой
`kubectl explain deploy`.
```

## StatefulSet

Использование ресурса StatefulSet - еще один из способов управление множественными репликами приложения.
Он подходит, если каждой реплике приложения необходимо внутри хранить какое-то состояние.
Характеризуется следующими особенностями:
* Стабильные и уникальные сетевые идентификаторы
* Стабильное постоянное хранилище
* Упорядоченный процесс развертывания и увеличения реплик
* Упорядоченный процесс обновления

Пример StatefulSet с Service типа Headless:
```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 80
    name: web
  clusterIP: None
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  selector:
    matchLabels:
      app: nginx
  serviceName: "nginx"
  podManagementPolicy: OrderedReady
  replicas: 3
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 100Mi
```

Здесь также как и в Deployment есть:
* `selector` - выбирает набор подов по их лейблам
* `replicas` - указывает количество необходимых реплик приложения
* `template` - описывает шаблон пода

Также есть и особые параметры:
* `volumeClaimTemplates` - позволяет описать шаблон для создания PersistentVolumeClaim, что позволяет
  создать для каждой реплики свое собственное хранилище и смонтировать внутрь пода (в кластере должен
  существовать StorageClass с именем `storageClassName`)
* `serviceName` - указывает имя ресурса Service для построения доменного имени пода вида
  `podName.serviceName.namespace.svc.cluster.local` или внутри неймспейса просто `podName.serviceName`
* `podManagementPolicy` - контролирует каким образом будет производиться создание новых и удаление
  старых реплик при создании StatefulSet, увеличении и уменьшении количества реплик
* `updateStrategy` - контролирует процесс обновления при изменении шаблона пода

### Идентификатор пода
Поды в StatefulSet имеют уникальные идентификаторы, которые создаются последовательно и позволяют
иметь постоянные сетевые имена. Имена состоят из имени StatefulSet и суффикса `-<n>`,
где n от 0 до replicas-1.

```console
$ k get pods
NAME    READY   STATUS    RESTARTS   AGE
web-0   1/1     Running   0          2m21s
web-1   1/1     Running   0          109s
web-2   1/1     Running   0          82s
```

Также каждый под имеет стабильное сетевое имя вида `$(pod).$(service).$(namespace).svc.cluster.local`:
```console
$ k exec web-0 -- /bin/sh -c 'echo $HOSTNAME > /usr/share/nginx/html/index.html'
$ k exec web-1 -- /bin/sh -c 'echo $HOSTNAME > /usr/share/nginx/html/index.html'
$ k exec web-2 -- /bin/sh -c 'echo $HOSTNAME > /usr/share/nginx/html/index.html'
$ k exec -it web-0 -- /bin/bash
root@web-0:/# curl web-0.nginx.default.svc.cluster.local
web-0
root@web-0:/# curl web-0.nginx
web-0
root@web-0:/# curl web-1.nginx
web-1
root@web-0:/# curl web-2.nginx
web-2
```

Имена регистрируются в кластерном DNS:
```console
root@web-0:/# getent hosts web-0.nginx
10.244.1.65     web-0.nginx.default.svc.cluster.local
root@web-0:/# getent hosts web-1.nginx
10.244.2.77     web-1.nginx.default.svc.cluster.local
root@web-0:/# getent hosts web-2.nginx
10.244.1.67     web-2.nginx.default.svc.cluster.local
root@web-0:/# getent hosts nginx # headless service
10.244.1.67     nginx.default.svc.cluster.local
10.244.1.65     nginx.default.svc.cluster.local
10.244.2.77     nginx.default.svc.cluster.local
```

### Pod Management Policies

#### OrderedReady
Данная политика используется по умолчанию и гарантирует:
* Для StatefulSet с N репликами при создании или увеличении реплик поды создаются последовательно
  от 0 до N-1
* При уменьшении количества реплик поды терминируются в обратном порядке от N-1 до 0
* Перед созданием нового пода все предыдущие должны находиться в статусе Running и Ready
* Перед удалением пода все предыдущие должны быть завершены и удалены

#### Parallel
При изменении количества реплик не происходит процесса ожидания пока другие поды окажутся в состоянии
Running и Ready или успешно терминирутся, операции с подами производятся параллельно.

### Update Strategies
* OnDelete - отключает автоматическое обновление подов в StatefulSet при изменении `spec.template`,
  для обновления требуется вручную удалить под
* RollingUpdate - стратегия с последовательным обновлением подов, используется по умолчанию

```{note}
С другими параметрами ресурса [StatefulSet можно также ознакомиться в документации][sts] или командой
`kubectl explain sts`.
```

## DaemonSet
Если требуется запустить приложение на всех нодах(или на некоторых), то для этого можно воспользоваться
ресурсом DaemonSet. При добавлении или удалении нод в кластере поды также будут добавляться и удаляться.
Типичные примеры использования:
* Запуск контроллера, управляющего хранилищем на ноде
* Запуск коллектора логов на ноде
* Запуск агента мониторинга на ноде

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app.kubernetes.io/component: exporter
    app.kubernetes.io/name: node-exporter
  name: node-exporter
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: exporter
      app.kubernetes.io/name: node-exporter
  template:
    metadata:
      labels:
        app.kubernetes.io/component: exporter
        app.kubernetes.io/name: node-exporter
    spec:
      containers:
      - args:
        - --path.sysfs=/host/sys
        - --path.rootfs=/host/root
        name: node-exporter
        image: prom/node-exporter
        ports:
          - containerPort: 9100
            protocol: TCP
        volumeMounts:
        - mountPath: /host/sys
          mountPropagation: HostToContainer
          name: sys
          readOnly: true
        - mountPath: /host/root
          mountPropagation: HostToContainer
          name: root
          readOnly: true
      volumes:
      - hostPath:
          path: /sys
        name: sys
      - hostPath:
          path: /
        name: root
```
В данном примере на всех нодах запускается экспортер метрик для возможности мониторинга нод кластера.
Конфигурация `selector`, `template` и `updateStrategy` будет аналогична ресурсу StatefulSet.
Для того, чтобы управлять набором нод, на которых требуется запустить поды, можно изменить конфигурацию
шаблона пода - `spec.template.spec.nodeSelector`, `spec.template.spec.affinity.nodeAffinity` и
`spec.template.spec.tolerations`.

```{note}
С другими параметрами ресурса [DaemonSet можно также ознакомиться в документации][ds] или командой
`kubectl explain ds`.
```

## Job/Cronjob
Если приложение не должно работать постоянно, а после запуска и выполнения своего функционала должно
завершиться, то для такой рабочей нагрузки подойдет ресурс Job. Job позволяет запустить один или
несколько подов с задачей, ожидая успешного завершения и повторяя запуск заданное количество раз
при неудачах.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: hello
spec:
  backoffLimit: 3
  activeDeadlineSeconds: 30
  template:
    spec:
      containers:
      - name: hello
        image: busybox
        command: ["/bin/sh", "-c", "echo hello"]
      restartPolicy: Never
```
Параметр `template` как и в других ресурсах задает шаблон запускаемого пода. Через параметр
`backoffLimit` можно указать количество попыток запуска пода, если запуск завершился неудачей. А через
параметр `activeDeadlineSeconds` можно указать максимальное время выполнения, после которого запуск
будет считаться неудачным.

```{note}
С другими параметрами ресурса [Job можно также ознакомиться в документации][job] или командой
`kubectl explain job`.
```

Ресурс CronJob позволяет запускать ресурсы Job регулярно по расписанию.

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: hello
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      backoffLimit: 3
      activeDeadlineSeconds: 30
      template:
        spec:
          containers:
          - name: hello
            image: busybox
            command: ["/bin/sh", "-c", "echo hello"]
          restartPolicy: Never
```

Здесь в параметре `jobTemplate` указывается шаблон создаваемого ресурса Job, а в параметре `schedule`
указывается расписание запуска в формате [cron][].

```{note}
С другими параметрами ресурса [CronJob можно также ознкомиться в документации][cronjob] или командой
`kubectl explain cronjob`.
```

## Rollout

Для управления процессом развертывания(обновления) существует команда `kubectl rollout`, с помощью
которой можно посмотреть статус развертывания, поставить на паузу, посмотреть историю развертываний и
сделать откат к предыдущей версии. Команда может быть применена к ресурсам Deployment, StatefulSet и
DaemonSet.

### Status

Для просмотра статуса есть команда `kubectl rollout status`, которая покажет статус последнего
развертывания. Если развертывание активно в текущий момент, то команда будет отслеживать его статус
до завершения.

```console
$ k set env deploy/nginx-deployment ENV="$(date)"
deployment.apps/nginx-deployment env updated
$ k rollout status deployment nginx-deployment
Waiting for deployment "nginx-deployment" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "nginx-deployment" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "nginx-deployment" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "nginx-deployment" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "nginx-deployment" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "nginx-deployment" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "nginx-deployment" rollout to finish: 1 old replicas are pending termination...
Waiting for deployment "nginx-deployment" rollout to finish: 1 old replicas are pending termination...
deployment "nginx-deployment" successfully rolled out
```

### History

Историю развертываний можно посмотреть командой `kubectl rollout history`:
```console
$ k rollout history deploy/nginx-deployment
deployment.apps/nginx-deployment
REVISION  CHANGE-CAUSE
1         <none>
2         <none>
```

REVISION и CHANGE-CAUSE берутся из аннотаций `deployment.kubernetes.io/revision` и
`kubernetes.io/change-cause` соответствующей ReplicaSet. Добавить аннотацию на последнюю ревизию
также можно путем добавления на Deployment:

```console
$ k set env deploy nginx-deployment ENV="$(date)"
deployment.apps/nginx-deployment env updated
$ k annotate deploy/nginx-deployment kubernetes.io/change-cause="new date 1"
deployment.apps/nginx-deployment annotated
$ k rollout history deploy/nginx-deployment
deployment.apps/nginx-deployment
REVISION  CHANGE-CAUSE
1         <none>
2         <none>
3         new date 1
```

### Pause/Resume

Приостановить и возобновить процесс развертывания можно командами `kubectl rollout pause` и
`kubectl rollout resume`:
```console
$ k rollout pause deploy/nginx-deployment
deployment.apps/nginx-deployment paused
$ k set env deploy nginx-deployment ENV="$(date)"
deployment.apps/nginx-deployment env updated
$ k annotate deployment nginx-deployment kubernetes.io/change-cause="new date 3"
deployment.apps/nginx-deployment annotated
$ k rollout history deploy/nginx-deployment
deployment.apps/nginx-deployment
REVISION  CHANGE-CAUSE
1         <none>
2         <none>
3         new date 1
4         new date 2
$ k get deploy
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     0            3           3d22h
$ k rollout resume deploy/nginx-deployment
deployment.apps/nginx-deployment resumed
$ k get deploy
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           3d22h
$ k rollout history deploy/nginx-deployment
deployment.apps/nginx-deployment
REVISION  CHANGE-CAUSE
1         <none>
2         <none>
3         new date 1
4         new date 2
5         new date 3
```

### Rollback

Откатиться на предыдущую ревизию можно с помощью команды `kubectl rollout undo`, при этом создается
новая ревизия:
```console
$ k rollout history deploy/nginx-deployment
deployment.apps/nginx-deployment
REVISION  CHANGE-CAUSE
1         new date 1
2         new date 2
3         new date 3
$ k rollout undo deploy/nginx-deployment
deployment.apps/nginx-deployment rolled back
$ k rollout history deploy/nginx-deployment
deployment.apps/nginx-deployment
REVISION  CHANGE-CAUSE
1         new date 1
3         new date 3
4         new date 2
```
Также можно явно задать ревизию через флаг `--to-revision=<n>`, где n - номер ревизии:
```console
$ k rollout undo deploy/nginx-deployment --to-revision 1
deployment.apps/nginx-deployment rolled back
$ k rollout history deploy/nginx-deployment
deployment.apps/nginx-deployment
REVISION  CHANGE-CAUSE
3         new date 3
4         new date 2
5         new date 1
```

### Restart

Если требуется просто пересоздать все поды без изменения его шаблона, то можно использовать команду
`kubectl rollout restart`, при этом также появится новая ревизия:
```console
$ k rollout restart deploy/nginx-deployment
deployment.apps/nginx-deployment restarted
$ k rollout history deploy/nginx-deployment
deployment.apps/nginx-deployment
REVISION  CHANGE-CAUSE
3         new date 3
4         new date 2
5         new date 1
6         new date 1
```

## Scaling

Управление количеством реплик ресурса производится путем запроса к сабресурсу `/scale` с указанием
количества необходимых реплик. Такую операцию поддерживают стандартные ресурсы Deployment, ReplicaSet и
StatefulSet.
```console
curl -XPATCH -H "Content-Type: application/merge-patch+json" \
  "https://${api}/apis/apps/v1/namespaces/default/deployments/nginx-deployment/scale" \
  -d '{"spec":{"replicas":2}}'
```

В утилите kubectl есть команда `kubectl scale` позволяющая производить данную операцию.
```console
$ k scale --replicas=10 deploy/nginx-deployment
deployment.apps/nginx-deployment scaled
$ k get po
NAME                               READY   STATUS    RESTARTS   AGE
nginx-deployment-76d6c9b8c-56lvx   1/1     Running   0          31s
nginx-deployment-76d6c9b8c-6n5cn   1/1     Running   0          31s
nginx-deployment-76d6c9b8c-8fbrs   1/1     Running   0          31s
nginx-deployment-76d6c9b8c-bzj82   1/1     Running   0          32s
nginx-deployment-76d6c9b8c-cbpzt   1/1     Running   0          17m
nginx-deployment-76d6c9b8c-ljtql   1/1     Running   0          31s
nginx-deployment-76d6c9b8c-qqdkt   1/1     Running   0          31s
nginx-deployment-76d6c9b8c-tblnb   1/1     Running   0          17m
nginx-deployment-76d6c9b8c-wkn5b   1/1     Running   0          30s
nginx-deployment-76d6c9b8c-zvkkr   1/1     Running   0          31s
$ k scale --replicas=0 deploy/nginx-deployment
deployment.apps/nginx-deployment scaled
$ k get po
NAME                               READY   STATUS        RESTARTS   AGE
nginx-deployment-76d6c9b8c-56lvx   1/1     Terminating   0          43s
nginx-deployment-76d6c9b8c-8fbrs   0/1     Terminating   0          43s
nginx-deployment-76d6c9b8c-bzj82   1/1     Terminating   0          44s
nginx-deployment-76d6c9b8c-cbpzt   1/1     Terminating   0          17m
nginx-deployment-76d6c9b8c-ljtql   1/1     Terminating   0          43s
nginx-deployment-76d6c9b8c-qqdkt   0/1     Terminating   0          43s
nginx-deployment-76d6c9b8c-tblnb   1/1     Terminating   0          17m
nginx-deployment-76d6c9b8c-zvkkr   1/1     Terminating   0          43s
$ k get po
No resources found in default namespace.
```

### Horizontal Pod Autoscaling

```{mermaid}
graph BT
  hpa[Horizontal Pod Autoscaler] --> scale[Scale]
  subgraph rc[RC / Deployment]
    scale
  end
  scale -.-> pod1[Pod 1]
  scale -.-> pod2[Pod 2]
  scale -.-> pod3[Pod N]

  classDef hpa fill:#D5A6BD,stroke:#1E1E1D,stroke-width:1px,color:#1E1E1D;
  classDef rc fill:#F9CB9C,stroke:#1E1E1D,stroke-width:1px,color:#1E1E1D;
  classDef scale fill:#B6D7A8,stroke:#1E1E1D,stroke-width:1px,color:#1E1E1D;
  classDef pod fill:#9FC5E8,stroke:#1E1E1D,stroke-width:1px,color:#1E1E1D;
  class hpa hpa;
  class rc rc;
  class scale scale;
  class pod1,pod2,pod3 pod
```

## Disruptions

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: nginx
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: nginx
```

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: nginx
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: nginx
```

[deploy]:https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
[conditions]:https://pkg.go.dev/k8s.io/api/apps/v1#DeploymentConditionType
[sts]:https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/
[ds]:https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
[job]:https://kubernetes.io/docs/concepts/workloads/controllers/job/
[cronjob]:https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/
[cron]:https://ru.wikipedia.org/wiki/Cron
