# k8s-arbitrage
## Requirements

* Docker
* Miniqube
* Kubectl
## Installing
```
$ eval $(minikube docker-env)
$ docker build . -t k8s-arbitrage:0.3
$ kubectl apply -f pv.yaml
$ kubectl apply -f pod.yaml
```

## Running

```
$ kubectl cp arbitrage.pcap k8s-arbitrage:/data/ar16.pcap
$ kubectl logs k8s-arbitrage
$ minikube ssh
# ls -la /data/arbitrage
...
-rw-r--r--. 1 docker 1000 370632 Mar 27 13:05 ar16.pcap
-rw-r--r--. 1 root   root  42022 Mar 27 13:05 ar16.pcap1616850356618407031pod0.json
```