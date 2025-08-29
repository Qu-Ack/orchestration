package deploy

import (
	"database/sql"

	"github.com/go-redis/redis"
)

type repo struct {
	db    *sql.DB
	redis *redis.Client
}

func (r *repo) NEW_DEPLOYMENT() {

}

func (r *repo) NEW_SERVICE() {

}

func (r *repo) NEW_ENV_VAR() {

}
