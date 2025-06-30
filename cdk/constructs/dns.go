package constructs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DnsProps struct {
	LoadBalancer     awselasticloadbalancingv2.ApplicationLoadBalancer
	Environment      string
	AppName          string
	ParentDomainName string // elca-web.com
	SubdomainName    string // todo-app.kit.elca-web.com
}

type Dns struct {
	Construct      constructs.Construct
	HostedZone     awsroute53.IHostedZone
	Certificate    awscertificatemanager.Certificate
	AliasRecord    awsroute53.ARecord
}

func NewDns(scope constructs.Construct, id string, props *DnsProps) *Dns {
	construct := constructs.NewConstruct(scope, &id)
	this := &Dns{
		Construct: construct,
	}

	// 既存の親ドメインのHosted Zoneを参照
	this.HostedZone = awsroute53.HostedZone_FromLookup(this.Construct, jsii.String("ParentHostedZone"), &awsroute53.HostedZoneProviderProps{
		DomainName: jsii.String(props.ParentDomainName),
	})

	// SSL証明書をACMで作成（us-east-1リージョンでCloudFront用も考慮）
	this.Certificate = awscertificatemanager.NewCertificate(this.Construct, jsii.String("Certificate"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String(props.SubdomainName),
		Validation: awscertificatemanager.CertificateValidation_FromDns(this.HostedZone),
		SubjectAlternativeNames: &[]*string{
			jsii.String("*." + props.SubdomainName), // ワイルドカード証明書も追加
		},
	})

	// A Record（Alias）を作成してALBにトラフィックを転送
	this.AliasRecord = awsroute53.NewARecord(this.Construct, jsii.String("AliasRecord"), &awsroute53.ARecordProps{
		Zone:       this.HostedZone,
		RecordName: jsii.String(props.SubdomainName),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewLoadBalancerTarget(props.LoadBalancer, nil)),
		Comment:    jsii.String("Alias record for " + props.AppName + " " + props.Environment + " environment"),
	})

	// 出力
	awscdk.NewCfnOutput(this.Construct, jsii.String("DomainName"), &awscdk.CfnOutputProps{
		Value:       jsii.String(props.SubdomainName),
		Description: jsii.String("Application domain name"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-domain"),
	})

	awscdk.NewCfnOutput(this.Construct, jsii.String("CertificateArn"), &awscdk.CfnOutputProps{
		Value:       this.Certificate.CertificateArn(),
		Description: jsii.String("SSL Certificate ARN"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-cert-arn"),
	})

	awscdk.NewCfnOutput(this.Construct, jsii.String("ApplicationURL"), &awscdk.CfnOutputProps{
		Value:       jsii.String("https://" + props.SubdomainName),
		Description: jsii.String("Application HTTPS URL"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-https-url"),
	})

	return this
}