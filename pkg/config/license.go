package config

import (
	"errors"
	"fmt"
	"time"
)

const (
	EditionFree       = "free"
	EditionCertified  = "certified"
	EditionEnterprise = "enterprise"
)

type Limits struct {
	MaxManagedClusters int `json:"max_managed_clusters"` // -1 = unlimited
	MaxTotalNodes      int `json:"max_total_nodes"`      // -1 = unlimited
}

type License struct {
	ID        string    `json:"license_id"`
	Customer  string    `json:"customer"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"` // 企业版可为空
	Edition   string    `json:"edition"`
	Limits    Limits    `json:"limits"`
	Features  []string  `json:"features,omitempty"`
}

// IsUnlimited returns true if the given limit is -1
func (l *Limits) IsUnlimited() bool {
	return l.MaxManagedClusters == -1 && l.MaxTotalNodes == -1
}

func (lic *License) Validate(clusterNumber, nodeNumber int) error {
	// 1. 检查是否过期（企业版可无 ExpiresAt）
	if !lic.ExpiresAt.IsZero() && time.Now().After(lic.ExpiresAt) {
		return errors.New("license has expired")
	}

	// 2. 检查集群数量
	if lic.Limits.MaxManagedClusters != -1 && clusterNumber > lic.Limits.MaxManagedClusters {
		return fmt.Errorf(
			"exceeded cluster limit: %d > %d",
			clusterNumber, lic.Limits.MaxManagedClusters,
		)
	}

	// 3. 检查总节点数
	if lic.Limits.MaxTotalNodes != -1 {
		if nodeNumber > lic.Limits.MaxTotalNodes {
			return fmt.Errorf(
				"exceeded total node limit: %d > %d",
				nodeNumber, lic.Limits.MaxTotalNodes,
			)
		}
	}

	return nil
}
