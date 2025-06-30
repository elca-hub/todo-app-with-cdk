package main

import (
	"cdk/stacks"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Get environment variables
	environment := getEnvOrDefault("ENVIRONMENT", "development")
	appName := getEnvOrDefault("APP_NAME", "todo-app")
	parentDomainName := getEnvOrDefault("PARENT_DOMAIN_NAME", "elca-web.com")
	subdomainName := getEnvOrDefault("SUBDOMAIN_NAME", "todo-app.kit.elca-web.com")

	// Create Todo App Stack
	stacks.NewTodoAppStack(app, "TodoAppStack", &stacks.TodoAppStackProps{
		StackProps: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Todo App infrastructure stack with ECR, ECS, RDS, ALB, and Route53"),
			Tags: &map[string]*string{
				"Environment": jsii.String(environment),
				"Application": jsii.String(appName),
				"ManagedBy":   jsii.String("CDK"),
			},
		},
		Environment:      environment,
		AppName:          appName,
		ParentDomainName: parentDomainName,
		SubdomainName:    subdomainName,
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// Use current CLI configuration for dev stacks
	// This is required for Route53 hosted zone lookup
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	region := os.Getenv("CDK_DEFAULT_REGION")

	// デフォルト値を設定（必要に応じて変更）
	if account == "" {
		account = "123456789012" // プレースホルダー
	}
	if region == "" {
		region = "ap-northeast-1" // 東京リージョン
	}

	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
