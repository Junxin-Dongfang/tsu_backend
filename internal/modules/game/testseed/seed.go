package testseed

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// StableUUID returns a deterministic UUID (v5-like) based on label to keep fixtures repeatable.
func StableUUID(label string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(label))
}

// ensureBaseClass inserts a minimal playable class used by tests.
func ensureBaseClass(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()

	classID := StableUUID("class:basic:test")
	_, err := db.Exec(`
		INSERT INTO game_config.classes (id, class_code, class_name, tier)
		VALUES ($1, $2, $3, 'basic')
		ON CONFLICT (id) DO NOTHING
	`, classID, "class-basic-test", "测试基础职业")
	require.NoError(t, err, "seed class failed")

	return classID
}

// EnsureUser creates a test user record if missing and returns its UUID.
func EnsureUser(t *testing.T, db *sql.DB, label string) uuid.UUID {
	t.Helper()

	userID := StableUUID("user:" + label)
	username := fmt.Sprintf("test_%s", strings.ReplaceAll(label, " ", "_"))
	email := fmt.Sprintf("%s@example.com", username)

	_, err := db.Exec(`
		INSERT INTO auth.users (id, username, email)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, userID, username, email)
	require.NoError(t, err, "seed user failed")

	return userID
}

// EnsureHero creates a hero bound to the given user with deterministic ID and base class.
func EnsureHero(t *testing.T, db *sql.DB, userID uuid.UUID, label string) uuid.UUID {
	t.Helper()

	classID := ensureBaseClass(t, db)
	heroID := StableUUID("hero:" + label)
	heroName := fmt.Sprintf("测试英雄-%s", label)

	_, err := db.Exec(`
		INSERT INTO game_runtime.heroes (
			id, user_id, class_id, hero_name,
			current_level, experience_total, experience_available, experience_spent, status)
		VALUES ($1, $2, $3, $4, 1, 0, 0, 0, 'active')
		ON CONFLICT (id) DO NOTHING
	`, heroID, userID, classID, heroName)
	require.NoError(t, err, "seed hero failed")

	return heroID
}

// CleanupTeamsByHero removes any team/membership records created with the given leader hero to keep tests isolated.
func CleanupTeamsByHero(t *testing.T, db *sql.DB, heroID uuid.UUID) {
	t.Helper()

	_, _ = db.Exec(`DELETE FROM game_runtime.team_members WHERE team_id IN (SELECT id FROM game_runtime.teams WHERE leader_hero_id = $1)`, heroID)
	_, _ = db.Exec(`DELETE FROM game_runtime.team_warehouses WHERE team_id IN (SELECT id FROM game_runtime.teams WHERE leader_hero_id = $1)`, heroID)
	_, _ = db.Exec(`DELETE FROM game_runtime.teams WHERE leader_hero_id = $1`, heroID)
}
