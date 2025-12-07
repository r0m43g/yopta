// backend/cmd/server/main.go
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yopta-template/internal/cache"
	"yopta-template/internal/handlers"
	"yopta-template/internal/models"
	"yopta-template/internal/utils"
	"yopta-template/internal/xss"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	authMiddleware "yopta-template/internal/middleware"
)

func main() {
	// Load environment variables from .env file
	// This ensures our application has access to configuration values without hardcoding
	if err := godotenv.Load(); err != nil {
		log.Fatalf(
			"Klaida įkeliant aplinkos kintamuosius: %v",
			err,
		) // Error loading environment variables
	}

	// Extract essential configuration from environment variables
	dsn := os.Getenv("DSN")                // Database Source Name for MySQL connection
	jwtSecret := os.Getenv("JWT_SECRET")   // Secret key for JWT token generation and validation
	jwtExpiry := os.Getenv("JWT_EXPIRY")   // Expiration time for JWT tokens
	csrfSecret := os.Getenv("CSRF_SECRET") // Secret for CSRF protection
	logType := os.Getenv("LOG_TYPE")       // Logging configuration: "file", "console" or "" (both)

	// Validate that all required environment variables are set
	if dsn == "" || jwtSecret == "" || jwtExpiry == "" || csrfSecret == "" {
		log.Fatal(
			"Ne visi būtini aplinkos kintamieji nustatyti (DSN, JWT_SECRET, JWT_EXPIRY ir CSRF_SECRET)", // Not all required environment variables are set
		)
	}

	// Initialize database connection using the DSN from environment
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Klaida jungiantis prie duomenų bazės: %v", err) // Error connecting to database
	}
	defer db.Close() // Ensure DB connection is closed when application exits

	// Verify the database connection is working with a ping
	if err := db.Ping(); err != nil {
		log.Fatalf(
			"Klaida tikrinant ryšį su duomenų baze: %v",
			err,
		) // Error checking database connection
	}

	// Initialize in-memory cache for performance optimization
	appCache := cache.NewCache()

	// Configure database connection for logging subsystem
	models.SetDBConnection(db)

	// Initialize rate limiter with configuration from environment variables
	// This protects against DoS attacks by limiting request frequency
	rateLimiter := utils.InitRateLimiter()

	// Create new Chi router instance
	// Chi is a lightweight, idiomatic and composable router for Go HTTP services
	r := chi.NewRouter()

	// Apply blacklist middleware to all requests
	r.Use(authMiddleware.BlacklistMiddleware(db))
	// Add security headers middleware to protect against various attacks
	r.Use(authMiddleware.SecurityHeaders())

	// Add XSS protection middleware to sanitize input and output
	r.Use(xss.XSSProtectionMiddleware())

	// Configure CORS (Cross-Origin Resource Sharing) for API security
	// This controls which domains can access our API
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Refresh-Token",
			"X-Csrf-Token",
			"X-Request-ID",
		},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Link", "X-Csrf-Token", "X-Request-ID"},
	}))

	// Add request ID middleware for tracing requests through the system
	// This helps with debugging and monitoring
	r.Use(authMiddleware.ContextWithRequestID)

	// Add standard middleware for logging, panic recovery
	r.Use(middleware.Logger)    // Logs HTTP requests
	r.Use(middleware.Recoverer) // Recovers from panics and returns 500 error

	// Add rate limiting middleware to prevent abuse
	r.Use(authMiddleware.RateLimitMiddleware(rateLimiter))

	// Serve static files from the ./static directory (frontend assets)
	fs := http.FileServer(http.Dir("./static"))
	imgfs := http.FileServer(http.Dir("./uploads"))
	r.Handle("/assets/*", fs)
	r.Handle("/avatars/*", imgfs)

	// Serve the SPA index.html for all routes not matched by other handlers
	// This allows the frontend router to handle client-side routing
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// CSRF middleware setup (commented out for now)
	/************************************************************
	csrfMiddleware := csrf.Protect(
		[]byte(csrfSecret),
		csrf.Secure(false),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.FieldName("X-Csrf-Token"),
		csrf.CookieName("X-Csrf-Token"),
		csrf.HttpOnly(false),
	)
	r.Use(csrfMiddleware)
	*************************************************************/

	// CSRF token endpoint setup (commented out for now)
	/************************************************************
	r.Get("/api/sanctum/csrf-cookie", func(w http.ResponseWriter, r *http.Request) {
		token := csrf.Token(r)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-CSRF-Token", token)
		// Можно и в body отправить
	})
	*************************************************************/

	// Public routes group - these don't require authentication
	r.Group(func(r chi.Router) {
		// Apply standard logging middleware to public routes
		r.Use(authMiddleware.LoggingMiddleware(logType))

		r.Route("/api/v1", func(r chi.Router) {
			r.Get("/system-settings", handlers.GetSystemSettings(db))
			r.Post("/register", handlers.Register(db, jwtSecret, jwtExpiry))
			r.Post("/login", handlers.Login(db, jwtSecret, jwtExpiry))
			r.Post("/refresh-token", handlers.RefreshToken(db, jwtSecret, jwtExpiry))
			r.Post("/verify-email", handlers.VerifyEmail(db))
			r.Get("/test", handlers.Test())
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RegistrationEnabledMiddleware(db))

		// Маршрут регистрации
		r.Post("/api/v1/register", handlers.Register(db, jwtSecret, jwtExpiry))
	})

	// Protected routes group - requires JWT authentication
	r.Group(func(r chi.Router) {
		// Apply JWT middleware to validate tokens
		r.Use(authMiddleware.JWTMiddleware(jwtSecret))
		// Apply enhanced logging middleware that includes user info from JWT
		r.Use(authMiddleware.LoggingMiddleware(logType))

		// Routes available to any authenticated user

		r.Post("/api/v1/auth-ping", handlers.AuthPing())
		r.Get("/api/v1/profile", handlers.GetProfile(db))
		r.Put("/api/v1/profile-theme", handlers.UpdateProfileTheme(db))
		r.Put("/api/v1/profile-avatar", handlers.UpdateProfileAvatar(db))
		r.Post("/api/v1/change-password", handlers.ChangePassword(db))
		r.Delete("/api/v1/profile-avatar", handlers.DeleteProfileAvatar(db))

		r.Post("/api/v1/client-logs", handlers.SaveClientLogs(db))
		r.Get("/api/v1/stations", handlers.GetAllStations(db))
		r.Get("/api/v1/stations/{id}", handlers.GetStationByID(db))

		// Admin-only routes group with additional role-based middleware
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RoleMiddleware("admin"))
			// r.Get("/api/v1/system-settings", handlers.GetSystemSettings(db))
			r.Put("/api/v1/system-settings", handlers.UpdateSystemSetting(db))
			r.Get("/api/v1/blacklisted-ips", handlers.GetBlacklistedIPs(db))
			r.Post("/api/v1/blacklisted-ips", handlers.AddBlacklistedIP(db))
			r.Delete("/api/v1/blacklisted-ips", handlers.RemoveBlacklistedIP(db))
			r.Get("/api/v1/users", handlers.Users(db, appCache))
			r.Post("/api/v1/add-user", handlers.AddUser(db))
			r.Put("/api/v1/update-user-role", handlers.UpdateUserRole(db))
			r.Put("/api/v1/update-user-status", handlers.UpdateUserStatus(db))
			r.Put("/api/v1/update-user-verification", handlers.UpdateUserVerification(db))
			r.Delete("/api/v1/delete-user/{id}", handlers.DeleteUser(db))

			// Client logs management endpoints
			r.Get("/api/v1/client-logs", handlers.GetClientLogs(db))
			r.Get("/api/v1/client-logs/{id}", handlers.GetClientLogByID(db))
			r.Get("/api/v1/client-logs/statistics", handlers.GetClientLogStatistics(db))
			r.Delete("/api/v1/client-logs", handlers.DeleteClientLogs(db))

			// System logs management endpoints
			r.Get("/api/v1/logs", handlers.GetLogs(db))
			r.Get("/api/v1/logs/statistics", handlers.GetLogStatistics(db, appCache))
			r.Delete("/api/v1/logs", handlers.ClearOldLogs(db))

			// Cache management
			r.Post("/api/v1/cache/clear", handlers.ClearCache(appCache))

			r.Post("/api/v1/stations", handlers.CreateStation(db))
			r.Put("/api/v1/stations/{id}", handlers.UpdateStation(db))
			r.Delete("/api/v1/stations/{id}", handlers.DeleteStation(db))
			r.Post("/api/v1/stations/{stationId}/tracks", handlers.AddTrack(db))
			r.Put("/api/v1/tracks/{trackId}", handlers.UpdateTrack(db))
			r.Delete("/api/v1/tracks/{trackId}", handlers.DeleteTrack(db))
		})
	})

	// Get server port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		log.Println(
			"PORT aplinkos kintamasis nenustatytas, naudojama numatytoji reikšmė: 8080", // PORT environment variable not set, using default value: 8080
		)
		port = "8080"
	}

	// TLS configuration (commented out for now)
	/************************************************************
		CertFile := os.Getenv("CERT_FILE")
		KeyFile := os.Getenv("KEY_FILE")
		serverTLSCert, err := tls.LoadX509KeyPair(CertFile, KeyFile)
		if err != nil {
			log.Fatalf("Klaida įkeliant sertifikatą: %v", err)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{serverTLSCert},
		}
	***************************************************************/

	// Create HTTP server with graceful shutdown capability
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
		//		TLSConfig: tlsConfig,
	}

	// Start server in a goroutine to allow for graceful shutdown
	go func() {
		log.Printf("Serveris paleistas adresu %s", srv.Addr) // Server started at address
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Serverio klaida: %v", err) // Server error
		}
	}()

	// Create channel to listen for interrupt signal (Ctrl+C)
	// This enables graceful shutdown when the server is interrupted
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Serveris išjungiamas...") // Server shutting down

	// Create timeout context for shutdown
	// This gives ongoing requests a chance to complete before shutting down
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Klaida baigiant serverio darbą: %v", err) // Error when finishing server work
	}

	log.Println("Serveris sėkmingai išjungtas") // Server successfully shut down
}
