package main

import (
	"encoding/json"
	"fmt"

	appcfg "github.com/vocbl/users-svc/internal/infrastructure/cfg"
)

func main() {
	cfg, err := appcfg.Load("../../example.env")
	if err != nil {
		fmt.Println(err)
	}

	m, _ := json.Marshal(cfg)
	fmt.Println(string(m))
}
