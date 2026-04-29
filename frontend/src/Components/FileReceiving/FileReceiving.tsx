import { useState, useEffect } from 'react';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import styled, { keyframes } from 'styled-components';
import folderIcon from '../../assets/images/icons/folder-icon.svg'

import { sanitizeUGC, log } from "../../util/util"

interface FileReceivingData {
  sessionId: string;
  fileId: string;
  fileName: string;
  fileSize?: number;
  transmissionId?: string;
}

interface FileInfo {
  id: string;
  fileName: string;
  size: number;
  fileType: string;
}

interface FileReceivingProps {
  sessionId: string;
  transferTitle: string;
  totalFiles: number;
  totalSize: number;
  files: FileInfo[];
  onComplete: () => void;
  onStop: () => void;
}

export function FileReceiving({ sessionId, 
  transferTitle, 
  totalFiles, 
  totalSize, 
  files, 
  onComplete,
  onStop
}: FileReceivingProps) {
  const [receivingFiles, setReceivingFiles] = useState<FileReceivingData[]>([]);
  const [completedFiles, setCompletedFiles] = useState<FileReceivingData[]>([]);
  const [failedFiles, setFailedFiles] = useState<FileReceivingData[]>([]);
  const [receivedSize, setReceivedSize] = useState(0);

  // TODO: cblgh(2026-02-20): these effects need to be improved because they are somehow being registered twice without
  // proper deregistration due to the debug behaviour of react (during dev, react runs all effects twice)
  // Initialize receiving files from props
  useEffect(() => {
    const initialFiles = files.map(file => ({
      sessionId: sessionId,
      fileId: file.id,
      fileName: file.fileName,
      fileSize: file.size
    }));
    setReceivingFiles(initialFiles);
    log("🔄 Initialized FileReceiving with:", {
      sessionId,
      transferTitle,
      totalFiles,
      totalSize,
      files: initialFiles
    });
  }, [sessionId, files]);

  // TODO: cblgh(2026-02-20): these effects need to be improved because they are somehow being registered twice without
  // proper deregistration due to the debug behaviour of react (during dev, react runs all effects twice). this
  // incorrectly causes the progress indicator to show double the amount of files processed (so 8 instead of 4, when 4
  // files have been received) and double the amount of data transferred.
  //
  // i have verified that this problem is not present in release builds
  useEffect(() => {
    log("Setting up FileReceiving event listeners for sessionId:", sessionId);

    // Listen for individual file upload start events
    const cleanupReceiving = EventsOn("file-receiving", (data) => {
      log("File receiving:", data);
      const fileData = data as FileReceivingData;
      
      if (fileData.sessionId === sessionId) {
        log("File receiving for our session:", fileData);
        setReceivingFiles(prev => {
          return prev.map(f => 
            f.fileId === fileData.fileId 
              ? { ...f, ...fileData, status: 'receiving' }
              : f
          );
        });
      }
    });

    // Listen for file upload completion events
    const cleanupReceived = EventsOn("file-received", (data) => {
      log("✅ File received:", data);
      const fileData = data as FileReceivingData;
      
      if (fileData.sessionId === sessionId) {
        log("✅ File completed for our session:", fileData);
        
        // Move from receiving to completed
        setReceivingFiles(prev => prev.filter(f => f.fileId !== fileData.fileId));
        setCompletedFiles(prev => {
          const exists = prev.some(f => f.fileId === fileData.fileId);
          if (!exists) {
            const newCompleted = [...prev, fileData];
            
            if (fileData.fileSize) {
              setReceivedSize(prevSize => {
                const newSize = prevSize + fileData.fileSize!;
                log(`Updated receivedSize: ${newSize}/${totalSize}`);
                return newSize;
              });
            }

            log(`Progress: ${failedFiles.length + newCompleted.length}/${totalFiles} files completed`);
            if ((failedFiles.length + newCompleted.length) === totalFiles && totalFiles > 0) {
              log("🎉 All files completed!");
              // TODO cblgh(2026-02-16): call down to the backend for a new fn "CleanupTransfer / AllFilesResolved". 
              setTimeout(() => onComplete(), 1000);
            }
            return newCompleted;
          }
          return prev;
        });
      }
    });

    // Listen for file upload failure events
    const cleanupFailed = EventsOn("file-receive-failed", (data) => {
      log("❌ File failed:", data);
      const fileData = data as FileReceivingData;
      
      if (fileData.sessionId === sessionId) {
        log("❌ File failed for our session:", fileData);
        
        // Move from receiving to failed
        setReceivingFiles(prev => prev.filter(f => f.fileId !== fileData.fileId));
        setFailedFiles(prev => {
          const exists = prev.some(f => f.fileId === fileData.fileId);
          if (!exists) {
            const newFailed = [...prev, fileData];
            
            if ((newFailed.length + completedFiles.length) === totalFiles && totalFiles > 0) {
              log("🎉 All files completed!");
              // TODO cblgh(2026-02-16): call down to the backend for a new fn "CleanupTransfer / AllFilesResolved". 
              setTimeout(() => onComplete(), 1000);
            }
            return newFailed;
          }
          return prev;
        });
      }
    });

    // Listen for file upload progress events
    const cleanupProgress = EventsOn("file-upload-progress", (data) => {
      log("📈 File upload progress:", data);
      const progressData = data as { 
        sessionId: string; 
        fileId: string; 
        bytesReceived: number; 
        totalBytes: number;
      };
      
    });

    // TODO cblgh(2026-02-16): event that is currently unused?
    // Listen for transfer cancellation
    const cleanupCancel = EventsOn("transfer-cancelled", (data) => {
      log("❌ Transfer cancelled:", data);
      const cancelData = data as { sessionId: string };
      
      if (cancelData.sessionId === sessionId) {
        setReceivingFiles([]);
        setCompletedFiles([]);
        onComplete();
      }
    });

    return () => {
      log("🧹 Cleaning up FileReceiving event listeners");
      cleanupReceiving();
      cleanupReceived();
      cleanupFailed();
      cleanupProgress();
      cleanupCancel();
    };
  }, [sessionId, onComplete, totalFiles, totalSize]);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const currentFileNumber = completedFiles.length + (receivingFiles.length > 0 ? 1 : 0);

  const handleCancelTransfer = () => {
    log("Cancel transfer requested for session:", sessionId);
    // TODO cblgh(2026-02-16): bubble up call to backend for ending the transfer
    onStop();
  };

  return (
    <Container>
      <StatusMessage>
        <SpinnerIcon />
        Receiving and encrypting files
      </StatusMessage>

      <TransferCard>
        <TransferHeader>
          <TransferName>
            <FolderIcon/>
            <TransferText>{sanitizeUGC(transferTitle)}</TransferText>
          </TransferName>
           <ProgressText>
            {Math.min(currentFileNumber, totalFiles)}/{totalFiles} files
          </ProgressText>
          <SizeInfo>
            {formatFileSize(receivedSize)}/{formatFileSize(totalSize)}
          </SizeInfo>
        </TransferHeader>

        <ButtonContainer>
          <CancelButton onClick={handleCancelTransfer}>
            <CancelIcon>✕</CancelIcon>
            STOP TRANSFER
          </CancelButton>
        </ButtonContainer>
      </TransferCard>

    </Container>
  );
}

