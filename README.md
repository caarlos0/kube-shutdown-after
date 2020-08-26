# kube-shutdown-after

Shuts down deployments after a given time.

It can be useful to scale down sandbox clusters after work hours.

## Deploy

Just run:

```sh
helm repo add carlos https://carlos-charts.storage.googleapis.com
helm repo update
helm install carlos/kube-shutdown-after
```

It will create a deployment in the `kube-system` namespace.

`kube-shutdown-after` will loop through deployments every ~1 minute.

## Usage

Annotate the deployments you want to be shutdown with `shutdown-after` and
a time in `HH:mm TZ` format:

```sh
kubectl annotate deploy whatever shutdown-after='19:00 -02'
```

This will make `kube-shutdown-after` keep this deployment with 0 replicas
after 19:00 GMT-2.
