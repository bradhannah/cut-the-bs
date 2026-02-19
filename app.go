package main

import (
	"context"
	"fmt"
)

// App is the main application struct that holds lifecycle state
// and serves as the Wails binding target.
type App struct {
	ctx context.Context
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails app starts.
// The context is saved for calling runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name.
// This is a placeholder binding method for initial setup verification.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
