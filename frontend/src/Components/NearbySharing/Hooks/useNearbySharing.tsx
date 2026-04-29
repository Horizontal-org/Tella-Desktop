import { useState, useEffect, useRef } from "react";
import { useNavigate } from 'react-router-dom';
import { GetLocalIPs, RejectRegistration, ConfirmRegistration, StopTransfer } from "../../../../wailsjs/go/app/App";
import { EventsOn } from "../../../../wailsjs/runtime/runtime";
import { useServer } from "../../../Contexts/ServerContext";

type FlowStep = 'intro' | 'connect' | 'accept' | 'receive' | 'results';
type ModalState = 'waiting' | 'confirm';

interface FileInfo {
  id: string;
  fileName: string;
  size: number;
  fileType: string;
}

interface TransferData {
  sessionId: string;
  title: string;
  files: FileInfo[];
  totalFiles: number;
  transferredFiles: number;
  totalSize: number;
}

export function useNearbySharing() {
  const navigate = useNavigate();
  const { isRunning: serverRunning, isStarting: isStartingServer, startServer, stopServer } = useServer();
  
  // Flow state
  const [currentStep, setCurrentStep] = useState<FlowStep>('intro');
  
  // Network state
  const [localIPs, setLocalIPs] = useState<string[]>([]);
  
  // Transfer state
  const [currentSessionId, setCurrentSessionId] = useState<string>('');
  const [transferData, setTransferData] = useState<TransferData | null>(null);
  
  // Certificate verification state
  const [showVerificationModal, setShowVerificationModal] = useState(false);
  const [certificateHash, setCertificateHash] = useState<string>('');
  const [modalState, setModalState] = useState<ModalState>('waiting');

  // Connection mode state
  const [isUsingQRMode, setIsUsingQRMode] = useState(true);
  const isUsingQRModeRef = useRef(true);

  // Initialize network info and event listeners
  useEffect(() => {
    const fetchNetworkInfo = async () => {
      try {
        const ips = await GetLocalIPs();
        setLocalIPs(ips);
      } catch (error) {
        console.error('Failed to get network info:', error);
      }
    };

    fetchNetworkInfo();

    const cleanupPingListener = EventsOn("ping-received", (data) => {
      console.log("Ping received from iOS device:", data);
      setShowVerificationModal(true);
      setModalState('waiting')
    });

    const cleanupRegisterListener = EventsOn("register-request-received", (data) => {
      console.log("Register request received:", data);
      console.log("Current QR mode:", isUsingQRModeRef.current);

      if (isUsingQRModeRef.current) {
        console.log("🔗 QR mode active - auto-confirming registration");
        // Auto-confirm for QR mode
        ConfirmRegistration()
          .then(() => {
            console.log("✅ QR registration confirmed successfully");
            setCurrentStep('accept');
          })
          .catch((error) => {
            console.error("❌ Failed to auto-confirm QR registration:", error);
            // Fall back to manual confirmation
            setModalState('confirm');
            setShowVerificationModal(true);
          });
      } else {
        console.log("📱 Manual mode - showing confirmation modal");
        // Manual mode - show certificate verification modal
        setModalState('confirm');
      }
    });

    const cleanupCertListener = EventsOn("certificate-hash", (data) => {
      console.log("Certificate hash received:", data);
      setCertificateHash(data.toString());
    });

    const cleanupPrepareRequest = EventsOn("prepare-upload-request", (data) => {
      console.log("📨 Received prepare upload request in parent:", data);
      const requestData = data as TransferData;
      setTransferData(requestData);
      setCurrentSessionId(requestData.sessionId);
    });

    const cleanupFileReceived = EventsOn("file-received", () => {
      setTransferData(prev => {
          if (prev !== null) {
              const newTransferData = { ...prev, transferredFiles: prev.transferredFiles + 1 }
              return newTransferData
          }
          return prev
      })
    })

    return () => {
      cleanupFileReceived();
      cleanupPingListener();
      cleanupRegisterListener();
      cleanupCertListener();
      cleanupPrepareRequest();
    };
  }, []);

  // Server management - now delegated to ServerContext
  const handleStartServer = async () => {
    const success = await startServer();
    if (success) {
      setCurrentStep('connect');
    }
    return success;
  };

  const handleStopServer = async () => {
    return await stopServer();
  };

  // Certificate verification handlers
  const handleVerificationConfirm = async () => {
    console.log("✅ Certificate verification CONFIRMED");
    try {
      await ConfirmRegistration();
      setShowVerificationModal(false);
      setCurrentStep('accept');
      return true;
    } catch (error) {
      console.error("Failed to confirm registration:", error);
      return false;
    }
  };

  const handleVerificationDiscard = async () => {
    console.log("❌ Certificate verification DISCARDED");
    try {
      await RejectRegistration();
    } catch (error) {
      console.error("Failed to reject registration:", error);
    }
    
    setShowVerificationModal(false);
    setModalState('waiting');
    await handleStopServer();
    setCurrentStep('intro');
  };

  // Flow navigation handlers
  const handleBack = async () => {
    if (serverRunning) {
      await handleStopServer();
    }
    
    resetState();
    navigate('/');
  };

  const handleContinue = async () => {
    if (currentStep === 'intro') {
      await handleStartServer();
    }
  };

  // File transfer handlers
  const handleFileRequestAccept = (sessionId: string) => {
    console.log("📝 File request accepted for session:", sessionId);
    setCurrentSessionId(sessionId);
    setCurrentStep('receive');
  };

  const handleFileRequestReject = () => {
    console.log("❌ File request rejected");
    setTransferData(null);
    // go back to previous screen and allow resending
    setCurrentStep('accept');
  };

  const handleFileReceiving = () => {
    console.log("📥 File receiving started");
    setCurrentStep('receive');
  };

  const handleReceiveComplete = async () => {
    console.log("✅ File receiving completed");
    // all files have been handled (either completely transferred or failed) we can close the transfer session
    await StopTransfer(currentSessionId);
    // the file receiving is complete, stop the server
    if (serverRunning) {
      await handleStopServer();
    }
    setCurrentStep('results');
  };

  // called when "stop transfer" is clicked in the middle of an ongoing transfer
  const handleStopTransfer = async () => {
    console.log("❌ File transfer stopped");
    // stop the http server
    if (serverRunning) {
      await handleStopServer();
    }
    await StopTransfer(currentSessionId);
    // TODO cblgh(2026-02-19): set currentStep to results-error?
    setCurrentStep('results');
  }

  const handleViewFiles = async () => {
    console.log("📁 View files clicked - stopping server and navigating");
    if (serverRunning) {
      await handleStopServer();
    }
    navigate('/');
  };

  // Reset all state
  const resetState = () => {
    setCurrentSessionId('');
    setTransferData(null);
    setShowVerificationModal(false);
    setCertificateHash('');
    setModalState('waiting');
    setCurrentStep('intro');
    setIsUsingQRMode(true);
    isUsingQRModeRef.current = true;
  };

  return {
    // State
    currentStep,
    serverRunning,
    isStartingServer,
    localIPs,
    currentSessionId,
    transferData,
    showVerificationModal,
    certificateHash,
    modalState,
    isUsingQRMode,
    
    // State setters

    // QR mode handler
    handleQRModeChange: (isQR: boolean) => {
      setIsUsingQRMode(isQR);
      isUsingQRModeRef.current = isQR;
      console.log("QR mode changed to:", isQR);
    },
    
    // Actions
    handleBack,
    handleContinue,
    handleVerificationConfirm,
    handleVerificationDiscard,
    handleFileRequestAccept,
    handleFileRequestReject,
    handleFileReceiving,
    handleStopTransfer,
    handleReceiveComplete,
    handleViewFiles,
    
    // Server actions (delegated to context)
    startServer: handleStartServer,
    stopServer: handleStopServer,
    resetState
  };
}
