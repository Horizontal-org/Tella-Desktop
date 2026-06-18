package server

import (
	"context"
	crand "crypto/rand"
	ctls "crypto/tls"
	"fmt"
	"crypto/x509"
	"crypto/sha256"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"errors"
	"sync"
	"time"

	"gomod.cblgh.org/cerca/limiter"

	"Tella-Desktop/backend/core/modules/filestore"
	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/transfer"
	"Tella-Desktop/backend/utils/network"
	"Tella-Desktop/backend/utils/nonces"
	"Tella-Desktop/backend/utils/tls"
	"Tella-Desktop/backend/utils/devlog"
)

var log = devlog.Logger("server")

type service struct {
	pinnedSenderCertificateHash		string // pinned fingerprint for sender
	// fingerprintCandidate is derived from incoming requests when 
	// tlsConfig.ClientAuth is set to RequestClientCert pre-mTLS establishment (post mTLS config is set to RequireAnyClientCert)
	fingerprintCandidate string 
	fingerprintCandidateLockedIn bool
	limitingMiddleware  http.Handler 
	nonceManager        *nonces.NonceManager
	limiter             *RateLimitingWare
	tlsConfig				    *ctls.Config
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
var errStart = errors.New("start error")
func (s *service) Start(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		log("server is already running")
		return errStart
	}

	// Generate new PIN for each start
	s.pin = generateRandomPIN()
	s.registrationService.SetPINCode(s.pin)

	ipStrings, err := network.GetLocalIPs()
	if err != nil {
		log("failed to get local IPs: %v", err)
		return errStart
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
		log("failed to generate TLS config: %v", err)
		return errStart
	}

	// TODO (2026-06-18): change behaviour to wait to send ping response until confirm & continue

	// do not require any client certs when server is freshly started
	tlsConfig.ClientAuth = ctls.RequestClientCert
	// NOTE cblgh(2026-03-15): set up a custom cert pool using the pinned cert?
	// to allow use of tls.Config.ClientAuth: tls.RequireAndVerifyClientCert
	// c.f https://stackoverflow.com/a/63317898
	tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		log("number of certs being passed in: %d", len(rawCerts))
		// currently pre-mtls and not strictly requiring cert
		if (tlsConfig.ClientAuth == ctls.RequestClientCert && len(rawCerts) == 0) {
			return nil
		}
		// we're post-mtls, this should not occur for legitimate requests
		if (tlsConfig.ClientAuth == ctls.RequireAnyClientCert && len(rawCerts) == 0) {
			return errors.New("Sender certificate missing")
		}

		log("VerifyPeerCertificate called")

		sha256CertHash := sha256.Sum256(rawCerts[0])
		hexSHA256CertHash := fmt.Sprintf("%x", sha256CertHash)

		// in this section we use a mutex because we want to be sure to never set s.fingerprintCandidate to something else while it is being
		// fetched by s.GetSenderFingerprintCandidate for being presented to the user
		s.mu.Lock()
		if (tlsConfig.ClientAuth == ctls.RequestClientCert) {
			// prevent multiple register POST (with different senderCertificateHash values) from being sent in succession once
			// hash is displayed 
			// NOTE: this is not our session-long pinning - we haven't definitively set s.pinnedFingerprintCandidate! 
			// we only set s.pinnedSenderCertificateHash on successful sender certificate hash
			// verification i.e. calling s.PinFingerprint() from registration/handler.go
			if !s.fingerprintCandidateLockedIn {
				// still pre-mtls but we have now have a candidate to use for senderFingerprint.
				s.fingerprintCandidate = hexSHA256CertHash
				log("sender fingerprint candidate %s", s.fingerprintCandidate)
			} else {
				if hexSHA256CertHash != s.fingerprintCandidate {
					s.mu.Unlock()
					return errors.New("Hash of incoming request certificate did not match candidate for sender certificate hash")
				}
			}
			s.mu.Unlock()
			return nil
		}
		s.mu.Unlock()

		// we should only reach this point in the routine once we have established mTLS and have a pinned sender certificate hash
		log("incoming cert hash\n%x\n", sha256CertHash)
		if hexSHA256CertHash != s.pinnedSenderCertificateHash {
			return errors.New("Hash of incoming request certificate did not pinned sender certificate hash")
		}
		return nil
	}

	s.tlsConfig = tlsConfig

	mux := http.NewServeMux()

	transferHandler := transfer.NewHandler(s.transferService, s.fileService, s.defaultFolderID, s.nonceManager)

	// TODO (2026-02-19): dhekra / iOS closes the server when the transfer is explicitly stopped

	handler := NewHandler(mux, s.registrationHandler, transferHandler)
	handler.SetupRoutes(s.PinFingerprint, s.GetSenderFingerprintCandidate)

	s.limitingMiddleware = s.limiter.Handler(mux)
	s.port = port
	err = s.startServer()
	time.Sleep(500 * time.Millisecond)
	if err != nil {
		return err
	}

	log("HTTPS Server started on port %d with PIN %s\n", port, s.pin)
	return nil
}

