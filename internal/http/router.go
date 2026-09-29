package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/sakid00/enmasse-be/internal/http/handlers"
	mw "github.com/sakid00/enmasse-be/internal/http/middleware"
	"github.com/sakid00/enmasse-be/internal/service"
)

type RouterOptions struct {
	CORSOrigins      string
	JWTAccessSecret  string
	JWTIssuer        string
	JWTServiceSecret string
}

func NewRouter(authSvc *service.AuthService, opts RouterOptions) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(mw.CORS(opts.CORSOrigins))

	authH := handlers.NewAuthHandler(authSvc)
	meH := handlers.NewMeHandler(authSvc)
	intH := handlers.NewInternalHandler(authSvc)

	r.Get("/health", handlers.Health)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.With(httprate.LimitByIP(10, 10*time.Minute)).Post("/register/artist", authH.RegisterArtist)
			r.With(httprate.LimitByIP(10, 10*time.Minute)).Post("/register/vendor", authH.RegisterVendor)
			r.With(httprate.LimitByIP(10, 10*time.Minute)).Post("/email/status", authH.EmailStatus)
			r.With(httprate.LimitByIP(5, 10*time.Minute)).Post("/login", authH.Login)
			r.With(httprate.LimitByIP(5, 10*time.Minute)).Post("/claim", authH.Claim)
			r.With(httprate.LimitByIP(5, 10*time.Minute)).Post("/password/claim", authH.ClaimPassword)
			r.With(httprate.LimitByIP(5, 10*time.Minute)).Post("/password/set", authH.SetPassword)
			r.With(mw.HumanAuth(opts.JWTAccessSecret, opts.JWTIssuer)).Post("/password/change", authH.ChangePassword)
		})

		r.Route("/me", func(r chi.Router) {
			r.Use(mw.HumanAuth(opts.JWTAccessSecret, opts.JWTIssuer))
			r.Get("/profile", meH.Get)
			r.Patch("/profile", meH.Patch)
		})

		r.Route("/internal", func(r chi.Router) {
			r.Use(mw.ServiceAuth(opts.JWTServiceSecret))
			r.Get("/profile-status/{id}", intH.ProfileStatus)
			r.Get("/artists/{id}", intH.Artist)
			r.Get("/vendors/{id}", intH.Vendor)
		})
	})

	return r
}
