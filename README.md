# Netangels DNS webhook service for cert-manager support
  
Эта служба может быть установлена рядом с cert manager и использоваться для обработки вызовов dns-01, предоставляемых cert manager. Всю документацию по настройке dns-01 chalanges можно найти по адресу   [cert-manager.io](https://cert-manager.io/docs/configuration/acme/dns01/webhook/)

### Version support:

### Platfom support:
`linux/amd64`, `linux/arm64`, `linux/arm`, `linux/arm/v6`, `linux/386` 

### Deploy
#### Helm chart: 
Add repo:
```shell
    helm repo add netangels-dns-webhook https://navigatore300.github.io/netangels-dns-webhook/
```
Then:
```shell
    helm install my-netangels-dns-webhook netangels-dns-webhook/netangels-dns-webhook --version <version>
```
#### As sub-chart:
```YAML
    dependencies:
        - name: netangels-dns-webhook
          version: <version>
          repository: https://mavigatore300.github.io/netangels-dns-webhook/
          alias: netangels-dns-webhook
```
### Usage:

**Credentials secret:**
Вы должны самостоятельно создать секрет, содержащий ваши netangels.ru api ключ, и его имя должно совпадать с именем secret ref, указанным в конфигурации эмитента cert-manager/clusterIssuer.

#### Issuer/ClusterIssuer:
```YAML
    apiVersion: cert-manager.io/v1
    kind: ClusterIssuer
    metadata:
        name: letsencrypt-nginx
    spec:
        acme:
            email: <your_acme_email>
            server: https://acme-v02.api.letsencrypt.org/directory
            privateKeySecretRef:
                name: letsencrypt-nginx-private-key
            solvers:
            - dns01:
                webhook:
                    groupName: com.github.mavigatore300.cert-manager-netangels-webhook
                    solverName: netangels-dns-solver
                    config:
                        secretName: netangels-credentials # notice the name
              selector:
                dnsZones:
                - '<your_domain>'
```

**Credentials in config:**
Вы можете использовать конфигурацию вебхука напрямую, как показано ниже.
**_(use it at your own risk)_**
```diff
-              secretName: netangels-credentials # notice the name
+              accountName: "<account-name>"
+              apiKey: "<api-key>"
```
#### Secret
```YAML
    apiVersion: v1
    kind: Secret
    data:
        account-name: <your_account_name>
        api-key: <your_api_key>
    metadata:
        name: netangels-credentials # notice the name
        namespace: <namespace-where-cert-manager-is-installed>
    type: Opaque
```
### cert-manager namespace:

Вы можете переопределить значения на свои собственные, если вы решили установить cert-manager в пользовательском пространстве имен следующим образом (это необходимо для правильной работы):
```YAML
    netangels-dns-webhook:
        certManager:
            namespace: <cert-manager-namespace>
            serviceAccountName: <cert-manager-namespace>
```
### Resources:
Выбор ресурсных ограничений я оставляю на ваше усмотрение, поскольку вы знаете, на чем запускать службу. ;)
```YAML
    netangels-dns-webhook:
        resources: 
            limits:
                cpu: 100m  
                memory: 128Mi
            requests:
                cpu: 100m
                memory: 128Mi
```

### Logging:
Вы можете поднять уровень регистрации до отладочного, установив следующие значения:
```YAML
    netangels-dns-webhook:
        logLevel: DEBUG
```
Уровень отладки дает вам немного больше контекста при отладке вашей установки. Уровень журнала по умолчанию - INFO.

### Running the test suite:

Обновите [config](testdata/netangels-dns-webhook/config.json) или [netangels-credentials](testdata/netangels-dns-webhook/netangels-credentials.yaml) secret на Ваш API key и запустите:

```bash
$ TEST_ZONE_NAME=example.com. make test
```

## Parameters 

В следующей таблице перечислены настраиваемые параметры диаграммы netangels-dns-webhook и их значения по умолчанию.

| Parameter                        | Description                       | Default                                          |
|----------------------------------|-----------------------------------|--------------------------------------------------|
| `groupName`                      | Group name for the webhook        | `com.github.mavigatore300.cert-manager-netangels-webhook` |
| `debugLevel`                     | Logging level                     | `INFO`                                           |
| `certManager.namespace`          | cert-manager namespace            | `cert-manager`                                   |
| `certManager.serviceAccountName` | cert-manager service account name | `cert-manager`                                   |
| `image.repository`               | Docker image repository           | ``         |
| `image.tag`                      | Docker image tag                  | ``                                         |
| `image.pullPolicy`               | Docker image pull policy          | `IfNotPresent`                                   |
| `nameOverride`                   | Name override for the chart       | `""`                                             |
| `fullnameOverride`               | Full name override for the chart  | `""`                                             |
| `service.type`                   | Service type                      | `ClusterIP`                                      |
| `service.port`                   | Service port                      | `443`                                            |
| `resources`                      | Pod resources                     | Check `values.yaml` file                         |
| `nodeSelector`                   | Node selector                     | `nil`                                            |
| `tolerations`                    | Node toleration                   | `nil`                                            |
| `affinity`                       | Node affinity                     | `nil`                                            |

##### Special credits to: **Keyhole Aps** 
