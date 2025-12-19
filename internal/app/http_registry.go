package app

import (
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	httptransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http"
	healthtransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/health"
)

func httpRouters(cfg config.Config) []httptransport.Router {
	return []httptransport.Router{
		healthtransport.NewRouter(cfg),
	}
}
