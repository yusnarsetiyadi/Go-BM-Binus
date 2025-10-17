package general

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// --- Helper AHP utilities ---

// CalculateAHP: hitung bobot & CR dari matriks perbandingan pairwise (n x n)
func CalculateAHP(matrix [][]float64) ([]float64, float64) {
	n := len(matrix)
	if n == 0 {
		return nil, 0
	}
	colSum := make([]float64, n)

	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			colSum[j] += matrix[i][j]
		}
	}

	normalized := make([][]float64, n)
	for i := 0; i < n; i++ {
		normalized[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			// protect division by zero
			if colSum[j] == 0 {
				normalized[i][j] = 0
			} else {
				normalized[i][j] = matrix[i][j] / colSum[j]
			}
		}
	}

	weights := make([]float64, n)
	for i := 0; i < n; i++ {
		sum := 0.0
		for j := 0; j < n; j++ {
			sum += normalized[i][j]
		}
		weights[i] = sum / float64(n)
	}

	// lambda_max
	lambdaMax := 0.0
	for i := 0; i < n; i++ {
		rowSum := 0.0
		for j := 0; j < n; j++ {
			rowSum += matrix[i][j] * weights[j]
		}
		// protect zero weight
		if weights[i] != 0 {
			lambdaMax += rowSum / weights[i]
		}
	}
	lambdaMax /= float64(n)

	CI := 0.0
	if n > 1 {
		CI = (lambdaMax - float64(n)) / (float64(n) - 1)
	}

	RI := map[int]float64{
		1: 0.00, 2: 0.00, 3: 0.58, 4: 0.90, 5: 1.12,
		6: 1.24, 7: 1.32, 8: 1.41, 9: 1.45, 10: 1.49,
	}
	CR := 0.0
	if val, ok := RI[n]; ok && val != 0 {
		CR = CI / val
	}
	return weights, CR
}

// clipSaaty: clip ratio ke rentang [1/9, 9]
func clipSaaty(r float64) float64 {
	if math.IsNaN(r) || math.IsInf(r, 0) {
		return 1.0 // anggap netral
	}
	if r <= 0 {
		return 1.0
	}
	if r > 9.0 {
		return 9.0
	}
	if r < 1.0/9.0 {
		return 1.0 / 9.0
	}
	return r
}

// BuildPairwiseFromScores: buat matriks pairwise dari skor numerik (scores[i] lebih "baik" jika lebih besar)
// matrix[i][j] = clip(scores[i] / scores[j])
func BuildPairwiseFromScores(scores []float64) [][]float64 {
	n := len(scores)
	matrix := make([][]float64, n)
	for i := 0; i < n; i++ {
		matrix[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			// jika kedua nilai 0 -> set 1
			if scores[j] == 0 {
				if scores[i] == 0 {
					matrix[i][j] = 1
				} else {
					matrix[i][j] = clipSaaty(scores[i] / 1e-9)
				}
			} else {
				matrix[i][j] = clipSaaty(scores[i] / scores[j])
			}
		}
	}
	return matrix
}

// --- Parsing complexity JSON ---
// ekspektasi payload.EventComplexity adalah JSON array objek { "id": <int>, "event_name": "...", "complexity": <1-5> }
type complexityItem struct {
	ID         int     `json:"id"`
	EventName  string  `json:"event_name"`
	Complexity float64 `json:"complexity"`
}

// parseComplexities mengembalikan map[id]complexity
func ParseComplexities(raw string) map[int]float64 {
	out := map[int]float64{}
	if raw == "" {
		return out
	}
	var arr []complexityItem
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return out
	}
	for _, it := range arr {
		out[it.ID] = it.Complexity
	}
	return out
}

// --- Utility: compute days between created_at and event_date_start (urgency) ---
// lower days => lebih urgent -> kita ingin skor yang lebih besar untuk lebih urgent,
// jadi kita akan ubah: urgencyScore = 1 / (days + 1) atau pakai transformasi lain.
func ComputeUrgencyScore(createdAt, eventStart time.Time) float64 {
	diff := eventStart.Sub(createdAt)
	diffHours := diff.Hours()
	diffDays := diff.Hours() / 24.0

	if diffHours < 0 {
		diffHours = 0
		diffDays = 0
	}

	var score float64

	switch {
	case diffDays < 3:
		score = 9.0 / (diffHours + 1.0)
	case diffDays <= 30:
		score = 9.0 / (diffDays + 1.0)
	default:
		score = 9.0 / math.Log(diffDays+2)
	}

	if score > 9 {
		score = 9
	}
	if score < 0.1 {
		score = 0.1
	}

	return score
}

// helper untuk ambil nama alternatif dari alts
func AlternatifNamesFromAlts(alts []AltRaw) []string {
	names := make([]string, len(alts))
	for i := range alts {
		names[i] = alts[i].EventName
	}
	return names
}

// printMatrix untuk bantu lihat matriks pairwise
func PrintMatrix(m [][]float64) {
	for _, row := range m {
		for _, val := range row {
			fmt.Printf("%8.4f ", val)
		}
		fmt.Println()
	}
}

// simpan seluruh data mentah juga untuk proses AHP
type AltRaw struct {
	ID                int
	UserID            int
	UserName          string
	EventName         string
	EventLocation     string
	EventDateStart    time.Time
	EventDateEnd      time.Time
	Description       string
	EventTypeID       int
	EventTypeName     string
	EventTypePriority int
	StatusID          int
	StatusName        string
	CountParticipant  int
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}
