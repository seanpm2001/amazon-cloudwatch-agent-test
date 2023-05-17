package dimension

import (
	"github.com/aws/amazon-cloudwatch-agent-test/environment/computetype"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type EMFECSDimensionProvider struct {
	Provider
}

func (p EMFECSDimensionProvider) IsApplicable() bool {
	return p.env.ComputeType == computetype.ECS
}

func (p EMFECSDimensionProvider) GetDimension(instruction Instruction) types.Dimension {
	if instruction.Key == "InstanceID" {
		return types.Dimension{
			Name:  aws.String("InstanceID"),
			Value: aws.String("INSTANCEID"),
		}
	}
	return types.Dimension{}
}

func (p EMFECSDimensionProvider) Name() string {
	return "EMFECSProvider"
}

var _ IProvider = (*EMFECSDimensionProvider)(nil)
