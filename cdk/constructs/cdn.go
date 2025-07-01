package constructs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CDNInfrastructureProps struct {
	Environment string
	AppName     string
}

type CDNInfrastructure struct {
	Construct    constructs.Construct
	Bucket       awss3.Bucket
	Distribution awscloudfront.Distribution
}

func NewCDNInfrastructure(scope constructs.Construct, id string, props *BaseInfrastructureProps) *CDNInfrastructure {
	construct := constructs.NewConstruct(scope, &id)
	this := &CDNInfrastructure{
		Construct: construct,
	}

	/* bucketの作成 */
	this.Bucket = awss3.NewBucket(this.Construct, jsii.String("Bucket"), &awss3.BucketProps{
		BucketName:         jsii.String(props.AppName + "-" + props.Environment + "-bucket"),
		EventBridgeEnabled: jsii.Bool(true),
		PublicReadAccess:   jsii.Bool(true),
		RemovalPolicy:      awscdk.RemovalPolicy_DESTROY,
	})

	/* CloudFrontの作成 */
	this.Distribution = awscloudfront.NewDistribution(this.Construct, jsii.String("CloudFront"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.NewS3StaticWebsiteOrigin(this.Bucket, nil),
		},
		PriceClass: awscloudfront.PriceClass_PRICE_CLASS_200,
	})

	awscdk.NewCfnOutput(this.Construct, jsii.String("DistributionId"), &awscdk.CfnOutputProps{
		Value:       this.Distribution.DistributionId(),
		Description: jsii.String("CloudFront distribution ID"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-cloudfront-distribution-id"),
	})

	awscdk.NewCfnOutput(this.Construct, jsii.String("DistributionDomainName"), &awscdk.CfnOutputProps{
		Value:       this.Distribution.DomainName(),
		Description: jsii.String("CloudFront distribution domain name"),
		ExportName:  jsii.String(props.AppName + "-" + props.Environment + "-cloudfront-distribution-domain-name"),
	})

	return this
}
