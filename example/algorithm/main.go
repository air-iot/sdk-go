package main

import (
	"context"
	"github.com/air-iot/sdk-go/v4/algorithm"
	"github.com/air-iot/sdk-go/v4/example/algorithm/app"
)

func main() {
	s := new(app.TestAlgorithm)
	s.Ctx, s.Cancel = context.WithCancel(context.Background())
	algorithm.NewApp().Start(s)
}
