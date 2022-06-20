// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resthandler

import (
	"github.com/gin-gonic/gin"
	"github.com/illa-family/builder-backend/pkg/app"
	"go.uber.org/zap"
	"net/http"
)

type AppRestHandler interface {
	CreateApp(c *gin.Context)
	DeleteApp(c *gin.Context)
	RenameApp(c *gin.Context)
	GetAllApp(c *gin.Context)
}

type AppRestHandlerImpl struct {
	logger     *zap.SugaredLogger
	appService app.AppService
}

func NewAppRestHandlerImpl(logger *zap.SugaredLogger, appService app.AppService) *AppRestHandlerImpl {
	return &AppRestHandlerImpl{
		logger:     logger,
		appService: appService,
	}
}

func (impl AppRestHandlerImpl) CreateApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) DeleteApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) RenameApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) GetAllApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}