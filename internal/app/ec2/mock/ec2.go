// Package mocks contains mock for EC2 client
package mocks

import (
	"context"

	"github.com/TeiNam/prometheus-rds-exporter/internal/app/ec2"
	aws_ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	aws_ec2_types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

//nolint:golint,gomnd
var InstanceT3Large = ec2.EC2InstanceMetrics{
	MaximumIops:       15700,
	MaximumThroughput: 347.5,
	Memory:            8,
	Vcpu:              2,
}

//nolint:golint,gomnd
var InstanceT3Small = ec2.EC2InstanceMetrics{
	MaximumIops:       11800,
	MaximumThroughput: 260.62,
	Memory:            2,
	Vcpu:              2,
}

type EC2Client struct{}

func (m EC2Client) DescribeInstanceTypes(ctx context.Context, input *aws_ec2.DescribeInstanceTypesInput, optFns ...func(*aws_ec2.Options)) (*aws_ec2.DescribeInstanceTypesOutput, error) {
	var instances []aws_ec2_types.InstanceTypeInfo

	for _, instanceType := range input.InstanceTypes {
		//nolint // Hide "missing cases in switch" alert because instanceType has many values. Mock with return empty result for unknown instances
		switch instanceType {
		case "t3.large":
			instances = append(instances, aws_ec2_types.InstanceTypeInfo{
				InstanceType: instanceType,
				VCpuInfo:     &aws_ec2_types.VCpuInfo{DefaultVCpus: &InstanceT3Large.Vcpu},
				MemoryInfo:   &aws_ec2_types.MemoryInfo{SizeInMiB: &InstanceT3Large.Memory},
				EbsInfo: &aws_ec2_types.EbsInfo{EbsOptimizedInfo: &aws_ec2_types.EbsOptimizedInfo{
					MaximumIops:             &InstanceT3Large.MaximumIops,
					MaximumThroughputInMBps: &InstanceT3Large.MaximumThroughput,
				}},
			})
		case "t3.small":
			instances = append(instances, aws_ec2_types.InstanceTypeInfo{
				InstanceType: instanceType,
				VCpuInfo:     &aws_ec2_types.VCpuInfo{DefaultVCpus: &InstanceT3Small.Vcpu},
				MemoryInfo:   &aws_ec2_types.MemoryInfo{SizeInMiB: &InstanceT3Small.Memory},
				EbsInfo: &aws_ec2_types.EbsInfo{EbsOptimizedInfo: &aws_ec2_types.EbsOptimizedInfo{
					MaximumIops:             &InstanceT3Small.MaximumIops,
					MaximumThroughputInMBps: &InstanceT3Small.MaximumThroughput,
				}},
			})
		}
	}

	return &aws_ec2.DescribeInstanceTypesOutput{InstanceTypes: instances}, nil
}
