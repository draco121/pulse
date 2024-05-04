package core

import (
	"context"
	"fmt"
	"github.com/draco121/horizon/models"
	"github.com/draco121/horizon/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"mime/multipart"
	"path"
	"slices"
	"trainingservice/repository"
)

type ITrainingService interface {
	UploadTrainingFiles(ctx context.Context, botId string, projectId string, files []*multipart.FileHeader) error
	DeleteFile(ctx context.Context, botId string, projectId string, fileId primitive.ObjectID) error
	GetFile(ctx context.Context, botId string, projectId string, fileId primitive.ObjectID) (string, string, error)
	AddTrainingData(ctx context.Context, trainingData *models.TrainingData) (*models.TrainingData, error)
	GetTrainingData(ctx context.Context, botId primitive.ObjectID, projectId primitive.ObjectID) (*models.TrainingData, error)
	UpdateTrainingData(ctx context.Context, trainingData *models.TrainingData) (*models.TrainingData, error)
	ResetTrainingData(ctx context.Context, botId primitive.ObjectID, projectId primitive.ObjectID) (*models.TrainingData, error)
}

type trainingService struct {
	client *mongo.Client
	repo   repository.ITrainingRepository
}

func NewTrainingService(client *mongo.Client, repo repository.ITrainingRepository) ITrainingService {
	return &trainingService{
		client: client,
		repo:   repo,
	}
}

func (s *trainingService) UploadTrainingFiles(ctx context.Context, botId string, projectId string, files []*multipart.FileHeader) error {
	mongoSession, err := s.client.StartSession()
	if err != nil {
		utils.Logger.Error("failed to start mongo mongoSession", "error: ", err.Error())
		return err
	}
	defer mongoSession.EndSession(ctx)
	err = mongoSession.StartTransaction()
	if err != nil {
		utils.Logger.Error("failed to start mongo transaction", "error: ", err.Error())
		return err
	}
	bid, err := primitive.ObjectIDFromHex(botId)
	if err != nil {
		utils.Logger.Error("unable to fetch bot training data wrong bot id error: ", err.Error())
		return err
	}
	pid, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		utils.Logger.Error("unable to fetch bot training data wrong project id error: ", err.Error())
		return err
	}
	trainingData, err := s.repo.FindOneByBotId(ctx, bid, pid)
	if err != nil {
		utils.Logger.Info("failed to find training data by bot id error: ", err.Error())
		utils.Logger.Info("creating new training data by bot id")
		trainingData = &models.TrainingData{
			BotId:     bid,
			ProjectId: pid,
			ID:        primitive.NewObjectID(),
		}
		var filesData []models.Files
		for _, file := range files {
			f := models.Files{
				FileId:    primitive.NewObjectID(),
				FileName:  file.Filename,
				Extension: path.Ext(file.Filename),
			}
			file.Filename = f.FileId.Hex() + f.Extension
			err = s.repo.SaveFile(ctx, botId, projectId, file)
			if err != nil {
				utils.Logger.Error("could not save file error ", err.Error())
				return err
			}
			filesData = append(filesData, f)
		}
		trainingData.Files = filesData
		_, err = s.repo.InsertOne(ctx, trainingData)
		if err != nil {
			utils.Logger.Error("failed to create training data error ", err.Error())
			return err
		} else {
			utils.Logger.Info("created training data successfully")
			_ = mongoSession.CommitTransaction(ctx)
			return nil
		}
	} else {
		var filesData []models.Files
		for _, file := range files {
			f := models.Files{
				FileId:    primitive.NewObjectID(),
				FileName:  file.Filename,
				Extension: path.Ext(file.Filename),
			}
			file.Filename = f.FileId.Hex() + f.Extension
			err = s.repo.SaveFile(ctx, botId, projectId, file)
			if err != nil {
				utils.Logger.Error("could not save file error ", err.Error())
				return err
			}
			filesData = append(filesData, f)
		}
		trainingData.Files = filesData
		_, err = s.repo.InsertOne(ctx, trainingData)
		if err != nil {
			utils.Logger.Error("failed to create training data error ", err.Error())
			return err
		} else {
			utils.Logger.Info("created training data successfully")
			_ = mongoSession.CommitTransaction(ctx)
			return nil
		}
	}
}

