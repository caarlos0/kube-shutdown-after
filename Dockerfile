FROM scratch
COPY kube-shutdown-after /kube-shutdown-after
ENTRYPOINT ["/kube-shutdown-after"]