const spin = keyframes`
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
`;

const slideIn = keyframes`
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
`;

const Container = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 600px;
  animation: ${slideIn} 0.3s ease-out;

  border: 1px solid #CFCFCF;
   border-radius: 8px;
`;

const StatusMessage = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  color: #6c757d;
  font-size: 1rem;
  padding: 2rem;
  border-bottom: 1px solid #CFCFCF; 
  width: 100%;
  text-align: center;
`;

const SpinnerIcon = styled.div`
  width: 20px;
  height: 20px;
  border: 2px solid #e9ecef;
  border-top: 2px solid #007bff;
  border-radius: 50%;
  animation: ${spin} 1s linear infinite;
`;

const TransferCard = styled.div`
  width: 100%;
  background: white;
  border-radius: 12px;
`;

const TransferHeader = styled.div`
  display: flex;
  align-items: center;
  gap: 2rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid #CFCFCF; 
  padding: 1.5rem;
`;

const FolderIcon = styled.div`
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
  background-image: url(${folderIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
`;

const TransferName = styled.div`
  display: flex;
  align-items: center;
  gap: .5rem;
`
const TransferText = styled.p`
  margin: 0;
  font-size: 12px;
  font-weight: 700;
  color: #2C2C2CBF;
`;

const ProgressText = styled.div`
  color: #6c757d;
  font-size: 0.875rem;
  margin-top: 0.25rem;
`;

const SizeInfo = styled.div`
  color: #6c757d;
  font-size: 0.875rem;
  text-align: right;
`;

const ButtonContainer = styled.div`
  display: flex;
  justify-content: center;
  padding-bottom: 1rem;
`
const CancelButton = styled.button`
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: none;
  border: 1px solid #D9D9D9;
  color: #8B8E8F;
  padding: 0.75rem 1.5rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 600;
  transition: all 0.2s ease;
`;

const CancelIcon = styled.span`
  font-size: 1rem;
  font-weight: bold;
`;
