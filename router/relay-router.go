// Copyright (c) 2025 Tethys Plex
//
// This file is part of Veloera.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package router

import (
	"github.com/gin-gonic/gin"
	"veloera/controller"
	"veloera/middleware"
	"veloera/relay"
)

func SetRelayRouter(router *gin.Engine) {
	router.Use(middleware.CORS())
	router.Use(middleware.DecompressRequestMiddleware())

	// 提取通用的 models 路由设置函数
	setupModelsRouter := func(group *gin.RouterGroup) {
		group.GET("", controller.ListModels)
		group.GET("/:model", controller.RetrieveModel)
	}

	// 提取通用的 v1 路由设置函数
	setupV1Router := func(v1Router *gin.RouterGroup) {
		// WebSocket 路由
		wsRouter := v1Router.Group("")
		wsRouter.Use(middleware.Distribute())
		wsRouter.GET("/realtime", controller.WssRelay)

		// HTTP 路由
		httpRouter := v1Router.Group("")
		httpRouter.Use(middleware.Distribute())
		httpRouter.POST("/messages", controller.RelayClaude)
		httpRouter.POST("/completions", controller.Relay)
		httpRouter.POST("/chat/completions", controller.Relay)
		httpRouter.POST("/edits", controller.Relay)
		httpRouter.POST("/images/generations", controller.Relay)
		httpRouter.POST("/images/edits", controller.RelayNotImplemented)
		httpRouter.POST("/images/variations", controller.RelayNotImplemented)
		httpRouter.POST("/embeddings", controller.Relay)
		httpRouter.POST("/engines/:model/embeddings", controller.Relay)
		httpRouter.POST("/audio/transcriptions", controller.Relay)
		httpRouter.POST("/audio/translations", controller.Relay)
		httpRouter.POST("/audio/speech", controller.Relay)
		httpRouter.POST("/responses", controller.Relay)
		httpRouter.GET("/files", controller.RelayNotImplemented)
		httpRouter.POST("/files", controller.RelayNotImplemented)
		httpRouter.DELETE("/files/:id", controller.RelayNotImplemented)
		httpRouter.GET("/files/:id", controller.RelayNotImplemented)
		httpRouter.GET("/files/:id/content", controller.RelayNotImplemented)
		httpRouter.POST("/fine-tunes", controller.RelayNotImplemented)
		httpRouter.GET("/fine-tunes", controller.RelayNotImplemented)
		httpRouter.GET("/fine-tunes/:id", controller.RelayNotImplemented)
		httpRouter.POST("/fine-tunes/:id/cancel", controller.RelayNotImplemented)
		httpRouter.GET("/fine-tunes/:id/events", controller.RelayNotImplemented)
		httpRouter.DELETE("/models/:model", controller.RelayNotImplemented)
		httpRouter.POST("/moderations", controller.Relay)
		httpRouter.POST("/rerank", controller.Relay)
	}

	// 设置 /v1/models 路由
	modelsRouter := router.Group("/v1/models")
	modelsRouter.Use(middleware.TokenAuth())
	setupModelsRouter(modelsRouter)

	// 设置 /hf/v1/models 路由
	hfModelsRouter := router.Group("/hf/v1/models")
	hfModelsRouter.Use(middleware.TokenAuth())
	setupModelsRouter(hfModelsRouter)

	// 设置 /v1 路由组
	relayV1Router := router.Group("/v1")
	relayV1Router.Use(middleware.TokenAuth())
	relayV1Router.Use(middleware.TokenRateLimit())
	relayV1Router.Use(middleware.ModelRequestRateLimit())
	setupV1Router(relayV1Router)

	// 设置 /hf/v1 路由组
	relayHfV1Router := router.Group("/hf/v1")
	relayHfV1Router.Use(middleware.TokenAuth())
	relayHfV1Router.Use(middleware.TokenRateLimit())
	relayHfV1Router.Use(middleware.ModelRequestRateLimit())
	setupV1Router(relayHfV1Router)

	playgroundRouter := router.Group("/pg")
	playgroundRouter.Use(middleware.UserAuth())
	{
		playgroundRouter.POST("/chat/completions", controller.Playground)
	}

	relayMjRouter := router.Group("/mj")
	registerMjRouterGroup(relayMjRouter)

	relayMjModeRouter := router.Group("/:mode/mj")
	registerMjRouterGroup(relayMjModeRouter)
	//relayMjRouter.Use()

	relaySunoRouter := router.Group("/suno")
	relaySunoRouter.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		relaySunoRouter.POST("/submit/:action", controller.RelayTask)
		relaySunoRouter.POST("/fetch", controller.RelayTask)
		relaySunoRouter.GET("/fetch/:id", controller.RelayTask)
	}

}

func registerMjRouterGroup(relayMjRouter *gin.RouterGroup) {
	relayMjRouter.GET("/image/:id", relay.RelayMidjourneyImage)
	relayMjRouter.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		relayMjRouter.POST("/submit/action", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/shorten", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/modal", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/imagine", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/change", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/simple-change", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/describe", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/blend", controller.RelayMidjourney)
		relayMjRouter.POST("/notify", controller.RelayMidjourney)
		relayMjRouter.GET("/task/:id/fetch", controller.RelayMidjourney)
		relayMjRouter.GET("/task/:id/image-seed", controller.RelayMidjourney)
		relayMjRouter.POST("/task/list-by-condition", controller.RelayMidjourney)
		relayMjRouter.POST("/insight-face/swap", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/upload-discord-images", controller.RelayMidjourney)
	}
}
