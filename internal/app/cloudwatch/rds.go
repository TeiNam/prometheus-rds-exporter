// Package cloudwatch implements methods to retrieve AWS Cloudwatch information
package cloudwatch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_cloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	aws_cloudwath_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

const (
	MaxQueriesPerCloudwatchRequest int   = 500
	CloudwatchUsagePeriod          int32 = 5
	Minute                         int32 = 60
)

var errUnknownMetric = errors.New("unknown metric")

type CloudWatchMetrics struct {
	Instances map[string]*RdsMetrics
}

type RdsMetrics struct {
	CPUUtilization            *float64
	DBLoad                    *float64
	DBLoadCPU                 *float64
	DBLoadNonCPU              *float64
	DatabaseConnections       *float64
	FreeStorageSpace          *float64
	FreeableMemory            *float64
	MaximumUsedTransactionIDs *float64
	ReadIOPS                  *float64
	ReadThroughput            *float64
	ReplicaLag                *float64
	ReplicationSlotDiskUsage  *float64
	SwapUsage                 *float64
	TransactionLogsDiskUsage  *float64
	WriteIOPS                 *float64
	WriteThroughput           *float64
	BufferCacheHitRatio       *float64
	Deadlocks                 *float64
	Queries                   *float64
	EngineUptime              *float64
	SumBinaryLogSize          *float64
	NumBinaryLogFiles         *float64
	AuroraBinlogReplicaLag    *float64
	BinLogDiskUsage           *float64
}

func (m *RdsMetrics) Update(field string, value float64) error {
	switch field {
	case "DBLoad":
		m.DBLoad = &value
	case "DBLoadCPU":
		m.DBLoadCPU = &value
	case "DBLoadNonCPU":
		m.DBLoadNonCPU = &value
	case "CPUUtilization":
		m.CPUUtilization = &value
	case "DatabaseConnections":
		m.DatabaseConnections = &value
	case "FreeStorageSpace":
		m.FreeStorageSpace = &value
	case "FreeableMemory":
		m.FreeableMemory = &value
	case "SwapUsage":
		m.SwapUsage = &value
	case "WriteIOPS":
		m.WriteIOPS = &value
	case "ReadIOPS":
		m.ReadIOPS = &value
	case "ReplicaLag":
		m.ReplicaLag = &value
	case "ReplicationSlotDiskUsage":
		m.ReplicationSlotDiskUsage = &value
	case "MaximumUsedTransactionIDs":
		m.MaximumUsedTransactionIDs = &value
	case "ReadThroughput":
		m.ReadThroughput = &value
	case "WriteThroughput":
		m.WriteThroughput = &value
	case "TransactionLogsDiskUsage":
		m.TransactionLogsDiskUsage = &value
	case "BufferCacheHitRatio":
		m.BufferCacheHitRatio = &value
	case "Deadlocks":
		m.Deadlocks = &value
	case "Queries":
		m.Queries = &value
	case "EngineUptime":
		m.EngineUptime = &value
	case "SumBinaryLogSize":
		m.SumBinaryLogSize = &value
	case "NumBinaryLogFiles":
		m.NumBinaryLogFiles = &value
	case "AuroraBinlogReplicaLag":
		m.AuroraBinlogReplicaLag = &value
	case "BinLogDiskUsage":
		m.BinLogDiskUsage = &value
	default:
		return fmt.Errorf("can't process '%s' metrics: %w", field, errUnknownMetric)
	}

	return nil
}

// getCloudWatchMetricsName returns names of Cloudwatch metrics to collect
func getCloudWatchMetricsName() [24]string {
	return [24]string{
		"CPUUtilization",
		"DBLoad",
		"DBLoadCPU",
		"DBLoadNonCPU",
		"DatabaseConnections",
		"FreeStorageSpace",
		"FreeableMemory",
		"MaximumUsedTransactionIDs",
		"ReadIOPS",
		"ReadThroughput",
		"ReplicaLag",
		"ReplicationSlotDiskUsage",
		"SwapUsage",
		"TransactionLogsDiskUsage",
		"WriteIOPS",
		"WriteThroughput",
		"BufferCacheHitRatio",
		"Deadlocks",
		"Queries",
		"EngineUptime",
		"SumBinaryLogSize",
		"NumBinaryLogFiles",
		"AuroraBinlogReplicaLag",
		"BinLogDiskUsage",
	}
}

