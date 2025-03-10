package project

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"test.com/project-api/pkg/model"
	"test.com/project-api/pkg/model/pro"
	"test.com/project-api/pkg/model/tasks"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-common/fs"
	"test.com/project-common/tms"
	"test.com/project-grpc/task"
)

type HandlerTask struct {
}

func NewHandlerTask() *HandlerTask {
	return &HandlerTask{}
}

func (t *HandlerTask) TaskStages(c *gin.Context) {
	result := &common.Result{}
	ctx := context.Background()
	projectCode := c.PostForm("projectCode")
	page := &model.Page{}
	page.Bind(c)
	msg := &task.TaskReqMessage{
		MemberId:    c.GetInt64("memberId"),
		ProjectCode: projectCode,
		Page:        page.Page,
		PageSize:    page.PageSize,
	}
	//调用grpc客户端，把page转换为输入格式
	stages, err := TaskServiceClient.TaskStages(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	//把grpc返回的结果转换为输出格式
	var list []*tasks.TaskStagesResp
	copier.Copy(&list, stages.List)
	if list == nil {
		list = []*tasks.TaskStagesResp{}
	}
	for _, v := range list {
		v.TasksLoading = true  //任务加载状态
		v.FixedCreator = false //添加任务按钮定位
		v.ShowTaskCard = false //是否显示创建卡片
		v.Tasks = []int{}
		v.DoneTasks = []int{}
		v.UnDoneTasks = []int{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  list,
		"total": stages.Total,
		"page":  page.Page,
	}))
}

func (t *HandlerTask) memberProjectList(c *gin.Context) {
	result := &common.Result{}
	ctx := context.Background()
	projectCode := c.PostForm("projectCode")
	page := &model.Page{}
	page.Bind(c)
	msg := &task.TaskReqMessage{
		MemberId:    c.GetInt64("memberId"),
		ProjectCode: projectCode,
		Page:        page.Page,
		PageSize:    page.PageSize,
	}
	//调用grpc返回所有与用户有关的个人信息
	resp, err := TaskServiceClient.MemberProjectList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*pro.MemberProjectResp
	copier.Copy(&list, resp.List)
	if list == nil {
		list = []*pro.MemberProjectResp{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  list,
		"total": resp.Total,
		"page":  page.Page,
	}))
}

func (t *HandlerTask) taskList(c *gin.Context) {
	result := &common.Result{}
	stageCode := c.PostForm("stageCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, err := TaskServiceClient.TaskList(ctx, &task.TaskReqMessage{StageCode: stageCode})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var taskDisplayList []*tasks.TaskDisplay
	copier.Copy(&taskDisplayList, list.List)
	if taskDisplayList == nil {
		taskDisplayList = []*tasks.TaskDisplay{}
	}
	for _, v := range taskDisplayList {
		if v.Tags == nil {
			v.Tags = []int{}
		}
		if v.ChildCount == nil {
			v.ChildCount = []int{}
		}
	}
	c.JSON(http.StatusOK, result.Success(taskDisplayList))
}
func (t *HandlerTask) saveTask(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSaveReq
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		ProjectCode: req.ProjectCode,
		Name:        req.Name,
		StageCode:   req.StageCode,
		AssignTo:    req.AssignTo,
		MemberId:    c.GetInt64("memberId"),
	}
	taskMessage, err := TaskServiceClient.SaveTask(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	td := &tasks.TaskDisplay{}
	copier.Copy(td, taskMessage)
	if td != nil {
		if td.Tags == nil {
			td.Tags = []int{}
		}
		if td.ChildCount == nil {
			td.ChildCount = []int{}
		}
	}
	c.JSON(http.StatusOK, result.Success(td))
}

func (t *HandlerTask) taskSort(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSortReq
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		PreTaskCode:  req.PreTaskCode,
		NextTaskCode: req.NextTaskCode,
		ToStageCode:  req.ToStageCode,
	}
	_, err := TaskServiceClient.TaskSort(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (t *HandlerTask) myTaskList(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.MyTaskReq
	c.ShouldBind(&req)
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		MemberId: memberId,
		TaskType: int32(req.TaskType),
		Type:     int32(req.Type),
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	myTaskListResponse, err := TaskServiceClient.MyTaskList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var myTaskList []*tasks.MyTaskDisplay
	copier.Copy(&myTaskList, myTaskListResponse.List)
	if myTaskList == nil {
		myTaskList = []*tasks.MyTaskDisplay{}
	}
	for _, v := range myTaskList {
		v.ProjectInfo = tasks.ProjectInfo{
			Name: v.ProjectName,
			Code: v.ProjectCode,
		}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  myTaskList,
		"total": myTaskListResponse.Total,
	}))
}
func (t *HandlerTask) readTask(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
	}
	taskMessage, err := TaskServiceClient.ReadTask(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	td := &tasks.TaskDisplay{}
	copier.Copy(td, taskMessage)
	if td != nil {
		if td.Tags == nil {
			td.Tags = []int{}
		}
		if td.ChildCount == nil {
			td.ChildCount = []int{}
		}
	}
	c.JSON(200, result.Success(td))
}
func (t *HandlerTask) listTaskMember(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	page := &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
		Page:     page.Page,
		PageSize: page.PageSize,
	}
	taskMemberResponse, err := TaskServiceClient.ListTaskMember(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var tms []*tasks.TaskMember
	copier.Copy(&tms, taskMemberResponse.List)
	if tms == nil {
		tms = []*tasks.TaskMember{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  tms,
		"total": taskMemberResponse.Total,
		"page":  page.Page,
	}))
}

