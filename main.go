package main

import (
	"context" // For pretty printing the output
	"fmt"
	"log"
	"os" // To read environment variables
	"sort"
	"strconv" // To convert string env var to float
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// --- Configuration ---
var reportGranularity types.Granularity = types.GranularityMonthly
var reportMetrics = []string{"UnblendedCost"}
var reportGroupBy = []types.GroupDefinition{
	{Type: types.GroupDefinitionTypeDimension, Key: aws.String("SERVICE")},
}
var reportFilter *types.Expression // = nil
var awsRegion = "us-east-1"

// ---------------------

// getCostData remains the same as before
func getCostData(ctx context.Context, client *costexplorer.Client, startDate, endDate string, granularity types.Granularity, metrics []string, groupBy []types.GroupDefinition, filter *types.Expression) ([]types.ResultByTime, error) {
	var allResults []types.ResultByTime
	var nextPageToken *string

	for {
		input := &costexplorer.GetCostAndUsageInput{
			TimePeriod: &types.DateInterval{
				Start: aws.String(startDate),
				End:   aws.String(endDate),
			},
			Granularity:   granularity,
			Metrics:       metrics,
			NextPageToken: nextPageToken,
		}
		if groupBy != nil {
			input.GroupBy = groupBy
		}
		if filter != nil {
			input.Filter = filter
		}

		output, err := client.GetCostAndUsage(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to get cost and usage for %s to %s: %w", startDate, endDate, err)
		}
		if output.ResultsByTime != nil {
			allResults = append(allResults, output.ResultsByTime...)
		}
		if output.NextPageToken == nil {
			break
		}
		nextPageToken = output.NextPageToken
	}
	return allResults, nil
}

// processAndDisplayResults applies markup but doesn't mention it in output labels
func processAndDisplayResults(periodLabel string, results []types.ResultByTime, metrics []string, groupBy []types.GroupDefinition, markupMultiplier float64) {
	// No margin suffix needed in the output titles/labels anymore
	fmt.Printf("--- %s Results ---\n", periodLabel)
	if len(results) == 0 {
		fmt.Println("No results found for this period.")
		fmt.Println(string(make([]byte, 30, 30))) // Separator
		return
	}

	// Optional: Display raw API response for debugging - comment out if not needed
	// resultsJSON, err := json.MarshalIndent(results, "", "  ")
	// if err != nil {
	//  log.Printf("Failed to marshal results to JSON for %s: %v", periodLabel, err)
	// } else {
	//  fmt.Println("Raw API Response (JSON):")
	//  fmt.Println(string(resultsJSON))
	// }
	// fmt.Println(string(make([]byte, 30, 30))) // Separator

	costMetric := metrics[0] // Assuming the first metric is the cost one we want to mark up

	if len(groupBy) == 0 { // Not grouping
		var totalCost float64 // This will hold the marked-up cost

		for _, periodResult := range results {
			if metricValue, ok := periodResult.Total[costMetric]; ok && metricValue.Amount != nil {
				amount, err := strconv.ParseFloat(*metricValue.Amount, 64)
				if err == nil {
					totalCost += amount * markupMultiplier // Apply markup here
				} else {
					log.Printf("Warning: Could not parse amount '%s' for %s total.", *metricValue.Amount, periodLabel)
				}
			}
		}
		// Display only the final (marked-up) cost without mentioning margin or raw cost
		fmt.Printf("Total %s for %s: %.2f\n", costMetric, periodLabel, totalCost)

	} else if len(groupBy) > 0 && *groupBy[0].Key == "SERVICE" { // Grouping by SERVICE
		// Label doesn't mention margin
		fmt.Printf("Cost Summary by Service (%s):\n", periodLabel)
		serviceTotals := make(map[string]float64) // Holds marked-up costs per service

		for _, timePeriod := range results {
			for _, group := range timePeriod.Groups {
				if len(group.Keys) > 0 {
					serviceName := group.Keys[0] // Assumes SERVICE is the first key
					if metricValue, ok := group.Metrics[costMetric]; ok && metricValue.Amount != nil {
						amount, err := strconv.ParseFloat(*metricValue.Amount, 64)
						if err == nil {
							serviceTotals[serviceName] += amount * markupMultiplier // Apply markup here
						} else {
							log.Printf("Warning: Could not parse amount '%s' for service %s in %s.", *metricValue.Amount, serviceName, periodLabel)
						}
					}
				}
			}
		}

		// Sort map by value descending (using the marked-up value for sorting)
		type kv struct {
			Key   string
			Value float64
		}
		var sortedServices []kv
		for k, v := range serviceTotals {
			sortedServices = append(sortedServices, kv{k, v})
		}
		sort.Slice(sortedServices, func(i, j int) bool {
			return sortedServices[i].Value > sortedServices[j].Value
		})

		var overallTotal float64
		// Update table header - remove Raw Cost column header
		fmt.Printf("  %-50s %15s\n", "Service", costMetric)
		fmt.Printf("  %-50s %15s\n", "--------------------------------------------------", "---------------")
		for _, kvPair := range sortedServices {
			// Update table row - remove Raw Cost value column
			fmt.Printf("  %-50s %15.2f\n", kvPair.Key, kvPair.Value)
			overallTotal += kvPair.Value
		}
		fmt.Printf("  %-50s %15s\n", "--------------------------------------------------", "---------------")
		// Update table footer - remove Raw Cost value column
		fmt.Printf("  %-50s %15.2f\n", "Total (Sum of Services)", overallTotal)
	}
	fmt.Println(string(make([]byte, 30, 30))) // Separator
}

func main() {
	// --- Read and Parse AWS_MARGIN Environment Variable ---
	marginStr := os.Getenv("AWS_MARGIN")
	var marginPercent float64
	var err error

	if marginStr == "" {
		// Keep this log message to inform the user if no margin is applied
		log.Println("INFO: AWS_MARGIN environment variable not set. Applying 0% margin.")
		marginPercent = 0.0
	} else {
		marginPercent, err = strconv.ParseFloat(marginStr, 64)
		if err != nil {
			// log.Printf("ERROR: Parsing AWS_MARGIN value '%s': %v. Applying 0% margin.", marginStr, err)
			marginPercent = 0.0
		} else if marginPercent < 0 {
			// Keep this log message to inform the user about the applied discount
			log.Printf("INFO: Applying negative AWS_MARGIN (discount) of %.2f%%", marginPercent)
		} else {
			// Keep this log message to inform the user about the applied markup
			// log.Printf("INFO: Applying AWS_MARGIN of %.2f%%", marginPercent)
		}
	}
	markupMultiplier := 1.0 + (marginPercent / 100.0)
	// ----------------------------------------------------

	// --- Date Calculation ---
	// Using hardcoded date for consistency since current time was provided
	// now := time.Now()
	now := time.Date(2025, time.May, 1, 12, 0, 8, 0, time.FixedZone("EDT", -4*60*60)) // Thursday, May 1, 2025 12:00:08 PM EDT
	log.Printf("INFO: Current script time set to: %s", now.Format(time.RFC1123Z))

	currentYear, currentMonth, currentDay := now.Date()
	currentMonthStart := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, now.Location())
	currentMonthEndAPI := currentMonthStart.AddDate(0, 1, 0).Format("2006-01-02") // 2025-06-01
	currentMonthStartStr := currentMonthStart.Format("2006-01-02")                // 2025-05-01
	previousMonthStart := currentMonthStart.AddDate(0, -1, 0)                     // 2025-04-01
	previousMonthEndAPI := currentMonthStartStr                                   // 2025-05-01
	previousMonthStartStr := previousMonthStart.Format("2006-01-02")              // 2025-04-01
	// ---------------------

	// Load default AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("ERROR: unable to load SDK config, %v", err)
	}

	// Create Cost Explorer client
	client := costexplorer.NewFromConfig(cfg)

	// --- Fetch and Display Previous Month Data ---
	fmt.Printf("Fetching Previous Month's Data (%s to %s)...\n", previousMonthStartStr, previousMonthEndAPI)
	prevMonthResults, err := getCostData(context.TODO(), client, previousMonthStartStr, previousMonthEndAPI, reportGranularity, reportMetrics, reportGroupBy, reportFilter)
	if err != nil {
		log.Printf("ERROR: fetching previous month's data: %v", err)
	} else {
		// Pass multiplier, but not marginPercent as it's not needed for display logic anymore
		processAndDisplayResults(fmt.Sprintf("Previous Month (%s)", previousMonthStart.Format("Jan 2006")), prevMonthResults, reportMetrics, reportGroupBy, markupMultiplier)
	}

	// --- Fetch and Display Current Month Data (handle day 1) ---
	if currentDay == 1 {
		fmt.Println("--- Current Month Results ---")
		// This message remains as it's about data availability, not margin.
		fmt.Println("Today is the 1st of the month. Check back tomorrow to see this month's usage data.")
		fmt.Println(string(make([]byte, 30, 30))) // Separator
	} else {
		fmt.Printf("Fetching Current Month's Data (%s to %s)...\n", currentMonthStartStr, currentMonthEndAPI)
		currentMonthResults, err := getCostData(context.TODO(), client, currentMonthStartStr, currentMonthEndAPI, reportGranularity, reportMetrics, reportGroupBy, reportFilter)
		if err != nil {
			log.Printf("ERROR: fetching current month's data: %v", err)
		} else {
			// Pass multiplier, but not marginPercent
			processAndDisplayResults(fmt.Sprintf("Current Month MTD (%s)", currentMonthStart.Format("Jan 2006")), currentMonthResults, reportMetrics, reportGroupBy, markupMultiplier)
		}
	}
}
