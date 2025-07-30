import { useState, useEffect } from "react";
import { useNavigate } from 'react-router-dom';
import { GetLocalIPs, RejectRegistration, ConfirmRegistration, GetWiFiNetworkName } from "../../../../wailsjs/go/app/App";
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
  totalSize: number;
}

export function useNearbySharing() {
  const navigate = useNavigate();
  const { isRunning: serverRunning, isStarting: isStartingServer, startServer, stopServer } = useServer();
  
  // Flow state
  const [currentStep, setCurrentStep] = useState<FlowStep>('intro');
  
  // Network state
  const [wifiNetwork, setWifiNetwork] = useState<string>('');
  const [isWifiConfirmed, setIsWifiConfirmed] = useState(false);
  const [localIPs, setLocalIPs] = useState<string[]>([]);
  const [isLoadingWifi, setIsLoadingWifi] = useState<boolean>(false);
  
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
        setIsLoadingWifi(true);

        try {
          const wifiName = await GetWiFiNetworkName();
          setWifiNetwork(wifiName);
        } catch (wifiErr) {
          console.error('Failed to get WiFi network name:', wifiErr);
        } finally {
          setIsLoadingWifi(false);
        }
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
    if (currentStep === 'intro' && isWifiConfirmed) {
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
      await handleStopServer();
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
    // State
    currentStep,
    serverRunning,
    isStartingServer,
    wifiNetwork,
    isLoadingWifi,
    isWifiConfirmed,
    localIPs,
    currentSessionId,
    transferData,
    showVerificationModal,
    certificateHash,
    modalState,
    
    // State setters
    setIsWifiConfirmed,
    
    // Actions
    handleBack,
    handleContinue,
    handleVerificationConfirm,
    handleVerificationDiscard,
    handleFileRequestAccept,
    handleFileRequestReject,
    handleFileReceiving,
    handleReceiveComplete,
    handleViewFiles,
    
    // Server actions (delegated to context)
    startServer: handleStartServer,
    stopServer: handleStopServer,
    resetState
  };
}