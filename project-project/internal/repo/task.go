package repo

import (
	"context"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database"
)

type TaskStagesTemplateRepo interface {
	FindInProTemIds(ctx context.Context, ids []int) ([]data.MsTaskStagesTemplate, error)
	FindByProjectTemplateId(ctx context.Context, projectTemplateCode int) (list []*data.MsTaskStagesTemplate, err error)
}

type TaskStagesRepo interface {
	SaveTaskStages(ctx context.Context, conn database.DbConn, ts *data.TaskStages) error
	FindStagesByProjectId(ctx context.Context, projectCode int64, page int64, pageSize int64) (list []*data.TaskStages, total int64, err error)
	FindById(ctx context.Context, stageCode int) (*data.TaskStages, error)
}

type TaskRepo interface {
	FindTaskByStageCode(ctx context.Context, stageCode int) (list []*data.Task, err error)
	FindTaskMemberByTaskId(ctx context.Context, taskCode int64, memberId int64) (task *data.TaskMember, err error)
	FindTaskMaxIdNum(ctx context.Context, projectCode int64) (v *int, err error)
	FindTaskSort(ctx context.Context, projectCode int64, stageCode int64) (v *int, err error)
	SaveTask(ctx context.Context, conn database.DbConn, ts *data.Task) error
	SaveTaskMember(ctx context.Context, conn database.DbConn, tm *data.TaskMember) error
	FindTaskById(ctx context.Context, TaskCode int64) (*data.Task, error)
	FindTaskByStageCodeLtSort(ctx context.Context, stageCode int, sort int) (ts *data.Task, err error)
	UpdateTaskSort(context context.Context, conn database.DbConn, v *data.Task) error
	FindTaskByAssignTo(ctx context.Context, id int64, i int, page int64, size int64) ([]*data.Task, int64, error)
	FindTaskByCreatBy(ctx context.Context, id int64, done int, page int64, size int64) ([]*data.Task, int64, error)
	FindTaskByMemberCode(ctx context.Context, memberId int64, done int, page int64, size int64) ([]*data.Task, int64, error)
	FindTaskMemberPage(ctx context.Context, taskCode int64, page int64, size int64) ([]*data.TaskMember, int64, error)
FindTaskByIds(ctx context.Context, tids []int64) (list []*data.Task, err error) 
}

type TaskWorkTimeRepo interface {
	FindWorkTimeByTaskCode(ctx context.Context, taskCode int64) (WorkTime []*data.TaskWorkTime, err error)
	Save(ctx context.Context, workTime *data.TaskWorkTime) error
}
