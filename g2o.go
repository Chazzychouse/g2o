package main

import (
	"github.com/chazzychouse/g2o/cmd"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cmd.Execute()
}
