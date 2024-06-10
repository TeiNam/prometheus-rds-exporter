package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// getAWSConfiguration 함수는 리전 매개변수를 받아 해당 리전의 구성을 반환합니다.
func getAWSConfiguration(logger *slog.Logger, roleArn, sessionName, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region), // 지정된 리전을 사용하여 기본 구성을 로드합니다.
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("can't create AWS session: %w", err)
	}

	if roleArn != "" {
		logger.Debug("Assume role", "role", roleArn)

		client := sts.NewFromConfig(cfg)
		creds := stscreds.NewAssumeRoleProvider(client, roleArn, func(o *stscreds.AssumeRoleOptions) {
			o.RoleSessionName = sessionName
		})
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}

	// Try to automatically find the current AWS region via AWS EC2 IMDS metadata
	if cfg.Region == "" {
		logger.Debug("search AWS region using IMDS")

		client := imds.NewFromConfig(cfg)

		response, err := client.GetRegion(context.TODO(), &imds.GetRegionInput{})
		if err == nil {
			cfg.Region = response.Region
			logger.Info("found AWS region via IMDS", "region", cfg.Region)
		}
	}

	return cfg, nil
}

func getAWSSessionInformation(cfg aws.Config) (string, string, error) {
	client := sts.NewFromConfig(cfg)

	output, err := client.GetCallerIdentity(context.TODO(), nil)
	if err != nil {
		return "", "", fmt.Errorf("can't fetch information about current session: %w", err)
	}

	return *output.Account, cfg.Region, nil
}
