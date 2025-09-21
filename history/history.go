package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"api-security-scanner/types"
	"api-security-scanner/logging"
)

// HistoricalData represents historical comparison configuration
type HistoricalData struct {
	Enabled           bool   `yaml:"enabled"`
	StoragePath       string `yaml:"storage_path"`
	RetentionDays     int    `yaml:"retention_days"`
	ComparePrevious   bool   `yaml:"compare_previous"`
	TrendAnalysis     bool   `yaml:"trend_analysis"`
}

// ScanResult represents a stored scan result
type ScanResult struct {
	Timestamp   time.Time         `json:"timestamp"`
	Results     []types.EndpointResult `json:"results"`
	Summary     ScanSummary       `json:"summary"`
}

// ScanSummary represents a summary of scan results
type ScanSummary struct {
	TotalEndpoints    int     `json:"total_endpoints"`
	AverageScore     float64 `json:"average_score"`
	CriticalVulns    int     `json:"critical_vulnerabilities"`
	HighVulns        int     `json:"high_vulnerabilities"`
	MediumVulns      int     `json:"medium_vulnerabilities"`
	LowVulns         int     `json:"low_vulnerabilities"`
}

// TrendData represents trending information
type TrendData struct {
	Period       string    `json:"period"`
	ScoreTrend   []float64 `json:"score_trend"`
	VulnTrend    []int     `json:"vulnerability_trend"`
	Timestamps   []string  `json:"timestamps"`
}

