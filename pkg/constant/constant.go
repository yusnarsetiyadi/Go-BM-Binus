package constant

const (
	DRIVE_FOLDER = "BM_Binus_App"

	ROLE_ID_ADMIN = 1
	ROLE_ID_STAFF = 2

	REDIS_REQUEST_IP_KEYS      = "reset-password:ip:%s"
	REDIS_REQUEST_MAX_ATTEMPTS = 5
	REDIS_REQUEST_IP_EXPIRE    = 240
	REDIS_KEY_USER_LOGIN       = "login_token_user_"
	REDIS_KEY_AUTO_LOGOUT      = "user_auto_logout"
	REDIS_KEY_REFRESH_TOKEN    = "refresh-token:%s"
	REDIS_MAX_REFRESH_TOKEN    = 30

	PATH_FILE_SAVED    = "../file_saved"
	PATH_ASSETS_IMAGES = "assets/images"
	PATH_SHARE         = "/var/www/html/bm_binus/share"
)

var (
	BASE_URL    string = ""
	BASE_URL_UI string = "https://bmbinus.my.id/"
)
