package project

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/account"
)

type HandlerAccount struct {
}

func NewHandlerAccount() *HandlerAccount {
	return &HandlerAccount{}
}

func (a *HandlerAccount) account(c *gin.Context) {
	result := &common.Result{}
	ctx := context.Background()
	//输入转换为rpc的请求
	var req *model.AccountReq
	_ = c.ShouldBind(&req)
	memberId := c.GetInt64("memberId")
	msg := &account.AccountReqMessage{
		MemberId:         memberId,
		OrganizationCode: c.GetString("organizationCode"),
		Page:             int64(req.Page),
		PageSize:         int64(req.PageSize),
		SearchType:       int32(req.SearchType),
		DepartmentCode:   req.DepartmentCode,
	}
	response, err := AccountServiceClient.Account(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	//拷贝到model中
	//用make已经分配了空间了吗，还会是nil吗？
	accountList := make([]*model.MemberAccount, 0)
	authList:=make([]*model.ProjectAuth,0)
	copier.Copy(&accountList, response.AccountList)
	copier.Copy(&authList, response.AuthList)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"total":       response.Total,
		"page":        req.Page,
		"list":     accountList,
		"authList": authList,
	}))
	
}
