package integration_test

import (
	_ "github.com/jackc/pgx/v5/stdlib" // Importing `pgx/v5/stdlib` is necessary for `sql.Open("pgx", s.dsn)`.
)
