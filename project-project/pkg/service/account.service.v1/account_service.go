package account_service_v1

import (
	"context"

	"github.com/jinzhu/copier"
	"test.com/project-common/encrypts"
	"test.com/project-grpc/account"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/database/tran"
	"test.com/project-project/internal/domain"
	"test.com/project-project/internal/repo"
)

type AccountService struct {
	account.UnimplementedAccountServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	accountDomain     domain.AccountDomain
	projectAuthDomain domain.ProjectAuthDomain
}

func New() *AccountService {
	return &AccountService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		accountDomain:     *domain.NewAccountDomain(),
		projectAuthDomain: *domain.NewProjectAuthDomain(),
	}
}
func (a *AccountService) Account(ctx context.Context, msg *account.AccountReqMessage) (*account.AccountResponse, error) {
	organization := msg.OrganizationCode
	department := msg.DepartmentCode
	memberId := msg.MemberId
	page := msg.Page
	pageSize := msg.PageSize
	searchType := msg.SearchType
	organizationCode := encrypts.DecryptNoErr(organization)
	accountList, total, err := a.accountDomain.AccountList(organization, memberId, page, pageSize, department, searchType)
	if err != nil {
		return nil, err
	}
	authList, err := a.projectAuthDomain.AuthList(organizationCode)
	if err != nil {
		return nil, err
	}
	var accountMsgList []*account.MemberAccount
	var authMsgList []*account.ProjectAuth
	copier.Copy(&accountMsgList, accountList)
	copier.Copy(&authMsgList, authList)
	return &account.AccountResponse{
		Total:       total,
		AccountList: accountMsgList,
		AuthList:    authMsgList,
	}, nil
}
