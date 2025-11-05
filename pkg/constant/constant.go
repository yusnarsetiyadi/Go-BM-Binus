package constant

const (
	DRIVE_FOLDER = "BM_Binus_App"

	ROLE_ID_STAF  = 1
	ROLE_ID_BM    = 2
	ROLE_ID_ADMIN = 3

	STATUS_ID_PENGAJUAN  = 1
	STATUS_ID_VALIDASI   = 2
	STATUS_ID_PROSES     = 3
	STATUS_ID_FINALISASI = 4
	STATUS_ID_SELESAI    = 5

	BLANK_REQUEST_ID = 1

	REDIS_REQUEST_IP_KEYS        = "bmbinus-reset-password:ip:%s"
	REDIS_REQUEST_MAX_ATTEMPTS   = 10
	REDIS_REQUEST_IP_EXPIRE      = 240
	REDIS_KEY_USER_LOGIN         = "bmbinus_login_token_user_"
	REDIS_KEY_AUTO_LOGOUT        = "bmbinus_user_auto_logout"
	REDIS_KEY_REFRESH_TOKEN      = "bmbinus-refresh-token:%s"
	REDIS_MAX_REFRESH_TOKEN      = 30
	REDIS_KEY_USE_PRIORITY_COUNT = "use_priority_count"

	PATH_FILE_SAVED    = "../file_saved"
	PATH_ASSETS_IMAGES = "assets/images"
	PATH_SHARE         = "/var/www/html/bm_binus/share"
)

var (
	BASE_URL    string = ""
	BASE_URL_UI string = "https://bmbinus.my.id/"
)
