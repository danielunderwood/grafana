package database

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This is what the db sets empty time settings to
var DefaultTimeSettings, _ = simplejson.NewJson([]byte(`{}`))

// Default time to pass in with seconds rounded
var DefaultTime = time.Now().UTC().Round(time.Second)

// GetPublicDashboard
func TestIntegrationGetPublicDashboard(t *testing.T) {
	var sqlStore *sqlstore.SQLStore
	var dashboardStore *DashboardStore
	var savedDashboard *models.Dashboard

	setup := func() {
		sqlStore = sqlstore.InitTestDB(t)
		dashboardStore = ProvideDashboardStore(sqlStore)
		savedDashboard = insertTestDashboard(t, dashboardStore, "testDashie", 1, 0, true)
	}

	t.Run("returns PublicDashboard and Dashboard", func(t *testing.T) {
		setup()
		pubdash, err := dashboardStore.SavePublicDashboardConfig(context.Background(), models.SavePublicDashboardConfigCommand{
			PublicDashboard: models.PublicDashboard{
				IsEnabled:    true,
				Uid:          "abc1234",
				DashboardUid: savedDashboard.Uid,
				OrgId:        savedDashboard.OrgId,
				TimeSettings: DefaultTimeSettings,
				CreatedAt:    DefaultTime,
				CreatedBy:    7,
				AccessToken:  "NOTAREALUUID",
			},
		})
		require.NoError(t, err)

		pd, d, err := dashboardStore.GetPublicDashboard(context.Background(), "NOTAREALUUID")
		require.NoError(t, err)

		assert.Equal(t, pd, pubdash)
		assert.Equal(t, d.Uid, pubdash.DashboardUid)
	})

	t.Run("returns ErrPublicDashboardNotFound with empty uid", func(t *testing.T) {
		setup()
		_, _, err := dashboardStore.GetPublicDashboard(context.Background(), "")
		require.Error(t, models.ErrPublicDashboardIdentifierNotSet, err)
	})

	t.Run("returns ErrPublicDashboardNotFound when PublicDashboard not found", func(t *testing.T) {
		setup()
		_, _, err := dashboardStore.GetPublicDashboard(context.Background(), "zzzzzz")
		require.Error(t, models.ErrPublicDashboardNotFound, err)
	})

	t.Run("returns ErrDashboardNotFound when Dashboard not found", func(t *testing.T) {
		setup()
		_, err := dashboardStore.SavePublicDashboardConfig(context.Background(), models.SavePublicDashboardConfigCommand{
			DashboardUid: savedDashboard.Uid,
			OrgId:        savedDashboard.OrgId,
			PublicDashboard: models.PublicDashboard{
				IsEnabled:    true,
				Uid:          "abc1234",
				DashboardUid: "nevergonnafindme",
				OrgId:        savedDashboard.OrgId,
				CreatedAt:    DefaultTime,
				CreatedBy:    7,
			},
		})
		require.NoError(t, err)
		_, _, err = dashboardStore.GetPublicDashboard(context.Background(), "abc1234")
		require.Error(t, models.ErrDashboardNotFound, err)
	})
}

// GetPublicDashboardConfig
func TestIntegrationGetPublicDashboardConfig(t *testing.T) {
	var sqlStore *sqlstore.SQLStore
	var dashboardStore *DashboardStore
	var savedDashboard *models.Dashboard

	setup := func() {
		sqlStore = sqlstore.InitTestDB(t)
		dashboardStore = ProvideDashboardStore(sqlStore)
		savedDashboard = insertTestDashboard(t, dashboardStore, "testDashie", 1, 0, true)
	}

	t.Run("returns isPublic and set dashboardUid and orgId", func(t *testing.T) {
		setup()
		pubdash, err := dashboardStore.GetPublicDashboardConfig(context.Background(), savedDashboard.OrgId, savedDashboard.Uid)
		require.NoError(t, err)
		assert.Equal(t, &models.PublicDashboard{IsEnabled: false, DashboardUid: savedDashboard.Uid, OrgId: savedDashboard.OrgId}, pubdash)
	})

	t.Run("returns dashboard errDashboardIdentifierNotSet", func(t *testing.T) {
		setup()
		_, err := dashboardStore.GetPublicDashboardConfig(context.Background(), savedDashboard.OrgId, "")
		require.Error(t, models.ErrDashboardIdentifierNotSet, err)
	})

	t.Run("returns isPublic along with public dashboard when exists", func(t *testing.T) {
		setup()
		// insert test public dashboard
		resp, err := dashboardStore.SavePublicDashboardConfig(context.Background(), models.SavePublicDashboardConfigCommand{
			DashboardUid: savedDashboard.Uid,
			OrgId:        savedDashboard.OrgId,
			PublicDashboard: models.PublicDashboard{
				IsEnabled:    true,
				Uid:          "pubdash-uid",
				DashboardUid: savedDashboard.Uid,
				OrgId:        savedDashboard.OrgId,
				TimeSettings: DefaultTimeSettings,
				CreatedAt:    DefaultTime,
				CreatedBy:    7,
			},
		})
		require.NoError(t, err)

		pubdash, err := dashboardStore.GetPublicDashboardConfig(context.Background(), savedDashboard.OrgId, savedDashboard.Uid)
		require.NoError(t, err)

		assert.True(t, assert.ObjectsAreEqualValues(resp, pubdash))
		assert.True(t, assert.ObjectsAreEqual(resp, pubdash))
	})
}

