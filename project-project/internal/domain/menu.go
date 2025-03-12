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

type ProjectMenuDomain struct {
	projectAuthRepo repo.ProjectAuthRepo
	userRpcDomain   *UserRpcDomain
	menuRepo        repo.MenuRepo
}

func NewMenuDomain() *ProjectMenuDomain {
	return &ProjectMenuDomain{
		projectAuthRepo: dao.NewProjectAuthDao(),
		userRpcDomain:   NewUserRpcDomain(),
		menuRepo:        dao.NewMenuDao(),
	}
}

func (m *ProjectMenuDomain) MenuList() (list []*data.ProjectMenuChild, err error) {
	pms, err := m.menuRepo.FindMenus(context.Background())
	if err != nil {
		zap.L().Error("Index db FindMenus error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	childs := data.CovertChild(pms)
	return childs, nil
}
