package request

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/gdrive"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"bm_binus/pkg/ws"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/drive/v3"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.RequestCreateRequest) (map[string]interface{}, error)
	Find(ctx *abstraction.Context, payload *dto.RequestFindRequest) (map[string]interface{}, error)
	FindById(ctx *abstraction.Context, payload *dto.RequestFindByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.RequestUpdateRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.RequestDeleteByIDRequest) (map[string]interface{}, error)
}

type service struct {
	RequestRepository      repository.Request
	EventTypeRepository    repository.EventType
	FileRepository         repository.File
	NotificationRepository repository.Notification
	UserRepository         repository.User
	StatusRepository       repository.Status

	DB     *gorm.DB
	sDrive *drive.Service
	fDrive *drive.File
}

func NewService(f *factory.Factory) Service {
	return &service{
		RequestRepository:      f.RequestRepository,
		EventTypeRepository:    f.EventTypeRepository,
		FileRepository:         f.FileRepository,
		NotificationRepository: f.NotificationRepository,
		UserRepository:         f.UserRepository,
		StatusRepository:       f.StatusRepository,

		DB:     f.Db,
		sDrive: f.GDrive.Service,
		fDrive: f.GDrive.FolderBM,
	}
}

func SendNotif(s *service, ctx *abstraction.Context, title string, message string, userId int, requestId int) error {
	modelNotification := &model.NotificationEntityModel{
		Context: ctx,
		NotificationEntity: model.NotificationEntity{
			Title:     title,
			Message:   message,
			IsRead:    false,
			UserId:    userId,
			RequestId: requestId,
		},
	}
	if err := s.NotificationRepository.Create(ctx, modelNotification).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.RequestCreateRequest) (map[string]interface{}, error) {
	var (
		allFileUploaded []string
		sendNotifTo     []int
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_STAF {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		parsedEventDateStart, err := general.Parse("2006-01-02 15:04:05", payload.EventDateStart)
		if err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "err parse event date start:"+err.Error())
		}

		parsedEventDateEnd, err := general.Parse("2006-01-02 15:04:05", payload.EventDateEnd)
		if err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "err parse event date end:"+err.Error())
		}

		eventTypeData, err := s.EventTypeRepository.FindById(ctx, payload.EventTypeId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if eventTypeData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "event type not found")
		}

		modelRequest := &model.RequestEntityModel{
			Context: ctx,
			RequestEntity: model.RequestEntity{
				UserId:           ctx.Auth.ID,
				EventName:        payload.EventName,
				EventLocation:    payload.EventLocation,
				EventDateStart:   parsedEventDateStart,
				EventDateEnd:     parsedEventDateEnd,
				Description:      payload.Description,
				EventTypeId:      payload.EventTypeId,
				CountParticipant: payload.CountParticipant,
				StatusId:         constant.STATUS_ID_PENGAJUAN,
				IsDelete:         false,
			},
		}
		if err = s.RequestRepository.Create(ctx, modelRequest).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		for _, file := range payload.Files {
			f, err := file.Open()
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			defer f.Close()

			isFileAvailable, fullFileName := general.ValidateFileUpload(file.Filename)
			if !isFileAvailable {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), fmt.Sprintf("file format for %s is not approved", file.Filename))
			}

			newFile, err := gdrive.CreateFile(s.sDrive, fullFileName, "application/octet-stream", f, s.fDrive.Id)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			allFileUploaded = append(allFileUploaded, newFile.Id)

			modelFile := &model.FileEntityModel{
				Context: ctx,
				FileEntity: model.FileEntity{
					RequestId: modelRequest.ID,
					File:      newFile.Id,
					FileName:  newFile.Name,
					IsDelete:  false,
				},
			}
			if err := s.FileRepository.Create(ctx, modelFile).Error; err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		userBM, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_BM, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		for _, v := range userBM {
			err := SendNotif(s, ctx, "Event Baru!", modelRequest.EventName, v.ID, modelRequest.ID)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			sendNotifTo = append(sendNotifTo, v.ID)
		}

		return nil
	}); err != nil {
		for _, v := range allFileUploaded {
			errDel := gdrive.DeleteFile(s.sDrive, v)
			if errDel != nil {
				logrus.Error("error delete file for error trxmanager:", errDel.Error())
			}
		}
		return nil, err
	}

	for _, v := range general.RemoveDuplicateArrayInt(sendNotifTo) {
		if err := ws.PublishNotificationWithoutTransaction(v, s.DB, ctx); err != nil {
			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
	}

	return map[string]interface{}{
		"message": "success create!",
	}, nil
}

