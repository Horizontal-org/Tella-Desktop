import { useState, useEffect } from 'react';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { AcceptTransfer, RejectTransfer } from '../../../wailsjs/go/app/App';
import styled from 'styled-components';

interface FileInfo {
  id: string;
  fileName: string;
  size: number;
  fileType: string;
}

interface FileRequestData {
  sessionId: string;
  title: string;
  files: FileInfo[];
  totalFiles: number;
  totalSize: number;
}

interface FileRequestProps {
  onAccept: (sessionId: string) => void;
  onReject: () => void;
  onReceiving: () => void;
}

export function FileRequest({ onAccept, onReject, onReceiving }: FileRequestProps) {
  const [requestData, setRequestData] = useState<FileRequestData | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);

  useEffect(() => {
    const cleanupPrepareRequest = EventsOn("prepare-upload-request", (data) => {
      console.log("Received prepare upload request:", data);
      setRequestData(data as FileRequestData);
    });

    const cleanupFileReceiving = EventsOn("file-receiving", (data) => {
      console.log("File receiving:", data);
      onReceiving();
    });

    return () => {
      cleanupPrepareRequest();
      cleanupFileReceiving();
    };
  }, [onReceiving]);

  const handleAccept = async () => {
    if (!requestData) return;
    
    setIsProcessing(true);
    try {
      await AcceptTransfer(requestData.sessionId);
      onAccept(requestData.sessionId);
      setRequestData(null);
    } catch (error) {
      console.error('Failed to accept transfer:', error);
    } finally {
      setIsProcessing(false);
    }
  };

  const handleReject = async () => {
    if (!requestData) return;
    
    setIsProcessing(true);
    try {
      await RejectTransfer(requestData.sessionId);
      onReject();
      setRequestData(null);
    } catch (error) {
      console.error('Failed to reject transfer:', error);
    } finally {
      setIsProcessing(false);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (!requestData) {
    return (
      <StepContent>
        <StepTitle>Waiting for the sender to send files</StepTitle>
        <LoadingContainer>
          <LoadingSpinner />
        </LoadingContainer>
        <StepSubtitle>
          This screen will automatically update when you've received a request to send files.
        </StepSubtitle>
      </StepContent>
    );
  }

  return (
    <StepContent>
      <StepTitle>
        The sender is trying to send you {requestData.totalFiles} files. Would you like to accept them?
      </StepTitle>
      
      <TransferCard>
        <TransferHeader>
          <TransferDetails>
            <TransferTitle>{requestData.title}</TransferTitle>
            <TransferStats>
              {requestData.totalFiles} files • {formatFileSize(requestData.totalSize)}
            </TransferStats>
          </TransferDetails>
        </TransferHeader>
      </TransferCard>

      <ButtonsContainer>
        <RejectButton 
          onClick={handleReject} 
          disabled={isProcessing}
        >
          ✕ REJECT
        </RejectButton>
        <AcceptButton 
          onClick={handleAccept} 
          disabled={isProcessing}
        >
          {isProcessing ? 'ACCEPTING...' : '✓ ACCEPT'}
        </AcceptButton>
      </ButtonsContainer>
    </StepContent>
  );
}

const StepContent = styled.div`
  max-width: 600px;
  width: 100%;
  text-align: center;
`;

const StepTitle = styled.h2`
  font-size: 1.2rem;
  font-weight: 600;
  color: #212529;
  margin-bottom: 1rem;
`;

const StepSubtitle = styled.p`
  font-size: 0.9rem;
  color: #6c757d;
  margin-bottom: 2rem;
`;

const LoadingContainer = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 3rem 0;
`;

const LoadingSpinner = styled.div`
  width: 48px;
  height: 48px;
  border: 4px solid #e9ecef;
  border-top: 4px solid #007bff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;

const TransferCard = styled.div`
  border: 1px solid #e9ecef;
  border-radius: 8px;
  margin-bottom: 2rem;
  background-color: white;
`;

const TransferHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem;
`;

const TransferDetails = styled.div`
  flex: 1;
  text-align: left;
`;

const TransferTitle = styled.div`
  font-weight: 600;
  color: #212529;
  margin-bottom: 0.25rem;
  font-size: 1rem;
`;

const TransferStats = styled.div`
  font-size: 0.875rem;
  color: #6c757d;
`;

const ProgressIndicator = styled.div`
  width: 60px;
  height: 4px;
  background-color: #e9ecef;
  border-radius: 2px;
  position: relative;
  
  &::after {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    height: 100%;
    width: 30%;
    background-color: #007bff;
    border-radius: 2px;
  }
`;

const ButtonsContainer = styled.div`
  display: flex;
  gap: 1rem;
  justify-content: center;
`;

const Button = styled.button`
  padding: 0.75rem 2rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  text-transform: uppercase;
  min-width: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const RejectButton = styled(Button)`
  background-color: white;
  color: #dc3545;
  border: 1px solid #dc3545;
  
  &:hover:not(:disabled) {
    background-color: #f8f9fa;
    border-color: #c82333;
    color: #c82333;
  }
`;

const AcceptButton = styled(Button)`
  background-color: #28a745;
  color: white;
  border: 1px solid #28a745;
  
  &:hover:not(:disabled) {
    background-color: #218838;
    border-color: #1e7e34;
  }
`;