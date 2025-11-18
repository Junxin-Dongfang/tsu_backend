package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"tsu-self/internal/modules/auth/client"
)

func main() {
	readURL := os.Getenv("KETO_READ_URL")
	writeURL := os.Getenv("KETO_WRITE_URL")

	if readURL == "" {
		readURL = "localhost:4466"
	}
	if writeURL == "" {
		writeURL = "localhost:4467"
	}

	fmt.Printf("尝试连接 Keto...\n")
	fmt.Printf("Read URL: %s\n", readURL)
	fmt.Printf("Write URL: %s\n", writeURL)

	ketoClient, err := client.NewKetoClient(readURL, writeURL)
	if err != nil {
		fmt.Printf("❌ Failed to create Keto client: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Keto client created successfully")

	// 测试初始化团队权限
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("\n测试初始化团队权限...")
	err = ketoClient.InitializeTeamPermissions(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to initialize team permissions: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Team permissions initialized successfully")

	// 测试添加团队成员
	fmt.Println("\n测试添加团队成员...")
	testTeamID := "test-team-001"
	testHeroID := "test-hero-001"
	err = ketoClient.AddTeamMember(ctx, testTeamID, testHeroID, "leader")
	if err != nil {
		fmt.Printf("❌ Failed to add team member: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Team member added successfully")

	// 测试检查权限
	fmt.Println("\n测试检查权限...")
	allowed, err := ketoClient.CheckTeamPermission(ctx, testTeamID, "disband_team", testHeroID)
	if err != nil {
		fmt.Printf("❌ Failed to check permission: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Permission check result: %v\n", allowed)

	// 清理测试数据
	fmt.Println("\n清理测试数据...")
	err = ketoClient.RemoveTeamMember(ctx, testTeamID, testHeroID, "leader")
	if err != nil {
		fmt.Printf("⚠️  Failed to remove test member: %v\n", err)
	} else {
		fmt.Println("✅ Test data cleaned up")
	}

	fmt.Println("\n🎉 所有测试通过!")
}
