package router

import (
	"Main/api"
	"Main/utils"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouter() {
	h := server.Default(server.WithHostPorts(utils.GetHost() + utils.GetPort()))

	g1 := h.Group("api/ingest")
	g1.POST("/new", api.InitNewIngest)
	g1.GET("/:id/status", api.InitCheckIngest)
	g1.POST("/delete", api.InitCancelIngest)

	g2 := h.Group("/api/chat")
	g2.POST("/new", api.InitNewChat)
	g2.POST("/continue", api.InitContinueChat)
	g2.POST("/delete", api.InitDeleteChat)

	h.Spin()
}
