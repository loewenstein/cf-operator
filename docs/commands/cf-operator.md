## cf-operator

cf-operator manages BOSH deployments on Kubernetes

### Synopsis

cf-operator manages BOSH deployments on Kubernetes

```
cf-operator [flags]
```

### Options

```
      --apply-crd                                (APPLY_CRD) If true, apply CRDs on start (default true)
      --bosh-dns-docker-image string             (BOSH_DNS_DOCKER_IMAGE) The docker image used for emulating bosh DNS (a CoreDNS image) (default "coredns/coredns:1.6.3")
  -n, --cf-operator-namespace string             (CF_OPERATOR_NAMESPACE) The operator namespace (default "default")
      --cluster-domain string                    (CLUSTER_DOMAIN) The Kubernetes cluster domain (default "cluster.local")
      --ctx-timeout int                          (CTX_TIMEOUT) context timeout for each k8s API request in seconds (default 30)
  -o, --docker-image-org string                  (DOCKER_IMAGE_ORG) Dockerhub organization that provides the operator docker image (default "cfcontainerization")
      --docker-image-pull-policy string          (DOCKER_IMAGE_PULL_POLICY) Image pull policy (default "IfNotPresent")
  -r, --docker-image-repository string           (DOCKER_IMAGE_REPOSITORY) Dockerhub repository that provides the operator docker image (default "cf-operator")
  -t, --docker-image-tag string                  (DOCKER_IMAGE_TAG) Tag of the operator docker image (default "0.0.1")
  -h, --help                                     help for cf-operator
  -c, --kubeconfig string                        (KUBECONFIG) Path to a kubeconfig, not required in-cluster
  -l, --log-level string                         (LOG_LEVEL) Only print log messages from this level onward (default "debug")
      --max-boshdeployment-workers int           (MAX_BOSHDEPLOYMENT_WORKERS) Maximum number of workers concurrently running BOSHDeployment controller (default 1)
      --max-extendedjob-workers int              (MAX_EXTENDEDJOB_WORKERS) Maximum number of workers concurrently running ExtendedJob controller (default 1)
      --max-extendedsecret-workers int           (MAX_EXTENDEDSECRET_WORKERS) Maximum number of workers concurrently running ExtendedSecret controller (default 5)
      --max-extendedstatefulset-workers int      (MAX_EXTENDEDSTATEFULSET_WORKERS) Maximum number of workers concurrently running ExtendedStatefulSet controller (default 1)
  -w, --operator-webhook-service-host string     (CF_OPERATOR_WEBHOOK_SERVICE_HOST) Hostname/IP under which the webhook server can be reached from the cluster
  -p, --operator-webhook-service-port string     (CF_OPERATOR_WEBHOOK_SERVICE_PORT) Port the webhook server listens on (default "2999")
  -x, --operator-webhook-use-service-reference   (CF_OPERATOR_WEBHOOK_USE_SERVICE_REFERENCE) If true the webhook service is targetted using a service reference instead of a URL
      --watch-namespace string                   (WATCH_NAMESPACE) Namespace to watch for BOSH deployments
```

### SEE ALSO

* [cf-operator util](cf-operator_util.md)	 - Calls a utility subcommand
* [cf-operator version](cf-operator_version.md)	 - Print the version number

###### Auto generated by spf13/cobra on 17-Oct-2019
