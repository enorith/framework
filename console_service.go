package framework

import "github.com/urfave/cli"

type ConsoleService struct {
	consoleApp *cli.App
}

// Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (cs *ConsoleService) Register(app *App) error {

	consoleApp := &cli.App{
		Name:    "Enorith console app",
		Version: AppVersion,
	}

	cs.consoleApp = consoleApp

	return nil
}

func (cs *ConsoleService) SetApp(setFn func(app *cli.App)) {
	setFn(cs.consoleApp)
}
