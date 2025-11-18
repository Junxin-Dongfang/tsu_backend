package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"tsu-self/internal/modules/game/service"
)

func main() {
	teamID := flag.String("team-id", "", "Team ID whose warehouse will receive loot")
	gold := flag.Int64("gold", 0, "Gold amount to add to the warehouse (required)")
	dungeonID := flag.String("dungeon-id", "manual-test-dungeon", "Optional dungeon ID for audit fields")
	flag.Parse()

	if *teamID == "" {
		log.Fatal("team-id is required")
	}
	if *gold <= 0 {
		log.Fatal("gold must be greater than 0")
	}

	dbURL := os.Getenv("TSU_GAME_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://tsu_game_user:tsu_game_password@tsu_postgres:5432/tsu_db?sslmode=disable&search_path=game_runtime,game_config,auth,public"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	warehouseService := service.NewTeamWarehouseService(db)
	err = warehouseService.AddLootToWarehouse(ctx, &service.AddLootToWarehouseRequest{
		TeamID:          *teamID,
		SourceDungeonID: *dungeonID,
		Gold:            *gold,
	})
	if err != nil {
		log.Fatalf("failed to add loot: %v", err)
	}

	fmt.Printf("Added %d gold to warehouse for team %s (dungeon: %s)\n", *gold, *teamID, *dungeonID)
}
