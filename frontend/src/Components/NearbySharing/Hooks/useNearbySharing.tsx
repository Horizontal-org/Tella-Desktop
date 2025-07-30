import { useState, useEffect } from "react";
import { useNavigate } from 'react-router-dom';
import { EventsOn } from "../../../../wailsjs/runtime/runtime";
import { StartServer, StopServer, GetLocalIPs, RejectRegistration, ConfirmRegistration } from "../../../../wailsjs/go/app/App";
const SERVER_PORT = 53317;

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
  totalSize: number;
}

export function useNearbySharing() {
  const navigate = useNavigate();
  
  // Flow state
  const [currentStep, setCurrentStep] = useState<FlowStep>('intro');
  
  // Server state
  const [serverRunning, setServerRunning] = useState(false);
  const [isStartingServer, setIsStartingServer] = useState(false);
  
  // Network state
  const [wifiNetwork, setWifiNetwork] = useState<string>('');
  const [isWifiConfirmed, setIsWifiConfirmed] = useState(false);
  const [localIPs, setLocalIPs] = useState<string[]>([]);
  
  // Transfer state
  const [currentSessionId, setCurrentSessionId] = useState<string>('');
  const [transferData, setTransferData] = useState<TransferData | null>(null);
  
  // Certificate verification state
  const [showVerificationModal, setShowVerificationModal] = useState(false);
  const [certificateHash, setCertificateHash] = useState<string>('');
  const [modalState, setModalState] = useState<ModalState>('waiting');

  // Initialize network info and event listeners
  useEffect(() => {
    const fetchNetworkInfo = async () => {
      try {
        const ips = await GetLocalIPs();
        setLocalIPs(ips);
        setWifiNetwork('dontstealmywifi');
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
      setModalState('confirm');
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

    return () => {
      cleanupPingListener();
      cleanupRegisterListener();
      cleanupCertListener();
      cleanupPrepareRequest();
    };
  }, []);

  // Server management
  const startServer = async () => {
    try {
      setIsStartingServer(true);
      await StartServer(SERVER_PORT);
      setServerRunning(true);
      return true;
    } catch (error) {
      console.error("Failed to start server:", error);
      return false;
    } finally {
      setIsStartingServer(false);
    }
  };

  const stopServer = async () => {
    if (serverRunning) {
      try {
        await StopServer();
        setServerRunning(false);
        return true;
      } catch (error) {
        console.error("Failed to stop server:", error);
        return false;
      }
    }
    return true;
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
    await stopServer();
    setCurrentStep('intro');
  };

  // Flow navigation handlers
  const handleBack = async () => {
    if (serverRunning) {
      await stopServer();
    }
    
    resetState();
    navigate('/');
  };

  const handleContinue = async () => {
    if (currentStep === 'intro' && isWifiConfirmed) {
      const success = await startServer();
      if (success) {
        setCurrentStep('connect');
      }
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
    setCurrentSessionId('');
    setCurrentStep('connect');
  };

  const handleFileReceiving = () => {
    console.log("📥 File receiving started");
    setCurrentStep('receive');
  };

  const handleReceiveComplete = () => {
    console.log("✅ File receiving completed");
    setCurrentStep('results');
  };

  const handleViewFiles = async () => {
    console.log("📁 View files clicked - stopping server and navigating");
    if (serverRunning) {
      await stopServer();
    }
    navigate('/');
  };

  // Reset all state
  const resetState = () => {
    setCurrentSessionId('');
    setTransferData(null);
    setIsWifiConfirmed(false);
    setShowVerificationModal(false);
    setCertificateHash('');
    setModalState('waiting');
    setCurrentStep('intro');
  };

  return {
    currentStep,
    serverRunning,
    isStartingServer,
    wifiNetwork,
    isWifiConfirmed,
    localIPs,
    currentSessionId,
    transferData,
    showVerificationModal,
    certificateHash,
    modalState,

    setIsWifiConfirmed,

    handleBack,
    handleContinue,
    handleVerificationConfirm,
    handleVerificationDiscard,
    handleFileRequestAccept,
    handleFileRequestReject,
    handleFileReceiving,
    handleReceiveComplete,
    handleViewFiles,
    startServer,
    stopServer,
    resetState
  };
}