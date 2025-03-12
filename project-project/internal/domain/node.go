package domain

import (
	"context"

	"go.uber.org/zap"
	"test.com/project-common/errs"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/repo"
	"test.com/project-project/pkg/model"
)

type Node struct {
	nodeRepo repo.NodeRepo
}

func NewNode() *Node {
	return &Node{
		nodeRepo: dao.NewNodeDao(),
	}
}

func (n *Node) FindNodes() ([]*data.ProjectNodeTree, *errs.BError) {
	ctx := context.Background()
	pnList, err := n.nodeRepo.FindNodes(ctx)
	if err != nil {
		zap.L().Error("FindNodes failed", zap.Error(err))
		return nil, model.DBError
	}
	pns := data.ToNodeTreeList(pnList)
	return pns, nil
}

func (d *Node) NodeList() ([]*data.ProjectNode, *errs.BError) {
	list, err := d.nodeRepo.FindNodes(context.Background())
	if err != nil {
		return nil, model.DBError
	}
	return list, nil
}
