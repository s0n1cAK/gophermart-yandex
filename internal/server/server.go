package server

import (
	"net/http"
	"os"
	"time"
	"yandex-diplom/internal/auth"
	config "yandex-diplom/internal/config/gophermart"
	"yandex-diplom/internal/gophermart"
	"yandex-diplom/internal/transport/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type server struct {
	logger *zap.Logger
	*http.Server
}

func New(cfg config.Config, svc gophermart.Service) (*server, error) {
	r := chi.NewRouter()

	logger := svc.GetLogger()

	//Middlewares
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(time.Second * 60))
	r.Use(Logging(logger))

	r.Mount("/api/user", userRoutes(svc))

	srv := &http.Server{
		Addr:    cfg.Address.Host,
		Handler: r,
	}

	return &server{
		logger: logger,
		Server: srv,
	}, nil
}

func userRoutes(svc gophermart.Service) chi.Router {
	r := chi.NewRouter()

	r.Post("/register", httpx.RegisterUser(svc))
	r.Post("/login", httpx.LoginUser(svc))

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(svc))
		r.Post("/orders", httpx.CreateOrder(svc))
		r.Get("/balance", httpx.GetBalance(svc))
		r.Post("/balance/withdraw", httpx.CreateWithdraw(svc))
		r.Group(func(r chi.Router) {
			r.Use(gzipCompession())
			r.Get("/withdrawals", httpx.GetWithdraws(svc))
			r.Get("/orders", httpx.GetOrders(svc))
		})

	})

	return r
}

func (s *server) logStartupInfo() {
	s.logger.Info("Starting server",
		zap.String("Address", s.Server.Addr),
	)
}

func (s *server) Start() *http.Server {
	s.logStartupInfo()

	_, exists := os.LookupEnv("SECRET")
	if !exists {
		s.logger.Info("Using default secret")
	}

	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("Error occurred while running server", zap.Error(err))
	}

	return s.Server
}
