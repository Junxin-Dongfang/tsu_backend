package impl

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

func TestMonsterRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMonsterRepository(db)
	ctx := context.Background()

	monster := &game_config.Monster{
		MonsterCode:  "TEST_MONSTER",
		MonsterName:  "测试怪物",
		MonsterLevel: 5,
		MaxHP:        100,
	}

	t.Run("成功创建怪物", func(t *testing.T) {
		returnCols := []string{
			"description", "hp_recovery", "max_mp", "mp_recovery", "base_str", "base_agi", "base_vit", "base_wlp",
			"base_int", "base_wis", "base_cha", "accuracy_attribute_code", "dodge_attribute_code", "initiative_attribute_code",
			"body_resist_attribute_code", "magic_resist_attribute_code", "mental_resist_attribute_code", "environment_resist_attribute_code",
			"damage_resistances", "passive_buffs", "drop_gold_min", "drop_gold_max", "drop_exp", "icon_url", "model_url",
			"is_active", "display_order", "deleted_at",
		}
		returnValues := make([]driver.Value, len(returnCols))
		rows := sqlmock.NewRows(returnCols).AddRow(returnValues...)

		mock.ExpectQuery(`INSERT INTO "game_config"\."monsters"`).
			WithArgs(
				sqlmock.AnyArg(), // id
				monster.MonsterCode,
				monster.MonsterName,
				monster.MonsterLevel,
				monster.MaxHP,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnRows(rows)

		err := repo.Create(ctx, monster)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMonsterRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMonsterRepository(db)
	ctx := context.Background()

	monsterID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("成功获取怪物", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "monster_code", "monster_name", "monster_level", "max_hp",
			"created_at", "updated_at",
		}).AddRow(
			monsterID, "TEST_MONSTER", "测试怪物", 5, 100,
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT .+ FROM "game_config"\."monsters" WHERE .*id = \$1.*deleted_at IS NULL`).
			WithArgs(monsterID).
			WillReturnRows(rows)

		monster, err := repo.GetByID(ctx, monsterID)
		assert.NoError(t, err)
		assert.NotNil(t, monster)
		assert.Equal(t, "TEST_MONSTER", monster.MonsterCode)
		assert.Equal(t, "测试怪物", monster.MonsterName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("怪物不存在", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM "game_config"\."monsters" WHERE .*id = \$1.*deleted_at IS NULL`).
			WithArgs(monsterID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		monster, err := repo.GetByID(ctx, monsterID)
		assert.Error(t, err)
		assert.Nil(t, monster)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMonsterRepository_GetByCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMonsterRepository(db)
	ctx := context.Background()

	t.Run("成功获取怪物", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "monster_code", "monster_name", "monster_level", "max_hp",
			"created_at", "updated_at",
		}).AddRow(
			"550e8400-e29b-41d4-a716-446655440000", "TEST_MONSTER", "测试怪物", 5, 100,
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT .+ FROM "game_config"\."monsters" WHERE .*monster_code = \$1.*deleted_at IS NULL`).
			WithArgs("TEST_MONSTER").
			WillReturnRows(rows)

		monster, err := repo.GetByCode(ctx, "TEST_MONSTER")
		assert.NoError(t, err)
		assert.NotNil(t, monster)
		assert.Equal(t, "TEST_MONSTER", monster.MonsterCode)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMonsterRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMonsterRepository(db)
	ctx := context.Background()

	t.Run("成功获取怪物列表", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "monster_code", "monster_name", "monster_level", "max_hp",
			"created_at", "updated_at",
		}).
			AddRow("id1", "MONSTER1", "怪物1", 1, 50, time.Now(), time.Now()).
			AddRow("id2", "MONSTER2", "怪物2", 2, 100, time.Now(), time.Now())

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)

		params := interfaces.MonsterQueryParams{
			Limit:  10,
			Offset: 0,
		}

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "game_config"\."monsters" WHERE`).
			WillReturnRows(countRows)

		mock.ExpectQuery(`SELECT .*FROM "game_config"\."monsters" WHERE`).
			WillReturnRows(rows)

		monsters, total, err := repo.List(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(monsters))
		assert.Equal(t, int64(2), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMonsterRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMonsterRepository(db)
	ctx := context.Background()

	t.Run("成功更新怪物", func(t *testing.T) {
		monster := &game_config.Monster{
			ID:           "550e8400-e29b-41d4-a716-446655440000",
			MonsterCode:  "TEST_MONSTER",
			MonsterName:  "更新后的名称",
			MonsterLevel: 5,
			MaxHP:        200,
		}

		mock.ExpectExec(`UPDATE "game_config"\."monsters" SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, monster)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMonsterRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMonsterRepository(db)
	ctx := context.Background()

	monsterID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("成功删除怪物", func(t *testing.T) {
		selectRows := sqlmock.NewRows([]string{"id", "monster_code", "monster_name", "monster_level", "max_hp", "created_at", "updated_at"}).
			AddRow(monsterID, "TEST_MONSTER", "测试怪物", 5, 100, time.Now(), time.Now())
		mock.ExpectQuery(`SELECT .*FROM "game_config"\."monsters" WHERE`).
			WithArgs(monsterID).
			WillReturnRows(selectRows)

		mock.ExpectExec(`UPDATE "game_config"\."monsters" SET "deleted_at"=`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, monsterID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMonsterRepository_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMonsterRepository(db)
	ctx := context.Background()

	t.Run("怪物代码存在", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "game_config"\."monsters" WHERE`).
			WithArgs("TEST_MONSTER").
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, "TEST_MONSTER")
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("怪物代码不存在", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "game_config"\."monsters" WHERE`).
			WithArgs("NONEXISTENT").
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, "NONEXISTENT")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