func (s *service) startServer() error {
	s.server = &http.Server{
		Addr:      fmt.Sprintf(":%d", s.port),
		Handler:   s.limitingMiddleware,
		TLSConfig: s.tlsConfig,
		// note cblgh(2026-02-16): the Timeout options were causing a a dysfunctional timeout behaviour for receiving large
		// files. the timeout would happen when having received ~150MB out of a 200MB large file. this is why they are set to 0.
		ReadTimeout:       0, // do not time out when reading body -- we will potentially be receiving multi gigabyte uploads
		ReadHeaderTimeout: 0, 
		WriteTimeout:      0,
		IdleTimeout:       0,
	}

	serverErrors := make(chan error)
	go func() {
		if err := s.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}
	}()
	// Check if there were any immediate startup errors
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server failed to start: %v", err)
	default:
		// server started successfully
	}
	s.running = true
	return nil
}
func (s *service) GetSenderFingerprintCandidate() string {
	var candidate string
	// use mutex because we want to be sure to never set s.fingerprintCandidate to something else while it is being
	// fetched for being presented to the user. 
	//
	// to limit attack surface, we only every allow setting one fingerprint
	// candidate once the sender fingerprint candidate has been displayed by the user
	s.mu.Lock()
	candidate = s.fingerprintCandidate
	log("sender fingerprint candidate locked to %s", candidate)
	s.fingerprintCandidateLockedIn = true
	s.mu.Unlock()
	return candidate
}

// PinFingerprint pins the hexadecimal-encoded SHA256 hash of the raw certificate bytes. The TLS config is changed to require a client cert, which necessitates restarting the https server instance.
func (s *service) PinFingerprint(senderFingerprint string) error {
	log("Pin fingerprint called")
	if len(senderFingerprint) != 64 {
		return errors.New("expected fingerprint string length of 64ch")
	}
	s.pinnedSenderCertificateHash = senderFingerprint
	// terminate the previous instance
	shutdownCtx, cancel := context.WithTimeout(s.ctx, 1500*time.Millisecond)
	defer cancel()
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log("Graceful shutdown failed: %v, forcing close\n", err)
	}
	log("Stopped server & restarting")
	// change the tls config to require client certs on connection going forward.
	// we use `tls.RequireAnyClientCert` as we do not have the full client cert
	// => can't create and pass a cert pool that will allow tls.VerifyCertificate to succeed
	s.tlsConfig.ClientAuth = ctls.RequireAnyClientCert
	// NOTE cblgh(2026-03-15): we need to allocate a new instance of http.Server and set the updated TLS config on it
	// before restarting the server for the config changes to take effect
	return s.startServer()
}

func (s *service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	log("Stopping HTTPS Server...\n")

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log("Graceful shutdown failed: %v, forcing close\n", err)
	}

	s.running = false
	s.server = nil
	s.tlsConfig = nil
	s.limitingMiddleware = nil
	s.pinnedSenderCertificateHash = ""
	s.fingerprintCandidate = ""
	s.fingerprintCandidateLockedIn = false

	log("HTTPS Server stopped\n")

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
