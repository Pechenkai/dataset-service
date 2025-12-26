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

type datasetFactResponse struct {
	DatasetID   uint64    `json:"dataset_id"`
	Fact        string    `json:"fact"`
	Length      int       `json:"length"`
	Source      string    `json:"source"`
	Mode        string    `json:"mode"`
	RetrievedAt time.Time `json:"retrieved_at"`
}

func TestE2E_DatasetFunFactFromExternalService(t *testing.T) {
	baseURL := strings.TrimRight(getBaseURL(t), "/")
	client := &http.Client{Timeout: 5 * time.Second}

	seed := seedDataset(t)

	status, fact := httpGetJSON[datasetFactResponse](t, client, fmt.Sprintf("%s/datasets/%d/fun-fact", baseURL, seed.datasetID))
	require.Equal(t, http.StatusOK, status)

	require.NotEmpty(t, fact.Fact)
	assert.Greater(t, fact.Length, 0)
	assert.Equal(t, seed.datasetID, fact.DatasetID)

	mode := strings.ToLower(os.Getenv("EXTERNAL_CATFACTS_MODE"))
	if mode == "" {
		mode = "mock"
	}
	assert.Equal(t, mode, strings.ToLower(fact.Mode))

	if mode == "mock" {
		mockURL := strings.TrimRight(os.Getenv("EXTERNAL_CATFACTS_MOCK_BASE_URL"), "/")
		if mockURL != "" {
			assert.Equal(t, mockURL, strings.TrimRight(fact.Source, "/"))
		}
	}
}
