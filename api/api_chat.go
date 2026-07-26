package api

import (
	"Main/model"
	"Main/sv"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

func InitNewChat(ctx context.Context, c *app.RequestContext) {
	chatNewModel := model.ChatNewModel{}
	err := c.BindJSON(&chatNewModel)
	if err != nil {
		c.JSON(400, model.FinalResponse{
			Status: "400",
			Info:   err.Error(),
			Data:   nil,
		})
		return
	}

	chatNewDTO, err := sv.NewChat(chatNewModel)
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
		Data:   chatNewDTO,
	})
}
func InitContinueChat(ctx context.Context, c *app.RequestContext) {
	continueModel := model.ChatContinueModel{}
	err := c.BindJSON(&continueModel)
	if err != nil {
		c.JSON(400, model.FinalResponse{
			Status: "400",
			Info:   err.Error(),
			Data:   nil,
		})
		return
	}
	continueDTO, err := sv.ContinueChat(continueModel)
	if err != nil {
		c.JSON(500, model.FinalResponse{
			Status: "400",
			Info:   err.Error(),
			Data:   nil,
		})
		return
	}
	c.JSON(200, model.FinalResponse{
		Status: "200",
		Info:   "success",
		Data:   continueDTO,
	})
}
func InitDeleteChat(ctx context.Context, c *app.RequestContext) {
	chatDeleteModel := model.ChatDeleteModel{}
	err := c.BindJSON(&chatDeleteModel)
	if err != nil {
		c.JSON(400, model.FinalResponse{
			Status: "400",
			Info:   err.Error(),
			Data:   nil,
		})
		return
	}
	err = sv.DeleteChat(chatDeleteModel)
	if err != nil {
		c.JSON(500, model.FinalResponse{
			Status: "500",
			Info:   err.Error(),
			Data:   nil,
		})
		return
	}
	c.JSON(200, model.FinalResponse{
		Status: "200",
		Info:   "success",
		Data:   nil,
	})
}