// HistoryManager manages historical scan data
type HistoryManager struct {
	config HistoricalData
	storageDir string
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(config HistoricalData) (*HistoryManager, error) {
	if config.StoragePath == "" {
		config.StoragePath = "./history"
	}
	if config.RetentionDays <= 0 {
		config.RetentionDays = 30
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history storage directory: %v", err)
	}

	return &HistoryManager{
		config:     config,
		storageDir: config.StoragePath,
	}, nil
}

// SaveScanResults saves scan results to history
func (h *HistoryManager) SaveScanResults(results []types.EndpointResult) error {
	if !h.config.Enabled {
		return nil
	}

	scanResult := ScanResult{
		Timestamp: time.Now(),
		Results:   results,
		Summary:   h.generateSummary(results),
	}

	// Save to file
	filename := h.generateFilename(scanResult.Timestamp)
	filePath := filepath.Join(h.storageDir, filename)

	data, err := json.MarshalIndent(scanResult, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scan results: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save scan results: %v", err)
	}

	// Clean up old files
	if err := h.cleanupOldFiles(); err != nil {
		logging.Warn("Failed to cleanup old history files", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logging.Info("Scan results saved to history", map[string]interface{}{
		"file_path": filePath,
		"timestamp": scanResult.Timestamp,
		"summary": scanResult.Summary,
	})

	return nil
}

// LoadPreviousResults loads previous scan results for comparison
func (h *HistoryManager) LoadPreviousResults() (*ScanResult, error) {
	if !h.config.Enabled || !h.config.ComparePrevious {
		return nil, nil
	}

	// Get the most recent scan result file
	files, err := h.getHistoryFiles()
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("no previous scan results found")
	}

	// Skip the most recent (current) and load the previous one
	if len(files) < 2 {
		return nil, fmt.Errorf("only one scan result found, no previous result to compare")
	}

	previousFile := files[1] // Second most recent
	data, err := os.ReadFile(filepath.Join(h.storageDir, previousFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read previous scan result: %v", err)
	}

	var result ScanResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal previous scan result: %v", err)
	}

	return &result, nil
}

// GenerateTrendAnalysis generates trend analysis data
func (h *HistoryManager) GenerateTrendAnalysis() (*TrendData, error) {
	if !h.config.Enabled || !h.config.TrendAnalysis {
		return nil, nil
	}

	files, err := h.getHistoryFiles()
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("no historical data available for trend analysis")
	}

	// Load last 10 scans for trend analysis
	limit := 10
	if len(files) < limit {
		limit = len(files)
	}

	var scores []float64
	var vulns []int
	var timestamps []string

	for i := 0; i < limit; i++ {
		data, err := os.ReadFile(filepath.Join(h.storageDir, files[i]))
		if err != nil {
			continue
		}

		var result ScanResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		scores = append(scores, result.Summary.AverageScore)
		vulns = append(vulns, result.Summary.CriticalVulns+result.Summary.HighVulns)
		timestamps = append(timestamps, result.Timestamp.Format("2006-01-02"))
	}

	return &TrendData{
		Period:     fmt.Sprintf("Last %d scans", len(scores)),
		ScoreTrend: scores,
		VulnTrend:  vulns,
		Timestamps: timestamps,
	}, nil
}

// CompareWithPrevious compares current results with previous results
func (h *HistoryManager) CompareWithPrevious(current []types.EndpointResult) (*ComparisonResult, error) {
	previous, err := h.LoadPreviousResults()
	if err != nil {
		return nil, err
	}

	if previous == nil {
		return nil, nil
	}

	currentSummary := h.generateSummary(current)
	return &ComparisonResult{
		PreviousScan:     previous.Timestamp,
		CurrentScan:      time.Now(),
		PreviousSummary:  previous.Summary,
		CurrentSummary:   currentSummary,
		ScoreChange:      currentSummary.AverageScore - previous.Summary.AverageScore,
		VulnChange:      (currentSummary.CriticalVulns + currentSummary.HighVulns) -
						(previous.Summary.CriticalVulns + previous.Summary.HighVulns),
		EndpointChanges:  h.compareEndpoints(previous.Results, current),
	}, nil
}

// ComparisonResult represents the comparison between two scans
type ComparisonResult struct {
	PreviousScan    time.Time       `json:"previous_scan"`
	CurrentScan     time.Time       `json:"current_scan"`
	PreviousSummary ScanSummary     `json:"previous_summary"`
	CurrentSummary  ScanSummary     `json:"current_summary"`
	ScoreChange     float64         `json:"score_change"`
	VulnChange      int             `json:"vulnerability_change"`
	EndpointChanges []EndpointChange `json:"endpoint_changes"`
}

// EndpointChange represents changes for a specific endpoint
type EndpointChange struct {
	URL           string  `json:"url"`
	PreviousScore int     `json:"previous_score"`
	CurrentScore  int     `json:"current_score"`
	ScoreChange   int     `json:"score_change"`
	NewVulns      []string `json:"new_vulnerabilities"`
	ResolvedVulns []string `json:"resolved_vulnerabilities"`
}

// Helper functions

func (h *HistoryManager) generateFilename(timestamp time.Time) string {
	return fmt.Sprintf("scan_%s.json", timestamp.Format("20060102_150405"))
}

func (h *HistoryManager) generateSummary(results []types.EndpointResult) ScanSummary {
	if len(results) == 0 {
		return ScanSummary{}
	}

	totalScore := 0
	criticalVulns := 0
	highVulns := 0
	mediumVulns := 0
	lowVulns := 0

	for _, result := range results {
		totalScore += result.Score

		for _, test := range result.Results {
			if !test.Passed {
				switch test.TestName {
				case "Injection Test", "NoSQL Injection Test":
					criticalVulns++
				case "XSS Test", "Auth Bypass Test":
					highVulns++
				case "Parameter Tampering Test":
					mediumVulns++
				case "Header Security Test":
					lowVulns++
				}
			}
		}
	}

	return ScanSummary{
		TotalEndpoints:   len(results),
		AverageScore:     float64(totalScore) / float64(len(results)),
		CriticalVulns:    criticalVulns,
		HighVulns:        highVulns,
		MediumVulns:      mediumVulns,
		LowVulns:         lowVulns,
	}
}

func (h *HistoryManager) getHistoryFiles() ([]string, error) {
	files, err := os.ReadDir(h.storageDir)
	if err != nil {
		return nil, err
	}

	var scanFiles []string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[:5] == "scan_" {
			scanFiles = append(scanFiles, file.Name())
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(scanFiles, func(i, j int) bool {
		infoI, _ := os.Stat(filepath.Join(h.storageDir, scanFiles[i]))
		infoJ, _ := os.Stat(filepath.Join(h.storageDir, scanFiles[j]))
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return scanFiles, nil
}

func (h *HistoryManager) cleanupOldFiles() error {
	if h.config.RetentionDays <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -h.config.RetentionDays)
	files, err := h.getHistoryFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := filepath.Join(h.storageDir, file)
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filePath); err != nil {
				logging.Warn("Failed to remove old history file", map[string]interface{}{
					"file": file,
					"error": err.Error(),
				})
			} else {
				logging.Debug("Removed old history file", map[string]interface{}{
					"file": file,
				})
			}
		}
	}

	return nil
}

