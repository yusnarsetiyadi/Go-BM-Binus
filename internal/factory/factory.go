package factory

import (
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/database"
	"bm_binus/pkg/gdrive"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/drive/v3"
	"gorm.io/gorm"
)

type Factory struct {
	Db *gorm.DB

	DbRedis *redis.Client

	GDrive GoogleDrive

	Repository_initiated
}

type Repository_initiated struct {
	UserRepository         repository.User
	RoleRepository         repository.Role
	StatusRepository       repository.Status
	NotificationRepository repository.Notification
	RequestRepository      repository.Request
	EventTypeRepository    repository.EventType
	FileRepository         repository.File
	CommentRepository      repository.Comment
	AhpHistoryRepository   repository.AhpHistory
}

type GoogleDrive struct {
	Service  *drive.Service
	FolderBM *drive.File
}

func NewFactory() *Factory {
	f := &Factory{}
	f.SetupDb()
	f.SetupDbRedis()
	f.SetupGoogleDrive()
	f.SetupRepository()
	return f
}

func (f *Factory) SetupDb() {
	db, err := database.Connection("MYSQL")
	if err != nil {
		panic("Failed setup db, connection is undefined")
	}

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	panic(err)
	// }
	// sqlDB.SetMaxIdleConns(5)
	// sqlDB.SetMaxOpenConns(20)
	// sqlDB.SetConnMaxLifetime(time.Hour)

	f.Db = db
}

func (f *Factory) SetupDbRedis() {
	dbRedis := database.InitRedis()
	f.DbRedis = dbRedis
}

func (f *Factory) SetupGoogleDrive() {
	service, err := gdrive.InitService()
	if err != nil {
		panic("Failed setup gdrive, connection is undefined")
	}
	folderBm, err := gdrive.InitFolder(service, constant.DRIVE_FOLDER, "root")
	if err != nil {
		logrus.Infof("Failed setup folder %s, cause: %s", constant.DRIVE_FOLDER, err.Error())
	}
	f.GDrive.Service = service
	f.GDrive.FolderBM = folderBm
}

func (f *Factory) SetupRepository() {
	if f.Db == nil {
		panic("Failed setup repository, db is undefined")
	}

	f.UserRepository = repository.NewUser(f.Db)
	f.RoleRepository = repository.NewRole(f.Db)
	f.StatusRepository = repository.NewStatus(f.Db)
	f.NotificationRepository = repository.NewNotification(f.Db)
	f.RequestRepository = repository.NewRequest(f.Db)
	f.EventTypeRepository = repository.NewEventType(f.Db)
	f.FileRepository = repository.NewFile(f.Db)
	f.CommentRepository = repository.NewComment(f.Db)
	f.AhpHistoryRepository = repository.NewAhpHistory(f.Db)
}
