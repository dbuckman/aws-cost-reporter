# AWS Cost and Usage Reporter (Go)

A command-line tool written in Go to fetch and display AWS costs for the current and previous months, applying a configurable cost adjustment, using the AWS Cost Explorer API.

## Description

This tool interacts with the AWS Cost Explorer service via the AWS SDK for Go v2. It performs the following actions:

1.  Fetches cost and usage data for the **full previous month**.
2.  Fetches cost and usage data for the **current month-to-date**.
    * *Note:* If run on the 1st day of the month, it will indicate that current month data is not yet available from AWS.
3.  Reads an optional `AWS_MARGIN` environment variable to apply a percentage-based cost adjustment (markup or discount) to the fetched costs. The percentage being applied is logged when the script starts.
4.  Displays the **final adjusted costs** for each period, optionally grouped by dimensions like Service. The output *does not* explicitly label the costs as marked up or show the original raw costs from AWS in the final report sections.
5.  Handles pagination automatically when retrieving data from the Cost Explorer API.

## Prerequisites

Before running this tool, ensure you have the following:

1.  **Go:** Version 1.15 or later installed. ([Download Go](https://golang.org/dl/))
2.  **AWS Account:** An active AWS account.
3.  **AWS Credentials:** Configured AWS credentials. The tool uses the default credential chain provided by the AWS SDK for Go v2 (environment variables, shared credential file, IAM role, etc.). You can configure credentials using the AWS CLI: `aws configure`.
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

1.  **Clone or Download:** Get the source code (`main.go`).
2.  **Navigate to Directory:** Open your terminal in the directory containing `main.go`.
3.  **Initialize Go Module (if needed):** If this is a new project directory:
    ```bash
    go mod init <your-module-name> # e.g., [example.com/aws-cost-reporter](https://example.com/aws-cost-reporter)
    ```
4.  **Install Dependencies:** Tidy up the dependencies to download the required AWS SDK modules:
    ```bash
    go mod tidy
    ```

## Configuration

There are two ways to configure the script:

1.  **Cost Adjustment (Environment Variable):**
    * Set the `AWS_MARGIN` environment variable to the percentage adjustment you want to apply.
    * Example: `export AWS_MARGIN="15"` for a 15% markup.
    * Example: `export AWS_MARGIN="-5"` for a 5% discount.
    * If the variable is not set or contains an invalid number, a 0% adjustment (no change) will be applied by default.
    * The script will log the percentage it detects and applies when it starts running.

2.  **Script Variables (Inside `main.go`):**
    * You can modify the global variables near the top of the `main.go` file:
        * `reportGranularity`: Change between `types.GranularityMonthly` (default) and `types.GranularityDaily`.
        * `reportMetrics`: Modify the list of metrics (e.g., `{"UnblendedCost"}`). The first metric is used for cost calculations.
        * `reportGroupBy`: Set to `nil` or an empty slice `[]types.GroupDefinition{}` for no grouping, or configure `types.GroupDefinition` structs to group by `DIMENSION` (like `SERVICE`) or `TAG`.
        * `reportFilter`: Set to `nil` for no filtering, or define a `types.Expression` to filter results.
        * `awsRegion`: Change the target AWS region if needed (usually `us-east-1` for Cost Explorer).

## Usage

1.  **Set Environment Variable (Optional):** Set the `AWS_MARGIN` if you want to apply a cost adjustment.
    * *Linux/macOS:* `export AWS_MARGIN="10.5"`
    * *Windows CMD:* `set AWS_MARGIN=10.5`
    * *Windows PowerShell:* `$env:AWS_MARGIN = "10.5"`
2.  **Run the Script:** Execute the script from its directory:
    ```bash
    go run main.go
    ```

## Output

The script will produce output similar to this:

1.  **Log Messages:** Initial lines indicating the `AWS_MARGIN` percentage being applied (if any).
2.  **Previous Month Section:**
    * A header like `--- Previous Month (Apr 2025) Results ---`.
    * Cost summary (either total or by service, depending on `reportGroupBy`). The costs shown include the adjustment from `AWS_MARGIN`.
3.  **Current Month Section:**
    * A header like `--- Current Month MTD (May 2025) Results ---`.
    * If run on the 1st of the month, a message indicating data is not yet available.
    * Otherwise, a cost summary similar to the previous month's, showing month-to-date costs including the adjustment.

**Note:** The final cost figures displayed in the "Results" sections *do not* explicitly mention the margin percentage applied or show the original AWS costs. They represent the final calculated cost after the adjustment specified by `AWS_MARGIN`. The initial log message serves as the notification of the adjustment percentage being used.

## Code Structure

* **Configuration:** Global variables and environment variable parsing in `main()`.
* **Date Calculation:** Determines start/end dates for current/previous months in `main()`.
* **`getCostData` Function:** Handles fetching data from AWS Cost Explorer, including pagination.
* **`processAndDisplayResults` Function:** Takes fetched data and the calculated markup multiplier, applies the adjustment to costs, and formats the output for display (totals or grouped by service).
* **Error Handling:** Uses standard Go error checking and `log` package messages.