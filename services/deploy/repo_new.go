package deploy

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const cacheTTL = 24 * time.Hour

func (s *Dservice) cacheDeployment(dep *Deployment_New) error {
	if dep == nil || dep.ID == "" {
		return fmt.Errorf("invalid deployment: missing ID")
	}

	key := s.getCacheKey(dep.ID, dep.UserID)

	data, err := json.Marshal(dep)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment: %w", err)
	}

	err = s.repo.redis.Set(key, data, cacheTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to cache deployment: %w", err)
	}

	return nil
}

func (s *Dservice) getCachedDeployment(id string, userid string) (*Deployment_New, error) {
	key := s.getCacheKey(id, userid)

	val, err := s.repo.redis.Get(key).Bytes()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get deployment from cache: %w", err)
	}

	var dep Deployment_New
	if err := json.Unmarshal(val, &dep); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached deployment: %w", err)
	}

	return &dep, nil
}

func (s *Dservice) getAllCachedDeploymentsOfUser(userid string) ([]*Deployment_New, error) {
	var deployments []*Deployment_New

	pattern := fmt.Sprintf("deployment:*:%s", userid)

	iter := s.repo.redis.Scan(0, pattern, 0).Iterator()
	for iter.Next() {
		key := iter.Val()

		data, err := s.repo.redis.Get(key).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		dep := new(Deployment_New)
		if err := json.Unmarshal(data, dep); err != nil {
			return nil, err
		}

		if dep.Status == STATUS_PENDING {
			dep.UserType = deploymentTypeMapServer[dep.Type]
			deployments = append(deployments, dep)
		}

	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return deployments, nil
}

func (s *Dservice) deleteCachedDeployment(id string, userid string) error {
	key := s.getCacheKey(id, userid)

	if err := s.repo.redis.Del(key).Err(); err != nil {
		return fmt.Errorf("failed to delete cached deployment: %w", err)
	}

	return nil
}
