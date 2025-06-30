package constructs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapplicationautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ContainerServiceProps struct {
	VPC               awsec2.IVpc
	ALBSecurityGroup  awsec2.SecurityGroup
	ECSSecurityGroup  awsec2.SecurityGroup
	TaskExecutionRole awsiam.Role
	TaskRole          awsiam.Role
	LogGroup          awslogs.LogGroup
	DatabaseSecret    awssecretsmanager.Secret
	DatabaseEndpoint  string
	RailsMasterKey    string
	Environment       string
	AppName           string
	Certificate       awscertificatemanager.ICertificate
	DomainName        string
}

type ContainerService struct {
	Construct      constructs.Construct
	ECRRepository  awsecr.Repository
	ECSCluster     awsecs.Cluster
	TaskDefinition awsecs.FargateTaskDefinition
	ECSService     awsecs.FargateService
	LoadBalancer   awselasticloadbalancingv2.ApplicationLoadBalancer
	TargetGroup    awselasticloadbalancingv2.ApplicationTargetGroup
	HttpListener   awselasticloadbalancingv2.ApplicationListener
	HttpsListener  awselasticloadbalancingv2.ApplicationListener
}

func NewContainerService(scope constructs.Construct, id string, props *ContainerServiceProps) *ContainerService {
	construct := constructs.NewConstruct(scope, &id)
	this := &ContainerService{
		Construct: construct,
	}

	// ECR Repository
	this.ECRRepository = awsecr.NewRepository(this.Construct, jsii.String("ECRRepository"), &awsecr.RepositoryProps{
		RepositoryName:  jsii.String(props.AppName + "-" + props.Environment),
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
		ImageScanOnPush: jsii.Bool(true),
		LifecycleRules: &[]*awsecr.LifecycleRule{
			{
				Description:   jsii.String("Keep last 10 images"),
				MaxImageCount: jsii.Number(10),
				RulePriority:  jsii.Number(1),
			},
		},
	})

	// ECS Cluster
	this.ECSCluster = awsecs.NewCluster(this.Construct, jsii.String("ECSCluster"), &awsecs.ClusterProps{
		Vpc:         props.VPC,
		ClusterName: jsii.String(props.AppName + "-" + props.Environment + "-cluster"),
	})

	// Fargate Task Definition
	this.TaskDefinition = awsecs.NewFargateTaskDefinition(this.Construct, jsii.String("TaskDefinition"), &awsecs.FargateTaskDefinitionProps{
		MemoryLimitMiB: jsii.Number(512), // 0.5GB - cost optimized
		Cpu:            jsii.Number(256), // 0.25 vCPU - cost optimized
		ExecutionRole:  props.TaskExecutionRole,
		TaskRole:       props.TaskRole,
		Family:         jsii.String(props.AppName + "-" + props.Environment),
	})

	// Container Definition
	container := this.TaskDefinition.AddContainer(jsii.String("AppContainer"), &awsecs.ContainerDefinitionOptions{
		Image: awsecs.ContainerImage_FromEcrRepository(this.ECRRepository, jsii.String("latest")),
		Logging: awsecs.LogDriver_AwsLogs(&awsecs.AwsLogDriverProps{
			LogGroup:     props.LogGroup,
			StreamPrefix: jsii.String("ecs"),
		}),
		Environment: &map[string]*string{
			"RAILS_ENV":        jsii.String("production"),
			"DB_HOST":          jsii.String(props.DatabaseEndpoint),
			"DB_PORT":          jsii.String("3306"),
			"DB_NAME":          jsii.String("app_production"),
			"RAILS_MASTER_KEY": jsii.String(props.RailsMasterKey),
		},
		Secrets: &map[string]awsecs.Secret{
			"APP_DATABASE_PASSWORD": awsecs.Secret_FromSecretsManager(props.DatabaseSecret, jsii.String("password")),
			"DB_USERNAME":           awsecs.Secret_FromSecretsManager(props.DatabaseSecret, jsii.String("username")),
		},
		HealthCheck: &awsecs.HealthCheck{
			Command:     jsii.Strings("CMD-SHELL", "curl -f http://localhost:3000/health || exit 1"),
			Interval:    awscdk.Duration_Seconds(jsii.Number(30)),
			Timeout:     awscdk.Duration_Seconds(jsii.Number(5)),
			Retries:     jsii.Number(2), // Set retry count to 2
			StartPeriod: awscdk.Duration_Seconds(jsii.Number(60)),
		},
		// Container restart policy
		Essential: jsii.Bool(true), // If this container stops, the task stops
	})

	container.AddPortMappings(&awsecs.PortMapping{
		ContainerPort: jsii.Number(3000),
		Protocol:      awsecs.Protocol_TCP,
	})

	// Application Load Balancer
	this.LoadBalancer = awselasticloadbalancingv2.NewApplicationLoadBalancer(this.Construct, jsii.String("ALB"), &awselasticloadbalancingv2.ApplicationLoadBalancerProps{
		Vpc:              props.VPC,
		InternetFacing:   jsii.Bool(true),
		SecurityGroup:    props.ALBSecurityGroup,
		LoadBalancerName: jsii.String(props.AppName + "-" + props.Environment + "-alb"),
	})

	// Target Group
	this.TargetGroup = awselasticloadbalancingv2.NewApplicationTargetGroup(this.Construct, jsii.String("TargetGroup"), &awselasticloadbalancingv2.ApplicationTargetGroupProps{
		Port:       jsii.Number(3000),
		Protocol:   awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		Vpc:        props.VPC,
		TargetType: awselasticloadbalancingv2.TargetType_IP,
		HealthCheck: &awselasticloadbalancingv2.HealthCheck{
			Path:                    jsii.String("/health"),
			Protocol:                awselasticloadbalancingv2.Protocol_HTTP,
			Port:                    jsii.String("3000"),
			HealthyThresholdCount:   jsii.Number(2),
			UnhealthyThresholdCount: jsii.Number(3),
			Timeout:                 awscdk.Duration_Seconds(jsii.Number(5)),
			Interval:                awscdk.Duration_Seconds(jsii.Number(30)),
		},
		TargetGroupName: jsii.String(props.AppName + "-" + props.Environment + "-tg"),
	})

	// HTTP Listener (HTTPからHTTPSへのリダイレクト)
	this.HttpListener = this.LoadBalancer.AddListener(jsii.String("HttpListener"), &awselasticloadbalancingv2.BaseApplicationListenerProps{
		Port:     jsii.Number(80),
		Protocol: awselasticloadbalancingv2.ApplicationProtocol_HTTP,
	})

	// リダイレクトアクションを追加
	this.HttpListener.AddAction(jsii.String("RedirectAction"), &awselasticloadbalancingv2.AddApplicationActionProps{
		Action: awselasticloadbalancingv2.ListenerAction_Redirect(&awselasticloadbalancingv2.RedirectOptions{
			Protocol:  jsii.String("HTTPS"),
			Port:      jsii.String("443"),
			Permanent: jsii.Bool(true),
		}),
	})

	// HTTPS Listener
	if props.Certificate != nil {
		this.HttpsListener = this.LoadBalancer.AddListener(jsii.String("HttpsListener"), &awselasticloadbalancingv2.BaseApplicationListenerProps{
			Port:     jsii.Number(443),
			Protocol: awselasticloadbalancingv2.ApplicationProtocol_HTTPS,
			Certificates: &[]awselasticloadbalancingv2.IListenerCertificate{
				awselasticloadbalancingv2.ListenerCertificate_FromCertificateManager(props.Certificate),
			},
			DefaultTargetGroups: &[]awselasticloadbalancingv2.IApplicationTargetGroup{
				this.TargetGroup,
			},
		})
	}

	// ECS Fargate Service
	this.ECSService = awsecs.NewFargateService(this.Construct, jsii.String("ECSService"), &awsecs.FargateServiceProps{
		Cluster:        this.ECSCluster,
		TaskDefinition: this.TaskDefinition,
		DesiredCount:   jsii.Number(2), // 分散システムなので
		SecurityGroups: &[]awsec2.ISecurityGroup{props.ECSSecurityGroup},
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
		ServiceName: jsii.String(props.AppName + "-" + props.Environment + "-service"),
		// Use Spot instances for cost optimization in non-production
		CapacityProviderStrategies: func() *[]*awsecs.CapacityProviderStrategy {
			if props.Environment != "production" {
				return &[]*awsecs.CapacityProviderStrategy{
					{
						CapacityProvider: jsii.String("FARGATE_SPOT"),
						Weight:           jsii.Number(1),
					},
				}
			}
			return &[]*awsecs.CapacityProviderStrategy{
				{
					CapacityProvider: jsii.String("FARGATE"),
					Weight:           jsii.Number(1),
				},
			}
		}(),
		EnableExecuteCommand: jsii.Bool(true), // For debugging
		AssignPublicIp:       jsii.Bool(false),
	})

	// Attach ECS Service to Target Group
	this.ECSService.AttachToApplicationTargetGroup(this.TargetGroup)

	// Auto Scaling (optional, for cost optimization)
	scaling := this.ECSService.AutoScaleTaskCount(&awsapplicationautoscaling.EnableScalingProps{
		MinCapacity: jsii.Number(1),
		MaxCapacity: jsii.Number(3),
	})

	// Scale based on CPU utilization
	scaling.ScaleOnCpuUtilization(jsii.String("CpuScaling"), &awsecs.CpuUtilizationScalingProps{
		TargetUtilizationPercent: jsii.Number(70),
		ScaleInCooldown:          awscdk.Duration_Minutes(jsii.Number(5)),
		ScaleOutCooldown:         awscdk.Duration_Minutes(jsii.Number(2)),
	})

	// Outputs
	awscdk.NewCfnOutput(this.Construct, jsii.String("ECRRepositoryURI"), &awscdk.CfnOutputProps{
		Value:       this.ECRRepository.RepositoryUri(),
		Description: jsii.String("ECR Repository URI"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-ecr-uri"),
	})

	awscdk.NewCfnOutput(this.Construct, jsii.String("LoadBalancerDNS"), &awscdk.CfnOutputProps{
		Value:       this.LoadBalancer.LoadBalancerDnsName(),
		Description: jsii.String("Application Load Balancer DNS name"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-alb-dns"),
	})

	// Application URL output (HTTPSまたはHTTP)
	if props.DomainName != "" {
		awscdk.NewCfnOutput(this.Construct, jsii.String("LoadBalancerURL"), &awscdk.CfnOutputProps{
			Value:       jsii.String("https://" + props.DomainName),
			Description: jsii.String("Application HTTPS URL"),
			ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-app-url"),
		})
	} else {
		awscdk.NewCfnOutput(this.Construct, jsii.String("LoadBalancerURL"), &awscdk.CfnOutputProps{
			Value:       jsii.String("http://" + *this.LoadBalancer.LoadBalancerDnsName()),
			Description: jsii.String("Application HTTP URL"),
			ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-app-url"),
		})
	}

	return this
}
