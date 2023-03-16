# external-csi-ha-controller
HA Controller for CSI Volumes
## Quickstart

You can install CRDs ExternalHaController into a cluster using the included make chart. run this in the root directory.

```
$ make install 
```

You can uninstall CRDs ExternalHaController from a cluster using the included make chart. run this in the root directory.

```
$ make uninstall 
```

You can build the ExternalHaController-operator docker image using the included make chart. run this in the root directory.

```
$ make docker-build 
```

You can push the ExternalHaController-operator docker image using the included make chart. run this in the root directory.

```
$ make docker-push 
```

You can deploy the ExternalHaController-operator using the included make chart. run this in the root directory.

```
$ make deploy 
```

You can undeploy the ExternalHaController-operator using the included make chart. run this in the root directory.

```
$ make undeploy 
```

You can then create a ExternalHaController to connect to the cluster.

```
apiVersion: csiplugins.spdbdev.io/v1
kind: ExternalHaController
metadata:
  name: externalhacontroller-sample
spec:
  deletePod: true
```

```
$ kubectl apply -f externalhacontroller-sample.yaml
```

You can then access your ExternalHaController cluster via the`External-Csi-Ha-Controller`service.

## Disclaimers
