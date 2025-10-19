package ahphistory

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.AhpHistoryCreateRequest) (map[string]interface{}, error)
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	FindById(ctx *abstraction.Context, payload *dto.AhpHistoryFindByIDRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.AhpHistoryDeleteByIDRequest) (map[string]interface{}, error)
}

type service struct {
	AhpHistoryRepository repository.AhpHistory
	RequestRepository    repository.Request

	DB *gorm.DB
}

func NewService(f *factory.Factory) Service {
	return &service{
		AhpHistoryRepository: f.AhpHistoryRepository,
		RequestRepository:    f.RequestRepository,

		DB: f.Db,
	}
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.AhpHistoryCreateRequest) (map[string]interface{}, error) {
	var resId int
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_BM {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		requestData, err := s.RequestRepository.FindById(ctx, payload.ReferenceRequest)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if requestData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request not found")
		}

		fmt.Println("=== [AHP DYNAMIC CALCULATION] ===")

		// --- [ Tambahan baru: konversi payload comparison ke matrix input ] ---

		// fungsi bantu: bangun matriks pairwise dari AhpComparisonRequest (kriteria_comparison)
		buildPairwiseFromComparison := func(items []string, comps []dto.AhpComparisonRequest) [][]float64 {
			n := len(items)
			matrix := make([][]float64, n)
			for i := range matrix {
				matrix[i] = make([]float64, n)
				for j := range matrix[i] {
					matrix[i][j] = 1
				}
			}
			for _, c := range comps {
				i1 := slices.Index(items, c.Item1)
				i2 := slices.Index(items, c.Item2)
				if i1 == -1 || i2 == -1 {
					continue
				}
				matrix[i1][i2] = c.Value
				matrix[i2][i1] = 1 / c.Value
			}
			return matrix
		}

		// fungsi bantu: ambil matrix alternatif per kriteria dari map
		buildAltMatrixFromMap := func(alternatif []string, comps []dto.AhpComparisonRequest) [][]float64 {
			n := len(alternatif)
			matrix := make([][]float64, n)
			for i := range matrix {
				matrix[i] = make([]float64, n)
				for j := range matrix[i] {
					matrix[i][j] = 1
				}
			}
			for _, c := range comps {
				i1 := slices.Index(alternatif, c.Item1)
				i2 := slices.Index(alternatif, c.Item2)
				if i1 == -1 || i2 == -1 {
					continue
				}
				matrix[i1][i2] = c.Value
				matrix[i2][i1] = 1 / c.Value
			}
			return matrix
		}

		// --- 1. Bangun pairwise matrix Kriteria ---
		kritMatrix := buildPairwiseFromComparison(payload.Kriteria, payload.KriteriaComparison)

		fmt.Println("\n[ Matriks Kriteria ]")
		general.PrintMatrix(kritMatrix)

		kritWeights, kritCR := general.CalculateAHP(kritMatrix)
		fmt.Println("Bobot Kriteria:", kritWeights)
		fmt.Printf("CR: %.4f\n", kritCR)

		// --- 2. Bangun pairwise matrix untuk setiap kriteria terhadap alternatif ---
		nAlt := len(payload.Alternatif)
		totalScore := make([]float64, nAlt)

		for kriteriaName, comps := range payload.AlternatifComparison {
			fmt.Printf("\n[ Matriks Alternatif terhadap %s ]\n", kriteriaName)

			mAlt := buildAltMatrixFromMap(payload.Alternatif, comps)
			general.PrintMatrix(mAlt)
			wAlt, crAlt := general.CalculateAHP(mAlt)
			fmt.Println("Bobot alternatif:", wAlt)
			fmt.Printf("CR: %.4f\n", crAlt)

			// cari index kriteria yang sesuai
			kIndex := -1
			for i, k := range payload.Kriteria {
				if k == kriteriaName {
					kIndex = i
					break
				}
			}
			if kIndex == -1 {
				continue
			}

			// akumulasi skor global
			for i := 0; i < nAlt; i++ {
				totalScore[i] += kritWeights[kIndex] * wAlt[i]
			}
		}

		// --- 3. Ranking hasil akhir ---
		fmt.Println("\n=== [HASIL AKHIR AHP] ===")
		type item struct {
			Name  string
			Score float64
		}
		results := []item{}
		for i, name := range payload.Alternatif {
			results = append(results, item{Name: name, Score: totalScore[i]})
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})

		for i, r := range results {
			fmt.Printf("%d. %s (Score: %.6f)\n", i+1, r.Name, r.Score)
		}

		// prepare data for save to db
		storedKriteriaComparison := map[string]interface{}{
			"matrix":  kritMatrix,
			"weights": kritWeights,
			"cr":      kritCR,
		}

		storedAlternatifComparison := map[string]interface{}{}
		for kriteriaName, comps := range payload.AlternatifComparison {
			mAlt := buildAltMatrixFromMap(payload.Alternatif, comps)
			wAlt, crAlt := general.CalculateAHP(mAlt)
			storedAlternatifComparison[kriteriaName] = map[string]interface{}{
				"matrix":  mAlt,
				"weights": wAlt,
				"cr":      crAlt,
			}
		}

		storedPriorityGlobal := map[string]interface{}{
			"alternatif": payload.Alternatif,
			"priority":   results, // hasil total skor (global priority)
		}

		// Encode ke JSON
		kritJSON, _ := json.Marshal(storedKriteriaComparison)
		altJSON, _ := json.Marshal(storedAlternatifComparison)
		prioJSON, _ := json.Marshal(storedPriorityGlobal)

		modelAhpHistory := &model.AhpHistoryEntityModel{
			Context: ctx,
			AhpHistoryEntity: model.AhpHistoryEntity{
				Kriteria:             strings.Join(payload.Kriteria, ","),
				KriteriaComparison:   string(kritJSON),
				Alternatif:           strings.Join(payload.Alternatif, ","),
				AlternatifComparison: string(altJSON),
				PriorityGlobal:       string(prioJSON),
				ReferenceRequest:     payload.ReferenceRequest,
				IsDelete:             false,
			},
		}
		if err := s.AhpHistoryRepository.Create(ctx, modelAhpHistory).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		resId = modelAhpHistory.ID

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "success create!",
		"id":      resId,
	}, nil
}

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	var res []map[string]interface{} = nil
	data, err := s.AhpHistoryRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.AhpHistoryRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	for _, v := range data {
		requestData, err := s.RequestRepository.FindById(ctx, v.ReferenceRequest)
		if err != nil && err.Error() != "record not found" {
			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if requestData == nil {
			return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request not found")
		}

		var kriteriaData map[string]interface{}
		var altData map[string]interface{}
		var globalData map[string]interface{}
		_ = json.Unmarshal([]byte(v.KriteriaComparison), &kriteriaData)
		_ = json.Unmarshal([]byte(v.AlternatifComparison), &altData)
		_ = json.Unmarshal([]byte(v.PriorityGlobal), &globalData)
		kriteriaVal := strings.Split(v.Kriteria, ",")
		alternatifVal := strings.Split(v.Alternatif, ",")

		globalSummary := []map[string]interface{}{}
		if arr, ok := globalData["priority"].([]interface{}); ok {
			for i, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					globalSummary = append(globalSummary, map[string]interface{}{
						"rank":  i + 1,
						"name":  m["Name"],
						"score": fmt.Sprintf("%.6f", m["Score"]),
					})
				}
			}
		}

		resData := map[string]interface{}{
			"id":         v.ID,
			"kriteria":   kriteriaVal,
			"alternatif": alternatifVal,
			"priority":   globalSummary,
			"reference_request": map[string]interface{}{
				"id":         requestData.ID,
				"user":       requestData.User.Name,
				"event_name": requestData.EventName,
			},
			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
		}

		res = append(res, resData)
	}

	resp := map[string]interface{}{
		"count": count,
		"data":  res,
	}

	return resp, nil
}

