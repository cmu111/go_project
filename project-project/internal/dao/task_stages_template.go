package dao

import (
	"context"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database/gorms"
)

type TaskStagesTemplateDao struct {
	conn *gorms.GormConn
}

func (t *TaskStagesTemplateDao) FindInProTemIds(ctx context.Context, ids []int) ([]data.MsTaskStagesTemplate, error) {
	var tsts []data.MsTaskStagesTemplate
	session := t.conn.Session(ctx)
	err := session.Where("project_template_code in ?", ids).Find(&tsts).Error
	return tsts, err
}

func (t *TaskStagesTemplateDao) FindByProjectTemplateId(ctx context.Context, id int) ([]*data.MsTaskStagesTemplate, error) {
	var tsts []*data.MsTaskStagesTemplate
	session := t.conn.Session(ctx)
	err := session.Where("project_template_code = ?", id).Order("sort desc,id asc").Find(&tsts).Error
	return tsts, err
}

func NewTaskStagesTemplateDao() *TaskStagesTemplateDao {
	return &TaskStagesTemplateDao{
		conn: gorms.New(),
	}
}
