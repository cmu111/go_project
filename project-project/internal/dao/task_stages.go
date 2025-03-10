package dao

import (
	"context"
	"fmt"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database"
	"test.com/project-project/internal/database/gorms"
)

type TaskStagesDao struct {
	conn *gorms.GormConn
}

func NewTaskStagesDao() *TaskStagesDao {
	return &TaskStagesDao{
		conn: gorms.New(),
	}
}

func (t *TaskStagesDao) SaveTaskStages(ctx context.Context, conn database.DbConn, ts *data.TaskStages) error {
	t.conn = conn.(*gorms.GormConn)
	return t.conn.Tx(ctx).Save(&ts).Error
}

func (t *TaskStagesDao) FindStagesByProjectId(ctx context.Context, projectCode int64, page int64, pageSize int64) (list []*data.TaskStages, total int64, err error) {
	conn := t.conn.Session(ctx)
	index := (page - 1) * pageSize
	sql := fmt.Sprintf("select * from ms_task_stages where project_code=?  order by sort limit ?,?")
	err = conn.Raw(sql, projectCode, index, pageSize).Scan(&list).Error
	err = conn.Model(&data.TaskStages{}).Where("project_code=?", projectCode).Count(&total).Error
	return
}

func (t *TaskStagesDao) FindById(ctx context.Context, stageCode int) (*data.TaskStages, error) {
	conn:= t.conn.Session(ctx)
	stage :=&data.TaskStages{}
	err:=conn.Where("id=?", stageCode).Find(&stage).Error
	return stage, err
}