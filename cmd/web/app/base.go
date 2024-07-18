package app

import (
	"context"
	"net/http"

	"github.com/elekram/matterhorn/cmd/web/ui"
	env "github.com/elekram/matterhorn/config"
)

func SignIn(w http.ResponseWriter, r *http.Request) {
	// _appName := env.Config.AppName

	if env.Config.DevMode {
		component := ui.SignIn()
		component.Render(context.Background(), w)
		return
	}

	component := ui.SignIn()
	component.Render(context.Background(), w)
}

func Root(w http.ResponseWriter, r *http.Request) {
	appName := env.Config.AppName

	if env.Config.DevMode {
		component := ui.Layout(ui.DevHead(appName))
		component.Render(context.Background(), w)
		return
	}

	component := ui.Layout(ui.Head(appName))
	component.Render(context.Background(), w)

}
