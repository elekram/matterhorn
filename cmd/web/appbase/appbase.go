package appBase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/elekram/matterhorn/cmd/web/ui"
	appcfg "github.com/elekram/matterhorn/config"
	"go.mongodb.org/mongo-driver/mongo"
)

func SignIn(cfg *appcfg.ConfigProperties, db *mongo.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// _appName := env.Config.AppName

		if cfg.DevMode {
			component := ui.SignIn()
			component.Render(context.Background(), w)
			return
		}

		component := ui.SignIn()
		component.Render(context.Background(), w)
	})
}

func Home(cfg *appcfg.ConfigProperties, db *mongo.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appName := cfg.AppName

		if cfg.DevMode {
			component := ui.Layout(ui.DevHead(appName))
			component.Render(context.Background(), w)
			return
		}

		component := ui.Layout(ui.Head(appName))
		component.Render(context.Background(), w)
	})
}

// func Root(w http.ResponseWriter, r *http.Request) {
// 	appName := cfg.AppName

// 	if cfg.DevMode {
// 		component := ui.Layout(ui.DevHead(appName))
// 		component.Render(context.Background(), w)
// 		return
// 	}

// 	component := ui.Layout(ui.Head(appName))
// 	component.Render(context.Background(), w)

// }

func Auth(w http.ResponseWriter, r *http.Request) {
	pBody := bodyMapper(r.Body)

	for key, value := range pBody {
		fmt.Println(key, value)
	}

	w.Write([]byte("Auth post"))
}

func bodyMapper(b io.ReadCloser) map[string]string {
	body, err := io.ReadAll(b)
	if err != nil {
		fmt.Println(err)
	}

	b.Close()

	bodyMap := map[string]string{}
	bodyElements := strings.Split(string(body), "&")

	for _, bodyElement := range bodyElements {
		elementParts := strings.Split(string(bodyElement), "=")

		name := elementParts[0]
		value := elementParts[1]

		unescapedValue, err := url.QueryUnescape(value)
		if err != nil {
			fmt.Println(err)
		}

		bodyMap[name] = unescapedValue
	}

	return bodyMap
}
