//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type datasetSummaryResponse struct {
	DatasetID uint64    `json:"dataset_id"`
	Summary   string    `json:"summary"`
	Model     string    `json:"model"`
	Mode      string    `json:"mode"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

func TestE2E_DatasetSummaryFromOpenAI(t *testing.T) {
	baseURL := strings.TrimRight(getBaseURL(t), "/")
	client := &http.Client{Timeout: 8 * time.Second}

	seed := seedDataset(t)

	status, summary := httpGetJSON[datasetSummaryResponse](t, client, fmt.Sprintf("%s/datasets/%d/summary", baseURL, seed.datasetID))
	require.Equal(t, http.StatusOK, status)

	assert.Equal(t, seed.datasetID, summary.DatasetID)
	require.NotEmpty(t, summary.Summary)
	require.NotEmpty(t, summary.Model)
	require.NotEmpty(t, summary.Source)

	mode := strings.ToLower(strings.TrimSpace(getEnvDefault("EXTERNAL_OPENAI_MODE", "mock")))
	assert.Equal(t, mode, strings.ToLower(summary.Mode))
}

func getEnvDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
