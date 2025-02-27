## Synopsis

![](https://i.imgur.com/waxVImv.png)
### [View all Roadmaps](https://github.com/nholuongut/all-roadmaps) &nbsp;&middot;&nbsp; [Best Practices](https://github.com/nholuongut/all-roadmaps/blob/main/public/best-practices/) &nbsp;&middot;&nbsp; [Questions](https://www.linkedin.com/in/nholuong/)
<br/>

*dregsy* lets you sync *Docker* images between registries, public or private. Several sync tasks can be defined, as one-off or periodic tasks (see *Configuration* section). Images are synced using [*Skopeo*](https://github.com/nholuongut/skopeo).

## Configuration
Sync tasks are defined in a YAML config file:

```yaml
# relay config sections
skopeo:
  # path to the skopeo binary; defaults to 'skopeo', in which case it needs to
  # be in PATH
  binary: skopeo
  # directory under which to look for client certs & keys, as well as CA certs
  # (see note below)
  certs-dir: /etc/skopeo/certs.d

# list of sync tasks
tasks:

  - name: task1 # required

    # interval in seconds at which the task should be run; when omitted,
    # the task is only run once at start-up
    interval: 60

    # determines whether for this task, more verbose output should be
    # produced; defaults to false when omitted
    verbose: true

    # does not copy the tag if it already exists in the target registry
    # This can cause problems for tags that are not immutable such
    # as 'latest'; defaults to false when omitted
    skipExistingTags: false

    # 'source' and 'target' are both required and describe the source and
    # target registries for this task:
    #  - 'registry' points to the server; required
    #  - 'auth' contains the base64 encoded credentials for the registry
    #    in JSON form {"username": "...", "password": "..."}
    #  - 'auth-refresh' specifies an interval for automatic retrieval of
    #    credentials; only for AWS ECR (see below)
    #  - 'skip-tls-verify' determines whether to skip TLS verification for the
    #    registry server (only for 'skopeo', see note below); defaults to false
    source:
      registry: source-registry.acme.com
      auth: eyJ1c2VybmFtZSI6ICJhbGV4IiwgInBhc3N3b3JkIjogInNlY3JldCJ9Cg==
    target:
      registry: dest-registry.acme.com
      auth: eyJ1c2VybmFtZSI6ICJhbGV4IiwgInBhc3N3b3JkIjogImFsc29zZWNyZXQifQo=
      skip-tls-verify: true

    # 'mappings' is a list of 'from':'to' pairs that define mappings of image
    # paths in the source registry to paths in the destination; 'from' is
    # required, while 'to' can be dropped if the path should remain the same as
    # 'from'. Additionally, the tags being synced for a mapping can be limited
    # by providing a 'tags' list. When omitted, all image tags are synced.
    mappings:
      - from: test/image
        to: archive/test/image
        tags: ['0.1.0', '0.1.1']
      - from: test/another-image
      - from: test/yet-another-image
        to: archive/test/yet-anotheerimage
        tags: ['>=v1.2.3'] 
        exludeTags: ['v1.5.*'] # All tags greater than 1.2.3 except 1.5.*
```

### Tags Filtering

Tags support simple logic:
 * Wildcards '\*' are allowed - for example 'v0.1.\*', or 'v1.\*.\*'
 * Comparison operators are supported - '>=v0.2', '>1', '<1.2', '<=1'. In this case the tag can not contain wildcards.


### Repository Validation & Client Authentication with TLS

When connecting to source and target repository servers, TLS validation is performed to verify the identity of a server. If you're using self-signed certificates for a repo server, or a server's certificate cannot be validated with the CA bundle available on your system, you need to provide the required CA certs. (The *dregsy* *Docker* image includes the CA bundle from the official `golang` image). Also, if a repo server requires client authentication, i.e. mutual TLS, you need to provide an appropriate client key & cert pair.

For this, specify the root folder with the `skopeo` setting `certs-dir` (defaults to `/etc/skopeo/certs.d`). However, it's important to note the following differences:

- When a repo server uses a non-standard port, the port number is included in image references when pulling and pushing. For TLS validation, `docker` will accordingly expect a `{registry host name}:{port}` folder. For `skopeo`, this is not the case, i.e. the port number is dropped from the folder name. This was a conscious decision to avoid pain when running *dregsy* in *Kubernetes* and mounting certs & keys from secrets: [mount paths must not contain `:`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#volumemount-v1-core).

- To skip TLS verification for a particular repo server when using the `docker` relay, you need to [configure the *Docker* daemon accordingly](https://docs.docker.com/registry/insecure/). With `skopeo`, you can easily set this in any source or target definition with the `skip-tls-verify` setting.


### *AWS ECR*

If a source or target is an *AWS ECR* registry, you need to retrieve the `auth` credentials via *AWS CLI*. They would however only be good for 12 hours, which is ok for one off tasks. For periodic tasks, or to avoid retrieving the credentials manually, you can specify an `auth-refresh` interval as a *Go* `Duration`, e.g. `10h`. If set, *dregsy* will initially and whenever the refresh interval has expired retrieve new access credentials. `auth` can be omitted when `auth-refresh` is set. Setting `auth-refresh` for anything other than an *AWS ECR* registry will raise an error.

Note however that you either need to set environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` for the *AWS* account you want to use and a user with sufficient permissions. Or if you're running *dregsy* on an *EC2* instance in your *AWS* account, the machine should have an appropriate instance profile. An according policy could look like this:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:CreateRepository"
      ],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:BatchCheckLayerAvailability",
        "ecr:DescribeRepositories",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload"
      ],
      "Resource": "arn:aws:ecr:<your_region>:<your_account>:repository/*"
    }
  ]
}
```


## Usage

```bash
dregsy -config= {path to config file}
```

If there are any periodic sync tasks defined (see *Configuration* above), *dregsy* remains running indefinitely. Otherwise, it will return once all one-off tasks have been processed.

### Running Natively
If you run *dregsy* natively on your system, with relay type `docker`, the *Docker* daemon of your system will be used as the relay for all sync tasks, so all synced images will wind up in the *Docker* storage of that daemon.

### Running Inside a *Docker* Container
You can use the [*dregsy* image on Dockerhub](https://hub.docker.com/r/yannhamon/dregsy/) for running *dregsy* containerized.

#### With `skopeo` relay
The image includes the `skopeo` binary, so all that's needed is:

```bash
docker run --rm -v {path to config file}:/config.yaml yannhamon/dregsy}
```

#### With `docker` relay
This will still use the local *Docker* daemon as the relay:

```bash
docker run --privileged --rm -v {path to config file}:/config.yaml -v /var/run/docker.sock:/var/run/docker.sock yannhamon/dregsy}
```

### Running On *Kubernetes*

When you run a *Docker* registry inside your *Kubernetes* cluster as an image cache, *dregsy* can come in handy as an automated updater for that cache. The example config below uses the `skopeo` relay:

```yaml
relay: skopeo
tasks:
  - name: task1
    interval: 60
    source:
      registry: registry.acme.com
      auth: eyJ1c2VybmFtZSI6ICJhbGV4IiwgInBhc3N3b3JkIjogInNlY3JldCJ9Cg==
    target:
      registry: registry.my-cluster
      auth: eyJ1c2VybmFtZSI6ICJhbGV4IiwgInBhc3N3b3JkIjogImFsc29zZWNyZXQifQo=
    mappings:
      - from: test/image
        to: archive/test/image
      - from: test/another-image
```

To keep your registry auth tokens in the config file secure, we are creating a Kubernetes _Secret_ instead of a _ConfigMap_:

```sh
kubectl create secret generic dregsy-config --from-file=./config.yaml
```

In addition, you will most likely want to mount client certs & keys, and CA certs from *Kubernetes* secrets into the pod for TLS validation to work. (The CA bundle from the official `golang` image is already included in the *dregsy* image.)

```yaml
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: kube-registry-updater
  namespace: kube-system
  labels:
    k8s-app: kube-registry-updater
    kubernetes.io/cluster-service: "true"
spec:
  serviceName: kube-registry-updater
  replicas: 1
  template:
    metadata:
      labels:
        k8s-app: kube-registry-updater
        kubernetes.io/cluster-service: "true"
    spec:
      containers:
      - name: dregsy
        image: yannhamon/dregsy
        command: ['dregsy', '-config=/config/config.yaml']
        resources:
          requests:
            cpu: 10m
            memory: 32Mi
        volumeMounts:
        - name: dregsy-config
          mountPath: /config
          readOnly: true
      volumes:
      - name: dregsy-config
        secret:
          secretName: dregsy-config
```


## Building

The `Makefile` has targets for building the binary and *Docker* image, and other stuff. Just run `make` to get a list. Note that for consistency, building is done inside a *Golang* build container, so you will need *Docker* to build. *dregsy*'s *Docker* image is based on *Alpine*, and installs *Skopeo* via `apk` during the image build.

# 🚀 I'm are always open to your feedback.  Please contact as bellow information:
### [Contact ]
* [Name: Nho Luong]
* [Skype](luongutnho_skype)
* [Github](https://github.com/nholuongut/)
* [Linkedin](https://www.linkedin.com/in/nholuong/)
* [Email Address](luongutnho@hotmail.com)
* [PayPal.me](https://www.paypal.com/paypalme/nholuongut)

![](https://i.imgur.com/waxVImv.png)
![](Donate.png)
[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/nholuong)

# License