func (s *service) FindById(ctx *abstraction.Context, payload *dto.AhpHistoryFindByIDRequest) (map[string]interface{}, error) {
	var res map[string]interface{} = nil
	data, err := s.AhpHistoryRepository.FindById(ctx, payload.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if data != nil {
		requestData, err := s.RequestRepository.FindById(ctx, data.ReferenceRequest)
		if err != nil && err.Error() != "record not found" {
			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if requestData == nil {
			return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request not found")
		}

		var kriteriaData map[string]interface{}
		var altData map[string]interface{}
		var globalData map[string]interface{}

		_ = json.Unmarshal([]byte(data.KriteriaComparison), &kriteriaData)
		_ = json.Unmarshal([]byte(data.AlternatifComparison), &altData)
		_ = json.Unmarshal([]byte(data.PriorityGlobal), &globalData)

		kriteriaVal := strings.Split(data.Kriteria, ",")
		_ = strings.Split(data.Alternatif, ",")

		// --- format ringkasan kriteria ---
		kritSummary := map[string]interface{}{
			"total": len(kriteriaVal),
			"list":  kriteriaVal,
			"cr":    kriteriaData["cr"],
			"weights": func() []map[string]interface{} {
				ws := []map[string]interface{}{}
				if arr, ok := kriteriaData["weights"].([]interface{}); ok {
					for i, w := range arr {
						ws = append(ws, map[string]interface{}{
							"name":   kriteriaVal[i],
							"weight": fmt.Sprintf("%.4f", w),
						})
					}
				}
				return ws
			}(),
			"matrix": kriteriaData["matrix"],
		}

		// --- format hasil alternatif per kriteria ---
		altSummary := []map[string]interface{}{}
		for k, v2 := range altData {
			if m, ok := v2.(map[string]interface{}); ok {
				altSummary = append(altSummary, map[string]interface{}{
					"kriteria": k,
					"cr":       m["cr"],
					"weights":  m["weights"],
					"matrix":   m["matrix"],
				})
			}
		}

		// --- format hasil global ranking ---
		globalSummary := []map[string]interface{}{}
		if arr, ok := globalData["priority"].([]interface{}); ok {
			for i, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					globalSummary = append(globalSummary, map[string]interface{}{
						"rank":  i + 1,
						"name":  m["Name"],
						"score": fmt.Sprintf("%.6f", m["Score"]),
					})
				}
			}
		}

		// --- build final response ---
		res = map[string]interface{}{
			"id":                 data.ID,
			"kriteria_summary":   kritSummary,
			"alternatif_summary": altSummary,
			"global_priority":    globalSummary,
			"reference_request": map[string]interface{}{
				"id":          requestData.ID,
				"user":        requestData.User.Name,
				"event_name":  requestData.EventName,
				"description": requestData.Description,
			},
			"created_at": general.FormatWithZWithoutChangingTime(data.CreatedAt),
		}
	}

	return map[string]interface{}{
		"data": res,
	}, nil
}

func (s *service) Delete(ctx *abstraction.Context, payload *dto.AhpHistoryDeleteByIDRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		ahpHistoryData, err := s.AhpHistoryRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if ahpHistoryData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "ahp history not found")
		}

		newAhpHistoryData := new(model.AhpHistoryEntityModel)
		newAhpHistoryData.Context = ctx
		newAhpHistoryData.ID = ahpHistoryData.ID
		newAhpHistoryData.IsDelete = true
		if err = s.AhpHistoryRepository.Update(ctx, newAhpHistoryData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}
