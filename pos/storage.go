package pos

import (
	"context"
	"database/sql"

	"github.com/drscre/pos_pipeline/pipeline"
)

type authorizationStorage struct {
	db *sql.DB
}

func (a authorizationStorage) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return a.db.BeginTx(ctx, opts)
}

func (a authorizationStorage) SelectForUpdate(conn pipeline.DBConn, id string) (pipeline.State, error) {
	return pipeline.State{
		Data: Authorization{}, // <-- deserialize into specific type (do we really need this?)
	}, nil
}

func (a authorizationStorage) Update(conn pipeline.DBConn, id string, newData interface{}, completedStep string) error {
	_ = newData.(Authorization)
	return nil
}
