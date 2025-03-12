package domain

import (
	"context"

	"test.com/project-common/errs"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/repo"
	"test.com/project-user/pkg/model"
)

type ProjectAuthDomain struct {
	projectAuthRepo       repo.ProjectAuthRepo
	userRpcDomain         *UserRpcDomain
	projectAuthNodeDomain *ProjectAuthNodeDomain
	nodeDomain            *Node
}

func NewProjectAuthDomain() *ProjectAuthDomain {
	return &ProjectAuthDomain{
		projectAuthRepo:       dao.NewProjectAuthDao(),
		userRpcDomain:         NewUserRpcDomain(),
		projectAuthNodeDomain: NewProjectAuthNodeDomain(),
		nodeDomain:            NewNode(),
	}
}

func (d *ProjectAuthDomain) AuthList(orgCode int64) ([]*data.ProjectAuthDisplay, *errs.BError) {
	ctx := context.Background()
	list, err := d.projectAuthRepo.FindAuthList(ctx, orgCode)
	if err != nil {
		return nil, model.DBError
	}
	if list == nil {
		return []*data.ProjectAuthDisplay{}, nil
	}
	displayList := make([]*data.ProjectAuthDisplay, 0)
	for _, v := range list {
		authDis := v.ToDisplay()
		displayList = append(displayList, authDis)
	}
	return displayList, nil
}

func (d *ProjectAuthDomain) AuthListWithPage(orgCode int64, page int64, pageSize int64) ([]*data.ProjectAuthDisplay, int64, *errs.BError) {
	ctx := context.Background()
	list, total, err := d.projectAuthRepo.FindAuthListPage(ctx, orgCode, page, pageSize)
	if err != nil {
		return nil, total, model.DBError
	}
	if list == nil {
		return []*data.ProjectAuthDisplay{}, 0, nil
	}
	displayList := make([]*data.ProjectAuthDisplay, 0)
	for _, v := range list {
		authDis := v.ToDisplay()
		displayList = append(displayList, authDis)
	}
	return displayList, total, nil
}

func (d *ProjectAuthDomain) AllNodeAndAuth(authId int64) ([]*data.ProjectNodeAuthTree, []string, *errs.BError) {
	nodeList, err := d.nodeDomain.NodeList()
	if err != nil {
		return nil, nil, err
	}
	checkedList, err := d.projectAuthNodeDomain.AuthNodeList(authId)
	if err != nil {
		return nil, nil, err
	}
	list := data.ToAuthNodeTreeList(nodeList, checkedList)
	return list, checkedList, nil
}