func (s *service) Find(ctx *abstraction.Context, payload *dto.RequestFindRequest) (map[string]interface{}, error) {
	var (
		res  []map[string]interface{} = nil
		alts []general.AltRaw
	)
	data, err := s.RequestRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.RequestRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	for _, v := range data {
		resData := map[string]interface{}{
			"id": v.ID,
			"user": map[string]interface{}{
				"id":   v.User.ID,
				"name": v.User.Name,
			},
			"event_name":       v.EventName,
			"event_location":   v.EventLocation,
			"event_date_start": general.FormatWithZWithoutChangingTime(v.EventDateStart),
			"event_date_end":   general.FormatWithZWithoutChangingTime(v.EventDateEnd),
			"event_type": map[string]interface{}{
				"id":       v.EventType.ID,
				"name":     v.EventType.Name,
				"priority": v.EventType.Priority,
			},
			"status": map[string]interface{}{
				"id":   v.Status.ID,
				"name": v.Status.Name,
			},
			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
		}
		res = append(res, resData)

		alts = append(alts, general.AltRaw{
			ID:                v.ID,
			UserID:            v.User.ID,
			UserName:          v.User.Name,
			EventName:         v.EventName,
			EventLocation:     v.EventLocation,
			EventDateStart:    v.EventDateStart,
			EventDateEnd:      v.EventDateEnd,
			Description:       v.Description,
			EventTypeID:       v.EventType.ID,
			EventTypeName:     v.EventType.Name,
			EventTypePriority: v.EventType.Priority,
			StatusID:          v.Status.ID,
			StatusName:        v.Status.Name,
			CountParticipant:  v.CountParticipant,
			CreatedAt:         v.CreatedAt,
			UpdatedAt:         v.UpdatedAt,
		})
	}

	// ahp
	if payload.UseAhp != nil && *payload.UseAhp == "yes" {
		fmt.Println("=== [AHP MODE AKTIF] ===")

		complexityMap := map[int]float64{}
		if payload.EventComplexity != nil && *payload.EventComplexity != "" {
			fmt.Println("-> Parsing EventComplexity JSON...")
			complexityMap = general.ParseComplexities(*payload.EventComplexity)
			fmt.Println("   Hasil parsing complexityMap:", complexityMap)
		}

		n := len(alts)
		if n > 0 {
			fmt.Printf("-> Jumlah alternatif: %d\n", n)

			// siapkan slice skor
			urgencyScores := make([]float64, n)
			importanceScores := make([]float64, n)
			participantScores := make([]float64, n)
			complexityScores := make([]float64, n)

			for i, a := range alts {
				urgencyScores[i] = general.ComputeUrgencyScore(a.CreatedAt, a.EventDateStart)
				priority := a.EventTypePriority
				if priority <= 0 {
					priority = 1
				}
				importanceScores[i] = float64(1) / float64(priority)
				participantScores[i] = float64(a.CountParticipant)

				compVal := 1.0
				if c, ok := complexityMap[a.ID]; ok {
					compVal = c
				}
				complexityScores[i] = 6.0 - compVal
			}

			fmt.Println("\n--- [SKOR AWAL SETIAP KRITERIA] ---")
			for i, a := range alts {
				fmt.Printf("%d. %s\n", i+1, a.EventName)
				fmt.Printf("   Urgency: %.4f\n", urgencyScores[i])
				fmt.Printf("   Importance: %.4f\n", importanceScores[i])
				fmt.Printf("   Participants: %.4f\n", participantScores[i])
				fmt.Printf("   Complexity: %.4f\n", complexityScores[i])
			}

			critNames := []string{"Urgency", "Importance", "Participants", "Complexity"}
			critImportanceRaw := []float64{5, 3, 2, 1}
			fmt.Println("\n--- [KRITERIA UTAMA] ---")
			fmt.Println("Nama:", critNames)
			fmt.Println("Bobot Awal:", critImportanceRaw)

			criteriaMatrix := general.BuildPairwiseFromScores(critImportanceRaw)
			fmt.Println("\nMatriks Perbandingan Kriteria:")
			general.PrintMatrix(criteriaMatrix)

			criteriaWeights, criteriaCR := general.CalculateAHP(criteriaMatrix)
			fmt.Printf("Bobot Kriteria: %.4f %.4f %.4f %.4f\n", criteriaWeights[0], criteriaWeights[1], criteriaWeights[2], criteriaWeights[3])
			fmt.Printf("CR (Consistency Ratio): %.4f\n", criteriaCR)

			// build pairwise alternative matrices
			mUrgency := general.BuildPairwiseFromScores(urgencyScores)
			wUrgency, crUrgency := general.CalculateAHP(mUrgency)
			fmt.Println("\n--- [AHP Urgency] ---")
			general.PrintMatrix(mUrgency)
			fmt.Println("Bobot alternatif:", wUrgency)
			fmt.Printf("CR: %.4f\n", crUrgency)

			mImportance := general.BuildPairwiseFromScores(importanceScores)
			wImportance, crImportance := general.CalculateAHP(mImportance)
			fmt.Println("\n--- [AHP Importance] ---")
			general.PrintMatrix(mImportance)
			fmt.Println("Bobot alternatif:", wImportance)
			fmt.Printf("CR: %.4f\n", crImportance)

			mParticipant := general.BuildPairwiseFromScores(participantScores)
			wParticipant, crParticipant := general.CalculateAHP(mParticipant)
			fmt.Println("\n--- [AHP Participants] ---")
			general.PrintMatrix(mParticipant)
			fmt.Println("Bobot alternatif:", wParticipant)
			fmt.Printf("CR: %.4f\n", crParticipant)

			mComplexity := general.BuildPairwiseFromScores(complexityScores)
			wComplexity, crComplexity := general.CalculateAHP(mComplexity)
			fmt.Println("\n--- [AHP Complexity] ---")
			general.PrintMatrix(mComplexity)
			fmt.Println("Bobot alternatif:", wComplexity)
			fmt.Printf("CR: %.4f\n", crComplexity)

			// hitung skor akhir
			finalScores := make([]float64, n)
			for i := 0; i < n; i++ {
				finalScores[i] = criteriaWeights[0]*wUrgency[i] +
					criteriaWeights[1]*wImportance[i] +
					criteriaWeights[2]*wParticipant[i] +
					criteriaWeights[3]*wComplexity[i]
			}

			fmt.Println("\n--- [FINAL SCORE SETIAP ALTERNATIF] ---")
			for i, a := range alts {
				fmt.Printf("%s: %.6f\n", a.EventName, finalScores[i])
			}

			// ranking
			type rItem struct {
				ID    int
				Name  string
				Score float64
			}
			var ranked []rItem
			for i, a := range alts {
				ranked = append(ranked, rItem{
					ID:    a.ID,
					Name:  a.EventName,
					Score: finalScores[i],
				})
			}
			sort.SliceStable(ranked, func(i, j int) bool {
				return ranked[i].Score > ranked[j].Score
			})

			fmt.Println("\n--- [RANKING AKHIR] ---")
			for i, r := range ranked {
				fmt.Printf("%d. %s (Score: %.6f)\n", i+1, r.Name, r.Score)
			}

			// siapkan hasil ahpResult
			altResults := []map[string]interface{}{}
			for idx, it := range ranked {
				altResults = append(altResults, map[string]interface{}{
					"rank":  idx + 1,
					"id":    it.ID,
					"name":  it.Name,
					"score": it.Score,
				})
			}

			if len(res) == len(alts) {
				// buat map untuk lookup skor AHP berdasarkan ID
				scoreMap := make(map[int]float64)
				for _, r := range ranked {
					scoreMap[r.ID] = r.Score
				}

				// normalisasi skor jadi persen (0–100)
				maxScore := ranked[0].Score
				minScore := ranked[len(ranked)-1].Score
				diff := maxScore - minScore
				if diff == 0 {
					diff = 1 // biar gak bagi 0
				}

				// tambahkan ahp_score ke setiap item res sesuai ID
				for _, r := range res {
					id, ok := r["id"].(int)
					if !ok {
						continue
					}
					score := scoreMap[id]
					total := 0.0
					for _, v := range scoreMap {
						total += v
					}
					percent := (score / total) * 100
					r["ahp_score"] = map[string]interface{}{
						"raw":     score,
						"percent": fmt.Sprintf("%.2f%%", percent),
					}
				}

				// urutkan ulang res berdasarkan ranking
				sort.SliceStable(res, func(i, j int) bool {
					idI, _ := res[i]["id"].(int)
					idJ, _ := res[j]["id"].(int)
					return scoreMap[idI] > scoreMap[idJ]
				})
			}
		}
	}

	resp := map[string]interface{}{
		"count": count,
		"data":  res,
	}

	return resp, nil
}

