package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_config"
)

func TestMonsterService_validateJSONFields(t *testing.T) {
	svc := &MonsterService{}

	monster := &game_config.Monster{}
	require.NoError(t, monster.DamageResistances.Marshal(map[string]interface{}{
		"FIRE_RESIST": 0.5,
		"ICE_DR":      10,
	}))
	require.NoError(t, monster.PassiveBuffs.Marshal([]interface{}{
		map[string]interface{}{"buff_id": "TEST"},
	}))

	require.NoError(t, svc.validateJSONFields(monster))
}

func TestMonsterService_validateJSONFields_Invalid(t *testing.T) {
	svc := &MonsterService{}
	monster := &game_config.Monster{}
	require.NoError(t, monster.DamageResistances.Marshal(map[string]interface{}{
		"FIRE_RESIST": 2.0,
	}))

	err := svc.validateJSONFields(monster)
	require.Error(t, err)
}
