# Kubernetes API Operations
## Examining requests

```
kubectl get pods --all-namespaces -v6
kubectl get pods --namespace default -v6
```


## Making requests
### Using kubectl as a proxy
```
kubectl proxy &
HOST=http://127.0.0.1:8001
```

### Creating a resource

```
cat > pod.yaml <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
EOF

curl $HOST/api/v1/namespaces/project1/pods \
    -H "Content-Type: application/yaml" \
    --data-binary @pod.yaml
```

### Getting information about a resource
```
curl -X GET \
    $HOST/api/v1/namespaces/project1/pods/nginx
```

### Getting the list of resources
```
curl $HOST/api/v1/pods
curl $HOST/api/v1/namespaces/project1/pods
```

### Filtering the result of a list
#### Using label selectors

```
kubectl run nginx1 --image nginx --labels mylabel=foo
kubectl run nginx2 --image nginx --labels mylabel=bar

curl $HOST/api/v1/namespaces/default/pods?labelSelector=mylabel

curl $HOST/api/v1/namespaces/default/pods?labelSelector=\!mylabel

curl $HOST/api/v1/namespaces/default/pods?labelSelector=mylabel==foo

curl $HOST/api/v1/namespaces/default/pods?labelSelector=mylabel=foo

curl $HOST/api/v1/namespaces/default/pods?labelSelector=mylabel\!=foo

curl $HOST/api/v1/namespaces/default/pods?labelSelector=mylabel+in+(foo,baz) 

curl $HOST/api/v1/namespaces/default/pods?labelSelector=mylabel+notin+(foo,baz) 

curl $HOST/api/v1/namespaces/default/pods?labelSelector=mylabel,otherlabel==bar

```

#### Using field selectors
```
curl $HOST/api/v1/namespaces/default/pods?fieldSelector=status.phase==Running

curl $HOST/api/v1/namespaces/default/pods?fieldSelector=status.phase=Running

curl $HOST/api/v1/namespaces/default/pods?fieldSelector=status.phase\!=Running

curl $HOST/api/v1/namespaces/default/pods?fieldSelector=status.phase==Running,spec.restartPolicy\!=Always
```

### Deleting a resource

```
curl -X DELETE \
    $HOST/api/v1/namespaces/project1/pods/nginx
```

### Deleting a collection of resources
```
curl -X DELETE \
	$HOST/api/v1/namespaces/project1/pods
```

### Updating a resource
```
cat > deploy.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
EOF

curl $HOST/apis/apps/v1/namespaces/project1/deployments \
    -H "Content-Type: application/yaml" \
    --data-binary @deploy.yaml

cat deploy.yaml | \
    sed 's/image: nginx/image: nginx:latest/' > \
    deploy2.yaml

curl -X PUT \
    $HOST/apis/apps/v1/namespaces/project1/deployments/nginx \
    -H "Content-Type: application/yaml" \
    --data-binary @deploy2.yaml
```

### Managing conflicts when updating a resource

```
cat > deploy.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
EOF

curl $HOST/apis/apps/v1/namespaces/project1/deployments \
    -H "Content-Type: application/yaml" \
    --data-binary @deploy.yaml

curl $HOST/apis/apps/v1/namespaces/project1/deployments/nginx

cat > deploy2.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  resourceVersion: "668867" # change this version
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
EOF

curl -X PUT \
    $HOST/apis/apps/v1/namespaces/project1/deployments/nginx \
    -H "Content-Type: application/yaml" \
    --data-binary @deploy2.yaml

sed -i 's/668867/668908/' deploy2.yaml

curl -X PUT \
    $HOST/apis/apps/v1/namespaces/project1/deployments/nginx \
    -H "Content-Type: application/yaml" \
    --data-binary @deploy2.yaml
```

### Using a strategic merge patch to update a resource
```
cat > deploy.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
EOF

curl $HOST/apis/apps/v1/namespaces/project1/deployments \
    -H "Content-Type: application/yaml" \
    --data-binary @deploy.yaml

cat > deploy-patch.json <<EOF
{
  "spec":{
    "template":{
      "spec":{
        "containers":[{
          "name":"nginx",
          "image":"nginx:alpine"
        }]
}}}}
EOF

curl -X PATCH \
    $HOST/apis/apps/v1/namespaces/project1/deployments/nginx \
    -H "Content-Type: application/strategic-merge-patch+json" \
    --data-binary @deploy-patch.json
```

### Applying resources server-side

```
cat > deploy.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        env:
        - name: key1
          value: value1
        - name: key2
          value: value2
        - name: key3
          value: value3
EOF

curl -X PATCH \
    $HOST/apis/apps/v1/namespaces/project1/deployments/nginx?\
    fieldManager=manager1 \
    -H "Content-Type: application/apply-patch+yaml" \
    --data-binary @deploy.yaml

cat > patch.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: key2
          value: value2bis
EOF

curl -X PATCH \
    "$HOST/apis/apps/v1/namespaces/project1/deployments/nginx? \
    fieldManager=manager2&force=true" \
    -H "Content-Type: application/apply-patch+yaml" \
    --data-binary @patch.yaml

cat > patch2.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        env:
        - name: key1
          value: value1
EOF

curl -X PATCH \
    "$HOST/apis/apps/v1/namespaces/project1/deployments/nginx? \
    fieldManager=manager1" \
    -H "Content-Type: application/apply-patch+yaml" \
    --data-binary @patch2.yaml
```

### Restarting a watch request

```
curl "$HOST/api/v1/namespaces/default/pods?watch=true"

curl "$HOST/api/v1/namespaces/default/pods?\
    watch=true&\
    resourceVersion=2436677" # change this version

curl "$HOST/api/v1/namespaces/default/pods?\
    watch=true&\
    resourceVersion=2436655" # change this version
```

### Allowing bookmarks to efficiently restart a watch request

```
curl "$HOST/api/v1/namespaces/project1/pods?\
    labelSelector=mylabel==foo&watch=true"

kubectl run nginx1 --image nginx --labels mylabel=foo
kubectl run nginx2 --image nginx --labels mylabel=bar

curl "$HOST/api/v1/namespaces/default/pods? \
    labelSelector=mylabel==foo& \
    watch=true& \
    allowWatchBookmarks=true"

kubectl delete pods nginx2

kubectl delete pods nginx1
kubectl run nginx2 --image nginx --labels mylabel=bar

curl "$HOST/api/v1/namespaces/default/pods? \
    labelSelector=mylabel==foo& \
    watch=true& \
    allowWatchBookmarks=true& \
    resourceVersion=2532566" # change this version
```

### Paginating results

```
curl "$HOST/api/v1/pods?limit=1"

curl "$HOST/api/v1/pods?limit=1&continue=<continue_token_1>"
```

## Getting results in different formats
### Getting result as table

```
curl $HOST/api/v1/pods \
    -H 'Accept: application/json;as=Table;g=meta.k8s.io;v=v1'

curl $HOST/api/v1/pods?includeObject=None \
    -H 'Accept: application/json;as=Table;g=meta.k8s.io;v=v1'
```

### Using the YAML format

```
curl $HOST/api/v1/pods -H 'Accept: application/yaml'

curl $HOST/api/v1/namespaces/default/pods \
    -H "Content-Type: application/yaml" \
    -H 'Accept: application/yaml' \
    --data-binary @pod.yaml
```

