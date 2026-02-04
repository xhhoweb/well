package snowflake

import (
	"sync"

	"github.com/bwmarrin/snowflake"
	"well_go/internal/core/config"
	"well_go/internal/core/logger"
)

var (
	node     *snowflake.Node
	nodeOnce sync.Once
)

// Init Initialize snowflake generator
func Init(cfg *config.SnowflakeConfig) error {
	var initErr error
	nodeOnce.Do(func() {
		var err error
		node, err = snowflake.NewNode(cfg.WorkerID)
		if err != nil {
			logger.Error("failed to initialize snowflake",
				logger.String("error", err.Error()),
				logger.Int64("worker_id", cfg.WorkerID))
			initErr = err
			return
		}
		logger.Info("snowflake initialized",
			logger.Int64("worker_id", cfg.WorkerID))
	})
	return initErr
}

// Generate Generate new snowflake ID
func Generate() int64 {
	return node.Generate().Int64()
}

// GetNode Get snowflake node
func GetNode() *snowflake.Node {
	return node
}
