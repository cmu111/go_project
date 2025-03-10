package repo

import (
	"context"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database"
)

type SourceLinkRepo interface {
	Save(ctx context.Context, conn database.DbConn, link *data.SourceLink) error
	FindByTaskCode(ctx context.Context, taskCode int64) (list []*data.SourceLink, err error)
}