func (h *HistoryManager) compareEndpoints(previous, current []types.EndpointResult) []EndpointChange {
	// Create a map of previous results by URL
	previousMap := make(map[string]types.EndpointResult)
	for _, result := range previous {
		previousMap[result.URL] = result
	}

	var changes []EndpointChange

	for _, currResult := range current {
		prevResult, exists := previousMap[currResult.URL]

		change := EndpointChange{
			URL:          currResult.URL,
			CurrentScore: currResult.Score,
		}

		if exists {
			change.PreviousScore = prevResult.Score
			change.ScoreChange = currResult.Score - prevResult.Score
			change.NewVulns = h.findNewVulnerabilities(prevResult.Results, currResult.Results)
			change.ResolvedVulns = h.findResolvedVulnerabilities(prevResult.Results, currResult.Results)
		} else {
			// New endpoint
			change.PreviousScore = 0
			change.ScoreChange = currResult.Score
			change.NewVulns = h.getFailedTests(currResult.Results)
		}

		changes = append(changes, change)
	}

	// Check for removed endpoints
	for _, prevResult := range previous {
		found := false
		for _, currResult := range current {
			if currResult.URL == prevResult.URL {
				found = true
				break
			}
		}

		if !found {
			changes = append(changes, EndpointChange{
				URL:           prevResult.URL,
				PreviousScore: prevResult.Score,
				CurrentScore:  0,
				ScoreChange:   -prevResult.Score,
				ResolvedVulns: h.getFailedTests(prevResult.Results),
			})
		}
	}

	return changes
}

func (h *HistoryManager) findNewVulnerabilities(previous, current []types.TestResult) []string {
	prevFailed := h.getFailedTestsMap(previous)
	currFailed := h.getFailedTestsMap(current)

	var newVulns []string
	for test := range currFailed {
		if !prevFailed[test] {
			newVulns = append(newVulns, test)
		}
	}

	return newVulns
}

func (h *HistoryManager) findResolvedVulnerabilities(previous, current []types.TestResult) []string {
	prevFailed := h.getFailedTestsMap(previous)
	currFailed := h.getFailedTestsMap(current)

	var resolvedVulns []string
	for test := range prevFailed {
		if !currFailed[test] {
			resolvedVulns = append(resolvedVulns, test)
		}
	}

	return resolvedVulns
}

func (h *HistoryManager) getFailedTests(results []types.TestResult) []string {
	var failed []string
	for _, result := range results {
		if !result.Passed {
			failed = append(failed, result.TestName)
		}
	}
	return failed
}

func (h *HistoryManager) getFailedTestsMap(results []types.TestResult) map[string]bool {
	failed := make(map[string]bool)
	for _, result := range results {
		if !result.Passed {
			failed[result.TestName] = true
		}
	}
	return failed
}

// GenerateHistoricalComparisonJSON generates JSON formatted historical comparison
func GenerateHistoricalComparisonJSON(comparison *ComparisonResult) {
	if comparison == nil {
		fmt.Println("{}")
		return
	}

	data, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		fmt.Printf("{\"error\": \"Failed to generate comparison: %v\"}\n", err)
		return
	}

	fmt.Println(string(data))
}

// GenerateHistoricalComparisonHTML generates HTML formatted historical comparison
func GenerateHistoricalComparisonHTML(comparison *ComparisonResult) {
	if comparison == nil {
		fmt.Println("<html><body><p>No comparison data available</p></body></html>")
		return
	}

	fmt.Printf(`
<html>
<head>
    <title>Historical Comparison Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .change { margin: 10px 0; padding: 10px; border-radius: 3px; }
        .positive { background-color: #d4edda; color: #155724; }
        .negative { background-color: #f8d7da; color: #721c24; }
        .neutral { background-color: #fff3cd; color: #856404; }
        table { width: 100%%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Historical Comparison Report</h1>
        <p>Previous Scan: %s</p>
        <p>Current Scan: %s</p>
    </div>

    <div class="summary">
        <h2>Summary</h2>
        <p>Score Change: <span class="%s">%.2f</span></p>
        <p>Vulnerability Change: <span class="%s">%+d</span></p>
    </div>

    <div class="endpoint-changes">
        <h2>Endpoint Changes</h2>
        <table>
            <tr>
                <th>Endpoint</th>
                <th>Previous Score</th>
                <th>Current Score</th>
                <th>Change</th>
                <th>New Vulnerabilities</th>
                <th>Resolved Vulnerabilities</th>
            </tr>`,
		comparison.PreviousScan.Format("2006-01-02 15:04:05"),
		comparison.CurrentScan.Format("2006-01-02 15:04:05"),
		getChangeClass(comparison.ScoreChange),
		comparison.ScoreChange,
		getChangeClass(float64(-comparison.VulnChange)),
		comparison.VulnChange,
	)

	for _, change := range comparison.EndpointChanges {
		newVulns := strings.Join(change.NewVulns, ", ")
		resolvedVulns := strings.Join(change.ResolvedVulns, ", ")
		if newVulns == "" {
			newVulns = "None"
		}
		if resolvedVulns == "" {
			resolvedVulns = "None"
		}

		fmt.Printf(`
            <tr>
                <td>%s</td>
                <td>%d</td>
                <td>%d</td>
                <td class="%s">%+d</td>
                <td>%s</td>
                <td>%s</td>
            </tr>`,
			change.URL,
			change.PreviousScore,
			change.CurrentScore,
			getChangeClass(float64(change.ScoreChange)),
			change.ScoreChange,
			newVulns,
			resolvedVulns,
		)
	}

	fmt.Println(`
        </table>
    </div>
</body>
</html>`)
}

