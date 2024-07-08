package db

import (
	"context"
)

type LavalinkNode struct {
	Name      string `db:"name"`
	SessionID string `db:"session_id"`
}

func (d *DB) GetLavalinkNodes(ctx context.Context) ([]LavalinkNode, error) {
	var nodes []LavalinkNode
	if err := d.dbx.SelectContext(ctx, &nodes, "SELECT * FROM lavalink_nodes"); err != nil {
		return nil, err
	}

	return nodes, nil
}

func (d *DB) AddLavalinkNodes(ctx context.Context, nodes []LavalinkNode) error {
	_, err := d.dbx.NamedExecContext(ctx, "INSERT INTO lavalink_nodes (name, session_id) VALUES (:name, :session_id) ON CONFLICT DO UPDATE SET session_id = :session_id", nodes)
	return err
}