// generateCloudWatchQueryForInstance return the cloudwatch query for a specific instance's metric
func generateCloudWatchQueryForInstance(queryID *string, metricName string, dbIdentifier string) CloudWatchMetricRequest {
	query := &aws_cloudwath_types.MetricDataQuery{
		Id: queryID,
		MetricStat: &aws_cloudwath_types.MetricStat{
			Metric: &aws_cloudwath_types.Metric{
				Namespace:  aws.String("AWS/RDS"),
				MetricName: aws.String(metricName),
				Dimensions: []aws_cloudwath_types.Dimension{
					{
						Name:  aws.String("DBInstanceIdentifier"),
						Value: aws.String(dbIdentifier),
					},
				},
			},
			Stat:   aws.String("Average"), // Specify the statistic to retrieve
			Period: aws.Int32(Minute),     // Specify the period of the metric. Granularity - 1 minute
		},
	}

	return CloudWatchMetricRequest{
		Dbidentifier: dbIdentifier,
		MetricName:   metricName,
		Query:        *query,
	}
}

// generateCloudWatchQueriesForInstances returns all cloudwatch queries for specified instances
func generateCloudWatchQueriesForInstances(dbIdentifiers []string) map[string]CloudWatchMetricRequest {
	queries := make(map[string]CloudWatchMetricRequest)

	metrics := getCloudWatchMetricsName()

	for i, dbIdentifier := range dbIdentifiers {
		for _, metricName := range metrics {
			queryID := aws.String(fmt.Sprintf("%s_%d", strings.ToLower(metricName), i))

			query := generateCloudWatchQueryForInstance(queryID, metricName, dbIdentifier)

			queries[*queryID] = query
		}
	}

	return queries
}

func NewRDSFetcher(client CloudWatchClient, logger slog.Logger) *RdsFetcher {
	return &RdsFetcher{
		client: client,
		logger: &logger,
	}
}

type RdsFetcher struct {
	client     CloudWatchClient
	statistics Statistics
	logger     *slog.Logger
}

func (c *RdsFetcher) GetStatistics() *Statistics {
	return &c.statistics
}

func (c *RdsFetcher) updateMetricsWithCloudWatchQueriesResult(metrics map[string]*RdsMetrics, requests map[string]CloudWatchMetricRequest, startTime *time.Time, endTime *time.Time, chunk []string) error {
	params := &aws_cloudwatch.GetMetricDataInput{
		StartTime:         startTime,
		EndTime:           endTime,
		ScanBy:            "TimestampDescending",
		MetricDataQueries: []aws_cloudwath_types.MetricDataQuery{},
	}

	for _, key := range chunk {
		query := requests[key].Query
		params.MetricDataQueries = append(params.MetricDataQueries, query)
	}

	resp, err := c.client.GetMetricData(context.TODO(), params)
	if err != nil {
		return fmt.Errorf("error calling GetMetricData: %w", err)
	}

	for _, m := range resp.MetricDataResults {
		if m.Values == nil {
			c.logger.Warn("cloudwatch value is empty", "metric", *m.Label)

			continue
		}

		val := requests[*m.Id]

		_, instanceMetricExists := metrics[val.Dbidentifier]
		if !instanceMetricExists {
			metrics[val.Dbidentifier] = &RdsMetrics{}
		}

		if len(m.Values) > 0 {
			err := metrics[val.Dbidentifier].Update(val.MetricName, m.Values[0])
			if err != nil {
				return fmt.Errorf("failed to process metrics %s: %w", val.MetricName, err)
			}
		}
	}

	return nil
}

func (c *RdsFetcher) GetRDSInstanceMetrics(dbIdentifiers []string) (CloudWatchMetrics, error) {
	metrics := make(map[string]*RdsMetrics)

	cloudWatchQueries := generateCloudWatchQueriesForInstances(dbIdentifiers)
	startTime := aws.Time(time.Now().Add(-3 * time.Minute)) // Start time - 1 hour ago
	endTime := aws.Time(time.Now())                         // End time - now
	chunkSize := MaxQueriesPerCloudwatchRequest

	cloudWatchAPICalls := float64(0)
	chunk := make([]string, 0, chunkSize)

	for query := range cloudWatchQueries {
		chunk = append(chunk, query)

		if len(chunk) == chunkSize {
			err := c.updateMetricsWithCloudWatchQueriesResult(metrics, cloudWatchQueries, startTime, endTime, chunk)
			if err != nil {
				return CloudWatchMetrics{}, fmt.Errorf("can't fetch Cloudwatch metrics: %w", err)
			}

			chunk = nil
			cloudWatchAPICalls += 1
		}
	}

	// Process last, potentially incomplete batch
	if len(chunk) > 0 {
		err := c.updateMetricsWithCloudWatchQueriesResult(metrics, cloudWatchQueries, startTime, endTime, chunk)
		if err != nil {
			return CloudWatchMetrics{}, fmt.Errorf("can't fetch Cloudwatch metrics: %w", err)
		}

		c.statistics.CloudWatchAPICall++
	}

	return CloudWatchMetrics{
		Instances: metrics,
	}, nil
}
