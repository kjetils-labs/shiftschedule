package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/shiftschedule/internal/models"
)

func mapRowToPersonnel(rows pgx.Rows) (*models.Personnel, error) {
	var p models.Personnel
	if err := rows.Scan(&p.ID, &p.Name); err != nil {
		return &models.Personnel{}, err
	}
	return &p, nil
}
