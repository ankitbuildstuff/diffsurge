package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tvc-org/tvc/internal/config"
	"github.com/tvc-org/tvc/pkg/logger"
)

type Router struct {
	routes []routeEntry
	log    *logger.Logger
}

type routeEntry struct {
	pathPrefix string
	target     *url.URL
	proxy      *httputil.ReverseProxy
	config     config.RouteConfig
}

func NewRouter(routes []config.RouteConfig, log *logger.Logger) *Router {
	r := &Router{log: log}

	for _, rc := range routes {
		target, err := url.Parse(rc.Target)
		if err != nil {
			log.Error().Err(err).Str("target", rc.Target).Msg("Invalid route target, skipping")
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			log.Error().Err(err).Str("target", target.String()).Str("path", req.URL.Path).Msg("Proxy error")
			http.Error(w, `{"error":"upstream unavailable"}`, http.StatusBadGateway)
		}

		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = target.Host
		}

		r.routes = append(r.routes, routeEntry{
			pathPrefix: rc.PathPrefix,
			target:     target,
			proxy:      proxy,
			config:     rc,
		})

		log.Info().Str("prefix", rc.PathPrefix).Str("target", rc.Target).Msg("Route registered")
	}

	return r
}

func (r *Router) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, route := range r.routes {
			if strings.HasPrefix(req.URL.Path, route.pathPrefix) {
				route.proxy.ServeHTTP(w, req)
				return
			}
		}

		if len(r.routes) == 1 {
			r.routes[0].proxy.ServeHTTP(w, req)
			return
		}

		http.Error(w, `{"error":"no matching route"}`, http.StatusNotFound)
	})
}

func (r *Router) RouteForPath(path string) *config.RouteConfig {
	for _, route := range r.routes {
		if strings.HasPrefix(path, route.pathPrefix) {
			return &route.config
		}
	}
	return nil
}
