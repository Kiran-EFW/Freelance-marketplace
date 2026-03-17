package review

import (
	"testing"
)

// ---------------------------------------------------------------------------
// TestEvaluateProviderLevel
// ---------------------------------------------------------------------------

func TestEvaluateProviderLevel(t *testing.T) {
	tests := []struct {
		name         string
		trustScore   float64
		totalJobs    int
		totalReviews int
		want         string
	}{
		// "new" — below all thresholds
		{
			name:       "brand new provider, no jobs",
			trustScore: 0,
			totalJobs:  0,
			want:       "new",
		},
		{
			name:       "low score, few jobs",
			trustScore: 1.5,
			totalJobs:  2,
			want:       "new",
		},
		{
			name:       "sufficient score but not enough jobs for active",
			trustScore: 3.0,
			totalJobs:  2,
			want:       "new",
		},

		// "active" — trust >= 2.0 AND jobs >= 3
		{
			name:       "exactly at active threshold",
			trustScore: 2.0,
			totalJobs:  3,
			want:       "active",
		},
		{
			name:       "above active but below trusted",
			trustScore: 3.0,
			totalJobs:  10,
			want:       "active",
		},
		{
			name:       "high score but only 3 jobs - capped at active",
			trustScore: 5.0,
			totalJobs:  3,
			want:       "active",
		},

		// "trusted" — trust >= 3.5 AND jobs >= 20
		{
			name:       "exactly at trusted threshold",
			trustScore: 3.5,
			totalJobs:  20,
			want:       "trusted",
		},
		{
			name:       "above trusted but below expert",
			trustScore: 3.8,
			totalJobs:  30,
			want:       "trusted",
		},
		{
			name:       "high score but only 20 jobs - capped at trusted",
			trustScore: 4.8,
			totalJobs:  20,
			want:       "trusted",
		},

		// "expert" — trust >= 4.0 AND jobs >= 50
		{
			name:       "exactly at expert threshold",
			trustScore: 4.0,
			totalJobs:  50,
			want:       "expert",
		},
		{
			name:       "above expert but not enough for champion",
			trustScore: 4.3,
			totalJobs:  70,
			want:       "expert",
		},

		// "local_champion" — trust >= 4.5 AND jobs >= 100
		{
			name:       "exactly at champion threshold",
			trustScore: 4.5,
			totalJobs:  100,
			want:       "local_champion",
		},
		{
			name:       "max level provider",
			trustScore: 5.0,
			totalJobs:  500,
			want:       "local_champion",
		},
		{
			name:       "perfect score but 99 jobs - stays expert",
			trustScore: 5.0,
			totalJobs:  99,
			want:       "expert",
		},

		// Edge cases
		{
			name:       "just below active score threshold",
			trustScore: 1.99,
			totalJobs:  100,
			want:       "new",
		},
		{
			name:       "just below trusted score threshold",
			trustScore: 3.49,
			totalJobs:  100,
			want:       "active",
		},
		{
			name:       "negative score",
			trustScore: -1.0,
			totalJobs:  0,
			want:       "new",
		},
		{
			name:       "score over 5.0 with enough jobs",
			trustScore: 5.5, // clamp in real code, but function handles it
			totalJobs:  200,
			want:       "local_champion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateProviderLevel(tt.trustScore, tt.totalJobs, tt.totalReviews)
			if got != tt.want {
				t.Errorf("evaluateProviderLevel(%f, %d, %d) = %q, want %q",
					tt.trustScore, tt.totalJobs, tt.totalReviews, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestLevelProgressionMonotonic
// ---------------------------------------------------------------------------

// Verify that increasing score and jobs always progresses levels in order.
func TestLevelProgressionMonotonic(t *testing.T) {
	levelOrder := map[string]int{
		"new":            0,
		"active":         1,
		"trusted":        2,
		"expert":         3,
		"local_champion": 4,
	}

	prevLevel := "new"
	// Increase score and jobs gradually.
	steps := []struct {
		score float64
		jobs  int
	}{
		{0.0, 0},
		{1.0, 1},
		{2.0, 3},
		{3.0, 10},
		{3.5, 20},
		{4.0, 50},
		{4.5, 100},
		{5.0, 200},
	}

	for _, step := range steps {
		level := evaluateProviderLevel(step.score, step.jobs, 0)
		if levelOrder[level] < levelOrder[prevLevel] {
			t.Errorf("level went backwards: score=%f jobs=%d gave %q, previous was %q",
				step.score, step.jobs, level, prevLevel)
		}
		prevLevel = level
	}
}

// ---------------------------------------------------------------------------
// TestTrustScoreWeightsSum
// ---------------------------------------------------------------------------

func TestTrustScoreWeightsSum(t *testing.T) {
	total := trustWeightRating + trustWeightCompletion + trustWeightResponse +
		trustWeightVolume + trustWeightRecency

	if total < 0.999 || total > 1.001 {
		t.Errorf("trust score weights sum to %f, want 1.0", total)
	}
}
