# AWS Cost and Usage Reporter (Go)

A command-line tool written in Go to fetch AWS Cost and Usage data for a specified date range using the AWS Cost Explorer API.

## Description

This tool interacts with the AWS Cost Explorer service via the AWS SDK for Go v2. It allows users to retrieve cost and usage metrics, optionally filtered and grouped by various dimensions (like Service or Tags), for a defined time period and granularity (Daily, Monthly). The results, including handling for paginated responses from the API, are printed to the console in JSON format, along with optional summary calculations.

## Prerequisites

Before running this tool, ensure you have the following:

1.  **Go:** Version 1.15 or later installed. ([Download Go](https://golang.org/dl/))
2.  **AWS Account:** An active AWS account.
3.  **AWS Credentials:** Configured AWS credentials. The tool uses the default credential chain provided by the AWS SDK for Go v2, which looks for credentials in this order:
    * Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`).
    * Shared credential file (`~/.aws/credentials`).
    * AWS config file (`~/.aws/config`).
    * IAM role attached to an EC2 instance or ECS task (if running in AWS).
    You can configure credentials easily using the AWS CLI: `aws configure`.
4.  **AWS Cost Explorer Enabled:** Cost Explorer must be enabled in your AWS account. Data may take up to 24 hours to populate after initial enablement.
5.  **IAM Permissions:** The IAM user or role executing the script needs permissions for `ce:GetCostAndUsage`. Attach a policy similar to this:
    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "ce:GetCostAndUsage"
                ],
                "Resource": "*"
            }
        ]
    }
    ```

## Installation & Setup

1.  **Clone or Download:** Get the source code. If you clone a repository:
    ```bash
    git clone <your-repo-url>
    cd <your-repo-directory>
    ```
    Or, save the `main.go` file provided previously into a new directory.

2.  **Initialize Go Module (if needed):** If you just downloaded the file, initialize Go modules:
    ```bash
    go mod init <your-module-name> # e.g., [example.com/aws-cost-reporter](https://example.com/aws-cost-reporter)
    ```

3.  **Install Dependencies:** Tidy up the dependencies to ensure the required AWS SDK modules are downloaded:
    ```bash
    go mod tidy
    ```
    This will download `github.com/aws/aws-sdk-go-v2/config`, `github.com/aws/aws-sdk-go-v2/service/costexplorer`, etc., based on the imports in `main.go`.

## Configuration

Modify the `// --- Configuration ---` section within the `main.go` file to customize the report parameters:

* **Date Range:**
    * Set specific `startDate` and `endDate` strings (`YYYY-MM-DD`). Remember `endDate` is *exclusive*.
    * Or use the provided logic to calculate the range for the last N days.
* **Granularity:** Change `granularity` to `types.GranularityDaily`, `types.GranularityMonthly`, or `types.GranularityHourly`. (Note: `HOURLY` requires a start date within the last 14 days).
* **Metrics:** Modify the `metrics` slice to include desired metrics like `"UnblendedCost"`, `"AmortizedCost"`, `"UsageQuantity"`, etc.
* **Grouping (`groupBy`):**
    * Set to `nil` for no grouping.
    * Add `types.GroupDefinition` structs to the slice to group by `DIMENSION` (like `SERVICE`, `REGION`) or `TAG`. Update the `Key` accordingly.
* **Filtering (`filter`):**
    * Set to `nil` for no filtering.
    * Create a `types.Expression` struct to filter by `Dimensions` or `Tags`. Uncomment and modify the examples provided in the code.
* **AWS Region (`awsRegion`):** Set the AWS region for the Cost Explorer client (usually `us-east-1` is appropriate for billing/CE, but can be changed if needed).

## Usage

Navigate to the project directory in your terminal and run the script using:

```bash
go run main.go
```

## Output
The script will:

1. Print the parameters being used for the API call (Date Range, Granularity, Metrics, Grouping/Filtering status).
2. Print the fetched cost and usage data results as a JSON array. Each element in the array typically represents a time period (based on granularity) and contains the costs/usage, potentially broken down by the specified groups.
3. If grouping is disabled or set to group by SERVICE, it will print a summary of the total cost for the period or a breakdown by service, respectively.

## Code Structure
- Configuration: Variables are defined at the top of main().
- AWS SDK Setup: Loads configuration and creates the Cost Explorer client.
- Pagination Loop: Iteratively calls GetCostAndUsage, handling the NextPageToken to retrieve all data pages.
- API Call: Constructs the input struct and calls client.GetCostAndUsage.
- Result Processing: Appends results from each page and prints the final aggregated data, including optional summary calculations.
- Error Handling: Uses standard Go error checking (if err != nil) and log.Fatalf for critical errors.