import { useState, useEffect, useRef } from "react";
import { useNavigate } from 'react-router-dom';
import { GetLocalIPs, RejectRegistration, ManualConfirmationReceiverForReceiver, ConfirmRegistration, StopTransfer } from "../../../../wailsjs/go/app/App";
import { EventsOn } from "../../../../wailsjs/runtime/runtime";
import { useServer } from "../../../Contexts/ServerContext";
import { log } from "../../../util/util"

type FlowStep = 'intro' | 'connect' | 'accept' | 'receive' | 'results' | 'interrupted';
type ManualConfirmationState = 'CONFIRM_RECEIVER' | 'CONFIRM_SENDER' 

interface OnReceiveCompleteProps {
    numFailed: number;
    numReceived: number;
    totalFiles: number;
}

interface FileInfo {
  id: string;
  fileName: string;
  size: number;
  fileType: string;
}

interface NearbySharingError {
    text: string;
    button: string;
    hasError: boolean;
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
  numInProgressFiles: number;
}

export function useNearbySharing() {
  const navigate = useNavigate();
  const { isRunning: serverRunning, isStarting: isStartingServer, startServer, stopServer } = useServer();
  
  // Flow state
  const [currentStep, setCurrentStep] = useState<FlowStep>('intro');
  
  // Network state
  const [localIPs, setLocalIPs] = useState<string[]>([]);

  // Error state text
  const [nearbySharingError, setNearbySharingError] = useState<NearbySharingError>({ text: "", button: "", hasError: false});
  
  // Transfer state
  const [currentSessionId, setCurrentSessionId] = useState<string>('');
  const [transferData, setTransferData] = useState<TransferData | null>(null);
  
  // Certificate verification state
  const [showVerificationModal, setShowVerificationModal] = useState(false);
  const [receiverCertificateHash, setReceiverCertificateHash] = useState<string>('');
  const [senderCertificateHash, setSenderCertificateHash] = useState<string>('');
  const [modalState, setModalState] = useState<ManualConfirmationState>('CONFIRM_RECEIVER');
  const [senderConfirmedReceiver, setSenderConfirmedReceiver] = useState<boolean>(false)

  // Stop dialog while receiving files
  const [showStopDialog, setShowStopDialog] = useState(false);

  // Initialize network info and event listeners
  useEffect(() => {
    const fetchNetworkInfo = async () => {
      try {
        const ips = await GetLocalIPs();
        setLocalIPs(ips);
      } catch (error) {
        console.error('Failed to get network info:', error);
        // TODO (2026-06-18): make sure that this will actually be seen (currently part of CertVerificationModal!)
        setNearbySharingError({ text: "Failed to get network info and could not start server.", button: "Start over", hasError: true } as NearbySharingError)
      }
    };

    fetchNetworkInfo();

    const cleanupPingListener = EventsOn("ping-received", (data) => {
      log("Ping received:", data);
      setShowVerificationModal(true);
      setModalState('CONFIRM_RECEIVER')
    });

    // TODO (2026-06-18): actually emit 'nearby-sharing-error' somewhere in the backend
    const cleanupErrorListener = EventsOn("nearby-sharing-error", (data) => {
      let err = data as NearbySharingError
      err.hasError = true
      setNearbySharingError(err)
    });

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

    // TODO (2026-06-22): implement event in backend and handler here in frontend that signals that register timed out or
    // max PIN registration attempts has been reached

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
      // NOTE (2026-07-06): can close connection sometimes be received in a race-like manner & we accidentally set
      // "interrupted" while we have received all files?
      // this should be remedied as of 61ccaad
      if (connectionData.transferOngoing) {
        if (connectionData.numInProgressFiles === 0) {
            setCurrentStep('results');
        } else {
            setCurrentStep('interrupted');
        }
      } else {
          setCurrentStep('intro');
      }
    });

    return () => {
      cleanupFileReceived();
      cleanupPingListener();
      cleanupErrorListener();
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

  // {Receiver, Sender} Certificate Hash verification handlers
  const handleReceiverConfirmReceiver = async () => {
      await ManualConfirmationReceiverForReceiver()
      setModalState("CONFIRM_SENDER")
  }

  const handleVerificationConfirm = async () => {
    log("✅ Sender Certificate Hash: verification CONFIRMED");
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
    log("❌ Verification DISCARDED");
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

  const handleTryAgain = async () => {
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

  const handleReceiveComplete = async ({ totalFiles, numReceived, numFailed }: OnReceiveCompleteProps) => {
    log("✅ File receiving completed");
    // all files have been handled (either completely transferred or failed) we can close the transfer session
    await StopTransfer(currentSessionId);
    // the file receiving is complete, stop the server
    if (serverRunning) {
      await handleStopServer();
    }
    log(`total ${totalFiles} numRecv ${numReceived} numFailed ${numFailed}`)
    if (numFailed > 0 || numReceived < totalFiles) {
        setCurrentStep('interrupted');
    } else {
        // if all files were received and none of them were failed:
        setCurrentStep('results');
    }
  };

  const handleClickStopTransfer = () => { 
      setShowStopDialog(true)
  }

  const handleHideStopDialog = () => { 
      setShowStopDialog(false)
  }

  // called when "stop transfer" is clicked in the middle of an ongoing transfer
  const handleStopTransfer = async () => {
    log("❌ File transfer stopped");
    // stop the http server
    if (serverRunning) {
      await handleStopServer();
    }
    await StopTransfer(currentSessionId);
    setShowStopDialog(false)
    setCurrentStep('interrupted');
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
    setShowStopDialog(false)
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
    nearbySharingError,
    showVerificationModal,
    receiverCertificateHash,
    senderCertificateHash,
    senderConfirmedReceiver,
    modalState,

    showStopDialog,
    handleHideStopDialog,

    // Actions
    handleBack,
    handleContinue,
    handleReceiverConfirmReceiver,
    handleVerificationConfirm,
    handleVerificationDiscard,
    handleTryAgain,
    handleFileRequestAccept,
    handleFileRequestReject,
    handleFileReceiving,
    handleStopTransfer,
    handleReceiveComplete,
    handleViewFiles,
    handleClickStopTransfer,
    
    // Server actions (delegated to context)
    startServer: handleStartServer,
    stopServer: handleStopServer,
    resetState
  };
}
