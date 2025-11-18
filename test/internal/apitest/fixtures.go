package apitest

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// FixtureFactory generates unique payloads to avoid data collisions across environments.
type FixtureFactory struct {
	RunID                string
	PlayerUsernamePrefix string
	PlayerEmailSuffix    string
}

// NewFixtureFactory builds a factory bound to the provided config.
func NewFixtureFactory(cfg Config) FixtureFactory {
	return FixtureFactory{
		RunID:                UniqueSuffix(),
		PlayerUsernamePrefix: cfg.PlayerUsernamePrefix,
		PlayerEmailSuffix:    cfg.PlayerEmailSuffix,
	}
}

// HeroName returns a unique hero name using the run id.
func (f FixtureFactory) HeroName(tag string) string {
	base := sanitizeUsername(tag)
	if len(base) > 10 {
		base = base[:10]
	}
	suffix := f.RunID
	if len(suffix) > 6 {
		suffix = suffix[len(suffix)-6:]
	}
	name := fmt.Sprintf("%s_%s", base, suffix)
	if len(name) > 20 {
		name = name[:20]
	}
	if len(name) < 2 {
		name = "h_" + suffix
	}
	return name
}

// TeamName returns a unique team name for the run.
func (f FixtureFactory) TeamName(tag string) string {
	base := sanitizeUsername(tag)
	if len(base) > 10 {
		base = base[:10]
	}
	suffix := f.RunID
	if len(suffix) > 6 {
		suffix = suffix[len(suffix)-6:]
	}
	name := fmt.Sprintf("%s_%s", base, suffix)
	if len(name) > 20 {
		name = name[:20]
	}
	if len(name) < 2 {
		name = "tm_" + suffix
	}
	return name
}

// UniquePlayerCredentials yields username/email/password triples for registration flows.
func (f FixtureFactory) UniquePlayerCredentials(tag string) (username, email, password string) {
	base := f.PlayerUsernamePrefix
	if tag != "" {
		base = base + "_" + tag
	}
	base = sanitizeUsername(base)
	suffix := UniqueSuffix()
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	username = fmt.Sprintf("%s_%s", base, suffix)
	if len(username) > 30 {
		username = username[:30]
	}
	email = fmt.Sprintf("%s@%s", username, f.PlayerEmailSuffix)
	password = "Admin123!"
	return
}

var usernameAllowed = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitizeUsername(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = usernameAllowed.ReplaceAllString(s, "")
	if s == "" {
		s = "user"
	}
	if len(s) > 20 {
		s = s[:20]
	}
	return s
}

// BuildCreateHeroRequest produces a valid hero creation payload.
func (f FixtureFactory) BuildCreateHeroRequest(classID string, description string) CreateHeroRequest {
	if description == "" {
		description = "自动化测试创建的英雄"
	}
	return CreateHeroRequest{
		ClassID:     classID,
		HeroName:    f.HeroName("hero"),
		Description: description,
	}
}

// BuildCreateTeamRequest generates a payload using the provided leader hero id.
func (f FixtureFactory) BuildCreateTeamRequest(leaderHeroID string) CreateTeamRequest {
	return CreateTeamRequest{
		HeroID:      leaderHeroID,
		TeamName:    f.TeamName("alpha"),
		Description: "自动化构建的测试团队",
	}
}

// BuildUpdateTeamInfoRequest returns a simple rename payload.
func (f FixtureFactory) BuildUpdateTeamInfoRequest() UpdateTeamInfoRequest {
	return UpdateTeamInfoRequest{
		Name:        fmt.Sprintf("%s-renamed", f.TeamName("alpha")),
		Description: "自动化更新描述",
	}
}

// BuildDistributeGoldRequest splits even gold among receivers.
func (f FixtureFactory) BuildDistributeGoldRequest(distributorID string, receivers []string, total int64) DistributeGoldRequest {
	alloc := make(map[string]int64, len(receivers))
	if total < int64(len(receivers)) {
		total = int64(len(receivers))
	}
	per := total / int64(len(receivers))
	for _, r := range receivers {
		alloc[r] = per
	}
	return DistributeGoldRequest{DistributorID: distributorID, Distributions: alloc}
}

// BuildSelectDungeonRequest prepares a simple dungeon selection payload.
func (f FixtureFactory) BuildSelectDungeonRequest(heroID, dungeonID string) SelectDungeonRequest {
	return SelectDungeonRequest{HeroID: heroID, DungeonID: dungeonID}
}

// BuildEnterDungeonRequest reuses hero ID; dungeon optional for random selection.
func (f FixtureFactory) BuildEnterDungeonRequest(heroID, dungeonID string) EnterDungeonRequest {
	return EnterDungeonRequest{HeroID: heroID, DungeonID: dungeonID}
}

// BuildCompleteDungeonRequest marks completion with optional loot payload stub.
func (f FixtureFactory) BuildCompleteDungeonRequest(heroID, dungeonID string) CompleteDungeonRequest {
	return CompleteDungeonRequest{
		HeroID:    heroID,
		DungeonID: dungeonID,
	}
}

// SampleUploadPath returns the canonical path to the placeholder upload asset.
func SampleUploadPath() string {
	return filepath.Join("test", "fixtures", "sample-upload.txt")
}