// SavePublicDashboardConfig
func TestIntegrationSavePublicDashboardConfig(t *testing.T) {
	var sqlStore *sqlstore.SQLStore
	var dashboardStore *DashboardStore
	var savedDashboard *models.Dashboard
	var savedDashboard2 *models.Dashboard

	setup := func() {
		sqlStore = sqlstore.InitTestDB(t, sqlstore.InitTestDBOpt{FeatureFlags: []string{featuremgmt.FlagPublicDashboards}})
		dashboardStore = ProvideDashboardStore(sqlStore)
		savedDashboard = insertTestDashboard(t, dashboardStore, "testDashie", 1, 0, true)
		savedDashboard2 = insertTestDashboard(t, dashboardStore, "testDashie2", 1, 0, true)
	}

	t.Run("saves new public dashboard", func(t *testing.T) {
		setup()
		resp, err := dashboardStore.SavePublicDashboardConfig(context.Background(), models.SavePublicDashboardConfigCommand{
			DashboardUid: savedDashboard.Uid,
			OrgId:        savedDashboard.OrgId,
			PublicDashboard: models.PublicDashboard{
				IsEnabled:    true,
				Uid:          "pubdash-uid",
				DashboardUid: savedDashboard.Uid,
				OrgId:        savedDashboard.OrgId,
				TimeSettings: DefaultTimeSettings,
				CreatedAt:    DefaultTime,
				CreatedBy:    7,
				AccessToken:  "NOTAREALUUID",
			},
		})
		require.NoError(t, err)

		pubdash, err := dashboardStore.GetPublicDashboardConfig(context.Background(), savedDashboard.OrgId, savedDashboard.Uid)
		require.NoError(t, err)

		//verify saved response and queried response are the same
		assert.Equal(t, resp, pubdash)

		// verify we have a valid uid
		assert.True(t, util.IsValidShortUID(pubdash.Uid))

		// verify we didn't update all dashboards
		pubdash2, err := dashboardStore.GetPublicDashboardConfig(context.Background(), savedDashboard2.OrgId, savedDashboard2.Uid)
		require.NoError(t, err)
		assert.False(t, pubdash2.IsEnabled)
	})
}

func TestIntegrationnUpdatePublicDashboard(t *testing.T) {
	var sqlStore *sqlstore.SQLStore
	var dashboardStore *DashboardStore
	var savedDashboard *models.Dashboard

	setup := func() {
		sqlStore = sqlstore.InitTestDB(t, sqlstore.InitTestDBOpt{FeatureFlags: []string{featuremgmt.FlagPublicDashboards}})
		dashboardStore = ProvideDashboardStore(sqlStore)
		savedDashboard = insertTestDashboard(t, dashboardStore, "testDashie", 1, 0, true)
	}

	t.Run("updates an existing dashboard", func(t *testing.T) {
		setup()

		pdUid := "asdf1234"

		pdSaved, err := dashboardStore.SavePublicDashboardConfig(context.Background(), models.SavePublicDashboardConfigCommand{
			DashboardUid: savedDashboard.Uid,
			OrgId:        savedDashboard.OrgId,
			PublicDashboard: models.PublicDashboard{
				Uid:          pdUid,
				DashboardUid: savedDashboard.Uid,
				OrgId:        savedDashboard.OrgId,
				IsEnabled:    true,
				CreatedAt:    DefaultTime,
				CreatedBy:    7,
				AccessToken:  "NOTAREALUUID",
			},
		})
		require.NoError(t, err)

		// update initial record
		pdUpdated, err := dashboardStore.UpdatePublicDashboardConfig(context.Background(), models.SavePublicDashboardConfigCommand{
			DashboardUid: savedDashboard.Uid,
			OrgId:        savedDashboard.OrgId,
			PublicDashboard: models.PublicDashboard{
				Uid:          pdUid,
				DashboardUid: savedDashboard.Uid,
				OrgId:        savedDashboard.OrgId,
				IsEnabled:    false,
				TimeSettings: simplejson.NewFromAny(map[string]interface{}{"from": "now-8", "to": "now"}),
				UpdatedAt:    time.Now().UTC().Round(time.Second),
				UpdatedBy:    8,
			},
		})
		require.NoError(t, err)

		pdRetrieved, err := dashboardStore.GetPublicDashboardConfig(context.Background(), savedDashboard.OrgId, savedDashboard.Uid)
		require.NoError(t, err)

		// Some of these tests are potentially testing Xorm, however, they've been
		// left in for future developers to derive intent.

		// OrgId/dashboardUid haven't changed
		assert.Equal(t, pdSaved.OrgId, pdRetrieved.OrgId)
		assert.Equal(t, pdSaved.DashboardUid, pdRetrieved.DashboardUid)
		assert.Equal(t, pdSaved.CreatedBy, pdRetrieved.CreatedBy)
		assert.Equal(t, pdSaved.CreatedAt, pdRetrieved.CreatedAt)

		// created hasn't changed
		assert.Equal(t, pdSaved.CreatedBy, pdRetrieved.CreatedBy)
		assert.Equal(t, pdSaved.CreatedAt, pdRetrieved.CreatedAt)

		// Enabled has changed
		assert.Equal(t, pdUpdated.IsEnabled, pdRetrieved.IsEnabled)

		// Updated has been set
		assert.Equal(t, pdUpdated.UpdatedBy, pdRetrieved.UpdatedBy)
		assert.Equal(t, pdUpdated.UpdatedAt, pdRetrieved.UpdatedAt)
	})
}
