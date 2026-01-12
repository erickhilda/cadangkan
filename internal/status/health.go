package status

import (
	"math"
	"time"

	"github.com/erickhilda/cadangkan/internal/backup"
)

const (
	// HealthScoreHealthy is the minimum score for healthy status
	HealthScoreHealthy = 80.0
	// HealthScoreWarning is the minimum score for warning status
	HealthScoreWarning = 50.0
	// HealthAnalysisDays is the number of days to analyze for health score
	HealthAnalysisDays = 30
	// RecencyMaxDays is the maximum days for recency calculation (7 days)
	RecencyMaxDays = 7
)

// CalculateHealthScore calculates the health score for a database based on its backup history.
func CalculateHealthScore(backups []backup.BackupListEntry) HealthScore {
	score := HealthScore{
		Recommendations: []string{},
		RecentBackups:   backups,
	}

	// Filter backups from last 30 days
	cutoff := time.Now().AddDate(0, 0, -HealthAnalysisDays)
	var recentBackups []backup.BackupListEntry
	for _, b := range backups {
		if b.CreatedAt.After(cutoff) {
			recentBackups = append(recentBackups, b)
		}
	}

	if len(recentBackups) == 0 {
		score.Recommendations = append(score.Recommendations, "No backups found in the last 30 days. Create your first backup.")
		return score
	}

	// Calculate success rate (50% of score)
	successfulCount := 0
	totalCount := len(recentBackups)
	for _, b := range recentBackups {
		if b.Status == backup.StatusCompleted || b.Status == "" {
			successfulCount++
		}
	}

	if totalCount > 0 {
		score.SuccessRate = (float64(successfulCount) / float64(totalCount)) * 50.0
	} else {
		score.SuccessRate = 0
	}

	if score.SuccessRate < 50.0 {
		score.Recommendations = append(score.Recommendations, "Some backups have failed. Check backup logs for errors.")
	}

	// Calculate recency score (30% of score)
	if len(backups) > 0 {
		latestBackup := backups[0] // Already sorted newest first
		daysSince := time.Since(latestBackup.CreatedAt).Hours() / 24.0

		if daysSince <= RecencyMaxDays {
			score.RecencyScore = ((RecencyMaxDays - daysSince) / RecencyMaxDays) * 30.0
		} else {
			score.RecencyScore = 0
		}

		if daysSince > RecencyMaxDays {
			score.Recommendations = append(score.Recommendations, "Last backup is more than 7 days old. Consider scheduling regular backups.")
		} else if daysSince > 3 {
			score.Recommendations = append(score.Recommendations, "Last backup is more than 3 days old. Consider more frequent backups.")
		}
	} else {
		score.RecencyScore = 0
		score.Recommendations = append(score.Recommendations, "No backups found. Create your first backup.")
	}

	// Calculate consistency score (20% of score)
	if len(recentBackups) >= 2 {
		// Calculate intervals between consecutive backups
		var intervals []float64
		for i := 0; i < len(recentBackups)-1; i++ {
			interval := recentBackups[i].CreatedAt.Sub(recentBackups[i+1].CreatedAt).Hours() / 24.0
			intervals = append(intervals, interval)
		}

		// Calculate mean interval
		var sum float64
		for _, interval := range intervals {
			sum += interval
		}
		meanInterval := sum / float64(len(intervals))

		// Calculate standard deviation
		var variance float64
		for _, interval := range intervals {
			diff := interval - meanInterval
			variance += diff * diff
		}
		stdDev := math.Sqrt(variance / float64(len(intervals)))

		// Consistency score: lower deviation relative to mean = higher score
		if meanInterval > 0 {
			coefficientOfVariation := stdDev / meanInterval
			// Normalize: 0 CV = 100% score, 1+ CV = 0% score
			consistencyRatio := 1.0 - math.Min(1.0, coefficientOfVariation)
			score.ConsistencyScore = consistencyRatio * 20.0
		} else {
			score.ConsistencyScore = 0
		}

		if score.ConsistencyScore < 10.0 {
			score.Recommendations = append(score.Recommendations, "Backup intervals are inconsistent. Consider scheduling regular backups.")
		}
	} else {
		// Not enough backups for consistency calculation
		score.ConsistencyScore = 0
		if len(recentBackups) == 1 {
			score.Recommendations = append(score.Recommendations, "Only one backup found. Create more backups to assess consistency.")
		}
	}

	// Calculate total score
	score.TotalScore = score.SuccessRate + score.RecencyScore + score.ConsistencyScore

	// Add general recommendations based on total score
	if score.TotalScore >= HealthScoreHealthy {
		// Healthy - no additional recommendations needed
	} else if score.TotalScore >= HealthScoreWarning {
		score.Recommendations = append(score.Recommendations, "Backup health is acceptable but could be improved.")
	} else {
		score.Recommendations = append(score.Recommendations, "Backup health needs attention. Review backup configuration and logs.")
	}

	return score
}

// GetHealthStatus returns a status string based on the health score.
func GetHealthStatus(score float64) string {
	if score >= HealthScoreHealthy {
		return "healthy"
	} else if score >= HealthScoreWarning {
		return "warning"
	}
	return "critical"
}
