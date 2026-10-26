package internal

import (
	"database/sql"
	"fmt"

	"github.com/anggerpputro/go-gator-cli/internal/config"
	"github.com/anggerpputro/go-gator-cli/internal/database"
)

type State struct {
	Config *config.Config
	DB     *database.Queries
}

func NewState() (*State, error) {
	cfg, err := config.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	dbQueries := database.New(db)

	return &State{
		Config: cfg,
		DB:     dbQueries,
	}, nil
}
