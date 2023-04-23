# Storage

В данном практическом занятии рассмотрим работу с постоянным хранилищем посредством ресурсов
[PersistentVolume][pv], [PersistentVolumeClaim][pvc] и [StorageClass][sc].

## Persistent Volume

Создадим ресурс `PersistentVolume`, который будет использовать файловую систему хоста, на котором он
будет расположен.
```bash
$ k apply -f - << EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-volume
spec:
  storageClassName: manual
  capacity:
    storage: 100Mi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data"
EOF
```

В данном примере в параметрах указано:
* `storageClassName: manual` - имя StorageClass, указано явно для будущего связывания с [pvc][]
* `capacity.storage` - размер тома
* `accessModes` - [режим подключения тома к подам][acc-modes]
* `hostPath.path` - путь на ноде, по которому будут храниться данные

После создания можно увидеть его состояние:
```console
$ k get pv
NAME        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM   STORAGECLASS   REASON   AGE
pv-volume   100Mi      RWO            Retain           Available           manual                  12s
```

Также из вывода видно, что reclaim policy задано как `Retain`, означающая, что после удаления [pv][]
данные в нем не будут удалены. А также status - `Available`, означающий, что том готов к подключению.

## Persistent Volume Claim

Теперь необходимо создать ресурс `PersistentVolumeClaim`, который можно рассматривать как запрос на
том с определенными характеристиками(размер, параметры доступа, тип хранилища и т.д.), который можно
использовать непосредственно в поде.

```bash
$ k apply -f - << EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pv-claim
spec:
  storageClassName: manual
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Mi
EOF
```

После создания можно увидеть состояние [pvc][]:
```console
$ k get pvc
NAME       STATUS   VOLUME      CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pv-claim   Bound    pv-volume   100Mi      RWO            manual         2s
```

Можно заметить интересную особенность: при создании [pvc][] мы указали требуемый нам размер - 50Mi, при
этом параметр capacity - 100Mi. Дело в том, что при создании [pvc][] ищется любой подходящий по параметрам
[pv][], в том числе большего размера. Если теперь посмотреть список [pv][], то мы увидим, что он теперь
связан с созданным [pvc][]:
```console
$ k get pv
NAME      CAPACITY ACCESS MODES RECLAIM POLICY STATUS CLAIM            STORAGECLASS REASON AGE
pv-volume 100Mi    RWO          Retain         Bound  default/pv-claim manual              21m
```

## PVC in Pod
Теперь созданный [pvc][] можно использовать в параметре `spec.volumes` пода:
```bash
$ k apply -f - << EOF
apiVersion: v1
kind: Pod
metadata:
  name: pv-pod
spec:
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - mountPath: "/usr/share/nginx/html"
      name: pv-storage
  volumes:
  - name: pv-storage
    persistentVolumeClaim:
      claimName: pv-claim
EOF
```

После создания можно создать файл в нашем [pv][] выполнив команду внутри контейнера:
```console
$ k exec -it pv-pod -- /bin/sh -c 'echo "hello" > /usr/share/nginx/html/index.html'
```

Далее можно убедиться, что файл корректно создан обратившись к nginx либо через `port-forward`, либо с
использованием `Service` и `Ingress`. Также можно воспользоваться специальным сабресурсом `/proxy` у пода,
сделав запрос через api kubernetes:
```console
$ k get --raw /api/v1/namespaces/default/pods/pv-pod/proxy
hello
```

Создадим еще один под, который также будет использовать данный [pvc][]:
```bash
$ k apply -f - << EOF
apiVersion: v1
kind: Pod
metadata:
  name: pv-pod-new
spec:
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - mountPath: "/usr/share/nginx/html"
      name: pv-storage
  volumes:
  - name: pv-storage
    persistentVolumeClaim:
      claimName: pv-claim
EOF
```

После его запуска также сделаем запрос к nginx:
```console
$ k get po
NAME         READY   STATUS    RESTARTS   AGE
pv-pod       1/1     Running   0          11m
pv-pod-new   1/1     Running   0          10s
$ k get --raw /api/v1/namespaces/default/pods/pv-pod-new/proxy
hello
```

Как видно, под запустился и также смонтировал данный том не смотря на режим `ReadWriteOnce`. Это связано
с тем, что режим `ReadWriteOnce` позволяет монтировать [pv][], если поды находятся на одной ноде. Если
требуется запретить данную возможность, то существует режим `ReadWriteOncePod`.

Также можем удалить созданные нами ресурсы и убедиться, что данные остались на ноде согласно политике
`Retain`:
```console
$ k delete pvc --all
persistentvolumeclaim "pv-claim" deleted
$ k delete pv --all
persistentvolume "pv-volume" deleted
$ k delete po --all
pod "pv-pod" deleted
pod "pv-pod-new" deleted
$ docker exec  -it kind-control-plane /bin/sh -c 'cat /mnt/data/index.html'
hello
```

## Storage Class

В [kind][] при создании кластера также устанавливается `local-path-provisioner`, который позволяет
динамически нарезать [pv][] на нодах. Для их использования создается StorageClass:
```console
$ k get sc
NAME               PROVISIONER           RECLAIMPOLICY VOLUMEBINDINGMODE    ALLOWVOLUMEEXPANSION AGE
standard (default) rancher.io/local-path Delete        WaitForFirstConsumer false                14d
```

Так как данный Storage Class используется по умолчанию, то можно создавать [pvc][] без указания параметра
`storageClassName`. Таким образом нам достаточно создать [pvc][] и под, а [pv][] создастся автоматически:

```bash
$ k apply -f - << EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pv-claim-sc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 150Mi
---
apiVersion: v1
kind: Pod
metadata:
  name: pv-pod-sc
spec:
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - mountPath: "/usr/share/nginx/html"
      name: pv-storage
  volumes:
  - name: pv-storage
    persistentVolumeClaim:
      claimName: pv-claim-sc
EOF
```

```console
$ k get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                 STORAGECLASS   REASON   AGE
pvc-828145c8-2e48-4046-b96f-d3cadbece5d8   150Mi      RWO            Delete           Bound    default/pv-claim-sc   standard                28s

$ k get pvc
NAME          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pv-claim-sc   Bound    pvc-828145c8-2e48-4046-b96f-d3cadbece5d8   150Mi      RWO            standard       36s
$ k get po
NAME        READY   STATUS    RESTARTS   AGE
pv-pod-sc   1/1     Running   0          38s
```


[pv]:https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistent-volumes
[pvc]:https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims
[sc]:https://kubernetes.io/docs/concepts/storage/storage-classes/
[acc-modes]:https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes
[kind]:https://kind.sigs.k8s.io/
