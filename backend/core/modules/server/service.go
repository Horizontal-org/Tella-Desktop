package server

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gomod.cblgh.org/cerca/limiter"

	"Tella-Desktop/backend/core/modules/filestore"
	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/transfer"
	"Tella-Desktop/backend/utils/network"
	"Tella-Desktop/backend/utils/nonces"
	"Tella-Desktop/backend/utils/tls"
)

type service struct {
	nonceManager        *nonces.NonceManager
	limiter             *RateLimitingWare
	server              *http.Server
	listener            net.Listener
	running             bool
	port                int
	pin                 string
	ctx                 context.Context
	registrationService registration.Service
	registrationHandler *registration.Handler
	transferService     transfer.Service
	fileService         filestore.Service
	defaultFolderID     int64
	mu                  sync.RWMutex
}

func NewService(
	ctx context.Context,
	registrationService registration.Service,
	registrationHandler *registration.Handler,
	transferService transfer.Service,
	fileService filestore.Service,
	defaultFolderID int64,
	nonceManager *nonces.NonceManager,
) Service {

	rateLimitingInstance := NewRateLimitingWare()
	srv := &service{
		nonceManager:        nonceManager,
		limiter:             rateLimitingInstance,
		ctx:                 ctx,
		running:             false,
		registrationService: registrationService,
		registrationHandler: registrationHandler,
		transferService:     transferService,
		fileService:         fileService,
		defaultFolderID:     defaultFolderID,
	}

	return srv
}

type RateLimitingWare struct {
	limiter *limiter.TimedRateLimiter
}

func NewRateLimitingWare() *RateLimitingWare {
	ware := RateLimitingWare{}
	// refresh one access every 30 seconds. forget about the requester after 24h of non-activity
	ware.limiter = limiter.NewTimedRateLimiter([]string{}, 30*time.Second, 24*time.Hour)
	// allow initial burst rate allowance to 1000 allow requests at once
	// NOTE cblgh(2026-03-13): different approach: start with small burst allowance and when progressing to upload, increase burst allowance? alt: use different rate limiter for upload
	ware.limiter.SetBurstAllowance(1000)
	ware.limiter.SetLimitAllRoutes(true)
	return &ware
}

func (ware *RateLimitingWare) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		portIndex := strings.LastIndex(req.RemoteAddr, ":")
		ip := req.RemoteAddr[:portIndex]
		// specific fix in case of using a reverse proxy setup
		if address, exists := req.Header["X-Real-Ip"]; ip == "127.0.0.1" && exists {
			ip = address[0]
		}
		isLimited := ware.limiter.IsLimited(ip, req.URL.String())
		if isLimited {
			http.Error(res, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(res, req)
	})
}

// TODO cblgh(2026-03-13): revamp backend to be stateful like frontend
// <zero state> -> [ping] -> [register] -> [prepare-upload] -> [upload] -> [close-connection] -> <end>
func (s *service) Start(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	// Generate new PIN for each start
	s.pin = generateRandomPIN()
	s.registrationService.SetPINCode(s.pin)

	ipStrings, err := network.GetLocalIPs()
	if err != nil {
		return fmt.Errorf("failed to get local IPs: %v", err)
	}

	// Parse strings ip into net.IP
	var ips []net.IP
	for _, ipStr := range ipStrings {
		if ip := net.ParseIP(ipStr); ip != nil {
			ips = append(ips, ip)
		}
	}

	tlsConfig, err := tls.GenerateTLSConfig(s.ctx, tls.Config{
		CommonName:   "Tella Desktop",
		Organization: []string{"Tella"},
		IPAddresses:  ips,
	})
	if err != nil {
		return fmt.Errorf("failed to generate TLS config: %v", err)
	}

	mux := http.NewServeMux()

	// TODO cblgh(2026-02-16): pass something (serverErrors? another channel?) to transfer's handler so that
	// close-connection can terminate the server
	transferHandler := transfer.NewHandler(s.transferService, s.fileService, s.defaultFolderID, s.nonceManager)

	// TODO (2026-02-19): dhekra / iOS closes the server when the transfer is explicitly stopped
	// TODO cblgh(2026-02-16): if using channel for close-connection then make sure, for all other paths, to drain <-done so that we don't have a goroutine leak
	// go func() {
	// 	<-done
	// 	s.Stop(context.TODO)
	// }()

	handler := NewHandler(mux, s.registrationHandler, transferHandler)
	handler.SetupRoutes()

	limitingMiddleware := s.limiter.Handler(mux)

	s.server = &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   limitingMiddleware,
		TLSConfig: tlsConfig,
		// TODO cblgh(2026-02-16): the Timeout options were causing a a dysfunctional timeout behaviour for receiving large
		// files. the timeout would happen when having received ~150MB out of a 200MB large file. this is why they are set
		// to 0
		ReadTimeout:       0, // do not time out when reading body -- we will potentially be receiving multi gigabyte uploads
		ReadHeaderTimeout: 0, //30 * time.Second, // this configures timeouts on the header portion of each request, which should be small :)
		WriteTimeout:      0,
		IdleTimeout:       0,
	}

	s.port = port

	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
			serverErrors <- err
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}
	}()

	// Give the server time to start up properly
	time.Sleep(500 * time.Millisecond)

	// Check if there were any immediate startup errors
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server failed to start: %v", err)
	default:
		// server started successfully
	}

	s.running = true
	fmt.Printf("HTTPS Server started on port %d with PIN %s\n", port, s.pin)
	return nil
}

func (s *service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	fmt.Printf("Stopping HTTPS Server...\n")

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Graceful shutdown failed: %v, forcing close\n", err)
	}

	s.running = false
	s.server = nil

	fmt.Printf("HTTPS Server stopped\n")

	// Add delay to ensure port is fully released
	time.Sleep(1 * time.Second)

	return nil
}

func (s *service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *service) GetPIN() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pin
}

const PIN_LEN = 6

func generateRandomPIN() string {
	maxN := big.NewInt(10)
	var sequence []string
	for i := 0; i < PIN_LEN; i++ {
		// crypto/rand.Int cannot return an error when using crypto/rand.Reader.
		bigN, _ := crand.Int(crand.Reader, maxN)
		sequence = append(sequence, strconv.FormatInt(bigN.Int64(), 10))
	}
	return strings.Join(sequence, "")
}