// GenerateHistoricalComparisonText generates text formatted historical comparison
func GenerateHistoricalComparisonText(comparison *ComparisonResult) {
	if comparison == nil {
		fmt.Println("No comparison data available")
		return
	}

	fmt.Printf(`
Historical Comparison Report
===========================

Previous Scan: %s
Current Scan:  %s

Summary:
- Score Change:     %+.2f
- Vulnerability Change: %+d

Endpoint Changes:
`,
		comparison.PreviousScan.Format("2006-01-02 15:04:05"),
		comparison.CurrentScan.Format("2006-01-02 15:04:05"),
		comparison.ScoreChange,
		comparison.VulnChange,
	)

	for _, change := range comparison.EndpointChanges {
		fmt.Printf("\nEndpoint: %s\n", change.URL)
		fmt.Printf("  Previous Score: %d\n", change.PreviousScore)
		fmt.Printf("  Current Score:  %d\n", change.CurrentScore)
		fmt.Printf("  Score Change:   %+d\n", change.ScoreChange)

		if len(change.NewVulns) > 0 {
			fmt.Printf("  New Vulnerabilities: %s\n", strings.Join(change.NewVulns, ", "))
		}

		if len(change.ResolvedVulns) > 0 {
			fmt.Printf("  Resolved Vulnerabilities: %s\n", strings.Join(change.ResolvedVulns, ", "))
		}
	}
}

// GenerateTrendAnalysisJSON generates JSON formatted trend analysis
func GenerateTrendAnalysisJSON(trendData *TrendData) {
	if trendData == nil {
		fmt.Println("{}")
		return
	}

	data, err := json.MarshalIndent(trendData, "", "  ")
	if err != nil {
		fmt.Printf("{\"error\": \"Failed to generate trend analysis: %v\"}\n", err)
		return
	}

	fmt.Println(string(data))
}

// GenerateTrendAnalysisHTML generates HTML formatted trend analysis
func GenerateTrendAnalysisHTML(trendData *TrendData) {
	if trendData == nil {
		fmt.Println("<html><body><p>No trend data available</p></body></html>")
		return
	}

	fmt.Printf(`
<html>
<head>
    <title>Trend Analysis Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .trend-container { margin: 20px 0; }
        .trend-item { margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 3px; }
        table { width: 100%%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Trend Analysis Report</h1>
        <p>Period: %s</p>
    </div>

    <div class="trend-container">
        <h2>Security Score Trend</h2>
        <table>
            <tr>
                <th>Date</th>
                <th>Score</th>
                <th>Critical + High Vulnerabilities</th>
            </tr>`,
		trendData.Period,
	)

	for i := 0; i < len(trendData.Timestamps); i++ {
		vulnCount := 0
		if i < len(trendData.VulnTrend) {
			vulnCount = trendData.VulnTrend[i]
		}

		score := 0.0
		if i < len(trendData.ScoreTrend) {
			score = trendData.ScoreTrend[i]
		}

		fmt.Printf(`
            <tr>
                <td>%s</td>
                <td>%.2f</td>
                <td>%d</td>
            </tr>`,
			trendData.Timestamps[i],
			score,
			vulnCount,
		)
	}

	fmt.Println(`
        </table>
    </div>
</body>
</html>`)
}

// GenerateTrendAnalysisText generates text formatted trend analysis
func GenerateTrendAnalysisText(trendData *TrendData) {
	if trendData == nil {
		fmt.Println("No trend data available")
		return
	}

	fmt.Printf(`
Trend Analysis Report
====================

Period: %s

Security Score Trend:
`,
		trendData.Period,
	)

	for i := 0; i < len(trendData.Timestamps); i++ {
		score := 0.0
		if i < len(trendData.ScoreTrend) {
			score = trendData.ScoreTrend[i]
		}

		vulnCount := 0
		if i < len(trendData.VulnTrend) {
			vulnCount = trendData.VulnTrend[i]
		}

		fmt.Printf("  %s: Score=%.2f, Critical+High Vulns=%d\n", trendData.Timestamps[i], score, vulnCount)
	}
}

// Helper function to determine CSS class for change indicators
func getChangeClass(change float64) string {
	if change > 0 {
		return "positive"
	} else if change < 0 {
		return "negative"
	}
	return "neutral"
}