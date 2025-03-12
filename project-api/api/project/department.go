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
	"test.com/project-grpc/department"
)

type HandlerDepartment struct {
}

func NewHandlerDepartment() *HandlerDepartment {
	return &HandlerDepartment{}
}

func (h *HandlerDepartment) list(c *gin.Context) {
	result := &common.Result{}
	var departmentReq *model.DepartmentReq
	c.ShouldBind(&departmentReq)
	msg := &department.DepartmentReqMessage{
		MemberId:             c.GetInt64("memberId"),
		Page:                 departmentReq.Page,
		PageSize:             departmentReq.PageSize,
		ParentDepartmentCode: departmentReq.Pcode,
		OrganizationCode:     c.GetString("organizationCode"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	listDepartmentMessage, err := DepartmentServiceClient.List(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*model.Department
	copier.Copy(&list, listDepartmentMessage.List)
	if list == nil {
		list = []*model.Department{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"total": listDepartmentMessage.Total,
		"page":  departmentReq.Page,
		"list":  list,
	}))
}

func (h *HandlerDepartment) save(c *gin.Context) {
	result := &common.Result{}
	var departmentReq *model.DepartmentReq
	c.ShouldBind(&departmentReq)
	msg := &department.DepartmentReqMessage{
		Name:                 departmentReq.Name,
		ParentDepartmentCode: departmentReq.Pcode,
		DepartmentCode:       departmentReq.DepartmentCode,
		OrganizationCode:     c.GetString("organizationCode"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	departmentMessage, err := DepartmentServiceClient.Save(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var department model.Department
	copier.Copy(&department, departmentMessage)
	c.JSON(http.StatusOK, result.Success(department))
}

func (h *HandlerDepartment) read(c *gin.Context) {
	result := &common.Result{}
	msg := &department.DepartmentReqMessage{
		DepartmentCode:   c.PostForm("departmentCode"),
		OrganizationCode: c.GetString("organizationCode"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	departmentMessage, err := DepartmentServiceClient.Read(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var departmentrsp model.Department
	copier.Copy(&departmentrsp, departmentMessage)
	c.JSON(http.StatusOK, result.Success(departmentrsp))

}
