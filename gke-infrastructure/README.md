# GKE Cluster Setup and Microservices Deployment
## 前提条件
- `gcloud` がインストールされており、ログイン済みであること。

## GKE クラスターの作成（アーキテクチャ: linux/amd64）
```bash
gcloud container clusters create go-microservice-k8s-cluster \
  --zone asia-northeast1-a \
  --num-nodes 3 \
  --project go-microservice-k8s
```

## クラスターへの認証
```bash
gcloud container clusters get-credentials go-microservice-k8s-cluster \
  --zone asia-northeast1-a \
  --project go-microservice-k8s
```

## 共有リソースの適用
```bash
kubectl apply -f shared-configmap.yaml
kubectl apply -f shared-mysql.yaml
```

## MySQL 権限の付与
```bash
kubectl exec -it mysql-0 -- /bin/bash

mysql -u root -p

GRANT ALL PRIVILEGES ON `microservice-k8s-demo-db`.* TO 'microservice-k8s-demo'@'%';
FLUSH PRIVILEGES;
```

## マイクロサービスのマニフェスト適用
### Cusotmer Service
```bash
kubectl apply -f customer.yaml
```

### Catalog Service
```bash
kubectl apply -f catalog.yaml
```

### Order Service
```bash
kubectl apply -f order.yaml
```

### Commerce Gateway
```bash
kubectl apply -f commerce-gateway.yaml
```

### Ingress
```bash
kubectl apply -f ingress.yaml
```
