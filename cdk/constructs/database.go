package constructs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DatabaseProps struct {
	VPC              awsec2.IVpc
	SecurityGroup    awsec2.SecurityGroup
	Environment      string
	AppName          string
	DatabaseName     string
	DatabaseUsername string
}

type Database struct {
	Construct        constructs.Construct
	DatabaseInstance awsrds.DatabaseInstance
	DatabaseSecret   awssecretsmanager.Secret
	SubnetGroup      awsrds.SubnetGroup
}

/*
*

RDS MySQLを作成します。
*/
func NewDatabase(scope constructs.Construct, id string, props *DatabaseProps) *Database {
	construct := constructs.NewConstruct(scope, &id)
	this := &Database{
		Construct: construct,
	}

	// Database credentials secret
	this.DatabaseSecret = awssecretsmanager.NewSecret(this.Construct, jsii.String("DatabaseSecret"), &awssecretsmanager.SecretProps{
		Description: jsii.String("Database credentials for " + props.AppName),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			SecretStringTemplate:    jsii.String(`{"username":"` + props.DatabaseUsername + `"}`),
			GenerateStringKey:       jsii.String("password"),
			PasswordLength:          jsii.Number(16),
			ExcludeCharacters:       jsii.String("\"@/\\"),
			RequireEachIncludedType: jsii.Bool(true),
		},
	})

	// DB Subnet Group
	this.SubnetGroup = awsrds.NewSubnetGroup(this.Construct, jsii.String("DBSubnetGroup"), &awsrds.SubnetGroupProps{
		Description: jsii.String("Subnet group for RDS database"),
		Vpc:         props.VPC,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
	})

	// Database instance type based on environment
	var instanceType awsec2.InstanceType
	if props.Environment == "production" {
		instanceType = awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_SMALL)
	} else {
		instanceType = awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_MICRO)
	}

	// RDS MySQL Instance
	this.DatabaseInstance = awsrds.NewDatabaseInstance(this.Construct, jsii.String("Database"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Mysql(&awsrds.MySqlInstanceEngineProps{
			Version: awsrds.MysqlEngineVersion_VER_8_0(),
		}),
		InstanceType:            instanceType,
		Vpc:                     props.VPC,
		SecurityGroups:          &[]awsec2.ISecurityGroup{props.SecurityGroup},
		SubnetGroup:             this.SubnetGroup,
		Credentials:             awsrds.Credentials_FromSecret(this.DatabaseSecret, jsii.String(props.DatabaseUsername)),
		DatabaseName:            jsii.String(props.DatabaseName),
		AllocatedStorage:        jsii.Number(20),
		MaxAllocatedStorage:     jsii.Number(100),
		StorageType:             awsrds.StorageType_GP2,
		BackupRetention:         awscdk.Duration_Days(jsii.Number(0)), // バックアップを無効化
		MultiAz:                 jsii.Bool(false),
		StorageEncrypted:        jsii.Bool(true),
		MonitoringInterval:      awscdk.Duration_Seconds(jsii.Number(0)), // お節約ですわー！
		AutoMinorVersionUpgrade: jsii.Bool(true),
		RemovalPolicy:           awscdk.RemovalPolicy_DESTROY, // Change to RETAIN for production
	})

	// Output database endpoint
	awscdk.NewCfnOutput(this.Construct, jsii.String("DatabaseEndpoint"), &awscdk.CfnOutputProps{
		Value:       this.DatabaseInstance.InstanceEndpoint().Hostname(),
		Description: jsii.String("RDS instance endpoint"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-db-endpoint"),
	})

	// Output database port
	awscdk.NewCfnOutput(this.Construct, jsii.String("DatabasePort"), &awscdk.CfnOutputProps{
		Value:       jsii.String("3306"),
		Description: jsii.String("RDS instance port"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-db-port"),
	})

	return this
}
