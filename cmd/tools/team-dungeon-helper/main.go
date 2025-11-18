package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"tsu-self/internal/modules/game/service"
)

var defaultDBURL = "postgres://tsu_game_user:tsu_game_password@tsu_postgres:5432/tsu_db?sslmode=disable&search_path=game_runtime,game_config,auth,public"

func main() {
	action := flag.String("action", "select", "Action to perform: select | enter | complete")
	teamID := flag.String("team-id", "", "Team ID (required)")
	heroID := flag.String("hero-id", "", "Hero ID performing the action (required)")
	dungeonID := flag.String("dungeon-id", "", "Dungeon ID (required)")
	gold := flag.Int64("gold", 0, "Gold to grant when action=complete")
	items := flag.String("items", "", "Comma separated loot items when action=complete. Format: item_id:quantity[:type]")
	flag.Parse()

	if *teamID == "" || *heroID == "" || *dungeonID == "" {
		log.Fatal("team-id, hero-id and dungeon-id are required")
	}

	dbURL := os.Getenv("TSU_GAME_DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultDBURL
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	dungeonService := service.NewTeamDungeonService(db, nil)

	switch strings.ToLower(*action) {
	case "select":
		if _, err := dungeonService.SelectDungeon(ctx, &service.SelectDungeonRequest{
			TeamID:    *teamID,
			HeroID:    *heroID,
			DungeonID: *dungeonID,
		}); err != nil {
			log.Fatalf("select dungeon failed: %v", err)
		}
		fmt.Printf("SelectDungeon succeeded (team=%s dungeon=%s)\n", *teamID, *dungeonID)
	case "enter":
		if _, err := dungeonService.EnterDungeon(ctx, &service.EnterDungeonRequest{
			TeamID:    *teamID,
			HeroID:    *heroID,
			DungeonID: *dungeonID,
		}); err != nil {
			log.Fatalf("enter dungeon failed: %v", err)
		}
		fmt.Printf("EnterDungeon succeeded (team=%s dungeon=%s)\n", *teamID, *dungeonID)
	case "complete":
		loot := service.LootData{Gold: *gold}
		if *items != "" {
			for _, raw := range strings.Split(*items, ",") {
				segments := strings.Split(raw, ":")
				if len(segments) < 2 {
					log.Fatalf("invalid loot item format: %s", raw)
				}
				item := service.LootItem{ItemID: segments[0]}
				if len(segments) >= 3 {
					item.ItemType = segments[2]
				}
				var qty int
				if _, err := fmt.Sscanf(segments[1], "%d", &qty); err != nil || qty <= 0 {
					log.Fatalf("invalid quantity for loot item %s: %v", raw, err)
				}
				item.Quantity = qty
				loot.Items = append(loot.Items, item)
			}
		}
		if _, err := dungeonService.CompleteDungeon(ctx, &service.CompleteDungeonRequest{
			TeamID:    *teamID,
			HeroID:    *heroID,
			DungeonID: *dungeonID,
			Loot:      loot,
		}); err != nil {
			log.Fatalf("complete dungeon failed: %v", err)
		}
		fmt.Printf("CompleteDungeon succeeded (team=%s dungeon=%s gold=%d)\n", *teamID, *dungeonID, *gold)
	default:
		log.Fatalf("unsupported action %s", *action)
	}
}
