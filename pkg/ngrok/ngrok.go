package ngrok

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.ngrok.com/ngrok"
	ngrokConfig "golang.ngrok.com/ngrok/config"
)

func Run() ngrok.Tunnel {
	listener, err := ngrok.Listen(context.Background(),
		ngrokConfig.HTTPEndpoint(
			ngrokConfig.WithDomain("oryx-credible-buzzard.ngrok-free.app"),
		),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		logrus.Error("Ngrok error: ", err.Error())
	}
	return listener
}
