package servicequotas_test

import (
	"testing"

	"github.com/TeiNam/prometheus-rds-exporter/internal/app/servicequotas"
	mock "github.com/TeiNam/prometheus-rds-exporter/internal/app/servicequotas/mock"
	converter "github.com/TeiNam/prometheus-rds-exporter/internal/app/unit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRDSQuotas(t *testing.T) {
	client := mock.ServiceQuotasClient{}

	result, err := servicequotas.NewFetcher(client).GetRDSQuotas()
	require.NoError(t, err, "GetRDSQuotas must succeed")
	assert.Equal(t, mock.DBinstancesQuota, result.DBinstances, "DbInstance quota is incorrect")
	assert.Equal(t, converter.GigaBytesToBytes(mock.TotalStorage), result.TotalStorage, "Total storage quota is incorrect")
	assert.Equal(t, mock.ManualDBInstanceSnapshots, result.ManualDBInstanceSnapshots, "Manual db instance snapshot quota is incorrect")
}
