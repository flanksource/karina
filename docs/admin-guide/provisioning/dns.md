### Dynamic DNS
```yaml
dns:
  nameserver: NAMESERVER:53
  algorithm: hmac-md5
  key: !!env DNS_KEY
  keyName: k8s.
  zone: k8s
```



### Route 53

