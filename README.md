# Todo App with CDK

低コストでRailsアプリケーションをAWS ECS上にデプロイするためのCDKプロジェクトです。

## アーキテクチャ

- **ECR**: Dockerイメージの管理
- **ECS Fargate**: コンテナアプリケーションの実行（Spot インスタンス使用でコスト削減）
- **RDS MySQL**: データベース（t3.micro/t3.small）
- **Application Load Balancer**: HTTP/HTTPSアクセス（HTTPは自動でHTTPSにリダイレクト）
- **Route53**: カスタムドメイン管理（todo-app.kit.elca-web.com）
- **ACM**: SSL証明書の自動管理
- **VPC**: プライベートネットワーク環境

## コスト最適化

- Fargate Spot インスタンス（開発環境）
- RDS t3.micro（開発環境）、t3.small（本番環境）
- 最小限のAuto Scaling設定（1-3インスタンス）
- ログ保持期間1週間
- ECRイメージライフサイクル（最新10個まで保持）

## プロジェクト構成

```
cdk/
├── constructs/          # 再利用可能なコンストラクト
│   ├── base.go         # VPC、セキュリティグループ、IAMロール
│   ├── database.go     # RDS MySQL
│   ├── container.go    # ECR、ECS、ALB
│   └── dns.go          # Route53、ACM証明書
├── stacks/
│   └── todo-app-stack.go # メインスタック
├── cdk.go              # エントリーポイント
├── go.mod              # Go依存関係
└── README.md

program/todo-app-develop/  # Railsアプリケーション
```

## デプロイ手順

### 1. 前提条件

```bash
# AWS CLI設定
aws configure
# または
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_DEFAULT_REGION=ap-northeast-1

# CDKのインストール
npm install -g aws-cdk
```

### 2. CDKの初期化

```bash
cd cdk
cdk bootstrap
```

### 3. Route53 Hosted Zone の確認

デプロイ前に、elca-web.com のHosted Zoneが存在することを確認してください：

```bash
aws route53 list-hosted-zones --query 'HostedZones[?Name==`elca-web.com.`]'
```

### 4. スタックのデプロイ

```bash
# 開発環境（デフォルト）
RAILS_MASTER_KEY=[master.key] CDK_DEFAULT_REGION=ap-northeast-1 cdk deploy

# 本番環境
ENVIRONMENT=production CDK_DEFAULT_REGION=ap-northeast-1 cdk deploy

# カスタムドメインを使用する場合
SUBDOMAIN_NAME=your-custom.kit.elca-web.com CDK_DEFAULT_REGION=ap-northeast-1 cdk deploy
```

### 5. Dockerイメージのビルド・プッシュ

```bash
# ECRリポジトリURIを取得
ECR_URI=$(aws cloudformation describe-stacks --stack-name TodoAppStack --query 'Stacks[0].Outputs[?OutputKey==`ECRRepositoryURI`].OutputValue' --output text)

# ECRログイン
aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin $ECR_URI

# Dockerイメージのビルド
cd program/todo-app-develop
docker build --platform linux/amd64 -t todo-app:latest .

# イメージにタグ付け
docker tag todo-app:latest $ECR_URI:latest

# ECRにプッシュ
docker push $ECR_URI:latest
```

### 6. ECSサービスの更新

```bash
# サービスを強制更新してイメージを反映
aws ecs update-service --cluster todo-app-development-cluster --service todo-app-development-service --force-new-deployment
```

## 環境変数

### CDKデプロイ時

- `ENVIRONMENT`: デプロイ環境（development/production）
- `APP_NAME`: アプリケーション名（デフォルト: todo-app）
- `PARENT_DOMAIN_NAME`: 親ドメイン名（デフォルト: elca-web.com）
- `SUBDOMAIN_NAME`: サブドメイン名（デフォルト: todo-app.kit.elca-web.com）
- `CDK_DEFAULT_REGION`: デプロイリージョン（推奨: ap-northeast-1）

### Railsアプリケーション（自動設定）

- `RAILS_ENV`: production
- `DB_HOST`: RDSエンドポイント
- `DB_PORT`: 3306
- `DB_NAME`: app_production
- `DB_USERNAME`: Secrets Managerから取得
- `APP_DATABASE_PASSWORD`: Secrets Managerから取得

## アクセス

デプロイ完了後、アプリケーションにアクセスできます：

### カスタムドメイン
```
https://todo-app.kit.elca-web.com
```

### ALBのDNS名でアクセス（フォールバック）
```bash
aws cloudformation describe-stacks --stack-name TodoAppStack --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' --output text
```

**注意**: HTTPアクセスは自動的にHTTPSにリダイレクトされます。

## 削除

```bash
cd cdk
cdk destroy
```

## 料金目安（東京リージョン）

### 開発環境（月額）
- ECS Fargate Spot: 約$5-15（使用時間により変動）
- RDS t3.micro: 約$15
- ALB: 約$20
- **合計: 約$40-50/月**

### 本番環境（月額）
- ECS Fargate: 約$15-30
- RDS t3.small: 約$30
- ALB: 約$20
- **合計: 約$65-80/月**

※ 実際の料金は使用量により変動します。