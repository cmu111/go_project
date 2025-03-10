package dao

import (
	"context"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database/gorms"
)

type TaskWorkTimeDao struct {
	conn *gorms.GormConn
}

func NewTaskWorkTimeDao() *TaskWorkTimeDao {
	return &TaskWorkTimeDao{
		conn: gorms.New(),
	}
}
func (d *TaskWorkTimeDao) FindWorkTimeByTaskCode(ctx context.Context, taskCode int64) (WorkTime []*data.TaskWorkTime, err error) {
	session := d.conn.Session(context.Background())
	err = session.Where("task_code = ?", taskCode).Find(&WorkTime).Error
	return
}
func (d *TaskWorkTimeDao) Save(ctx context.Context, workTime *data.TaskWorkTime) error {
	session := d.conn.Session(context.Background())
	return session.Save(workTime).Error
}
