// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

//go:build !windows

package metric_value_benchmark

import (
	_ "embed"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/qri-io/jsonschema"

	"github.com/aws/amazon-cloudwatch-agent-test/environment"
	"github.com/aws/amazon-cloudwatch-agent-test/internal/awsservice"
	"github.com/aws/amazon-cloudwatch-agent-test/test/metric"
	"github.com/aws/amazon-cloudwatch-agent-test/test/metric/dimension"
	"github.com/aws/amazon-cloudwatch-agent-test/test/status"
	"github.com/aws/amazon-cloudwatch-agent-test/test/test_runner"
)

//go:embed agent_resources/container_insights_eks_telemetry.json
var eksContainerInsightsSchema string

type EKSDaemonTestRunner struct {
	test_runner.BaseTestRunner
}

var _ IEKSTestRunner = (*EKSDaemonTestRunner)(nil)

func (e *EKSDaemonTestRunner) validate(eks *environment.MetaData) status.TestGroupResult {
	metrics := e.getMeasuredMetrics()
	testResults := make([]status.TestResult, 0)
	for _, name := range metrics {
		testResults = append(testResults, e.validateInstanceMetrics(name))
	}

	testResults = append(testResults, e.validateLogs(eks))
	return status.TestGroupResult{
		Name:        e.getTestName(),
		TestResults: testResults,
	}
}

func (e *EKSDaemonTestRunner) validateInstanceMetrics(name string) status.TestResult {
	testResult := status.TestResult{
		Name:   name,
		Status: status.FAILED,
	}

	dims, failed := e.DimensionFactory.GetDimensions([]dimension.Instruction{
		{
			Key:   "ClusterName",
			Value: dimension.UnknownDimensionValue(),
		},
	})

	if len(failed) > 0 {
		log.Println("failed to get dimensions")
		return testResult
	}

	fetcher := metric.MetricValueFetcher{}
	values, err := fetcher.Fetch("ContainerInsights", name, dims, metric.AVERAGE, test_runner.HighResolutionStatPeriod)
	if err != nil {
		log.Println("failed to fetch metrics", err)
		return testResult
	}

	if !isAllValuesGreaterThanOrEqualToExpectedValue(name, values, 0) {
		return testResult
	}

	testResult.Status = status.SUCCESSFUL
	return testResult
}

func (e *EKSDaemonTestRunner) validateLogs(eks *environment.MetaData) status.TestResult {
	testResult := status.TestResult{
		Name:   "emf-logs",
		Status: status.FAILED,
	}

	rs := jsonschema.Must(eksContainerInsightsSchema)

	now := time.Now()
	group := fmt.Sprintf("/aws/containerinsights/%s/performance", eks.EKSClusterName)

	// need to get the instances used for the EKS cluster
	EKSInstances, err := awsservice.GetEKSInstances(eks.EKSClusterName)
	if err != nil {
		log.Println("failed to get EKS instances")
		return testResult
	}

	for _, instance := range EKSInstances {
		validateLogContents := func(s string) bool {
			return strings.Contains(s, fmt.Sprintf("\"NodeName\":\"%s\"", *instance.InstanceName))
		}

		var ok bool
		stream := *instance.InstanceName
		log.Println(fmt.Sprintf("EKS instance is: %s", stream))
		ok, err = awsservice.ValidateLogs(group, stream, nil, &now, func(logs []string) bool {
			if len(logs) < 1 {
				log.Println(fmt.Sprintf("failed to get logs for instance: %s", stream))
				return false
			}

			for _, l := range logs {
				if !awsservice.MatchEMFLogWithSchema(l, rs, validateLogContents) {
					log.Println("failed to match logs with schema")
					return false
				}
			}
			return true
		})

		if err != nil || !ok {
			return testResult
		}
	}

	testResult.Status = status.SUCCESSFUL
	return testResult
}

func (e *EKSDaemonTestRunner) getTestName() string {
	return "EKSContainerInstance"
}

func (e *EKSDaemonTestRunner) getAgentConfigFileName() string {
	return "" // TODO: maybe not needed?
}

func (e *EKSDaemonTestRunner) getAgentRunDuration() time.Duration {
	return time.Minute * 3
}

func (e *EKSDaemonTestRunner) getMeasuredMetrics() []string {
	return []string{
		"node_cpu_reserved_capacity",
		"node_cpu_utilization",
		"node_network_total_bytes",
		"node_filesystem_utilization",
		"node_number_of_running_pods",
		"node_number_of_running_containers",
		"node_memory_utilization",
		"node_memory_reserved_capacity",
	}
}
