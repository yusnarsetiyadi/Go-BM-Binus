package general

import (
	"bm_binus/internal/abstraction"
	"bm_binus/pkg/constant"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// validation email
func IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// validation phone
func IsValidPhone(phone string) bool {
	phoneNumberRegex := `^\+[1-9]\d{1,14}$`
	re := regexp.MustCompile(phoneNumberRegex)
	return re.MatchString(phone)
}

// Now ...
func Now() *time.Time {
	now := time.Now()
	return &now
}

// NowUTC ...
func NowUTC() *time.Time {
	now := time.Now().UTC()
	return &now
}

// NowLocal ...
func NowLocal() *time.Time {
	now := time.Now().UTC().Add(time.Hour * 7)
	return &now
}

// NowWithLocation ...
func NowWithLocation() *time.Time {
	now := time.Now().In(Location())
	return &now
}

// Location ...
func Location() *time.Location {
	return time.FixedZone("Asia/Jakarta", 7*60*60)
}

func Parse(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, Location())
}

// LastWeek ...
func LastWeek(now time.Time) (start time.Time, end time.Time) {
	end = StartOfWeek(now).Add(-1)

	oneWeek := (24 * 6) * time.Hour
	start = StartOfDay(end.Add(-oneWeek))
	return
}

// LastMonth ...
func LastMonth(now time.Time) (time.Time, time.Time) {
	end := StartOfMonth(now).Add(-time.Nanosecond)
	return StartOfMonth(end), end
}

// StartOfMonth ...
func StartOfMonth(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

// StartOfWeek ...
func StartOfWeek(now time.Time) time.Time {
	wd := now.Weekday()
	if wd == time.Sunday {
		now = now.AddDate(0, 0, -6)
	} else {
		now = now.AddDate(0, 0, -int(wd)+1)
	}
	return StartOfDay(now)
}

// StartOfDay ...
func StartOfDay(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// EndOfDay ...
func EndOfDay(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-1), now.Location())
}

func IsToday(t time.Time) bool {
	// Dapatkan tanggal hari ini
	now := NowLocal()

	// Bandingkan tahun, bulan, dan hari
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

func RandSeq(n int) string {
	var letters = []rune("123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func StringInSlice(text string, data []string) bool {
	for _, row := range data {
		if row == text {
			return true
		}
	}
	return false
}

// generate random password
func GeneratePassword(passwordLength, minSpecialChar, minNum, minUpperCase, minLowerCase int) string {
	var password strings.Builder
	var lowerCharSet string = "abcdedfghijklmnopqrstuvwxyz"
	var upperCharSet string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var specialCharSet string = "!@#$%&*"
	var numberSet string = "0123456789"
	var allCharSet string = lowerCharSet + upperCharSet + specialCharSet + numberSet

	//Set special character
	for i := 0; i < minSpecialChar; i++ {
		random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random]))
	}

	//Set numeric
	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	//Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	//Set lowercase
	for i := 0; i < minLowerCase; i++ {
		random := rand.Intn(len(lowerCharSet))
		password.WriteString(string(lowerCharSet[random]))
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase - minLowerCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

func SanitizeStringOfAlphabet(input string) string {
	// Menghapus karakter yang bukan huruf, underscore
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' {
			return r
		}
		return -1
	}, input)
}

func SanitizeStringOfNumber(input string) string {
	// Menghapus karakter yang bukan angka
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, input)
}

func SanitizeString(input string) string {
	// Define regex to remove dangerous characters
	re := regexp.MustCompile(`[%'";()=<>` + "`" + `#\-\[\]]`)
	sanitized := re.ReplaceAllString(input, "")

	// Allow letters, numbers, underscores, spaces, and additional safe characters
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' ||
			r == ' ' ||
			r == '-' ||
			r == '.' ||
			r == ':' ||
			r == '/' ||
			r == '@' {
			return r
		}
		return -1
	}, sanitized)
}

