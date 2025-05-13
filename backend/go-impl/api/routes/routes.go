package routes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sourcegraph/conc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	banner = `
    __    _             ________          __ 
   / /   (_)___  ____  / ____/ /_  ____ _/ /_
  / /   / / __ \/ __ \/ /   / __ \/ __ \/ __/
 / /___/ / / / / /_/ / /___/ / / / /_/ / /_  
/_____/_/_/ /_/\__, /\____/_/ /_/\__,_/\__/  
              /____/                         

Server Started at %s
`
)

func NewEngine() (*gin.Engine, error) {
	r := gin.New()
	r.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	return r, nil
}

type HttpEngine struct {
	Engine *gin.Engine
	Addr   string

	wsHandler http.Handler
	// ChatRouteV1 *v1.ChatRoute
	routes []Route
}

func NewHTTPEngine(addr string, wsHandler http.Handler, routes ...Route) *HttpEngine {

	engine, err := NewEngine()
	if err != nil {
		log.Fatal(err)
	}

	return &HttpEngine{
		Engine:    engine,
		Addr:      addr,
		wsHandler: wsHandler,
		// ChatRouteV1: chatRoute,
		routes: routes,
	}
}

type Route interface {
	RegisterRoute(r *gin.RouterGroup)
}

func (h *HttpEngine) RegisterRoutes() {
	eg := h.Engine.Group("/api")

	// Register WebSocket route
	h.Engine.GET("/ws", gin.WrapH(h.wsHandler))

	// Serve openapi.yaml
	h.Engine.StaticFile("/openapi.yaml", "openapi.yaml")
	// Swagger UI, pointing to the /openapi.yaml endpoint
	h.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/openapi.yaml"),
		ginSwagger.DefaultModelsExpandDepth(-1), // Hide models by default
	))

	// v := reflect.ValueOf(*h)
	// for i := 0; i < v.NumField(); i++ {
	// 	if router, ok := v.Field(i).Interface().(Route); ok {
	// 		router.RegisterRoute(eg)
	// 	}
	// }
	for _, route := range h.routes {
		route.RegisterRoute(eg)
	}
}

func (h *HttpEngine) Run() (*http.Server, error) {
	h.RegisterRoutes()
	srv := &http.Server{
		Addr:    h.Addr,
		Handler: h.Engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	fmt.Printf(banner, h.Addr)

	// 监控结束指令
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 停止服务
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	var wg conc.WaitGroup
	wg.Go(func() {
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown", "err", err)
		}
	})
	if r := wg.WaitAndRecover(); r != nil {
		log.Fatal("Server Shutdown", "wait err", r.String())
	}

	log.Println("server exiting")

	return srv, nil
}
