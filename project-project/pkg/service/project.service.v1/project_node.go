package project_service_v1

import (
	"context"

	"github.com/jinzhu/copier"
	"test.com/project-grpc/project"
)

func (p *ProjectService) NodeList(context.Context, *project.ProjectRpcMessage) (*project.ProjectNodeResponseMessage, error) {
	pns, err := p.nodeDomain.FindNodes()
	if err != nil {
		return nil, err
	}
	var nodesList []*project.ProjectNodeMessage
	copier.Copy(&nodesList, pns)
	return &project.ProjectNodeResponseMessage{Nodes: nodesList}, nil
}
