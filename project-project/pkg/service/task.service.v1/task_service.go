package taskservicev1

import (
	"context"
	"time"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-common/tms"
	"test.com/project-grpc/task"
	"test.com/project-grpc/user/login"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database"
	"test.com/project-project/internal/database/tran"
	"test.com/project-project/internal/repo"
	"test.com/project-project/internal/rpc"
	"test.com/project-project/pkg/model"
)

type TaskService struct {
	task.UnimplementedTaskServiceServer
	cache                  repo.Cache
	transaction            tran.Transaction
	menuRepo               repo.MenuRepo
	projectRepo            repo.ProjectRepo
	projectTemplateRepo    repo.ProjectTemplateRepo
	taskStagesTemplateRepo repo.TaskStagesTemplateRepo
	taskStagesRepo         repo.TaskStagesRepo
	taskRepo               repo.TaskRepo
	projectLogRepo         repo.ProjectLogRepo
	taskWorkTimerepo       repo.TaskWorkTimeRepo
	fileRepo               repo.FileRepo
	sourceLinkRepo         repo.SourceLinkRepo
}

func New() *TaskService {
	return &TaskService{
		cache:                  dao.Rc,
		transaction:            dao.NewTransaction(),
		menuRepo:               dao.NewMenuDao(),
		projectRepo:            dao.NewProjectDao(),
		projectTemplateRepo:    dao.NewProjectTemplateDao(),
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(),
		taskStagesRepo:         dao.NewTaskStagesDao(),
		taskRepo:               dao.NewTaskDao(),
		projectLogRepo:         dao.NewProjectLogDao(),
		taskWorkTimerepo:       dao.NewTaskWorkTimeDao(),
		fileRepo:               dao.NewFileDao(),
		sourceLinkRepo:         dao.NewSourceLinkDao(),
	}
}

