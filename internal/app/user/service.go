package contact

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/gomail"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-redis/redis/v8"
	"github.com/jung-kurt/gofpdf"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.UserCreateRequest) (map[string]interface{}, error)
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	FindById(ctx *abstraction.Context, payload *dto.UserFindByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.UserUpdateRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.UserDeleteByIDRequest) (map[string]interface{}, error)
	ChangePassword(ctx *abstraction.Context, payload *dto.UserChangePasswordRequest) (map[string]interface{}, error)
	Export(ctx *abstraction.Context, payload *dto.UserExportRequest) (string, *bytes.Buffer, string, error)
	Info(ctx *abstraction.Context) (map[string]interface{}, error)
}

type service struct {
	UserRepository repository.User
	RoleRepository repository.Role

	DB      *gorm.DB
	DbRedis *redis.Client
}

func NewService(f *factory.Factory) Service {
	return &service{
		UserRepository: f.UserRepository,
		RoleRepository: f.RoleRepository,

		DB:      f.Db,
		DbRedis: f.DbRedis,
	}
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.UserCreateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_BM {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		userEmail, err := s.UserRepository.FindByEmail(ctx, payload.Email)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userEmail != nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "email already exist")
		}

		roleData, err := s.RoleRepository.FindById(ctx, payload.RoleId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if roleData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "role not found")
		}

		passwordString := general.GeneratePassword(8, 1, 1, 1, 1)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordString), bcrypt.DefaultCost)
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		modelUser := &model.UserEntityModel{
			Context: ctx,
			UserEntity: model.UserEntity{
				Name:     payload.Name,
				Email:    payload.Email,
				Password: string(hashedPassword),
				RoleId:   payload.RoleId,
				IsDelete: false,
			},
		}
		if err = s.UserRepository.Create(ctx, modelUser).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		if err = gomail.SendMail(payload.Email, "Welcome to Building Management Binus (Login Information)", general.ParseTemplateEmailToHtml("./assets/html/email/notif_login_info.html", struct {
			NAME     string
			EMAIL    string
			PASSWORD string
			LINK     string
		}{
			NAME:     payload.Name,
			EMAIL:    payload.Email,
			PASSWORD: passwordString,
			LINK:     constant.BASE_URL_UI,
		})); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success create!",
	}, nil
}

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	var res []map[string]interface{} = nil
	data, err := s.UserRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.UserRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	for _, v := range data {
		res = append(res, map[string]interface{}{
			"id":         v.ID,
			"name":       v.Name,
			"email":      v.Email,
			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*v.UpdatedAt),
			"role": map[string]interface{}{
				"id":   v.Role.ID,
				"name": v.Role.Name,
			},
		})
	}
	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}

func (s *service) FindById(ctx *abstraction.Context, payload *dto.UserFindByIDRequest) (map[string]interface{}, error) {
	var res map[string]interface{} = nil
	data, err := s.UserRepository.FindById(ctx, payload.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if data != nil {
		res = map[string]interface{}{
			"id":         data.ID,
			"name":       data.Name,
			"email":      data.Email,
			"is_delete":  data.IsDelete,
			"created_at": general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
			"role": map[string]interface{}{
				"id":   data.Role.ID,
				"name": data.Role.Name,
			},
		}

	}
	return map[string]interface{}{
		"data": res,
	}, nil
}

func (s *service) Update(ctx *abstraction.Context, payload *dto.UserUpdateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = payload.ID
		if payload.Name != nil {
			newUserData.Name = *payload.Name
		}
		if payload.Email != nil {
			newUserData.Email = *payload.Email
		}
		if payload.RoleId != nil {
			roleData, err := s.RoleRepository.FindById(ctx, *payload.RoleId)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if roleData == nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "role not found")
			}
			newUserData.RoleId = *payload.RoleId
		}

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success update!",
	}, nil
}

func (s *service) Delete(ctx *abstraction.Context, payload *dto.UserDeleteByIDRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_BM {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = userData.ID
		newUserData.IsDelete = true

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}

func (s *service) ChangePassword(ctx *abstraction.Context, payload *dto.UserChangePasswordRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.ID != payload.ID {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this user is not permitted")
		}

		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		if err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(payload.OldPassword)); err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, err, "old password is wrong")
		}

		if payload.OldPassword == payload.NewPassword {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "the new password cannot be the same as the old password")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = userData.ID
		newUserData.Password = string(hashedPassword)

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "success change password!",
	}, nil
}

