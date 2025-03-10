package repo

import (
	"context"

	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database"
)

type FileRepo interface {
	Save(ctx context.Context, conn database.DbConn, file *data.File) error
	FindByIds(background context.Context, ids []int64) (list []*data.File, err error)
}