func SanitizeStringDateBetween(input string) string {
	// Define regex untuk format tanggal yang diinginkan: YYYY-MM-DD_YYYY-MM-DD
	re := regexp.MustCompile(`[^0-9\-_]`)
	// Hapus semua karakter yang tidak sesuai dengan format yang diinginkan
	sanitized := re.ReplaceAllString(input, "")

	// Pastikan bahwa input sesuai dengan format 'YYYY-MM-DD_YYYY-MM-DD'
	// regex untuk mencocokkan tanggal dengan format yang benar
	dateFormat := `^\d{4}-\d{2}-\d{2}_\d{4}-\d{2}-\d{2}$`
	dateRe := regexp.MustCompile(dateFormat)

	// Jika format tidak sesuai, kembalikan string kosong atau bisa diubah sesuai kebutuhan
	if !dateRe.MatchString(sanitized) {
		return ""
	}

	return sanitized
}

func ProcessWhereParam(ctx *abstraction.Context, searchType string, whereStr string) (string, map[string]interface{}) {
	var (
		where      = "1=@where"
		whereParam = map[string]interface{}{
			"where": 1,
			"false": false,
			"true":  true,
		}
	)

	if whereStr != "" {
		where += " AND " + whereStr
	}

	// fill query search
	if ctx.QueryParam("search") != "" {
		val := "%" + SanitizeString(ctx.QueryParam("search")) + "%"
		switch searchType {
		case "user":
			where += " AND (LOWER(name) LIKE @search_name OR LOWER(email) LIKE @search_email)"
			whereParam["search_name"] = val
			whereParam["search_email"] = val
		case "role":
			where += " AND (LOWER(name) LIKE @search_name)"
			whereParam["search_name"] = val
		case "status":
			where += " AND (LOWER(name) LIKE @search_name)"
			whereParam["search_name"] = val
		case "event_type":
			where += " AND (LOWER(name) LIKE @search_name)"
			whereParam["search_name"] = val
		case "notification":
			where += " AND (LOWER(title) LIKE @search_title OR LOWER(message) LIKE @search_message)"
			whereParam["search_title"] = val
			whereParam["search_message"] = val
		case "request":
			where += " AND (LOWER(event_name) LIKE @search_event_name OR LOWER(event_location) LIKE @search_event_location)"
			whereParam["search_event_name"] = val
			whereParam["search_event_location"] = val
		case "ahp_history":
			where += " AND (LOWER(kriteria) LIKE @search_kriteria OR LOWER(alternatif) LIKE @search_alternatif OR LOWER(kriteria_comparison) LIKE @search_kriteria_comparison OR LOWER(alternatif_comparison) LIKE @search_alternatif_comparison OR LOWER(priority_global) LIKE @search_priority_global)"
			whereParam["search_kriteria"] = val
			whereParam["search_alternatif"] = val
			whereParam["search_kriteria_comparison"] = val
			whereParam["search_alternatif_comparison"] = val
			whereParam["search_priority_global"] = val
		}
	}

	// fill query filter
	if ctx.QueryParam("id") != "" {
		val, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("id")))
		where += " AND id = @id"
		whereParam["id"] = val
	}
	if ctx.QueryParam("name") != "" {
		val := "%" + SanitizeString(ctx.QueryParam("name")) + "%"
		where += " AND LOWER(name) LIKE @name"
		whereParam["name"] = val
	}
	if ctx.QueryParam("email") != "" {
		val := "%" + SanitizeString(ctx.QueryParam("email")) + "%"
		where += " AND LOWER(email) LIKE @email"
		whereParam["email"] = val
	}
	if ctx.QueryParam("event_name") != "" {
		val := "%" + SanitizeString(ctx.QueryParam("event_name")) + "%"
		where += " AND LOWER(event_name) LIKE @event_name"
		whereParam["event_name"] = val
	}
	if ctx.QueryParam("event_location") != "" {
		val := "%" + SanitizeString(ctx.QueryParam("event_location")) + "%"
		where += " AND LOWER(event_location) LIKE @event_location"
		whereParam["event_location"] = val
	}
	if ctx.QueryParam("role_id") != "" {
		val, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("role_id")))
		where += " AND role_id = @role_id"
		whereParam["role_id"] = val
	}
	if ctx.QueryParam("event_type_id") != "" {
		val, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("event_type_id")))
		where += " AND event_type_id = @event_type_id"
		whereParam["event_type_id"] = val
	}
	if ctx.QueryParam("user_id") != "" {
		val, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("user_id")))
		where += " AND user_id = @user_id"
		whereParam["user_id"] = val
	}
	if ctx.QueryParam("status_id") != "" {
		val, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("status_id")))
		where += " AND status_id = @status_id"
		whereParam["status_id"] = val
	}
	if ctx.QueryParam("reference_request") != "" {
		val, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("reference_request")))
		where += " AND reference_request = @reference_request"
		whereParam["reference_request"] = val
	}
	if ctx.QueryParam("created_by") != "" {
		val, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("created_by")))
		where += " AND created_by = @created_by"
		whereParam["created_by"] = val
	}
	if ctx.QueryParam("updated_by") != "" {
		val, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("updated_by")))
		where += " AND updated_by = @updated_by"
		whereParam["updated_by"] = val
	}
	if ctx.QueryParam("is_read") != "" {
		where += " AND is_read = @" + SanitizeStringOfAlphabet(ctx.QueryParam("is_read"))
	}
	if ctx.QueryParam("created_at") != "" {
		val := SanitizeStringDateBetween(ctx.QueryParam("created_at"))
		valDate := strings.Split(val, "_")
		where += " AND created_at BETWEEN @start_created_at AND @end_created_at"
		whereParam["start_created_at"] = valDate[0] + " 00:00:00"
		whereParam["end_created_at"] = valDate[1] + " 23:59:59"
	}
	if ctx.QueryParam("updated_at") != "" {
		val := SanitizeStringDateBetween(ctx.QueryParam("updated_at"))
		valDate := strings.Split(val, "_")
		where += " AND updated_at BETWEEN @start_updated_at AND @end_updated_at"
		whereParam["start_updated_at"] = valDate[0] + " 00:00:00"
		whereParam["end_updated_at"] = valDate[1] + " 23:59:59"
	}
	if ctx.QueryParam("event_date_start") != "" {
		val := SanitizeStringDateBetween(ctx.QueryParam("event_date_start"))
		valDate := strings.Split(val, "_")
		where += " AND event_date_start BETWEEN @start_event_date_start AND @end_event_date_start"
		whereParam["start_event_date_start"] = valDate[0] + " 00:00:00"
		whereParam["end_event_date_start"] = valDate[1] + " 23:59:59"
	}
	if ctx.QueryParam("event_date_end") != "" {
		val := SanitizeStringDateBetween(ctx.QueryParam("event_date_end"))
		valDate := strings.Split(val, "_")
		where += " AND event_date_end BETWEEN @start_event_date_end AND @end_event_date_end"
		whereParam["start_event_date_end"] = valDate[0] + " 00:00:00"
		whereParam["end_event_date_end"] = valDate[1] + " 23:59:59"
	}

	// fill random filter
	if ctx.QueryParam("for") != "" {
		val := SanitizeString(ctx.QueryParam("for"))
		switch val {
		case "staf":
			where += " AND status_id IN (1,2,3,4,5)"
		case "bm":
			where += " AND status_id IN (1,2,3,4,5)"
		case "admin":
			where += " AND status_id IN (3,4,5)"
		}
	}

	return where, whereParam
}

