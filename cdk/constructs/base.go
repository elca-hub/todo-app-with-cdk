package constructs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type BaseInfrastructureProps struct {
	Environment string
	AppName     string
}

type BaseInfrastructure struct {
	Construct         constructs.Construct
	VPC               awsec2.Vpc
	ALBSecurityGroup  awsec2.SecurityGroup
	ECSSecurityGroup  awsec2.SecurityGroup
	RDSSecurityGroup  awsec2.SecurityGroup
	TaskExecutionRole awsiam.Role
	TaskRole          awsiam.Role
	LogGroup          awslogs.LogGroup
}

func NewBaseInfrastructure(scope constructs.Construct, id string, props *BaseInfrastructureProps) *BaseInfrastructure {
	construct := constructs.NewConstruct(scope, &id)
	this := &BaseInfrastructure{
		Construct: construct,
	}

	// VPCの設定
	this.VPC = awsec2.NewVpc(this.Construct, jsii.String("VPC"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:                jsii.String("PublicSubnet"),
				SubnetType:          awsec2.SubnetType_PUBLIC,
				CidrMask:            jsii.Number(24),
				MapPublicIpOnLaunch: jsii.Bool(false), // Elastic IP料金節約ですわー！
			},
			{
				Name:       jsii.String("PrivateSubnet"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
		},
		NatGateways: jsii.Number(1), // コスト節約ですわー！
	})

	// Security Group for ALB
	this.ALBSecurityGroup = awsec2.NewSecurityGroup(this.Construct, jsii.String("ALBSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              this.VPC,
		Description:      jsii.String("Security group for Application Load Balancer"),
		AllowAllOutbound: jsii.Bool(true),
	})
	this.ALBSecurityGroup.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(80)),
		jsii.String("Allow HTTP traffic from anywhere"),
		jsii.Bool(false),
	)
	this.ALBSecurityGroup.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(443)),
		jsii.String("Allow HTTPS traffic from anywhere"),
		jsii.Bool(false),
	)

	// Security Group for ECS Tasks
	this.ECSSecurityGroup = awsec2.NewSecurityGroup(this.Construct, jsii.String("ECSSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              this.VPC,
		Description:      jsii.String("Security group for ECS tasks"),
		AllowAllOutbound: jsii.Bool(true),
	})
	this.ECSSecurityGroup.AddIngressRule(
		this.ALBSecurityGroup,
		awsec2.Port_Tcp(jsii.Number(3000)),
		jsii.String("Allow traffic from ALB"),
		jsii.Bool(false),
	)

	// Security Group for RDS
	this.RDSSecurityGroup = awsec2.NewSecurityGroup(this.Construct, jsii.String("RDSSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              this.VPC,
		Description:      jsii.String("Security group for RDS MySQL"),
		AllowAllOutbound: jsii.Bool(false),
	})
	this.RDSSecurityGroup.AddIngressRule(
		this.ECSSecurityGroup,
		awsec2.Port_Tcp(jsii.Number(3306)),
		jsii.String("Allow MySQL access from ECS"),
		jsii.Bool(false),
	)

	// ECS Task Execution Role
	this.TaskExecutionRole = awsiam.NewRole(this.Construct, jsii.String("TaskExecutionRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AmazonECSTaskExecutionRolePolicy")),
		},
	})

	// ECS Task Role
	this.TaskRole = awsiam.NewRole(this.Construct, jsii.String("TaskRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
	})

	// CloudWatch Log Group
	this.LogGroup = awslogs.NewLogGroup(this.Construct, jsii.String("LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/ecs/" + props.AppName + "-" + props.Environment),
		Retention:     awslogs.RetentionDays_ONE_WEEK,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	return this
}
