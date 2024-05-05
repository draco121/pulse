package controllers

import (
	"fmt"
	"github.com/draco121/horizon/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"pulse/core"
)

type Controllers struct {
	service core.ITrainingService
}

func NewControllers(service core.ITrainingService) Controllers {
	c := Controllers{
		service: service,
	}
	return c
}

func (s Controllers) UploadTrainingData(c *gin.Context) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("projectId is required"))
		return
	}
	botId := c.Param("botId")
	if botId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("botId is required"))
		return
	}
	err := c.Request.ParseMultipartForm(30 << 20) // 10 MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	files := c.Request.MultipartForm.File["files"]
	err = s.service.UploadTrainingFiles(c, projectId, botId, files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.Status(http.StatusCreated)
	}
}

func (s Controllers) DeleteFile(c *gin.Context) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("projectId is required"))
		return
	}
	botId := c.Param("botId")
	if botId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("botId is required"))
		return
	}
	fId := c.Param("fileId")
	if fId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("fileId is required"))
		return
	}
	fileId, err := primitive.ObjectIDFromHex(fId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	err = s.service.DeleteFile(c, botId, projectId, fileId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	} else {
		c.Status(http.StatusNoContent)
		return
	}
}

func (s Controllers) GetFile(c *gin.Context) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("projectId is required"))
		return
	}
	botId := c.Param("botId")
	if botId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("botId is required"))
		return
	}
	fId := c.Param("fileId")
	if fId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("fileId is required"))
		return
	}
	fileId, err := primitive.ObjectIDFromHex(fId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	fileName, filePath, err := s.service.GetFile(c, botId, projectId, fileId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, err)
		return
	}
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	c.File(filePath)
}

func (s Controllers) AddTrainingData(c *gin.Context) {
	var trainingData *models.TrainingData
	if err := c.ShouldBind(&trainingData); err != nil {
		c.JSON(400, err.Error())
	} else {
		res, err := s.service.AddTrainingData(c, trainingData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else {
			c.JSON(http.StatusCreated, res)
		}
	}
}

func (s Controllers) GetTrainingData(c *gin.Context) {
	bId := c.Query("botId")
	if bId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("botId is required"))
		return
	}
	botId, err := primitive.ObjectIDFromHex(bId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	pId := c.Query("projectId")
	if pId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("botId is required"))
		return
	}
	projectId, err := primitive.ObjectIDFromHex(pId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	trainingData, err := s.service.GetTrainingData(c, botId, projectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, trainingData)
	return
}

func (s Controllers) UpdateTrainingData(c *gin.Context) {
	var trainingData *models.TrainingData
	if err := c.ShouldBind(&trainingData); err != nil {
		c.JSON(http.StatusBadRequest, err)

	} else {
		res, err := s.service.UpdateTrainingData(c, trainingData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else {
			c.JSON(http.StatusOK, res)
		}
	}
}

func (s Controllers) DeleteTrainingData(c *gin.Context) {
	bId := c.Query("botId")
	if bId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("botId is required"))
		return
	}
	botId, err := primitive.ObjectIDFromHex(bId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	pId := c.Query("projectId")
	if pId == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("botId is required"))
		return
	}
	projectId, err := primitive.ObjectIDFromHex(pId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	trainingData, err := s.service.ResetTrainingData(c, botId, projectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, trainingData)
	return
}
