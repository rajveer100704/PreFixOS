package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"prefixos/internal/config"
	"prefixos/internal/eviction"
	"prefixos/internal/memory"
	"prefixos/internal/persistence"
	"prefixos/internal/radix"
	"prefixos/internal/server"
	pb "prefixos/proto/v1"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "Path to YAML configuration file")
	flag.Parse()

	log.Println("Bootstrapping PrefixOS Control Plane Engine...")

	// 1. Load Configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// 2. Initialize Memory Manager (Zero-Allocation Slab Allocator)
	bm := memory.NewBlockManager(cfg.Memory.TotalBlocks)

	// 3. Initialize Radix Tree Engine
	tree := radix.NewTree(bm)

	// 4. Initialize Eviction Policy
	ev := eviction.NewLRUEvictionPolicy()

	// 5. Initialize Persistence Engine
	var pe *persistence.Engine
	if cfg.Persistence.Enabled {
		var err error
		pe, err = persistence.NewEngine("data", tree, bm, cfg.Persistence.SyncImmediate)
		if err != nil {
			log.Printf("warning: persistence engine initialization failed: %v", err)
		} else {
			defer pe.Close()
		}
	}

	// 6. Initialize gRPC Server
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen on gRPC address %s: %v", grpcAddr, err)
	}

	grpcServer := grpc.NewServer()
	prefixOSServer := server.NewPrefixOSServer(tree, bm, ev, pe)
	pb.RegisterPrefixOSServiceServer(grpcServer, prefixOSServer)

	go func() {
		log.Printf("PrefixOS gRPC Server listening on %s", grpcAddr)
		if err := grpcServer.Serve(grpcLis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// 7. Initialize REST HTTP & Metrics Server
	httpAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort)
	httpServer := server.NewHTTPServer(httpAddr, tree, bm)

	go func() {
		log.Printf("PrefixOS REST & Metrics Server listening on http://%s", httpAddr)
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down PrefixOS Control Plane Engine...")
	grpcServer.GracefulStop()
	_ = httpServer.Stop()
	log.Println("PrefixOS cleanly stopped.")
}
