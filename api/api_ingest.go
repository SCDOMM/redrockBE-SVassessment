package api

import (
	"Main/model"
	"Main/sv"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

func InitNewIngest(ctx context.Context, c *app.RequestContext) {
	ingestModel := model.IngestNewModel{}
	err := c.BindJSON(&ingestModel)
	if err != nil {
		c.JSON(400, model.FinalResponse{
			Status: "400",
			Info:   err.Error(),
			Data:   nil,
		})
		return
	}
	ingestDTO, err := sv.NewIngest(ingestModel.RepoUrl)
	if err != nil {
		c.JSON(400, model.FinalResponse{
			Status: "500",
			Info:   err.Error(),
			Data:   nil,
		})
		return
	}
	c.JSON(200, model.FinalResponse{
		Status: "200",
		Info:   "success",
		Data:   ingestDTO,
	})
}
func InitCheckIngest(ctx context.Context, c *app.RequestContext) {
	taskId := c.Param("id")
	checkDTO, err := sv.CheckIngest(taskId)
	if err != nil {
		c.JSON(400, model.FinalResponse{
			Status: "500",
			Info:   err.Error(),
			Data:   nil,
		})
	}
	c.JSON(200, model.FinalResponse{
		Status: "200",
		Info:   "success",
		Data:   checkDTO,
	})
}
