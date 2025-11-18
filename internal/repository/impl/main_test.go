package impl

import (
    "fmt"
    "os"
    "testing"
)

func TestMain(m *testing.M) {
    if os.Getenv("RUN_REPOSITORY_TESTS") == "1" {
        os.Exit(m.Run())
    }
    fmt.Println("[SKIP] internal/repository/impl tests require RUN_REPOSITORY_TESTS=1 and a reachable Postgres instance")
    os.Exit(0)
}
