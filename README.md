# kube-shutdown-after

Shuts down deployments after a given time.

It can be useful to scale down sandbox clusters after work hours.

## Deploy

```sh
kubectl create -f https://raw.githubusercontent.com/caarlos0/kube-shutdown-after/master/deployment.yaml
```

## Usage

Annotate the deployments you want to be shutdown with `shutdown-after` and
a time in `HH:mm` format:

```sh
kubectl annotate deploy whatever shutdown-after=19:00
```