func (s *trainingService) DeleteFile(ctx context.Context, botId string, projectId string, fileId primitive.ObjectID) error {
	mongoSession, err := s.client.StartSession()
	if err != nil {
		utils.Logger.Error("failed to start mongo mongoSession", "error: ", err.Error())
		return err
	}
	defer mongoSession.EndSession(ctx)
	err = mongoSession.StartTransaction()
	if err != nil {
		utils.Logger.Error("failed to start mongo transaction", "error: ", err.Error())
		return err
	}
	bid, err := primitive.ObjectIDFromHex(botId)
	if err != nil {
		utils.Logger.Error("unable to fetch bot training data wrong bot id error: ", err.Error())
		return err
	}
	pid, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		utils.Logger.Error("unable to fetch bot training data wrong project id error: ", err.Error())
		return err
	}
	trainingData, err := s.repo.FindOneByBotId(ctx, bid, pid)
	if err != nil {
		utils.Logger.Error("no training data found error ", err.Error())
		return err
	} else {
		if len(trainingData.Files) <= 0 {
			return fmt.Errorf("no files are uploaded yet")
		} else {
			for i, file := range trainingData.Files {
				if file.FileId == fileId {
					err = s.repo.DeleteFile(ctx, botId, projectId, fileId)
					if err != nil {
						utils.Logger.Error("failed to delete file error ", err.Error())
						return err
					} else {
						utils.Logger.Info("file deleted successfully")
						trainingData.Files = slices.Delete(trainingData.Files, i, i+1)
						_, err = s.repo.UpdateOne(ctx, trainingData)
						if err != nil {
							utils.Logger.Error("failed to update training data after deleting file, error ", err.Error())
						} else {
							_ = mongoSession.CommitTransaction(ctx)
							utils.Logger.Info("file deleted successfully")
							return nil
						}
					}
				}
			}
		}
		utils.Logger.Debug("file not found")
		return fmt.Errorf("file not found with given Id")
	}
}

func (s *trainingService) GetFile(ctx context.Context, botId string, projectId string, fileId primitive.ObjectID) (string, string, error) {
	mongoSession, err := s.client.StartSession()
	if err != nil {
		utils.Logger.Error("failed to start mongo mongoSession", "error: ", err.Error())
		return "", "", err
	}
	defer mongoSession.EndSession(ctx)
	err = mongoSession.StartTransaction()
	if err != nil {
		utils.Logger.Error("failed to start mongo transaction", "error: ", err.Error())
		return "", "", err
	}
	bid, err := primitive.ObjectIDFromHex(botId)
	if err != nil {
		utils.Logger.Error("unable to fetch bot training data wrong bot id error: ", err.Error())
		return "", "", err
	}
	pid, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		utils.Logger.Error("unable to fetch bot training data wrong project id error: ", err.Error())
		return "", "", err
	}
	trainingData, err := s.repo.FindOneByBotId(ctx, bid, pid)
	if err != nil {
		utils.Logger.Error("no training data found error ", err.Error())
		return "", "", err
	} else {
		if len(trainingData.Files) <= 0 {
			return "", "", fmt.Errorf("no files are uploaded yet")
		} else {
			for _, file := range trainingData.Files {
				if file.FileId == fileId {
					botSpace, err := utils.CreateBotSpace(botId, projectId)
					if err != nil {
						utils.Logger.Error("unable to access bot storage space")
						return "", "", err
					} else {
						filePath := path.Join(botSpace, file.FileId.Hex()+file.Extension)
						_ = mongoSession.CommitTransaction(ctx)
						utils.Logger.Info("successfully fetched the file details")
						return file.FileName, filePath, nil
					}
				}
			}
		}
		utils.Logger.Debug("file not found")
		return "", "", fmt.Errorf("file not found with given Id")
	}
}

