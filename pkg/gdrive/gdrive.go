package gdrive

import (
	"bm_binus/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// connection
func InitService() (*drive.Service, error) {
	credentialsJson := config.Get().Drive.CredentialsDrive

	config, err := google.ConfigFromJSON([]byte(credentialsJson), drive.DriveScope)

	if err != nil {
		return nil, err
	}

	client := getClient(config)

	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		logrus.Printf("Cannot create the Google Drive service: %v\n", err)
		return nil, err
	}

	logrus.Info("Google Drive ready!")
	return service, nil
}

func InitFolder(service *drive.Service, folder string, parentId string) (*drive.File, error) {
	folderUsed, err := CheckFolderByName(service, folder, parentId)
	if err != nil {
		logrus.Printf("Cannot check folder by name from Google Drive: %v\n", err)
		return nil, err
	}

	if folderUsed == nil {
		folderCreated, err := CreateFolder(service, folder, parentId)
		if err != nil {
			logrus.Printf("Cannot create folder to Google Drive: %v\n", err)
			return nil, err
		}

		logrus.Infof("Folder %s created!", folder)
		return folderCreated, nil
	} else {
		query := fmt.Sprintf("'%s' in parents and trashed = false and mimeType != 'application/vnd.google-apps.folder'", folderUsed.Id)
		fileList, err := service.Files.List().Q(query).Fields("files(id)").Do()
		if err != nil {
			logrus.Printf("Cannot list files in folder: %v\n", err)
		} else {
			logrus.Infof("Folder %s ready! Contains %d file(s).", folder, len(fileList.Files))
		}
		return folderUsed, nil
	}
}

func getClient(config *oauth2.Config) *http.Client {
	tok, err := tokenFromEnv()
	if err != nil {
		tok = getTokenFromWeb(config)
		saveTokenToEnv(tok)
		logrus.Info("Regenerate token Google Drive!")
	}
	logrus.Info("Google Drive client found!")
	return config.Client(context.Background(), tok)
}

func tokenFromEnv() (*oauth2.Token, error) {
	tokenJSON := os.Getenv("TOKEN_DRIVE")
	if tokenJSON == "" {
		return nil, fmt.Errorf("TOKEN_DRIVE environment variable is not set")
	}
	tok := &oauth2.Token{}
	err := json.Unmarshal([]byte(tokenJSON), tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	refreshToken := os.Getenv("REFRESH_DRIVE")
	if refreshToken == "" {
		logrus.Printf("REFRESH_TOKEN environment variable is not set")
		return nil
	}

	tok := &oauth2.Token{RefreshToken: refreshToken}
	tokSource := config.TokenSource(context.Background(), tok)

	newToken, err := tokSource.Token()
	if err != nil {
		logrus.Printf("Unable to retrieve token from web: %v", err)
		return nil
	}

	return newToken
}

func saveTokenToEnv(token *oauth2.Token) {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		logrus.Printf("Failed to marshal token: %v", err)
		return
	}
	os.Setenv("TOKEN_DRIVE", string(tokenJSON))
}

// action
func CheckFolderByName(service *drive.Service, name string, parentId string) (*drive.File, error) {
	query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and '%s' in parents and trashed=false", name, parentId)

	filesList, err := service.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search folder: %v", err)
	}

	if len(filesList.Files) > 0 {
		return filesList.Files[0], nil
	}

	return nil, nil
}

func CreateFolder(service *drive.Service, name string, parentId string) (*drive.File, error) {
	d := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentId},
	}

	folder, err := service.Files.Create(d).Do()

	if err != nil {
		logrus.Println("Could not create dir: " + err.Error())
		return nil, err
	}

	permission := &drive.Permission{
		Role: "reader",
		Type: "anyone",
	}
	_, err = service.Permissions.Create(folder.Id, permission).Do()
	if err != nil {
		logrus.Println("Could not set permission: " + err.Error())
		return nil, err
	}

	return folder, nil
}

func CreateFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (*drive.File, error) {
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}
	file, err := service.Files.Create(f).Media(content).Do()

	if err != nil {
		logrus.Println("Could not create file: " + err.Error())
		return nil, err
	}

	permission := &drive.Permission{
		Role: "reader",
		Type: "anyone",
	}
	_, err = service.Permissions.Create(file.Id, permission).Do()
	if err != nil {
		logrus.Println("Could not set permission: " + err.Error())
		return nil, err
	}

	logrus.Println("File successfully created")
	return file, nil
}

func GetFile(service *drive.Service, fileID string) (*drive.File, error) {
	file, err := service.Files.Get(fileID).Fields("*").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve file with ID %s: %v", fileID, err)
	}
	return file, nil
}

func DeleteFile(service *drive.Service, fileID string) error {
	err := service.Files.Delete(fileID).Do()
	if err != nil {
		return fmt.Errorf("unable to delete file with ID %s: %v", fileID, err)
	}
	logrus.Println("File successfully deleted")
	return nil
}

func RenameFile(service *drive.Service, fileID, newName string) (*drive.File, error) {
	file := &drive.File{Name: newName}

	updatedFile, err := service.Files.Update(fileID, file).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to rename file with ID %s: %v", fileID, err)
	}

	return updatedFile, nil
}
