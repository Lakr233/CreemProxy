package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Lakr233/CreemProxy/src"
)

func main() {
	fmt.Println("======== CreemProxy Starting ========")
	fmt.Printf("Environment: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Go Version: %s\n", runtime.Version())

	now := time.Now()
	fmt.Printf("Start Time (UTC): %s\n", now.UTC().Format(time.RFC3339))
	fmt.Printf("Start Time (Local): %s\n", now.Format(time.RFC3339))
	fmt.Println("====================================")
	fmt.Println()

	src.PopulateEnv()
	src.PrepareCertifications()
	src.PrepareSigningAsset()
	src.Serve()

	fmt.Println()
	fmt.Println("======== CreemProxy Stopped ========")
	fmt.Printf("Stop Time (UTC): %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("Stop Time (Local): %s\n", time.Now().Format(time.RFC3339))
	fmt.Println("====================================")
}
