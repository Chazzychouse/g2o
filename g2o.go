package main

import (
	"github.com/chazzy/g2o/cmd"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cmd.Execute()
}
