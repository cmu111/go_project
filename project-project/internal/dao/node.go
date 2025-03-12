package dao

import (
	"context"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database/gorms"
)

type NodeDao struct {
	conn *gorms.GormConn
}

func (m *NodeDao) FindNodes(ctx context.Context) (pns []*data.ProjectNode, err error) {
	session := m.conn.Session(ctx)
	err = session.Find(&pns).Error
	return
}

func NewNodeDao() *NodeDao {
	return &NodeDao{
		conn: gorms.New(),
	}
}
