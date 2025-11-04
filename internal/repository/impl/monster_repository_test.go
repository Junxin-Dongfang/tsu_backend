package impl

import (
	"context"
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
		mock.ExpectExec("INSERT INTO game_config.monsters").
			WithArgs(
				monster.MonsterCode,
				monster.MonsterName,
				monster.MonsterLevel,
				sqlmock.AnyArg(), // description
				monster.MaxHP,
				sqlmock.AnyArg(), // hp_recovery
				sqlmock.AnyArg(), // max_mp
				sqlmock.AnyArg(), // mp_recovery
				sqlmock.AnyArg(), // base_str
				sqlmock.AnyArg(), // base_agi
				sqlmock.AnyArg(), // base_vit
				sqlmock.AnyArg(), // base_wlp
				sqlmock.AnyArg(), // base_int
				sqlmock.AnyArg(), // base_wis
				sqlmock.AnyArg(), // base_cha
				sqlmock.AnyArg(), // accuracy_formula
				sqlmock.AnyArg(), // dodge_formula
				sqlmock.AnyArg(), // initiative_formula
				sqlmock.AnyArg(), // body_resist_formula
				sqlmock.AnyArg(), // magic_resist_formula
				sqlmock.AnyArg(), // mental_resist_formula
				sqlmock.AnyArg(), // environment_resist_formula
				sqlmock.AnyArg(), // damage_resistances
				sqlmock.AnyArg(), // passive_buffs
				sqlmock.AnyArg(), // drop_gold_min
				sqlmock.AnyArg(), // drop_gold_max
				sqlmock.AnyArg(), // drop_exp
				sqlmock.AnyArg(), // icon_url
				sqlmock.AnyArg(), // model_url
				sqlmock.AnyArg(), // is_active
				sqlmock.AnyArg(), // display_order
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

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

		mock.ExpectQuery("SELECT .+ FROM game_config.monsters WHERE id = \\$1 AND deleted_at IS NULL").
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
		mock.ExpectQuery("SELECT .+ FROM game_config.monsters WHERE id = \\$1 AND deleted_at IS NULL").
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

		mock.ExpectQuery("SELECT .+ FROM game_config.monsters WHERE monster_code = \\$1 AND deleted_at IS NULL").
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

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM game_config.monsters WHERE deleted_at IS NULL").
			WillReturnRows(countRows)

		mock.ExpectQuery("SELECT .+ FROM game_config.monsters WHERE deleted_at IS NULL ORDER BY monster_level ASC LIMIT \\$1 OFFSET \\$2").
			WithArgs(params.Limit, params.Offset).
			WillReturnRows(rows)

		monsters, total, err := repo.List(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(monsters))
		assert.Equal(t, 2, total)
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

		mock.ExpectExec("UPDATE game_config.monsters SET").
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
		mock.ExpectExec("UPDATE game_config.monsters SET deleted_at = NOW\\(\\) WHERE id = \\$1").
			WithArgs(monsterID).
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
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM game_config.monsters WHERE monster_code = \\$1 AND deleted_at IS NULL\\)").
			WithArgs("TEST_MONSTER").
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, "TEST_MONSTER")
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("怪物代码不存在", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM game_config.monsters WHERE monster_code = \\$1 AND deleted_at IS NULL\\)").
			WithArgs("NONEXISTENT").
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, "NONEXISTENT")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
