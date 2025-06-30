package stacks

import (
	"cdk/constructs"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	constructs_lib "github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type TodoAppStackProps struct {
	awscdk.StackProps
	Environment      string
	AppName          string
	ParentDomainName string // elca-web.com
	SubdomainName    string // todo-app.kit.elca-web.com
}

type TodoAppStack struct {
	awscdk.Stack
	BaseInfrastructure *constructs.BaseInfrastructure
	Database           *constructs.Database
	Dns                *constructs.Dns
	ContainerService   *constructs.ContainerService
}

func NewTodoAppStack(scope constructs_lib.Construct, id string, props *TodoAppStackProps) *TodoAppStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Default values
	environment := props.Environment
	if environment == "" {
		environment = getEnvOrDefault("ENVIRONMENT", "development")
	}

	appName := props.AppName
	if appName == "" {
		appName = getEnvOrDefault("APP_NAME", "todo-app")
	}

	// ドメイン設定
	parentDomainName := props.ParentDomainName
	if parentDomainName == "" {
		parentDomainName = getEnvOrDefault("PARENT_DOMAIN_NAME", "elca-web.com")
	}

	subdomainName := props.SubdomainName
	if subdomainName == "" {
		subdomainName = getEnvOrDefault("SUBDOMAIN_NAME", "todo-app.kit.elca-web.com")
	}

	// Base Infrastructure
	baseInfra := constructs.NewBaseInfrastructure(stack, "BaseInfrastructure", &constructs.BaseInfrastructureProps{
		Environment: environment,
		AppName:     appName,
	})

	// Database
	database := constructs.NewDatabase(stack, "Database", &constructs.DatabaseProps{
		VPC:              baseInfra.VPC,
		SecurityGroup:    baseInfra.RDSSecurityGroup,
		Environment:      environment,
		AppName:          appName,
		DatabaseName:     "app_production",
		DatabaseUsername: "app",
	})

	railsMasterKey := getEnvOrDefault("RAILS_MASTER_KEY", "")

	// 最初にContainer Service（証明書なし）を作成
	containerService := constructs.NewContainerService(stack, "ContainerService", &constructs.ContainerServiceProps{
		VPC:               baseInfra.VPC,
		ALBSecurityGroup:  baseInfra.ALBSecurityGroup,
		ECSSecurityGroup:  baseInfra.ECSSecurityGroup,
		TaskExecutionRole: baseInfra.TaskExecutionRole,
		TaskRole:          baseInfra.TaskRole,
		LogGroup:          baseInfra.LogGroup,
		DatabaseSecret:    database.DatabaseSecret,
		DatabaseEndpoint:  *database.DatabaseInstance.InstanceEndpoint().Hostname(),
		Environment:       environment,
		AppName:           appName,
		Certificate:       nil, // 後で更新
		DomainName:        subdomainName,
		RailsMasterKey:    railsMasterKey,
	})

	// DNS と証明書の設定
	dns := constructs.NewDns(stack, "Dns", &constructs.DnsProps{
		LoadBalancer:     containerService.LoadBalancer,
		Environment:      environment,
		AppName:          appName,
		ParentDomainName: parentDomainName,
		SubdomainName:    subdomainName,
	})

	// Container ServiceにHTTPS Listenerを追加（証明書が作成された後）
	if dns.Certificate != nil {
		containerService.HttpsListener = containerService.LoadBalancer.AddListener(jsii.String("HttpsListener"), &awselasticloadbalancingv2.BaseApplicationListenerProps{
			Port:     jsii.Number(443),
			Protocol: awselasticloadbalancingv2.ApplicationProtocol_HTTPS,
			Certificates: &[]awselasticloadbalancingv2.IListenerCertificate{
				awselasticloadbalancingv2.ListenerCertificate_FromCertificateManager(dns.Certificate),
			},
			DefaultTargetGroups: &[]awselasticloadbalancingv2.IApplicationTargetGroup{
				containerService.TargetGroup,
			},
		})
	}

	// Stack Outputs
	awscdk.NewCfnOutput(stack, jsii.String("StackEnvironment"), &awscdk.CfnOutputProps{
		Value:       jsii.String(environment),
		Description: jsii.String("Deployment environment"),
		ExportName:  jsii.String(appName + "-" + environment + "-environment"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("AppName"), &awscdk.CfnOutputProps{
		Value:       jsii.String(appName),
		Description: jsii.String("Application name"),
		ExportName:  jsii.String(appName + "-" + environment + "-app-name"),
	})

	return &TodoAppStack{
		Stack:              stack,
		BaseInfrastructure: baseInfra,
		Database:           database,
		Dns:                dns,
		ContainerService:   containerService,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