func (s *trainingService) AddTrainingData(ctx context.Context, trainingData *models.TrainingData) (*models.TrainingData, error) {
	mongoSession, err := s.client.StartSession()
	if err != nil {
		utils.Logger.Error("failed to start mongo mongoSession", "error: ", err.Error())
		return nil, err
	}
	defer mongoSession.EndSession(ctx)
	err = mongoSession.StartTransaction()
	if err != nil {
		utils.Logger.Error("failed to start mongo transaction", "error: ", err.Error())
		return nil, err
	}
	td, err := s.repo.InsertOne(ctx, trainingData)
	if err != nil {
		utils.Logger.Error("failed to insert training data into db", "error: ", err.Error())
		return nil, err
	} else {
		_ = mongoSession.CommitTransaction(ctx)
		utils.Logger.Info("successfully inserted training data into db")
		return td, nil
	}
}

func (s *trainingService) GetTrainingData(ctx context.Context, botId primitive.ObjectID, projectId primitive.ObjectID) (*models.TrainingData, error) {
	mongoSession, err := s.client.StartSession()
	if err != nil {
		utils.Logger.Error("failed to start mongo mongoSession", "error: ", err.Error())
		return nil, err
	}
	defer mongoSession.EndSession(ctx)
	err = mongoSession.StartTransaction()
	if err != nil {
		utils.Logger.Error("failed to start mongo transaction", "error: ", err.Error())
		return nil, err
	}
	td, err := s.repo.FindOneByBotId(ctx, botId, projectId)
	if err != nil {
		utils.Logger.Error("failed to fetch training data from db", "error: ", err.Error())
		return nil, err
	} else {
		_ = mongoSession.CommitTransaction(ctx)
		utils.Logger.Info("successfully inserted training data into db")
		return td, nil
	}
}

func (s *trainingService) UpdateTrainingData(ctx context.Context, trainingData *models.TrainingData) (*models.TrainingData, error) {
	mongoSession, err := s.client.StartSession()
	if err != nil {
		utils.Logger.Error("failed to start mongo mongoSession", "error: ", err.Error())
		return nil, err
	}
	defer mongoSession.EndSession(ctx)
	err = mongoSession.StartTransaction()
	if err != nil {
		utils.Logger.Error("failed to start mongo transaction", "error: ", err.Error())
		return nil, err
	}
	td, err := s.repo.UpdateOne(ctx, trainingData)
	if err != nil {
		utils.Logger.Error("failed to update training data from db", "error: ", err.Error())
		return nil, err
	} else {
		_ = mongoSession.CommitTransaction(ctx)
		utils.Logger.Info("successfully updated training data into db")
		return td, nil
	}
}

func (s *trainingService) ResetTrainingData(ctx context.Context, botId primitive.ObjectID, projectId primitive.ObjectID) (*models.TrainingData, error) {
	mongoSession, err := s.client.StartSession()
	if err != nil {
		utils.Logger.Error("failed to start mongo mongoSession", "error: ", err.Error())
		return nil, err
	}
	defer mongoSession.EndSession(ctx)
	err = mongoSession.StartTransaction()
	if err != nil {
		utils.Logger.Error("failed to start mongo transaction", "error: ", err.Error())
		return nil, err
	}
	td, err := s.repo.DeleteOneByBotId(ctx, botId, projectId)
	if err != nil {
		utils.Logger.Error("failed to delete training data from db", "error: ", err.Error())
		return nil, err
	} else {
		_ = mongoSession.CommitTransaction(ctx)
		utils.Logger.Info("successfully deleted training data into db")
		return td, nil
	}
}
