package main

import (
	"go.uber.org/fx"

	"casbin-demo/internal/app"
)

func main() {
	fx.New(app.Module()).Run()
}
