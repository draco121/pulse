package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/draco121/common/models"
	"github.com/draco121/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"mime/multipart"
	"os"
	"path"
	"slices"
)

type ITrainingRepository interface {
	FindOneByBotId(ctx context.Context, botId primitive.ObjectID, projectId primitive.ObjectID) (*models.TrainingData, error)
	UpdateOne(ctx context.Context, trainingData *models.TrainingData) (*models.TrainingData, error)
	DeleteOneByBotId(ctx context.Context, botId primitive.ObjectID, projectId primitive.ObjectID) (*models.TrainingData, error)
	InsertOne(ctx context.Context, trainingData *models.TrainingData) (*models.TrainingData, error)
	SaveFile(ctx context.Context, botId string, projectId string, file *multipart.FileHeader) error
	DeleteFile(ctx context.Context, botId string, projectId string, fileId primitive.ObjectID) error
	GetFile(ctx context.Context, botId string, projectId string, fileId primitive.ObjectID) (string, error)
}

type trainingRepository struct {
	ITrainingRepository
	db *mongo.Database
}

func NewTrainingRepository(db *mongo.Database) ITrainingRepository {
	return &trainingRepository{
		db: db,
	}
}

func (ur *trainingRepository) InsertOne(ctx context.Context, trainingData *models.TrainingData) (*models.TrainingData, error) {
	ownerId := ctx.Value("UserId").(primitive.ObjectID)
	result, _ := ur.FindOneByBotId(ctx, trainingData.BotId, trainingData.ProjectId)
	if result != nil {
		return nil, fmt.Errorf("record exists")
	} else {
		trainingData.ID = primitive.NewObjectID()
		trainingData.Owner = ownerId
		_, err := ur.db.Collection("training-data").InsertOne(ctx, trainingData)
		if err != nil {
			return nil, err
		}
	}
	return trainingData, nil
}

func (ur *trainingRepository) UpdateOne(ctx context.Context, trainingData *models.TrainingData) (*models.TrainingData, error) {
	ownerId := ctx.Value("UserId").(primitive.ObjectID)
	filter := bson.D{{Key: "_id", Value: trainingData.ID}, {Key: "owner", Value: ownerId}}
	trainingData.Owner = ownerId
	update := bson.M{"$set": trainingData}
	result := models.TrainingData{}
	err := ur.db.Collection("training-data").FindOneAndUpdate(ctx, filter, update).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	} else {
		return &result, nil
	}
}

func (ur *trainingRepository) FindOneByBotId(ctx context.Context, botId primitive.ObjectID, projectId primitive.ObjectID) (*models.TrainingData, error) {
	userId := ctx.Value("UserId").(primitive.ObjectID)
	filter := bson.D{{Key: "botId", Value: botId}, {Key: "owner", Value: userId}, {Key: "projectId", Value: projectId}}
	result := models.TrainingData{}
	err := ur.db.Collection("training-data").FindOne(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	} else {
		return &result, nil
	}

}

func (ur *trainingRepository) DeleteOneByBotId(ctx context.Context, botId primitive.ObjectID, projectId primitive.ObjectID) (*models.TrainingData, error) {
	ownerId := ctx.Value("UserId").(primitive.ObjectID)
	filter := bson.D{{Key: "botId", Value: botId}, {Key: "owner", Value: ownerId}, {Key: "projectId", Value: projectId}}
	result := models.TrainingData{}
	err := ur.db.Collection("training-data").FindOneAndDelete(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	} else {
		return &result, nil
	}
}

func (ur *trainingRepository) SaveFile(ctx context.Context, botId string, projectId string, file *multipart.FileHeader) error {
	botPath, err := utils.CreateBotSpace(botId, projectId)
	if err != nil {
		return err
	} else {
		newFile, err := os.Create(path.Join(botPath, file.Filename))
		if err != nil {
			return err
		} else {
			reader, err := file.Open()
			if err != nil {
				return err
			} else {
				_, err = io.Copy(newFile, reader)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}
}

func (ur *trainingRepository) DeleteFile(ctx context.Context, botId string, projectId string, fileId primitive.ObjectID) error {
	bid, err := primitive.ObjectIDFromHex(botId)
	if err != nil {
		return err
	} else {
		pid, err := primitive.ObjectIDFromHex(projectId)
		if err != nil {
			return err
		}
		td, err := ur.FindOneByBotId(ctx, bid, pid)
		if err != nil {
			return err
		} else {
			for _, j := range td.Files {
				if slices.Contains(td.Files, j) {
					botSpace, err := utils.CreateBotSpace(projectId, botId)
					if err != nil {
						return err
					} else {
						filePath := path.Join(botSpace, fileId.Hex()+j.Extension)
						err = os.Remove(filePath)
						if err != nil {
							return err
						} else {
							return nil
						}
					}
				} else {
					return fmt.Errorf("file not exists")
				}
			}
		}
		return nil
	}
}

func (ur *trainingRepository) GetFile(ctx context.Context, botId string, projectId string, fileId primitive.ObjectID) (string, error) {
	bid, err := primitive.ObjectIDFromHex(botId)
	if err != nil {
		return "", err
	} else {
		pid, err := primitive.ObjectIDFromHex(projectId)
		if err != nil {
			return "", err
		}
		td, err := ur.FindOneByBotId(ctx, bid, pid)
		if err != nil {
			return "", err
		} else {
			for _, j := range td.Files {
				if slices.Contains(td.Files, j) {
					botSpace, err := utils.CreateBotSpace(projectId, botId)
					if err != nil {
						return "", err
					} else {
						filePath := path.Join(botSpace, fileId.Hex()+j.Extension)
						return filePath, nil
					}
				} else {
					return "", fmt.Errorf("file not exists")
				}
			}
		}
		return "", nil
	}
}