func (s *service) Export(ctx *abstraction.Context, payload *dto.UserExportRequest) (string, *bytes.Buffer, string, error) {
	data, err := s.UserRepository.Find(ctx, true)
	if err != nil && err.Error() != "record not found" {
		return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	if payload.Format == "pdf" {
		pdf := gofpdf.New("L", "mm", "A4", "")
		pdf.SetMargins(10, 10, 10)
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, "Building Management Binus - Laporan Data Pengguna")
		pdf.Ln(12)

		pdf.SetFont("Arial", "B", 10)
		header := []string{
			"No", "Nama", "Email", "Peran", "Tanggal Terdaftar",
		}
		colWidths := []float64{10, 50, 60, 50, 60}

		for i, str := range header {
			pdf.CellFormat(colWidths[i], 8, str, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 9)
		lineHeight := 5.0

		for i, v := range data {
			no := fmt.Sprintf("%d", i+1)
			row := []string{
				no,
				v.Name,
				v.Email,
				v.Role.Name,
				general.ConvertDateTimeToIndonesian(v.CreatedAt.Format("2006-01-02 15:04:05")),
			}

			startX := pdf.GetX()
			startY := pdf.GetY()

			maxHeight := 0.0
			for j, txt := range row {
				lines := pdf.SplitLines([]byte(txt), colWidths[j])
				h := float64(len(lines)) * lineHeight
				if h > maxHeight {
					maxHeight = h
				}
			}

			x := startX
			for j, txt := range row {
				pdf.Rect(x, startY, colWidths[j], maxHeight, "")
				pdf.SetXY(x, startY)
				pdf.MultiCell(colWidths[j], lineHeight, txt, "", "", false)

				x += colWidths[j]
			}

			pdf.SetXY(startX, startY+maxHeight)
		}

		var buf bytes.Buffer
		if err := pdf.Output(&buf); err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		filename := "Building Management Binus - Laporan Data Pengguna.pdf"
		return filename, &buf, "pdf", nil
	} else {
		f := excelize.NewFile()
		sheet := "BM Binus"
		index, err := f.NewSheet(general.TruncateSheetName(sheet))
		if err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		f.DeleteSheet("Sheet1")
		f.SetActiveSheet(index)
		f.SetCellValue(sheet, "A1", "No")
		f.SetCellValue(sheet, "B1", "Nama")
		f.SetCellValue(sheet, "C1", "Email")
		f.SetCellValue(sheet, "D1", "Peran")
		f.SetCellValue(sheet, "E1", "Tanggal Terdaftar")
		for i, v := range data {
			colA := fmt.Sprintf("A%d", i+2)
			colB := fmt.Sprintf("B%d", i+2)
			colC := fmt.Sprintf("C%d", i+2)
			colD := fmt.Sprintf("D%d", i+2)
			colE := fmt.Sprintf("E%d", i+2)
			no := i + 1
			f.SetCellValue(sheet, colA, no)
			f.SetCellValue(sheet, colB, v.Name)
			f.SetCellValue(sheet, colC, v.Email)
			f.SetCellValue(sheet, colD, v.Role.Name)
			f.SetCellValue(sheet, colE, general.ConvertDateTimeToIndonesian(v.CreatedAt.Format("2006-01-02 15:04:05")))
		}

		styleID, _ := f.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				WrapText: true,
				Vertical: "top",
			},
		})
		f.SetCellStyle(sheet, "A1", fmt.Sprintf("E%d", len(data)+1), styleID)

		cols := []string{"A", "B", "C", "D", "E"}
		lastRow := len(data) + 1

		for _, col := range cols {
			maxLen := 0
			headerCell := fmt.Sprintf("%s%d", col, 1)
			if val, err := f.GetCellValue(sheet, headerCell); err == nil {
				l := utf8.RuneCountInString(val)
				if l > maxLen {
					maxLen = l
				}
			}
			for r := 2; r <= lastRow; r++ {
				cell := fmt.Sprintf("%s%d", col, r)
				if val, err := f.GetCellValue(sheet, cell); err == nil {
					lines := strings.Split(val, "\n")
					for _, ln := range lines {
						l := utf8.RuneCountInString(ln)
						if l > maxLen {
							maxLen = l
						}
					}
				}
			}
			factor := 1.1
			padding := 2.0
			minWidth := 10.0
			maxWidth := 80.0

			width := float64(maxLen)*factor + padding
			if width < minWidth {
				width = minWidth
			}
			if width > maxWidth {
				width = maxWidth
			}

			_ = f.SetColWidth(sheet, col, col, width)
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		filename := "Building Management Binus - Laporan Data Pengguna.xlsx"
		return filename, &buf, "excel", nil
	}
}

func (s *service) Info(ctx *abstraction.Context) (map[string]interface{}, error) {
	var res map[string]interface{} = nil
	data, err := s.UserRepository.FindById(ctx, ctx.Auth.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if data != nil {
		res = map[string]interface{}{
			"id":         data.ID,
			"name":       data.Name,
			"email":      data.Email,
			"created_at": general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
			"role": map[string]interface{}{
				"id":   data.Role.ID,
				"name": data.Role.Name,
			},
		}

	}
	return map[string]interface{}{
		"data": res,
	}, nil
}