func ProcessLimitOffset(ctx *abstraction.Context, no_paging bool) (int, int) {
	var (
		limit  = 10
		offset = 0
	)
	if ctx.QueryParam("limit") != "" {
		l, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("limit")))
		limit = l
	}
	if ctx.QueryParam("offset") != "" {
		o, _ := strconv.Atoi(SanitizeStringOfNumber(ctx.QueryParam("offset")))
		offset = o
	}
	if no_paging {
		limit = math.MaxInt64
	} else if ctx.QueryParam("no_paging") != "" {
		if ctx.QueryParam("no_paging") == "yes" {
			limit = math.MaxInt64
		}
	}
	return limit, offset
}

func ProcessOrder(ctx *abstraction.Context) string {
	var (
		order string
		o     = "id"
		ob    = "ASC"
	)
	if ctx.QueryParam("order") != "" {
		o = ValidationOrder(ctx.QueryParam("order"))
	}
	if ctx.QueryParam("order_by") != "" {
		ob = ValidationOrderBy(ctx.QueryParam("order_by"))
	}
	order = o + " " + ob
	return order
}

func ValidationOrder(str string) string {
	str = SanitizeString(str)
	str = strings.ToLower(str)
	orderStack := []string{"id", "name", "email", "priority", "created_at", "event_name", "event_location", "event_type_id", "count_participant", "status_id", "is_read"} // fill query order
	for _, item := range orderStack {
		if item == str {
			return str
		}
	}
	return "id"
}

