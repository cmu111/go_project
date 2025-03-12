package menu_service_v1

import (
	"context"

	"github.com/jinzhu/copier"
	"test.com/project-grpc/menu"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/database/tran"
	"test.com/project-project/internal/domain"
	"test.com/project-project/internal/repo"
)

type MenuService struct {
	menu.UnimplementedMenuServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	projectMenuDomain *domain.ProjectMenuDomain
}

func New() *MenuService {
	return &MenuService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		projectMenuDomain: domain.NewMenuDomain(),
	}
}

func (m *MenuService) MenuList(ctx context.Context, req *menu.MenuReqMessage) (*menu.MenuResponseMessage, error) {
	childs, err := m.projectMenuDomain.MenuList()
	if err != nil {
		return nil, err
	}
	var menuList []*menu.MenuMessage
	copier.Copy(&menuList, childs)
	return &menu.MenuResponseMessage{
		List: menuList,
	}, nil

}
