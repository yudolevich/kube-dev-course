## Workloads

```{image} img/workloads.svg
:width: 200px
```

###

```{revealjs-fragments}
* **Deployment** и **ReplicaSet**
* **StatefulSet**
* **DaemonSet**
* **Job** и **CronJob**
```

### Deployment/ReplicaSet

```{image} img/deploy-rs-pod.svg
:width: 500px
```

### ReplicaSet ![](img/rs.svg)

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|8|9-11|12-19
---
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

### Deployment ![](img/deploy.svg)

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|8|9-10|11|12-14|15-22
---
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

### Selector

```{revealjs-code-block} yaml
  selector:
    matchLabels:
      app: nginx
      pod-template-hash: 76d6c9b8c
  template:
    metadata:
      labels:
        app: nginx
        pod-template-hash: 76d6c9b8c
```

### Update

```{revealjs-fragments}
* **Recreate**
* **RollingUpdate**
```

### Recreate

```{revealjs-code-block} console
---
data-line-numbers: 1-6|7-8|9-11|12-16|17-20|21-25|26-30
---
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

### RollingUpdate

```{revealjs-code-block} yaml
---
data-line-numbers: 1-6|5|6
---
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 25%
      maxSurge: 25%
```

### RollingUpdate

```{revealjs-code-block} console
---
data-line-numbers: 1-26|3|5-7|8-10|11-13|14-16|17-19|20-22|23-26
---
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

### Conditions

```{revealjs-code-block} console
---
data-line-numbers: 1|8|22-26|27-28|29-31
---
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

### Conditions

```
Replicas: 3 desired | 2 updated | 4 total | 3 available | 1 unavailable
```

```{revealjs-fragments}
* **desired** - ожидаемое количество реплик
* **updated** - количество реплик с новой версией
* **total** - общее количество существующих реплик(старой и новой версии)
* **available** - количество реплик в состоянии Ready
* **unavailable** - количество реплик в состоянии NotReady
```

### Conditions

```
OldReplicaSets:  nginx-deployment-76d6c9b8c (2/2 replicas created)
NewReplicaSet:   nginx-deployment-bc8fc6c46 (2/2 replicas created)
```

### Conditions

```{revealjs-fragments}
* **Available**
* **Progressing**
* **ReplicaFailure**
```

### StatefulSet ![](img/sts.svg)

### StatefulSet

```{revealjs-fragments}
* Стабильные и уникальные сетевые идентификаторы
* Стабильное постоянное хранилище
* Упорядоченный процесс развертывания и увеличения реплик
* Упорядоченный процесс обновления
```

### StatefulSet
```{revealjs-code-block} yaml
---
data-line-numbers: 2-14|16-19|21-23|24|25|26|27-28|29-42|43-51
---
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
  replicas: 3
  podManagementPolicy: OrderedReady
  updateStrategy:
    type: RollingUpdate
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

### Pod Identity

```{revealjs-code-block} console
$ k get pods
NAME    READY   STATUS    RESTARTS   AGE
web-0   1/1     Running   0          2m21s
web-1   1/1     Running   0          109s
web-2   1/1     Running   0          82s
```

### DNS

```{revealjs-code-block} console
---
data-line-numbers: 1-3|4-6|7-12
---
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

### Pod Management Policies

```{revealjs-fragments}
* OrderedReady
  * Увеличение реплик от 0 до N-1
  * Уменьшение реплик от N-1 до 0
  * Перед созданием предыдущие Running и Ready
  * Перед удалением предыдущие завершены и удалены
* Parallel
```

### Update

```{revealjs-fragments}
* **OnDelete**
* **RollingUpdate**
```

### DaemonSet ![](img/ds.svg)

### DaemonSet
```{revealjs-fragments}
* Запуск контроллера, управляющего хранилищем на ноде
* Запуск коллектора логов на ноде
* Запуск агента мониторинга на ноде
```

### DaemonSet

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|9-12|13-17|20-24|28-43
---
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

### Pod Placement

```{revealjs-fragments}
* spec.template.spec.nodeSelector
* spec.template.spec.affinity.nodeAffinity
* spec.template.spec.tolerations
```

### Job ![](img/job.svg)

### Job

```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|6|7|8-14
---
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

### CronJob ![](img/cronjob.svg)

### CronJob

```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|6|7-17
---
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

### Rollout ![](img/rollout.svg)

```console
$ kubectl rollout
```

### Status

```{revealjs-code-block} console
---
data-line-numbers: 3|4-12
---
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

```{revealjs-code-block} console
---
data-line-numbers: 5|6-10|3-4
---
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

```{revealjs-code-block} console
---
data-line-numbers: 1-2|3-6|7-13|14-16|17-18|19-29
---
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

```{revealjs-code-block} console
---
data-line-numbers: 1-6|7-8|9-14|15-16|17-22
---
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

```{revealjs-code-block} console
---
data-line-numbers: 1-2|3-9
---
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

### Scaling ![](img/scaling.svg)

```console
$ kubectl rollout
```

### Scaling

```console
$ curl -XPATCH -H "Content-Type: application/merge-patch+json" \
  "https://${api}/apis/apps/v1/namespaces/default/deployments/nginx-deployment/scale" \
  -d '{"spec":{"replicas":2}}'
```

### Scaling

```{revealjs-code-block} console
---
data-line-numbers: 1-2|3-14|15-16|17-26|27-28
---
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

```{image} img/hpa-mermaid.svg
:width: 500px
```

### Horizontal Pod Autoscaling

```{revealjs-code-block} console
$ kubectl autoscale deployment nginx-deployment --min=2 --max=5 --cpu-percent=80
horizontalpodautoscaler.autoscaling/nginx-deployment autoscaled
```


### Horizontal Pod Autoscaling

```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|6|7|8-14|15-18
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: nginx-deployment
spec:
  minReplicas: 2
  maxReplicas: 5
  metrics:
  - resource:
      name: cpu
      target:
        averageUtilization: 80
        type: Utilization
    type: Resource
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx-deployment
```

### Disruptions

### Involuntary

```{revealjs-fragments}
* Отказ железа на физической машине
* Удаление ВМ по ошибке
* Kernel panic
* Отключение ноды от кластера из-за сетевых проблем
```

### Voluntary

```{revealjs-fragments}
* Обновление шаблона пода в Deployment
* Выселение подов для обновления ноды
* Автоскейлинг нод
* Удаление пода с ноды для высвыбождения ресурсов
```

### Dealing with disruptions

```{revealjs-fragments}
* Правильно выставить ресурсы(requests/limits)
* Использовать несколько реплик
* Распределять реплики на разных нодах/зонах
```

### Pod disruption budgets

```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|6|7-9
---
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

### Pod disruption budgets

```{revealjs-code-block} yaml
---
data-line-numbers: 6
---
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
