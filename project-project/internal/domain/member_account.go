package domain

import (
	"context"
	"fmt"

	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/repo"
	"test.com/project-project/pkg/model"
)

type AccountDomain struct {
	accountRepo      repo.AccountRepo
	userRpcDomain    *UserRpcDomain
	departmentDomain *DepartmentDomain
}

func NewAccountDomain() *AccountDomain {
	return &AccountDomain{
		accountRepo:      dao.NewAccountDao(),
		userRpcDomain:    NewUserRpcDomain(),
		departmentDomain: NewDepartmentDomain(),
	}
}

func (d *AccountDomain) AccountList(
	organizationCode string,
	memberId int64,
	page int64,
	pageSize int64,
	departmentCode string,
	searchType int32) ([]*data.MemberAccountDisplay, int64, *errs.BError) {
	var condition string
	departmentCodeId := encrypts.DecryptNoErr(departmentCode)
	organizationCodeId := encrypts.DecryptNoErr(organizationCode)
	switch searchType {
	case 1:
		condition = "status = 1"
	case 2:
		condition = "department_code = NULL"
	case 3:
		condition = "status = 0"
	case 4:
		condition = fmt.Sprintf("status = 1 and department_code = %d", departmentCodeId)
	default:
		condition = "status = 1"
	}
	ctx := context.Background()
	accountList, total, err := d.accountRepo.FindList(ctx, condition, organizationCodeId, departmentCodeId, page, pageSize)
	if err != nil {
		return nil, 0, model.DBError
	}
	if accountList == nil {
		return []*data.MemberAccountDisplay{}, 0, model.DBError
	}
	var dList []*data.MemberAccountDisplay
	for _, v := range accountList {
		display := v.ToDisplay()
		memberInfo, _ := d.userRpcDomain.MemberInfo(ctx, v.MemberCode)
		display.Avatar = memberInfo.Avatar
		if v.DepartmentCode > 0 {
			department, err := d.departmentDomain.FindDepartmentById(v.DepartmentCode)
			if err != nil {
				return nil, 0, err
			}
			display.Departments = department.Name
		}
		dList = append(dList, display)
	}
	return dList, total, nil

}
