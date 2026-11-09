package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Hostname  string    `json:"hostname"`
	OS        string    `json:"os"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ServerRepository struct {
	db *pgxpool.Pool
}

func NewServerRepository(db *pgxpool.Pool) *ServerRepository {
	return &ServerRepository{
		db: db,
	}
}

func (r *ServerRepository) FindAll(ctx context.Context, userID string) ([]Server, error) {
	query := `
		SELECT
			id::text,
			name,
			COALESCE(hostname, ''),
			COALESCE(os, ''),
			status,
			created_at
		FROM servers
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := make([]Server, 0)

	for rows.Next() {
		var server Server

		err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.Hostname,
			&server.OS,
			&server.Status,
			&server.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

// Create inserts a new server into PostgreSQL.
func (r *ServerRepository) Create(
	ctx context.Context,
	userID string,
	name string,
	hostname string,
	os string,
) (*Server, error) {

	query := `
		INSERT INTO servers (
			user_id,
			name,
			hostname,
			os
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			NULLIF($4, '')
		)
		RETURNING
			id::text,
			name,
			COALESCE(hostname, ''),
			COALESCE(os, ''),
			status,
			created_at
	`

	var server Server

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		name,
		hostname,
		os,
	).Scan(
		&server.ID,
		&server.Name,
		&server.Hostname,
		&server.OS,
		&server.Status,
		&server.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (r *ServerRepository) FindByID(
	ctx context.Context,
	serverID string,
	userID string,
) (*Server, error) {

	query := `
        SELECT
            id::text,
            name,
            COALESCE(hostname, ''),
            COALESCE(os, ''),
            status,
            created_at
        FROM servers
        WHERE id = $1
		AND user_id = $2
    `

	var server Server

	err := r.db.QueryRow(
		ctx,
		query,
		serverID,
		userID,
	).Scan(
		&server.ID,
		&server.Name,
		&server.Hostname,
		&server.OS,
		&server.Status,
		&server.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (r *ServerRepository) Delete(
	ctx context.Context,
	serverID string,
	userID string,
) (bool, error) {
	result, err := r.db.Exec(ctx, `
		DELETE FROM servers
		WHERE id = $1 AND user_id = $2
	`, serverID, userID)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}
