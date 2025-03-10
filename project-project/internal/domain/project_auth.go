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
	projectAuthRepo repo.ProjectAuthRepo
	userRpcDomain   *UserRpcDomain
}

func NewProjectAuthDomain() *ProjectAuthDomain {
	return &ProjectAuthDomain{
		projectAuthRepo: dao.NewProjectAuthDao(),
		userRpcDomain:   NewUserRpcDomain(),
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