func ValidationOrderBy(str string) string {
	str = SanitizeStringOfAlphabet(str)
	str = strings.ToUpper(str)
	orderStack := []string{"ASC", "DESC"}
	for _, item := range orderStack {
		if item == str {
			return str
		}
	}
	return "ASC"
}

func ParseTemplateEmailToHtml(templateFileName string, data interface{}) string {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		logrus.Error("Error paring template email: ", err.Error())
		return ""
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		logrus.Error("Error paring template email: ", err.Error())
		return ""
	}
	return buf.String()
}

func ParseTemplateEmailToPlainText(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		if n.Type == html.ElementNode {
			switch n.Data {
			case "br", "p", "div", "hr":
				buf.WriteString("\n")
			case "a":
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						buf.WriteString(" [" + attr.Val + "]")
						break
					}
				}
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	output := buf.String()
	output = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(output, "\n\n")
	return strings.TrimSpace(output)
}

func ProcessHTMLResponseEmail(filePath, placeholder, value string) string {
	content, _ := ioutil.ReadFile(filePath)
	return strings.Replace(string(content), placeholder, value, -1)
}

func ValidateImage(filename string) (bool, string) {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg"}

	ext := strings.ToLower(filepath.Ext(filename))
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	safeName := strings.ReplaceAll(nameWithoutExt, ".", "")
	safeName = strings.ReplaceAll(safeName, "/", "")
	safeName = strings.ReplaceAll(safeName, "\\", "")
	safeName = strings.ReplaceAll(safeName, " ", "_")

	timestamp := time.Now().Format("20060102150405")
	fullFileName := safeName + "_" + timestamp + ext

	for _, validExt := range imageExtensions {
		if ext == validExt {
			return true, fullFileName
		}
	}
	return false, fullFileName
}

func StringToArrayInt(ids *string) []int {
	var idArray []string

	if ids != nil {
		idArray = strings.Split(*ids, ",")
	}

	var idInts []int
	for _, id := range idArray {
		if idInt, err := strconv.Atoi(strings.TrimSpace(id)); err == nil {
			idInts = append(idInts, idInt)
		} else {
			logrus.Println("Kesalahan mengonversi ID:", err)
		}
	}

	return idInts
}

func ArrayIntToString(idInts []int) string {
	var idStrings []string
	for _, id := range idInts {
		idStrings = append(idStrings, strconv.Itoa(id))
	}

	return strings.Join(idStrings, ",")
}

