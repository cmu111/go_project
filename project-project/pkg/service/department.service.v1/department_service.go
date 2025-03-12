package department_service_v1

import (
	"context"

	"github.com/jinzhu/copier"
	"test.com/project-common/encrypts"
	"test.com/project-grpc/department"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/database/tran"
	"test.com/project-project/internal/domain"
	"test.com/project-project/internal/repo"
)

type DepartmentService struct {
	department.UnimplementedDepartmentServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	accountDomain     domain.AccountDomain
	projectAuthDomain domain.ProjectAuthDomain
	departmentDomain  domain.DepartmentDomain
}

func New() *DepartmentService {
	return &DepartmentService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		accountDomain:     *domain.NewAccountDomain(),
		projectAuthDomain: *domain.NewProjectAuthDomain(),
		departmentDomain:  *domain.NewDepartmentDomain(),
	}
}

func (d *DepartmentService) List(ctx context.Context, req *department.DepartmentReqMessage) (*department.ListDepartmentMessage, error) {
	organizationCode := encrypts.DecryptNoErr(req.OrganizationCode)
	page := req.Page
	pageSize := req.PageSize
	var pcode int64
	if req.ParentDepartmentCode != "" {
		pcode = encrypts.DecryptNoErr(req.ParentDepartmentCode)
	}
	departmentList, total, err := d.departmentDomain.List(organizationCode, pcode, page, pageSize)
	if err != nil {
		return nil, err
	}
	var departmentMessages []*department.DepartmentMessage
	copier.Copy(&departmentMessages, departmentList)
	return &department.ListDepartmentMessage{
		List:  departmentMessages,
		Total: total,
	}, nil
}

func (d *DepartmentService) Save(ctx context.Context, req *department.DepartmentReqMessage) (*department.DepartmentMessage, error) {
	var departmentCode int64
	if req.DepartmentCode != "" {
		departmentCode = encrypts.DecryptNoErr(req.DepartmentCode)
	}
	var Pcode int64
	if req.ParentDepartmentCode != "" {
		Pcode = encrypts.DecryptNoErr(req.ParentDepartmentCode)
	}
	organizationCode := encrypts.DecryptNoErr(req.OrganizationCode)
	dpDisplay, err := d.departmentDomain.Save(organizationCode, departmentCode, Pcode, req.Name)
	if err != nil {
		return nil, err
	}
	var departmentMessage department.DepartmentMessage
	copier.Copy(&departmentMessage, dpDisplay)
	return &departmentMessage, nil
}

func (d *DepartmentService) Read(ctx context.Context, req *department.DepartmentReqMessage) (*department.DepartmentMessage, error) {
	var departmentCode int64
	if req.DepartmentCode != "" {
		departmentCode = encrypts.DecryptNoErr(req.DepartmentCode)
	}
	dpDisplay, err := d.departmentDomain.FindDepartmentById(departmentCode)
	if err != nil {
		return nil, err
	}
	var departmentMessage department.DepartmentMessage
	copier.Copy(&departmentMessage, dpDisplay)
	return &departmentMessage, nil
}
