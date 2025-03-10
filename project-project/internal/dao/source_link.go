package dao

import (
	"context"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database"
	"test.com/project-project/internal/database/gorms"
)

type SourceLinkDao struct {
	conn *gorms.GormConn
}

func (s *SourceLinkDao) Save(ctx context.Context, conn database.DbConn, link *data.SourceLink) error {
	s.conn = conn.(*gorms.GormConn)
	session := s.conn.Tx(ctx)
	return session.Save(&link).Error
}

func (s *SourceLinkDao) FindByTaskCode(ctx context.Context, taskCode int64) (list []*data.SourceLink, err error) {
	session := s.conn.Session(ctx)
	err = session.Model(&data.SourceLink{}).Where("link_type=? and link_code=?", "task", taskCode).Find(&list).Error
	return
}

func NewSourceLinkDao() *SourceLinkDao {
	return &SourceLinkDao{
		conn: gorms.New(),
	}
}
