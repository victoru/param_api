package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type ssmClient struct {
	client *ssm.SSM
}

func NewClient(region string) *ssm.SSM {
	session := session.Must(session.NewSession())
	if DebugMode {
		session.Config.WithRegion(region).WithLogLevel(aws.LogDebugWithHTTPBody) //.WithMaxRetries(2)
	} else {
		session.Config.WithRegion(region)
	}
	return ssm.New(session)
}

func (s ssmClient) ParamList(names ...string) (*ssm.GetParametersOutput, error) {
	//limit of 50 parameters, unless extra logic is added to paginate
	var ptrNames []*string
	for _, name := range names {
		ptrNames = append(ptrNames, aws.String(name))
	}
	params := &ssm.GetParametersInput{
		Names:          ptrNames,
		WithDecryption: aws.Bool(true),
	}
	return s.client.GetParameters(params)
}
