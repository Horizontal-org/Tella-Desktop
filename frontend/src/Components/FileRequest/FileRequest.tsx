import { useState, useEffect } from 'react';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { AcceptTransfer, RejectTransfer } from '../../../wailsjs/go/app/App';
import styled from 'styled-components';
import folderIcon from '../../assets/images/icons/folder-icon.svg'
import downloadIcon from '../../assets/images/icons/download-icon.svg'
import clockIcon from '../../assets/images/icons/clock-icon.svg'

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
        <TitleContainer>
          <ClockIcon /> 
          <StepTitle>Waiting for the sender to send files</StepTitle>
        </TitleContainer>
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
      <TitleContainer>
        <DownloadIcon />
        <StepTitle>
          The sender is trying to send you {requestData.totalFiles} files. Would you like to accept them?
        </StepTitle>
      </TitleContainer>
      
      <TransferCard>
        <TransferNameContainer>
          <FolderIcon />
          <TransferTitle>{requestData.title}</TransferTitle>
        </TransferNameContainer>
        <TransferStats>
          {requestData.totalFiles} files
        </TransferStats>
        <TransferStats>
          {formatFileSize(requestData.totalSize)}
        </TransferStats>
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

const TitleContainer = styled.div`
  justify-content: center;
  display: flex;
  border-bottom: 1px solid #CFCFCF;
  align-items: center;
  padding: 1rem 1.5rem;
  gap: 0.5rem;
`

const ClockIcon = styled.div`
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
  background-image: url(${clockIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
`;
const DownloadIcon = styled.div`
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
  background-image: url(${downloadIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
`;
const StepContent = styled.div`
  max-width: 600px;
  width: 100%;
  text-align: center;
  border: 1px solid #CFCFCF;
  border-radius: 8px;
`;

const StepTitle = styled.p`
  font-size: 0.9rem;
  font-weight: 600;
  color: #5F6368;
  text-align: center;
  padding-left: 0.8rem;
`;

const StepSubtitle = styled.p`
  border-top: 1px solid #CFCFCF;
  padding: 1.5rem;
  font-size: 0.9rem;
  color: #6c757d;
  text-align: center;
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem;
  margin-bottom: .5rem;
`;

const TransferNameContainer = styled.div`
  display: flex;
  align-items: center;
  gap: .5rem;
`

const FolderIcon = styled.div`
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
  background-image: url(${folderIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
`;


const TransferTitle = styled.div`
  font-weight: 600;
  color: #5F6368;
  margin-bottom: 0.25rem;
  font-size: 1rem;
`;

const TransferStats = styled.div`
  font-size: 0.875rem;
  color: #5F6368;
`;

const ButtonsContainer = styled.div`
  display: flex;
  gap: 1rem;
  justify-content: center;
  padding: 1.5rem;
  border-top: 1px solid #CFCFCF;
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
  color: #8B8E8F;
  border: 1px solid #D9D9D9;
`;

const AcceptButton = styled(Button)`
  background-color: transparent;
  color: #8B8E8F;
  border: 1px solid #D9D9D9;
`;