func (t *HandlerTask) listTaskLog(c *gin.Context) {
	result := &common.Result{}
	var req *model.TaskLogReq
	c.ShouldBind(&req)
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode: req.TaskCode,
		MemberId: c.GetInt64("memberId"),
		Page:     int64(req.Page),
		PageSize: int64(req.PageSize),
		All:      int32(req.All),
		Comment:  int32(req.Comment),
	}
	taskLogResponse, err := TaskServiceClient.TaskLog(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var tms []*model.ProjectLogDisplay
	copier.Copy(&tms, taskLogResponse.List)
	if tms == nil {
		tms = []*model.ProjectLogDisplay{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  tms,
		"total": taskLogResponse.Total,
		"page":  req.Page,
	}))
}
func (t *HandlerTask) taskWorkTimeList(c *gin.Context) {
	taskCode := c.PostForm("taskCode")
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
	}
	taskWorkTimeResponse, err := TaskServiceClient.TaskWorkTimeList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var tms []*model.TaskWorkTime
	copier.Copy(&tms, taskWorkTimeResponse.List)
	if tms == nil {
		tms = []*model.TaskWorkTime{}
	}
	c.JSON(http.StatusOK, result.Success(tms))
}

func (t *HandlerTask) saveTaskWorkTime(c *gin.Context) {
	result := &common.Result{}
	var req *model.SaveTaskWorkTimeReq
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode:  req.TaskCode,
		MemberId:  c.GetInt64("memberId"),
		Content:   req.Content,
		Num:       int32(req.Num),
		BeginTime: tms.ParseTime(req.BeginTime),
	}
	_, err := TaskServiceClient.SaveTaskWorkTime(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (t *HandlerTask) uploadFiles(c *gin.Context) {
	result := &common.Result{}
	req := &model.UploadFileReq{}
	c.ShouldBind(&req)
	multipartFile, err := c.MultipartForm()
	var key string
	if err != nil {
		zap.L().Error("c.MultipartForm() err", zap.Error(err))
		return
	}
	file := multipartFile.File
	//将file存入本地
	if req.ChunkNumber == 1 {
		//没分片直接存
		path := "upload/" + req.ProjectCode + "/" + req.TaskCode + "/" + tms.FormatYMD(time.Now())
		if fs.IsPathExists(path) == false {
			os.MkdirAll(path, os.ModePerm)
		}
		filePath := path + "/" + req.Filename
		osfile, _ := os.Create(filePath)
		defer osfile.Close()
		fileReader, _ := file["file"][0].Open()
		_, err = io.Copy(osfile, fileReader)
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		key = filePath
	} else {
		path := "upload/" + req.ProjectCode + "/" + req.TaskCode + "/" + tms.FormatYMD(time.Now())
		if fs.IsPathExists(path) == false {
			os.MkdirAll(path, os.ModePerm)
		}
		//以读写状态打开文件，写入新内容后存入
		fileName := path + "/" + req.Filename
		openFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		defer openFile.Close()
		fileReader, _ := file["file"][0].Open()
		_, err = io.Copy(openFile, fileReader)
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		key = fileName
	}
	//调用服务 存入file表
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	fileUrl := "http://localhost/" + key
	msg := &task.TaskFileReqMessage{
		TaskCode:         req.TaskCode,
		ProjectCode:      req.ProjectCode,
		OrganizationCode: c.GetString("organizationCode"),
		PathName:         key,
		FileName:         req.Filename,
		Size:             int64(req.TotalSize),
		Extension:        path.Ext(key),
		FileUrl:          fileUrl,
		FileType:         file["file"][0].Header.Get("Content-Type"),
		MemberId:         c.GetInt64("memberId"),
	}
	if req.TotalChunks == req.ChunkNumber {
		_, err = TaskServiceClient.SaveTaskFile(ctx, msg)
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
		}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"file":        key,
		"hash":        "",
		"key":         key,
		"url":         "http://localhost/" + key,
		"projectName": req.ProjectName,
	}))
	return
}
func (t *HandlerTask) taskSources(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sources, err := TaskServiceClient.TaskSources(ctx, &task.TaskReqMessage{TaskCode: taskCode})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var slList []*model.SourceLink
	copier.Copy(&slList, sources.List)
	if slList == nil {
		slList = []*model.SourceLink{}
	}
	c.JSON(http.StatusOK, result.Success(slList))
}

func (t *HandlerTask) createComment(c *gin.Context) {
	result := &common.Result{}
	req := model.CommentReq{}
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode:       req.TaskCode,
		CommentContent: req.Comment,
		Mentions:       req.Mentions,
		MemberId:       c.GetInt64("memberId"),
	}
	_, err := TaskServiceClient.CreateComment(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success(true))
}
