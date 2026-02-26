package dashboard

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardStats_JSONSerialization(t *testing.T) {
	stats := DashboardStats{
		Posts: PostStats{
			Total:     150,
			Published: 120,
			Draft:     20,
			Scheduled: 10,
		},
		Users: UserStats{
			Total:    25,
			Active:   23,
			Inactive: 2,
		},
		Comments: CommentStats{
			Total:    300,
			Pending:  15,
			Approved: 280,
			Spam:     5,
		},
		Media: MediaStats{
			Total:       500,
			StorageUsed: 2684354560, // 2.5GB
		},
	}

	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var decoded DashboardStats
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, stats, decoded)
}

func TestDashboardStats_JSONFieldNames(t *testing.T) {
	stats := DashboardStats{
		Media: MediaStats{StorageUsed: 1024},
	}

	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	// Verify snake_case field naming
	assert.Contains(t, string(data), `"storage_used"`)
	assert.NotContains(t, string(data), `"StorageUsed"`)
}
