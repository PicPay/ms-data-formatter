package http

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/Depado/ginprom"
	"github.com/PicPay/ms-data-formatter/pkg/health"
	"github.com/PicPay/ms-data-formatter/pkg/log"
	"github.com/PicPay/ms-data-formatter/pkg/validator"
	"github.com/gin-contrib/pprof"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	binding "github.com/gin-gonic/gin/binding"
)

type Endpoint interface {
	Router(*gin.RouterGroup)
}

type Server struct {
	*http.Server
	Router *gin.Engine
}

func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"message": "route not found"})
}

func BadRequest(c *gin.Context, err error) {
	log.Error("Bad request", err, nil)
	c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
}

func InternalServerError(c *gin.Context, err error) {
	log.Error("Internal server error", err, nil)
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
}

func Ok(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, obj)
}

func NewServer(addr, promName, promPath, healthPath string) *Server {
	gin.SetMode(gin.ReleaseMode)
	binding.Validator = validator.New("binding")

	router := gin.New()
	router.Use(health.GinHandler(healthPath))
	router.Use(ginzap.Ginzap(log.ZapLogger, time.RFC3339, true))
	router.Use(cors.Default())

	p := ginprom.New(
		ginprom.Engine(router),
		ginprom.Subsystem(promName),
		ginprom.Path(promPath),
	)
	router.Use(p.Instrument())

	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &Server{
		Router: router,
		Server: httpServer,
	}
}

func (s *Server) Debug() {
	pprof.Register(s.Router)
}

func (s *Server) Group(path string) *gin.RouterGroup {
	return s.Router.Group(path)
}

func (s *Server) Start() error {
	s.Router.NoRoute(NotFound)
	s.Server.Handler = s.Router
	return s.Server.ListenAndServe()
}

func (s *Server) Stop() {
	s.Close()
}