func (s *service) FindById(ctx *abstraction.Context, payload *dto.RequestFindByIDRequest) (map[string]interface{}, error) {
	var res map[string]interface{} = nil
	data, err := s.RequestRepository.FindById(ctx, payload.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if data != nil {
		res = map[string]interface{}{
			"id": data.ID,
			"user": map[string]interface{}{
				"id":   data.User.ID,
				"name": data.User.Name,
			},
			"event_name":       data.EventName,
			"event_location":   data.EventLocation,
			"event_date_start": general.FormatWithZWithoutChangingTime(data.EventDateStart),
			"event_date_end":   general.FormatWithZWithoutChangingTime(data.EventDateEnd),
			"description":      data.Description,
			"event_type": map[string]interface{}{
				"id":       data.EventType.ID,
				"name":     data.EventType.Name,
				"priority": data.EventType.Priority,
			},
			"count_participant": data.CountParticipant,
			"status": map[string]interface{}{
				"id":   data.Status.ID,
				"name": data.Status.Name,
			},
			"created_at": general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
		}
	}
	return map[string]interface{}{
		"data": res,
	}, nil
}

func (s *service) Update(ctx *abstraction.Context, payload *dto.RequestUpdateRequest) (map[string]interface{}, error) {
	var (
		sendNotifTo      []int
		statusesForAdmin = []int{
			constant.STATUS_ID_PROSES,
			constant.STATUS_ID_FINALISASI,
			constant.STATUS_ID_SELESAI,
		}
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		requestData, err := s.RequestRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if requestData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request not found")
		}

		reloadData := false
		newRequestData := new(model.RequestEntityModel)
		newRequestData.Context = ctx
		newRequestData.ID = payload.ID
		if payload.EventName != nil {
			newRequestData.EventName = *payload.EventName
			reloadData = true
		}
		if payload.EventLocation != nil {
			newRequestData.EventLocation = *payload.EventLocation
		}
		if payload.EventDateStart != nil {
			parsedEventDateStart, err := general.Parse("2006-01-02 15:04:05", *payload.EventDateStart)
			if err != nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "err parse event date start:"+err.Error())
			}
			newRequestData.EventDateStart = parsedEventDateStart
		}
		if payload.EventDateEnd != nil {
			parsedEventDateEnd, err := general.Parse("2006-01-02 15:04:05", *payload.EventDateEnd)
			if err != nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "err parse event date end:"+err.Error())
			}
			newRequestData.EventDateEnd = parsedEventDateEnd
		}
		if payload.Description != nil {
			newRequestData.Description = *payload.Description
		}
		if payload.EventTypeId != nil {
			eventTypeData, err := s.EventTypeRepository.FindById(ctx, *payload.EventTypeId)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if eventTypeData == nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "event type not found")
			}
			newRequestData.EventTypeId = *payload.EventTypeId
		}
		if payload.CountParticipant != nil {
			newRequestData.CountParticipant = *payload.CountParticipant
		}
		if payload.StatusId != nil {
			statusData, err := s.StatusRepository.FindById(ctx, *payload.StatusId)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if statusData == nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "status not found")
			}
			newRequestData.StatusId = *payload.StatusId
			reloadData = true
		}
		if err = s.RequestRepository.Update(ctx, newRequestData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		if reloadData {
			requestData, err = s.RequestRepository.FindById(ctx, payload.ID)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		userBM, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_BM, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		userAdmin, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_ADMIN, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		addIDs := func(users []*model.UserEntityModel) {
			for _, v := range users {
				sendNotifTo = append(sendNotifTo, v.ID)
			}
		}
		switch ctx.Auth.RoleID {
		case constant.ROLE_ID_STAF:
			addIDs(userBM)
			if slices.Contains(statusesForAdmin, requestData.StatusId) {
				addIDs(userAdmin)
			}
		case constant.ROLE_ID_BM:
			sendNotifTo = append(sendNotifTo, requestData.UserId)
			if slices.Contains(statusesForAdmin, requestData.StatusId) {
				addIDs(userAdmin)
			}
		case constant.ROLE_ID_ADMIN:
			sendNotifTo = append(sendNotifTo, requestData.UserId)
			addIDs(userBM)
		}

		for _, v := range sendNotifTo {
			err = SendNotif(s, ctx, "Event diperbarui!", requestData.EventName, v, requestData.ID)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, v := range general.RemoveDuplicateArrayInt(sendNotifTo) {
		if err := ws.PublishNotificationWithoutTransaction(v, s.DB, ctx); err != nil {
			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
	}

	return map[string]interface{}{
		"message": "success update!",
	}, nil
}

func (s *service) Delete(ctx *abstraction.Context, payload *dto.RequestDeleteByIDRequest) (map[string]interface{}, error) {
	var (
		sendNotifTo      []int
		statusesForAdmin = []int{
			constant.STATUS_ID_PROSES,
			constant.STATUS_ID_FINALISASI,
			constant.STATUS_ID_SELESAI,
		}
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		requestData, err := s.RequestRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if requestData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request not found")
		}

		if ctx.Auth.ID != requestData.UserId {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		newRequestData := new(model.RequestEntityModel)
		newRequestData.Context = ctx
		newRequestData.ID = requestData.ID
		newRequestData.IsDelete = true
		if err = s.RequestRepository.Update(ctx, newRequestData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userBM, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_BM, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		userAdmin, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_ADMIN, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		addIDs := func(users []*model.UserEntityModel) {
			for _, v := range users {
				sendNotifTo = append(sendNotifTo, v.ID)
			}
		}
		switch ctx.Auth.RoleID {
		case constant.ROLE_ID_STAF:
			addIDs(userBM)
			if slices.Contains(statusesForAdmin, requestData.StatusId) {
				addIDs(userAdmin)
			}
		case constant.ROLE_ID_BM:
			sendNotifTo = append(sendNotifTo, requestData.UserId)
			if slices.Contains(statusesForAdmin, requestData.StatusId) {
				addIDs(userAdmin)
			}
		case constant.ROLE_ID_ADMIN:
			sendNotifTo = append(sendNotifTo, requestData.UserId)
			addIDs(userBM)
		}

		for _, v := range sendNotifTo {
			err = SendNotif(s, ctx, "Event dihapus!", requestData.EventName, v, requestData.ID)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, v := range general.RemoveDuplicateArrayInt(sendNotifTo) {
		if err := ws.PublishNotificationWithoutTransaction(v, s.DB, ctx); err != nil {
			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
	}

	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}