func ValidateFileUpload(filename string) (bool, string) {
	fileExtensions := []string{
		// file image
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg",
		// file document
		".txt", ".pdf", ".csv", ".docx", ".xlsx", ".pptx",
		// file data
		".json", ".xml", ".yml", ".yaml", ".ini", ".log",
		// file media
		".mp3", ".wav", ".ogg", ".flac", ".mp4", ".avi", ".mkv", ".webm",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	safeName := strings.ReplaceAll(nameWithoutExt, ".", "")
	safeName = strings.ReplaceAll(safeName, "/", "")
	safeName = strings.ReplaceAll(safeName, "\\", "")
	safeName = strings.ReplaceAll(safeName, " ", "_")

	timestamp := time.Now().Format("20060102150405")
	fullFileName := safeName + "_" + timestamp + ext

	for _, validExt := range fileExtensions {
		if ext == validExt {
			return true, fullFileName
		}
	}
	return false, fullFileName
}

func GenerateKeyWatchTask(userId, taskId int) string {
	return fmt.Sprintf("user_%d_task_%d", userId, taskId)
}

func DiffIntSlices(oldSlice, newSlice []int) (added []int, removed []int) {
	oldMap := make(map[int]bool)
	newMap := make(map[int]bool)

	for _, v := range oldSlice {
		oldMap[v] = true
	}

	for _, v := range newSlice {
		newMap[v] = true
	}

	for _, v := range newSlice {
		if !oldMap[v] {
			added = append(added, v)
		}
	}

	for _, v := range oldSlice {
		if !newMap[v] {
			removed = append(removed, v)
		}
	}

	return added, removed
}

func FormatNamesFromArray(names []string) string {
	n := len(names)
	switch n {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " dan " + names[1]
	}
	return strings.Join(names[:n-1], ", ") + ", dan " + names[n-1]
}

type Color struct {
	Name    string
	R, G, B int
}

var colorList = []Color{
	{"Hitam", 0, 0, 0},
	{"Putih", 255, 255, 255},
	{"Merah", 255, 0, 0},
	{"Hijau", 0, 128, 0},
	{"Biru", 0, 0, 255},
	{"Kuning", 255, 255, 0},
	{"Oranye", 255, 165, 0},
	{"Abu-abu", 128, 128, 128},
	{"Coklat", 139, 69, 19},
	{"Hijau Tua", 31, 90, 51},
}

func GetColorNameFromCode(codeStr string) string {
	code, err := strconv.ParseUint(codeStr, 10, 32)
	if err != nil {
		return "Kode tidak valid"
	}
	argb := uint32(code)

	r := int((argb >> 16) & 0xFF)
	g := int((argb >> 8) & 0xFF)
	b := int(argb & 0xFF)

	closestName := ""
	minDistance := 1<<31 - 1 // max int
	for _, c := range colorList {
		dr := r - c.R
		dg := g - c.G
		db := b - c.B
		dist := dr*dr + dg*dg + db*db
		if dist < minDistance {
			minDistance = dist
			closestName = c.Name
		}
	}

	return closestName
}

func FormatWithZWithoutChangingTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05") + "Z"
}

func ProcessLoginFrom(input, currentLoginFrom string) string {
	if currentLoginFrom == "" {
		return input
	}
	if strings.Contains(currentLoginFrom, input) {
		return currentLoginFrom
	}
	return currentLoginFrom + " & " + input
}

func ProcessLogoutFrom(loginFrom, logoutFrom string) string {
	if logoutFrom == "web" {
		loginFrom = strings.ReplaceAll(loginFrom, "web", "")
	}
	if logoutFrom == "mobile" {
		loginFrom = strings.ReplaceAll(loginFrom, "mobile", "")
	}
	loginFrom = strings.TrimSpace(loginFrom)
	if loginFrom == "" {
		return ""
	}
	loginFrom = strings.ReplaceAll(loginFrom, " &", "")
	loginFrom = strings.ReplaceAll(loginFrom, "& ", "")
	loginFrom = strings.TrimSpace(loginFrom)
	if loginFrom == "" {
		return ""
	}
	return loginFrom
}

