package auth_service_v1

import (
	"context"

	"github.com/jinzhu/copier"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-grpc/auth"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/database/tran"
	"test.com/project-project/internal/domain"
	"test.com/project-project/internal/repo"
)

type AuthService struct {
	auth.UnimplementedAuthServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	accountDomain     domain.AccountDomain
	projectAuthDomain domain.ProjectAuthDomain
}

func New() *AuthService {
	return &AuthService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		accountDomain:     *domain.NewAccountDomain(),
		projectAuthDomain: *domain.NewProjectAuthDomain(),
	}
}

func (as *AuthService) AuthList(ctx context.Context, authMsg *auth.AuthReqMessage) (*auth.ListAuthMessage, error) {
	organizationCode := encrypts.DecryptNoErr(authMsg.OrganizationCode)
	page := authMsg.Page
	pageSize := authMsg.PageSize
	projectAuthList, total, err := as.projectAuthDomain.AuthListWithPage(organizationCode, page, pageSize)
	if err != nil {
		return nil, err
	}
	listAuth := []*auth.ProjectAuth{}
	copier.Copy(&listAuth, projectAuthList)
	return &auth.ListAuthMessage{
		List:  listAuth,
		Total: total,
	}, nil
}
func (a *AuthService) Apply(ctx context.Context, msg *auth.AuthReqMessage) (*auth.ApplyResponse, error) {
	if msg.Action == "getnode" {
		//获取列表
		list, checkedList, err := a.projectAuthDomain.AllNodeAndAuth(msg.AuthId)
		if err != nil {
			return nil, errs.GrpcError(err)
		}
		var prList []*auth.ProjectNodeMessage
		copier.Copy(&prList, list)
		return &auth.ApplyResponse{List: prList, CheckedList: checkedList}, nil
	}
	// if msg.Action == "save" {
	// 	//先删除 project_auth_node表 在新增  事务
	// 	//保存
	// 	nodes := msg.Nodes
	// 	//先删在存 加事务
	// 	authId := msg.AuthId
	// 	err := a.transaction.Action(func(conn database.DbConn) error {
	// 		err := a.projectAuthDomain.Save(conn, authId, nodes)
	// 		return err
	// 	})
	// 	if err != nil {
	// 		return nil, errs.GrpcError(err.(*errs.BError))
	// 	}
	// }
	return &auth.ApplyResponse{}, nil
}
