package repo

import (
	"context"

	"test.com/project-project/internal/data"
)

type NodeRepo interface {
	FindNodes(ctx context.Context) (pns []*data.ProjectNode, err error)
}
