# irsa-manager

TODO

- [x] delete secret
- [x] delete s3
- [x] only once secret creation (status check)
- [x] no update secret
- [x] no update bucket object
- [x] delete idp
- [x] issue: when irsasetup was deleted, resource remained with some error occured
- [x] certificate
- [ ] check keys.json keyid has to be empty or not
- [ ] IRSA api
- [ ] use with cert-manager

- [ ] validation webhook (invalid to change)

```
kubectl get secret -n kube-system irsa-manager-key -o jsonpath="{.data.ssh-privatekey}" | base64 --decode | sudo tee /etc/kubernetes/pki/irsa-manager.key > /dev/null
kubectl get secret -n kube-system irsa-manager-key -o jsonpath="{.data.ssh-publickey}" | base64 --decode | sudo tee /etc/kubernetes/pki/irsa-manager.pub > /dev/null
```