func (t *TaskService) TaskStages(c context.Context, msg *task.TaskReqMessage) (*task.TaskStagesResponse, error) {
	projectId := encrypts.DecryptNoErr(msg.ProjectCode)
	page := msg.Page
	pageSize := msg.PageSize
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	stages, total, err := t.taskStagesRepo.FindStagesByProjectId(ctx, projectId, page, pageSize)
	if err != nil {
		zap.L().Error("task FindStagesByProjectId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var tsMessages []*task.TaskStagesMessage
	copier.Copy(&tsMessages, stages)
	if tsMessages == nil {
		return &task.TaskStagesResponse{
			Total: 0,
			List:  tsMessages,
		}, nil
	}
	stagesMap := data.ToTaskStagesMap(stages)
	for _, v := range tsMessages {
		taskStages := stagesMap[int(v.Id)]
		v.Code = encrypts.EncryptNoErr(int64(v.Id))
		v.CreateTime = tms.FormatByMill(taskStages.CreateTime)
		v.ProjectCode = msg.ProjectCode
	}
	return &task.TaskStagesResponse{
		Total: int64(total),
		List:  tsMessages,
	}, nil
}

func (t *TaskService) MemberProjectList(c context.Context, msg *task.TaskReqMessage) (*task.MemberProjectResponse, error) {
	//1.根据projectcode查memberid
	ctx := context.Background()
	projectId := encrypts.DecryptNoErr(msg.ProjectCode)
	projectMembers, err := t.projectRepo.FindProjectMemberByPid(ctx, projectId)
	if err != nil {
		zap.L().Error("MemberProjectList FindProjectMemberByPid error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if projectMembers == nil || len(projectMembers) <= 0 {
		return &task.MemberProjectResponse{List: nil, Total: 0}, nil
	}
	//2.根据memberid查memberlist
	var mIds []int64
	pmMap := make(map[int64]*data.ProjectMember)
	for _, v := range projectMembers {
		mIds = append(mIds, v.MemberCode)
		pmMap[v.MemberCode] = v
	}
	//与user有关的信息由user模块管理，要交予他处理
	membersMsg := &login.UserMessage{
		MIds: mIds,
	}
	//var members *login.MemberMessageList
	members, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, membersMsg)
	if err != nil {
		zap.L().Error("MemberProjectList FindMemInfoByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if members == nil || len(members.List) <= 0 {
		return &task.MemberProjectResponse{List: nil, Total: 0}, nil
	}
	//3.组装返回数据
	var mpList []*task.MemberProjectMessage
	for _, v := range members.List {
		mp := &task.MemberProjectMessage{
			Name:   v.Name,
			Avatar: v.Avatar,
			Code:   encrypts.EncryptNoErr(v.Id),
			Email:  v.Email,
		}
		if pmMap[v.Id].IsOwner == v.Id {
			mp.IsOwner = 1
		}
		mpList = append(mpList, mp)
	}
	return &task.MemberProjectResponse{List: mpList, Total: int64(len(mpList))}, nil
}

func (t *TaskService) TaskList(c context.Context, msg *task.TaskReqMessage) (*task.TaskListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//输入转化为查询条件
	stagecode := encrypts.DecryptNoErr(msg.StageCode)
	tasklist, err := t.taskRepo.FindTaskByStageCode(ctx, int(stagecode))
	if err != nil {
		zap.L().Error("TaskList FindTaskByStageCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//返回值转化成可展示的格式
	var taskDisplayList []*data.TaskDisplay
	var mIds []int64
	for _, v := range tasklist {
		display := v.ToTaskDisplay()
		//当为隐私状态，且不在task成员表中时，不可见
		if v.Private == 1 {
			//代表隐私模式
			taskMember, err := t.taskRepo.FindTaskMemberByTaskId(ctx, v.Id, msg.MemberId)
			if err != nil {
				zap.L().Error("project task TaskList taskRepo.FindTaskMemberByTaskId error", zap.Error(err))
				return nil, errs.GrpcError(model.DBError)
			}
			if taskMember != nil {
				display.CanRead = model.CanRead
			} else {
				display.CanRead = model.NoRead
			}
		} else {
			display.CanRead = model.CanRead
		}
		//该task指派给谁,后续要取出该member的信息
		mIds = append(mIds, v.AssignTo)
		taskDisplayList = append(taskDisplayList, display)
	}
	if mIds == nil || len(mIds) <= 0 {
		return &task.TaskListResponse{List: nil}, nil
	}
	memberList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIds})
	if err != nil {
		zap.L().Error("TaskList FindMemInfoByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	memberMap := make(map[int64]*login.MemberMessage)
	for _, v := range memberList.List {
		memberMap[v.Id] = v
	}
	var executor data.Executor
	for _, v := range taskDisplayList {
		executor.Name = memberMap[encrypts.DecryptNoErr(v.AssignTo)].Name
		executor.Avatar = memberMap[encrypts.DecryptNoErr(v.AssignTo)].Avatar
		v.Executor = executor
	}
	//组装数据，直接使用copy即可，已经在ToTaskDisplay中处理过
	var taskList []*task.TaskMessage
	copier.Copy(&taskList, taskDisplayList)
	return &task.TaskListResponse{List: taskList}, nil
}

func (t *TaskService) SaveTask(c context.Context, msg *task.TaskReqMessage) (*task.TaskMessage, error) {
	//根据现有信息存入task与task_member表
	//校验输入是否正确，是否有对应的project或者task
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if msg.Name == "" {
		return nil, errs.GrpcError(model.TaskNameNotNull)
	}
	stageCode := encrypts.DecryptNoErr(msg.StageCode)
	taskStage, err := t.taskStagesRepo.FindById(ctx, int(stageCode))
	if err != nil {
		zap.L().Error("SaveTask FindById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskStage == nil {
		return nil, errs.GrpcError(model.TaskStagesNotNull)
	}
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	project, err := t.projectRepo.FindProjectById(ctx, projectCode)
	if err != nil {
		zap.L().Error("project task SaveTask projectRepo.FindProjectById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if project == nil || project.Deleted == model.Deleted {
		return nil, errs.GrpcError(model.ProjectAlreadyDeleted)
	}
	//校验完成，使用传入信息与数据库获取信息，转换成要存入的格式
	//获取当前project成员数(task表中当前project总人数)与排序数(当前task表中处于这个stage的人数)，要获取当前stage的总人数
	maxIdNum, err := t.taskRepo.FindTaskMaxIdNum(ctx, projectCode)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskMaxIdNum error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if maxIdNum == nil {
		a := 0
		maxIdNum = &a
	}
	maxSort, err := t.taskRepo.FindTaskSort(ctx, projectCode, stageCode)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskSort error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if maxSort == nil {
		a := 0
		maxSort = &a
	}
	//将数据转换为对应格式并组装
	assignTo := encrypts.DecryptNoErr(msg.AssignTo)
	ts := &data.Task{
		Name:        msg.Name,
		CreateTime:  time.Now().UnixMilli(),
		CreateBy:    msg.MemberId,
		AssignTo:    assignTo,
		ProjectCode: projectCode,
		StageCode:   int(stageCode),
		IdNum:       *maxIdNum + 1,
		Private:     project.OpenTaskPrivate,
		Sort:        *maxSort + 65536,
		BeginTime:   time.Now().UnixMilli(),
		EndTime:     time.Now().Add(2 * 24 * time.Hour).UnixMilli(),
	}
	//存入数据库，使用事务
	//调用action，写一个函数作为参数传入
	err = t.transaction.Action(func(conn database.DbConn) error {
		//1.存入task表
		err := t.taskRepo.SaveTask(ctx, conn, ts)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTask error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		//2.存入task_member表,判断是否与提交人memberid一致
		tm := &data.TaskMember{
			MemberCode: assignTo,
			TaskCode:   ts.Id,
			JoinTime:   time.Now().UnixMilli(),
			IsOwner:    model.Owner,
		}
		if msg.MemberId == assignTo {
			tm.IsExecutor = model.Executor
		}
		err = t.taskRepo.SaveTaskMember(ctx, conn, tm)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTaskMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}

		return nil
	})
	//这里不需要errs.GrpcError(model.DBError)是因为事务失败返回的就是这个错误
	if err != nil {
		return nil, err
	}
	display := ts.ToTaskDisplay()
	member, err := rpc.LoginServiceClient.FindMemInfoById(ctx, &login.UserMessage{MemId: assignTo})
	if err != nil {
		return nil, err
	}
	display.Executor = data.Executor{
		Name:   member.Name,
		Avatar: member.Avatar,
		Code:   member.Code,
	}
	tm := &task.TaskMessage{}
	copier.Copy(tm, display)
	createProjectLog(t.projectLogRepo, ts.ProjectCode, ts.Id, ts.Name, ts.AssignTo, "create", "task")
	return tm, nil
}

func createProjectLog(
	logRepo repo.ProjectLogRepo,
	projectCode int64,
	taskCode int64,
	taskName string,
	toMemberCode int64,
	logType string,
	actionType string) {
	remark := ""
	if logType == "create" {
		remark = "创建了任务"
	}
	pl := &data.ProjectLog{
		MemberCode:  toMemberCode,
		SourceCode:  taskCode,
		Content:     taskName,
		Remark:      remark,
		ProjectCode: projectCode,
		CreateTime:  time.Now().UnixMilli(),
		Type:        logType,
		ActionType:  actionType,
		Icon:        "plus",
		IsComment:   0,
		IsRobot:     0,
	}
	logRepo.SaveProjectLog(pl)
}

func (t *TaskService) TaskSort(c context.Context, msg *task.TaskReqMessage) (*task.TaskSortResponse, error) {

	preTaskCode := encrypts.DecryptNoErr(msg.PreTaskCode)
	toStageCode := encrypts.DecryptNoErr(msg.ToStageCode)
	if msg.PreTaskCode == msg.NextTaskCode {
		return &task.TaskSortResponse{}, nil
	}
	err := t.sortTask(preTaskCode, msg.NextTaskCode, toStageCode)
	if err != nil {
		return nil, err
	}
	return &task.TaskSortResponse{}, nil
}

func (t *TaskService) sortTask(preTaskCode int64, nextTaskCode string, toStageCode int64) error {
	//1.根据当前taskcode获取task
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	task, err := t.taskRepo.FindTaskById(ctx, preTaskCode)
	if err != nil {
		zap.L().Error("project task sortTask taskRepo.FindTaskById error", zap.Error(err))
		return errs.GrpcError(model.DBError)
	}
	//涉及保存起事务
	err = t.transaction.Action(func(conn database.DbConn) error {
		task.StageCode = int(toStageCode)
		if nextTaskCode != "" {
			nextTaskId := encrypts.DecryptNoErr(nextTaskCode)
			next, err := t.taskRepo.FindTaskById(ctx, nextTaskId)
			if err != nil {
				zap.L().Error("project task sortTask taskRepo.FindTaskById error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			//找到小于nextTaskCode的最大sort
			maxLtSort, err := t.taskRepo.FindTaskByStageCodeLtSort(ctx, int(toStageCode), next.Sort)
			if err != nil {
				zap.L().Error("project task sortTask taskRepo.FindTaskByStageCodeLtSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			if maxLtSort == nil {
				task.Sort = (0 + next.Sort) / 2
				maxLtSort = &data.Task{Sort: 0}
			} else {
				task.Sort = (maxLtSort.Sort + next.Sort) / 2
			}
			//更新task表
			err = t.taskRepo.UpdateTaskSort(ctx, conn, task)
			if err != nil {
				zap.L().Error("project task sortTask taskRepo.UpdateTaskSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			if task.Sort-maxLtSort.Sort < 50 {
				//如果距离上一个task过近，则重新排序
				t.resetSort(toStageCode)
				return nil
			}
		} else {
			//当前即为第一个stage的task,或者是当前stage的最后一个task
			maxSort, err := t.taskRepo.FindTaskSort(ctx, task.ProjectCode, toStageCode)
			if err != nil {
				zap.L().Error("project task sortTask taskRepo.FindTaskSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			if maxSort == nil {
				task.Sort = 65536
			} else {
				task.Sort = *maxSort + 65536
			}
			err = t.taskRepo.UpdateTaskSort(ctx, conn, task)
			if err != nil {
				zap.L().Error("project task sortTask taskRepo.UpdateTaskSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
		}
		return nil
	})
	return err
}

func (t *TaskService) resetSort(stageCode int64) error {
	list, err := t.taskRepo.FindTaskByStageCode(context.Background(), int(stageCode))
	if err != nil {
		return err
	}
	return t.transaction.Action(func(conn database.DbConn) error {
		iSort := 65536
		for index, v := range list {
			v.Sort = (index + 1) * iSort
			return t.taskRepo.UpdateTaskSort(context.Background(), conn, v)
		}
		return nil
	})
}

func (t *TaskService) MyTaskList(ctx context.Context, msg *task.TaskReqMessage) (*task.MyTaskListResponse, error) {
	var tsList []*data.Task
	var err error
	var total int64
	if msg.TaskType == 1 {
		//获取当前的task列表
		tsList, total, err = t.taskRepo.FindTaskByAssignTo(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByAssignTo error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	} else if msg.TaskType == 2 {
		//从member表取，再根据id来取task
		tsList, total, err = t.taskRepo.FindTaskByMemberCode(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByMemberCode error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	} else if msg.TaskType == 3 {
		tsList, total, err = t.taskRepo.FindTaskByCreatBy(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByCreatBy error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	} else {
		return nil, errs.GrpcError(model.TaskTypeInvalid)
	}
	//组装数据
	var mIds []int64
	var pIds []int64
	for _, v := range tsList {
		mIds = append(mIds, v.AssignTo)
		pIds = append(pIds, v.ProjectCode)
	}
	memberCnl := make(chan map[int64]*login.MemberMessage)
	projectCnl := make(chan map[int64]*data.Project)
	membererr := make(chan error)
	projecterr := make(chan error)
	defer close(memberCnl)
	defer close(projectCnl)
	go func() {
		memberList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIds})
		if err != nil {
			membererr <- err
			return
		}
		memberMap := make(map[int64]*login.MemberMessage)
		for _, v := range memberList.List {
			memberMap[v.Id] = v
		}
		memberCnl <- memberMap
	}()
	go func() {
		projectList, err := t.projectRepo.FindProjectByIds(ctx, pIds)
		if err != nil {
			zap.L().Error("project task MyTaskList FindProjectByIds error", zap.Error(err))
			projecterr <- err
			return
		}
		projectMap := data.ToProjectMap(projectList)
		projectCnl <- projectMap
	}()
	var memberMap map[int64]*login.MemberMessage
	var projectMap map[int64]*data.Project
	select {
	case memberMap = <-memberCnl:
	case err := <-membererr:
		return nil, err
	}
	select {
	case projectMap = <-projectCnl:
	case err := <-projecterr:
		return nil, err
	}
	var mtdList []*data.MyTaskDisplay
	for _, v := range tsList {
		mtd := v.ToMyTaskDisplay(projectMap[v.ProjectCode], memberMap[v.AssignTo].Name, memberMap[v.AssignTo].Avatar)
		mtdList = append(mtdList, mtd)
	}
	var myMsg []*task.MyTaskMessage
	copier.Copy(&myMsg, mtdList)
	return &task.MyTaskListResponse{List: myMsg, Total: total}, nil
}

func (t *TaskService) ReadTask(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskMessage, error) {
	//根据taskcode与memid查
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	memId := msg.MemberId
	ts, err := t.taskRepo.FindTaskById(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task ReadTask taskRepo.FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if ts == nil {
		return &task.TaskMessage{}, nil
	}
	taskDisplay := ts.ToTaskDisplay()
	if ts.Private == int(model.Private) {
		//代表隐私模式
		taskMember, err := t.taskRepo.FindTaskMemberByTaskId(ctx, ts.Id, memId)
		if err != nil {
			zap.L().Error("project task TaskList taskRepo.FindTaskMemberByTaskId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		if taskMember != nil {
			taskDisplay.CanRead = model.CanRead
		} else {
			taskDisplay.CanRead = model.NoRead
		}
	} else {
		taskDisplay.CanRead = model.CanRead
	}
	member, err := rpc.LoginServiceClient.FindMemInfoById(ctx, &login.UserMessage{MemId: ts.AssignTo})
	project, err := t.projectRepo.FindProjectById(ctx, ts.ProjectCode)
	stage, err := t.taskStagesRepo.FindById(ctx, ts.StageCode)
	if err != nil {
		zap.L().Error("project task ReadTask FindMemInfoById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	taskDisplay.Executor = data.Executor{
		Name:   member.Name,
		Avatar: member.Avatar,
	}
	taskDisplay.ProjectName = project.Name
	taskDisplay.StageName = stage.Name
	tm := &task.TaskMessage{}
	copier.Copy(tm, taskDisplay)
	return tm, nil

}
func (t *TaskService) ListTaskMember(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskMemberList, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	//taskMemberList:= []*data.TaskMember{}
	taskMemberList, total, err := t.taskRepo.FindTaskMemberPage(ctx, taskCode, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project task ListTaskMember taskRepo.FindTaskMemberPage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var mIds []int64
	memberMap := make(map[int64]*data.TaskMember)
	for _, v := range taskMemberList {
		mIds = append(mIds, v.MemberCode)
		memberMap[v.MemberCode] = v
	}
	memberList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIds})
	if err != nil {
		return nil, err
	}
	var taskMemberListMsg []*task.TaskMemberMessage
	for _, v := range memberList.List {
		var taskMemberMsg task.TaskMemberMessage
		taskMemberMsg.Id = memberMap[v.Id].Id
		taskMemberMsg.Name = v.Name
		taskMemberMsg.Avatar = v.Avatar
		taskMemberMsg.Code = encrypts.EncryptNoErr(memberMap[v.Id].MemberCode)
		taskMemberMsg.IsExecutor = int32(memberMap[v.Id].IsExecutor)
		taskMemberMsg.IsOwner = int32(memberMap[v.Id].IsOwner)
		taskMemberListMsg = append(taskMemberListMsg, &taskMemberMsg)
	}
	return &task.TaskMemberList{List: taskMemberListMsg, Total: total}, nil
}

func (t *TaskService) TaskLog(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskLogList, error) {

	//根据taskcode去tasklog表查即可
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	isAll := msg.All
	isComment := msg.Comment
	var logList []*data.ProjectLog
	var err error
	var total int64
	if isAll == 1 {
		logList, total, err = t.projectLogRepo.FindLogByTaskCode(ctx, taskCode, int(isComment))
	} else if isAll == 0 {
		logList, total, err = t.projectLogRepo.FindLogByTaskCodePage(ctx, taskCode, int(isComment), int(msg.Page), int(msg.PageSize))
	} else {
		return nil, errs.GrpcError(model.LogNumInvalid)
	}
	if err != nil {
		zap.L().Error("project task TaskLog projectLogRepo.FindLogByTaskCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if total == 0 {
		return &task.TaskLogList{}, nil
	}
	var displayList []*data.ProjectLogDisplay
	var mIdList []int64
	for _, v := range logList {
		mIdList = append(mIdList, v.MemberCode)
	}
	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	for _, v := range logList {
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := data.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}
	var l []*task.TaskLog
	copier.Copy(&l, displayList)
	return &task.TaskLogList{List: l, Total: total}, nil

}
func (t *TaskService) TaskWorkTimeList(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskWorkTimeResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	taskWorkTimeList, err := t.taskWorkTimerepo.FindWorkTimeByTaskCode(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task TaskWorkTimeList taskWorkTimeRepo.FindTaskWorkTimeByTaskCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskWorkTimeList == nil {
		return &task.TaskWorkTimeResponse{}, nil
	}
	var mIds []int64
	WorkTimeMap := make(map[int64]*data.TaskWorkTime)
	for _, v := range taskWorkTimeList {
		mIds = append(mIds, v.MemberCode)
		WorkTimeMap[v.MemberCode] = v
	}
	memberList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIds})
	if err != nil {
		return nil, err
	}
	var taskWorkTimeDisList []*data.TaskWorkTimeDisplay
	for _, v := range memberList.List {
		taskWorkTimeDis := WorkTimeMap[v.Id].ToDisplay()
		var member data.Member
		member.Name = v.Name
		member.Avatar = v.Avatar
		taskWorkTimeDis.Member = member
		taskWorkTimeDisList = append(taskWorkTimeDisList, taskWorkTimeDis)
	}
	var taskWorkTimeMsg []*task.TaskWorkTime
	copier.Copy(&taskWorkTimeMsg, taskWorkTimeDisList)
	//FIXME:这里没传total，感觉没必要
	return &task.TaskWorkTimeResponse{List: taskWorkTimeMsg}, nil
}

func (t *TaskService) SaveTaskWorkTime(ctx context.Context, msg *task.TaskReqMessage) (*task.SaveTaskWorkTimeResponse, error) {
	tmt := &data.TaskWorkTime{}
	tmt.BeginTime = msg.BeginTime
	tmt.Num = int(msg.Num)
	tmt.Content = msg.Content
	tmt.TaskCode = encrypts.DecryptNoErr(msg.TaskCode)
	tmt.MemberCode = msg.MemberId
	err := t.taskWorkTimerepo.Save(ctx, tmt)
	if err != nil {
		zap.L().Error("project task SaveTaskWorkTime taskWorkTimeRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &task.SaveTaskWorkTimeResponse{}, nil

}
func (t *TaskService) SaveTaskFile(ctx context.Context, msg *task.TaskFileReqMessage) (*task.TaskFileResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	err := t.transaction.Action(func(conn database.DbConn) error {
		//存file表
		f := &data.File{
			PathName:         msg.PathName,
			Title:            msg.FileName,
			Extension:        msg.Extension,
			Size:             int(msg.Size),
			ObjectType:       "",
			OrganizationCode: encrypts.DecryptNoErr(msg.OrganizationCode),
			TaskCode:         encrypts.DecryptNoErr(msg.TaskCode),
			ProjectCode:      encrypts.DecryptNoErr(msg.ProjectCode),
			CreateBy:         msg.MemberId,
			CreateTime:       time.Now().UnixMilli(),
			Downloads:        0,
			Extra:            "",
			Deleted:          model.NoDeleted,
			FileType:         msg.FileType,
			FileUrl:          msg.FileUrl,
			DeletedTime:      0,
		}
		err := t.fileRepo.Save(context.Background(), conn, f)
		if err != nil {
			zap.L().Error("project task SaveTaskFile fileRepo.Save error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		//存入source_link
		sl := &data.SourceLink{
			SourceType:       "file",
			SourceCode:       f.Id,
			LinkType:         "task",
			LinkCode:         taskCode,
			OrganizationCode: encrypts.DecryptNoErr(msg.OrganizationCode),
			CreateBy:         msg.MemberId,
			CreateTime:       time.Now().UnixMilli(),
			Sort:             0,
		}
		err = t.sourceLinkRepo.Save(context.Background(), conn, sl)
		if err != nil {
			zap.L().Error("project task SaveTaskFile sourceLinkRepo.Save error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &task.TaskFileResponse{}, nil
}
func (t *TaskService) TaskSources(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskSourceResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	//先查source_link表,再根据file_id查file表
	sourceLinkList, err := t.sourceLinkRepo.FindByTaskCode(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task TaskSources sourceLinkRepo.FindSourceLinkByLinkCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	fileLinkMap := make(map[int64]*data.SourceLink)
	fileIds := make([]int64, len(sourceLinkList))
	for _, v := range sourceLinkList {
		fileIds = append(fileIds, v.SourceCode)
		fileLinkMap[v.SourceCode] = v
	}
	fileList, err := t.fileRepo.FindByIds(ctx, fileIds)
	if err != nil {
		zap.L().Error("project task TaskSources fileRepo.FindByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var taskSourceList []*data.SourceLinkDisplay
	for _, v := range fileList {
		sourceLinkDis := fileLinkMap[v.Id].ToDisplay(v)
		taskSourceList = append(taskSourceList, sourceLinkDis)
	}
	var taskSourceMsg []*task.TaskSourceMessage
	copier.Copy(&taskSourceMsg, taskSourceList)
	return &task.TaskSourceResponse{List: taskSourceMsg}, nil
}
func (t *TaskService) CreateComment(ctx context.Context, msg *task.TaskReqMessage) (*task.CreateCommentResponse, error) {
	//把输入转为log格式
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	taskById, err := t.taskRepo.FindTaskById(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task CreateComment taskRepo.FindById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	log := &data.ProjectLog{
		MemberCode:   msg.MemberId,
		Content:      msg.CommentContent,
		Remark:       msg.CommentContent,
		Type:         "createComment",
		CreateTime:   time.Now().UnixMilli(),
		SourceCode:   taskCode,
		ActionType:   "task",
		ToMemberCode: 0,
		IsComment:    model.Comment,
		ProjectCode:  taskById.ProjectCode,
		Icon:         "plus",
		IsRobot:      0,
	}
	t.projectLogRepo.SaveProjectLog(log)
	return &task.CreateCommentResponse{}, nil
}
