package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "6193", "服务端口号")
	dbPath := flag.String("db", "aquarium.db", "SQLite数据库文件路径")
	flag.Parse()

	err := InitDB(*dbPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	SetupRoutes(mux)

	addr := ":" + *port
	fmt.Printf("水族馆管理系统启动中...\n")
	fmt.Printf("监听地址: http://localhost%s\n", addr)
	fmt.Printf("数据库文件: %s\n", *dbPath)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
