package project

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/menu"
)

type HandlerMenu struct {
}

func NewHandlerMenu() *HandlerMenu {
	return &HandlerMenu{}
}

func (h *HandlerMenu) GetMenu(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*2)
	defer cancel()
	menuReq := &menu.MenuReqMessage{}
	menuList, err := MenuServiceClient.MenuList(ctx, menuReq)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(200, result.Fail(code, msg))
		return
	}
	var menuListRsp []*model.Menu
	if menuList == nil {
		menuListRsp = []*model.Menu{}
		c.JSON(http.StatusOK, result.Success(menuListRsp))
		return
	}
	copier.Copy(&menuListRsp, menuList.List)
	c.JSON(http.StatusOK, result.Success(menuListRsp))
}