func SaveFileFromDriveLink(driveURL string) {
	req, _ := http.NewRequest("GET", driveURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; custom-downloader/1.0)")
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logrus.Error("Failed to download file: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Error("Failed to download file: status ", resp.Status)
		return
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition == "" {
		logrus.Error("Unable to detect file name from Content-Disposition")
		return
	}

	var fileName string
	parts := strings.Split(contentDisposition, "filename=")
	if len(parts) > 1 {
		fileName = strings.Trim(parts[1], `"`)
	} else {
		logrus.Error("File name not found in Content-Disposition")
		return
	}

	savePath := constant.PATH_FILE_SAVED
	err = os.MkdirAll(savePath, os.ModePerm)
	if err != nil {
		logrus.Error("Failed to create folder: ", err)
		return
	}

	localFilePath := filepath.Join(savePath, fileName)

	if _, err := os.Stat(localFilePath); err == nil {
		logrus.Info("File already exists: ", fileName)
		return
	}

	out, err := os.Create(localFilePath)
	if err != nil {
		logrus.Error("Failed to create file: ", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		logrus.Error("Failed to save file: ", err)
		return
	}

	logrus.Info("File downloaded and saved: ", fileName)
}

func ConvertLinkToFileSaved(driveLink, fileName, ext string) string {
	go SaveFileFromDriveLink(driveLink)
	return fmt.Sprintf("%s%s%s", constant.BASE_URL, "/file_saved/", EnsureFileExtension(fileName, ext))
}

func EnsureFileExtension(filename, ext string) string {
	currentExt := filepath.Ext(filename)
	if currentExt == "" {
		cleanExt := strings.TrimPrefix(ext, ".")
		return filename + "." + cleanExt
	}
	return filename
}

func GenerateInitial(name string) string {
	words := strings.Fields(name)
	if len(words) == 0 {
		return ""
	}
	switch len(words) {
	case 1:
		return strings.ToUpper(string([]rune(words[0])[0]))
	case 2:
		return strings.ToUpper(string([]rune(words[0])[0]) + string([]rune(words[1])[0]))
	default:
		first := string([]rune(words[0])[0])
		last := string([]rune(words[len(words)-1])[0])
		return strings.ToUpper(first + last)
	}
}

func IsValidURL(input string) string {
	parsedURL, err := url.ParseRequestURI(input)
	if err != nil {
		return ""
	}
	if (parsedURL.Scheme == "http" || parsedURL.Scheme == "https") && parsedURL.Host != "" {
		return parsedURL.String()
	}

	return ""
}

func CapitalizeEachWord(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

func GenerateRedisKeyUserLogin(userId int) string {
	return fmt.Sprintf("%s%d", constant.REDIS_KEY_USER_LOGIN, userId)
}

func GetRedisUUIDArray(client *redis.Client, key string) []string {
	val, err := client.Get(context.Background(), key).Result()
	if err != nil || val == "" {
		return []string{}
	}
	return strings.Split(val, "/")
}

func AppendUUIDToRedisArray(client *redis.Client, key string, newUUID string) {
	ctx := context.Background()
	val, err := client.Get(ctx, key).Result()
	if err != nil || val == "" {
		client.Set(ctx, key, newUUID, 0)
		return
	}
	updated := val + "/" + newUUID
	client.Set(ctx, key, updated, 0)
}

func RemoveUUIDFromRedisArray(client *redis.Client, key string, targetUUID string) {
	ctx := context.Background()
	val, err := client.Get(ctx, key).Result()
	if err != nil || val == "" {
		return
	}
	uuids := strings.Split(val, "/")
	var filtered []string
	for _, uuid := range uuids {
		if uuid != targetUUID && uuid != "" {
			filtered = append(filtered, uuid)
		}
	}
	newVal := strings.Join(filtered, "/")
	client.Set(ctx, key, newVal, 0)
}

func RemoveDuplicateArrayInt(input []int) []int {
	seen := make(map[int]bool)
	var result []int

	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func TruncateSheetName(name string) string {
	var invalidChars = regexp.MustCompile(`[\\/*?:\[\]]`)
	name = invalidChars.ReplaceAllString(name, "_")
	if utf8.RuneCountInString(name) > 31 {
		return string([]rune(name)[:31])
	}
	return name
}

func ConvertDateTimeToIndonesian(datetimeStr string) string {
	t, _ := time.Parse("2006-01-02 15:04:05", datetimeStr)
	days := []string{
		"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu",
	}
	months := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	dayName := days[int(t.Weekday())]
	monthName := months[int(t.Month())]
	formatted := fmt.Sprintf("%s, %d %s %d %02d:%02d",
		dayName, t.Day(), monthName, t.Year(), t.Hour(), t.Minute())

	return formatted
}
