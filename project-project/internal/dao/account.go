package dao

import (
	"context"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database/gorms"
)

type AccountDao struct {
	conn gorms.GormConn
}

func NewAccountDao() *AccountDao {
	return &AccountDao{
		conn: *gorms.New(),
	}
}

func (a *AccountDao) FindList(ctx context.Context, condition string, organizationCode int64, departmentCode int64, page int64, pageSize int64) (list []*data.MemberAccount, total int64, err error) {
	session := a.conn.Session(ctx)
	offset := (page - 1) * pageSize
	err = session.Model(&data.MemberAccount{}).
		Where("organization_code=?", organizationCode).
		Where(condition).Limit(int(pageSize)).Offset(int(offset)).Find(&list).Error
	err = session.Model(&data.MemberAccount{}).
		Where("organization_code=?", organizationCode).
		Where(condition).Count(&total).Error
	return
}
