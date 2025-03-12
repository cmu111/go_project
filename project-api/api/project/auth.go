package project

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/auth"
)

type HandlerAuth struct {
}

func NewHandlerAuth() *HandlerAuth {
	return &HandlerAuth{}
}

func (h *HandlerAuth) Auth(c *gin.Context) {
	result := &common.Result{}
	organizationCode := c.GetString("organizationCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var orgreq *model.AccountReq
	c.ShouldBind(&orgreq)
	orgMsg := auth.AuthReqMessage{
		OrganizationCode: organizationCode,
		Page:             int64(orgreq.Page),
		PageSize:         int64(orgreq.PageSize),
	}
	authList, err := AuthServiceClient.AuthList(ctx, &orgMsg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(200, result.Fail(code, msg))
		return
	}
	var authListData []*model.ProjectAuth
	if authList.List == nil {
		authListData = []*model.ProjectAuth{}
	}
	copier.Copy(&authListData, authList.List)
	pagestr := strconv.Itoa(orgreq.Page)
	c.JSON(200, result.Success(gin.H{
		"list":  authListData,
		"total": authList.Total,
		"page":  pagestr,
	}))
}

func (a *HandlerAuth) apply(c *gin.Context) {
	result := &common.Result{}
	var req *model.ProjectAuthReq
	c.ShouldBind(&req)
	var nodes []string
	if req.Nodes != "" {
		json.Unmarshal([]byte(req.Nodes), &nodes)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &auth.AuthReqMessage{
		Action: req.Action,
		AuthId: req.Id,
		Nodes:  nodes,
	}
	applyResponse, err := AuthServiceClient.Apply(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*model.ProjectNodeAuthTree
	copier.Copy(&list, applyResponse.List)
	var checkedList []string
	copier.Copy(&checkedList, applyResponse.CheckedList)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":        list,
		"checkedList": checkedList,
	}))
}
