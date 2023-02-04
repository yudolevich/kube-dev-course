@startmindmap

* core
 * what/why
 * components
  * api
  * etcd
  * sched
  * cm
  * kubelet
 * rest api(crud)
  * resource
   * yaml
   * meta
   * spec
   * sub(status)
 * auth
  * rbac

* resources
 * core
  * namespace
  * pod
  * deploy
  * replicaset
  * configmap
  * secret
  * serviceaccount
 * network
  * svc
  * ingress
 * batch
  * job
  * cronjob

* deploy
 * resource/limit
 * readiness/liveness
 * command/args
 * envs
 * volumes/volumeMounts
 * serviceAccount
 * securityContext
 * containers
  * init
  * sidecar
 * nodeselector
  * taints/tolerations
  * affinity
 * helm
  * go-template
   * values
   * helpers
   * funcs
  * package
  * hooks
  * history/rollback
  * test

* cli
 * kubectl
  * apply/create/get/describe/replace/patch/delete
  * logs
  * exec/attach/cp/rsync
  * port-forward/proxy
  * explain
  * api-resources
  * debug
  * config
   * kubeconfig
 * kind
 * minikube

* other
 * resourcequota
 * limitrange
 * debug container
 * testing
  * bluegreen
  * canary

@endmindmap
