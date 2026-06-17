import { useState, useEffect, useRef } from "react";
import { useNavigate } from 'react-router-dom';
import { GetLocalIPs, RejectRegistration, ConfirmRegistration, StopTransfer } from "../../../../wailsjs/go/app/App";
import { EventsOn } from "../../../../wailsjs/runtime/runtime";
import { useServer } from "../../../Contexts/ServerContext";
import { log } from "../../../util/util"

type FlowStep = 'intro' | 'connect' | 'accept' | 'receive' | 'results';
type ManualConfirmationState = 'CONFIRM_RECEIVER' | 'CONFIRM_SENDER' 

// TODO (2026-06-16): with state transitions etc, make sure to also handle if "sender confirmed before receiver!"

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

interface CloseConnectionData {
  sessionId: string;
  transferOngoing: boolean;
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
  const [receiverCertificateHash, setReceiverCertificateHash] = useState<string>('');
  const [senderCertificateHash, setSenderCertificateHash] = useState<string>('');
  const [modalState, setModalState] = useState<ManualConfirmationState>('CONFIRM_RECEIVER');
  const [senderConfirmedReceiver, setSenderConfirmedReceiver] = useState<boolean>(false)

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
      log("Ping received:", data);
      setShowVerificationModal(true);
      setModalState('CONFIRM_RECEIVER')
    });

    // TODO (2026-06-17):
    // * handle early confirm by sender in way that doesn't fuck up effects
    // * pass senderCertificateHash from golang in the Emit
    const cleanupRegisterListener = EventsOn("register-request-received", (data) => {
      log("Register request received:", data);
      setSenderCertificateHash(data.senderCertificateHash);
      setSenderConfirmedReceiver(true);
    });

    const cleanupCertListener = EventsOn("receiver-certificate-hash", (data) => {
      log("Receiver Certificate hash received:", data);
      setReceiverCertificateHash(data.toString());
    });

    const cleanupPrepareRequest = EventsOn("prepare-upload-request", (data) => {
      log("📨 Received prepare upload request in parent:", data);
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

    const cleanupCloseConnection = EventsOn("close-connection", async (data) => {
      log("XX Received close-connection", data);
      const connectionData = data as CloseConnectionData;
      await stopServer();
      if (connectionData.transferOngoing) {
          // TODO cblgh(2026-04-29): set currentStep to something like results-error?
          setCurrentStep('results');
      } else {
          setCurrentStep('intro');
      }
    });

    return () => {
      cleanupFileReceived();
      cleanupPingListener();
      cleanupRegisterListener();
      cleanupCertListener();
      cleanupPrepareRequest();
      cleanupCloseConnection();
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

  const handleReceiverConfirmReceiver = async () => {
      setModalState("CONFIRM_SENDER")
  }
  // Receiver Certificate verification handlers
  const handleVerificationConfirm = async () => {
    log("✅ Receiver Certificate verification CONFIRMED");
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

  // TODO (2026-06-15): revise copy of handleVerificationDiscard to make sure is implemented properly
  const handleWaitingForSenderCancel = async () => {
    log("❌ Waiting for sender CANCELED");
    try {
      await RejectRegistration();
    } catch (error) {
      console.error("Failed to reject registration:", error);
    }
    // if rejected, reset state
    setShowVerificationModal(false);
    setModalState('CONFIRM_RECEIVER');
    setSenderConfirmedReceiver(false);

    await handleStopServer();
    setCurrentStep('intro');
  };

  const handleVerificationDiscard = async () => {
    log("❌ Receiver Certificate verification DISCARDED");
    try {
      await RejectRegistration();
    } catch (error) {
      console.error("Failed to reject registration:", error);
    }
    
    // reset state
    setShowVerificationModal(false);
    setModalState('CONFIRM_RECEIVER');
    setSenderConfirmedReceiver(false);

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
    log("📝 File request accepted for session:", sessionId);
    setCurrentSessionId(sessionId);
    setCurrentStep('receive');
  };

  const handleFileRequestReject = () => {
    log("❌ File request rejected");
    setTransferData(null);
    // go back to previous screen and allow resending
    setCurrentStep('accept');
  };

  const handleFileReceiving = () => {
    log("📥 File receiving started");
    setCurrentStep('receive');
  };

  const handleReceiveComplete = async () => {
    log("✅ File receiving completed");
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
    log("❌ File transfer stopped");
    // stop the http server
    if (serverRunning) {
      await handleStopServer();
    }
    await StopTransfer(currentSessionId);
    // TODO cblgh(2026-02-19): set currentStep to results-error?
    setCurrentStep('results');
  }

  const handleViewFiles = async () => {
    log("📁 View files clicked - stopping server and navigating");
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
    setReceiverCertificateHash('');
    setSenderCertificateHash('');
    setModalState('CONFIRM_RECEIVER');
    setCurrentStep('intro');
    setSenderConfirmedReceiver(false);
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
    receiverCertificateHash,
    senderCertificateHash,
    senderConfirmedReceiver,
    modalState,

    // Actions
    handleBack,
    handleContinue,
    handleReceiverConfirmReceiver,
    handleVerificationConfirm,
    handleVerificationDiscard,
    handleWaitingForSenderCancel,
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
