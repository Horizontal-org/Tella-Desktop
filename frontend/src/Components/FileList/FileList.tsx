import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { GetFilesInFolder, OpenFileByID } from '../../../wailsjs/go/app/App';
import styled from 'styled-components';

interface FileInfo {
  id: number;
  name: string;
  mimeType: string;
  timestamp: string;
  size: number;
}

interface FolderInfo {
  id: number;
  name: string;
}

const formatTimestamp = (timestamp: string): string => {
  try {
    const date = new Date(timestamp);
    return date.toLocaleString();
  } catch (error) {
    return timestamp;
  }
};

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const getFriendlyMimeType = (mimeType: string): string => {
  const mimeMap: Record<string, string> = {
    'text/plain': 'Text',
    'text/html': 'HTML',
    'text/css': 'CSS',
    'text/javascript': 'JavaScript',
    'application/json': 'JSON',
    'application/xml': 'XML',
    'application/pdf': 'PDF',
    'application/msword': 'Word',
    'application/vnd.ms-excel': 'Excel',
    'application/vnd.ms-powerpoint': 'PowerPoint',
    'image/jpeg': 'JPEG Image',
    'image/png': 'PNG Image',
    'image/gif': 'GIF Image',
    'image/svg+xml': 'SVG Image',
    'audio/mpeg': 'MP3 Audio',
    'audio/wav': 'WAV Audio',
    'video/mp4': 'MP4 Video',
    'video/mpeg': 'MPEG Video',
    'application/octet-stream': 'Binary File',
  };

  return mimeMap[mimeType] || mimeType;
};

const getFileIcon = (mimeType: string): string => {
  if (mimeType.startsWith('image/')) return '🖼️';
  if (mimeType.startsWith('video/')) return '🎥';
  if (mimeType.startsWith('audio/')) return '🎵';
  if (mimeType.includes('pdf')) return '📄';
  if (mimeType.includes('word')) return '📝';
  if (mimeType.includes('excel') || mimeType.includes('spreadsheet')) return '📊';
  if (mimeType.includes('powerpoint') || mimeType.includes('presentation')) return '📊';
  if (mimeType.startsWith('text/')) return '📄';
  return '📁';
};

interface FileListProps {
  folderId?: number;
  folderName?: string;
}

export function FileList({ folderId: propFolderId, folderName: propFolderName }: FileListProps) {
  const { folderId: paramFolderId } = useParams<{ folderId: string }>();
  const navigate = useNavigate();
  
  const folderId = propFolderId || (paramFolderId ? parseInt(paramFolderId, 10) : null);
  
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [folderInfo, setFolderInfo] = useState<FolderInfo | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFiles = async () => {
    if (!folderId) {
      setError('No folder ID provided');
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const response = await GetFilesInFolder(folderId);
      setFiles(response.files);
      setFolderInfo({ id: folderId, name: propFolderName || response.folderName });
    } catch (err) {
      console.error('Failed to fetch files:', err);
      setError('Failed to fetch files. Please ensure you are logged in.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFiles();
  }, [folderId]);

  const handleFileClick = async (fileId: number) => {
    try {
      await OpenFileByID(fileId);
    } catch (err) {
      console.error('Failed to open file:', err);
    }
  };

  const handleBack = () => {
    navigate('/');
  };

  if (loading) {
    return (
      <FilesContainer>
        <Header>
          <BackButton onClick={handleBack}>← Back</BackButton>
          <Title>Loading...</Title>
        </Header>
        <LoadingMessage>Loading files...</LoadingMessage>
      </FilesContainer>
    );
  }

  if (error) {
    return (
      <FilesContainer>
        <Header>
          <BackButton onClick={handleBack}>← Back</BackButton>
          <Title>Error</Title>
        </Header>
        <ErrorMessage>{error}</ErrorMessage>
      </FilesContainer>
    );
  }

  if (!files || files.length === 0) {
    return (
      <FilesContainer>
        <Header>
          <BackButton onClick={handleBack}>← Back</BackButton>
          <Title>{folderInfo?.name || 'Folder'}</Title>
        </Header>
        <NoFilesMessage>
          No files found in this folder.
        </NoFilesMessage>
      </FilesContainer>
    );
  }

  return (
    <FilesContainer>
      <Header>
        <BackButton onClick={handleBack}>← Back</BackButton>
        <Title>{folderInfo?.name || 'Folder'}</Title>
        <FileCount>{files.length} files</FileCount>
      </Header>
      
      <FilesTable>
        <thead>
          <tr>
            <TableHeader></TableHeader>
            <TableHeader>Name</TableHeader>
            <TableHeader>Type</TableHeader>
            <TableHeader>Size</TableHeader>
            <TableHeader>Date Added</TableHeader>
          </tr>
        </thead>
        <tbody>
          {files.map((file) => (
            <FileRow 
              key={file.id} 
              onClick={() => handleFileClick(file.id)}
            >
              <TableCell>
                <FileIcon>{getFileIcon(file.mimeType)}</FileIcon>
              </TableCell>
              <TableCell>
                <FileName>{file.name}</FileName>
              </TableCell>
              <TableCell>{getFriendlyMimeType(file.mimeType)}</TableCell>
              <TableCell>{formatFileSize(file.size)}</TableCell>
              <TableCell>{formatTimestamp(file.timestamp)}</TableCell>
            </FileRow>
          ))}
        </tbody>
      </FilesTable>
    </FilesContainer>
  );
}

const FilesContainer = styled.div`
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
`;

const Header = styled.div`
  display: flex;
  align-items: center;
  margin-bottom: 2rem;
  gap: 1rem;
`;

const BackButton = styled.button`
  background: none;
  border: none;
  color: ${({ theme }) => theme.colors.lightGray};
  cursor: pointer;
  font-size: 1rem;
  padding: 0.5rem;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  transition: background-color 0.2s;
  
  &:hover {
    background-color: rgba(255, 255, 255, 0.1);
    color: ${({ theme }) => theme.colors.darkGray};
  }
`;

const Title = styled.h1`
  color: ${({ theme }) => theme.colors.darkGray};
  margin: 0;
  font-size: 1.8rem;
  font-weight: 600;
`;

const FileCount = styled.span`
  color: ${({ theme }) => theme.colors.lightGray};
  font-size: 1rem;
  margin-left: auto;
`;

const FilesTable = styled.table`
  width: 100%;
  border-collapse: collapse;
  margin-top: 1rem;
  background-color: rgba(255, 255, 255, 0.02);
  border-radius: ${({ theme }) => theme.borderRadius.default};
  overflow: hidden;
`;

const TableHeader = styled.th`
  padding: 1rem 0.75rem;
  text-align: left;
  border-bottom: 2px solid rgba(255, 255, 255, 0.1);
  color: ${({ theme }) => theme.colors.darkGray};
  font-weight: 600;
  background-color: rgba(255, 255, 255, 0.05);
  
  &:first-child {
    width: 50px;
  }
`;

const FileRow = styled.tr`
  cursor: pointer;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }
`;

const TableCell = styled.td`
  padding: 1rem 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  color: ${({ theme }) => theme.colors.lightGray};
  vertical-align: middle;
`;

const FileIcon = styled.span`
  font-size: 1.5rem;
  display: flex;
  justify-content: center;
  align-items: center;
`;

const FileName = styled.span`
  font-weight: 500;
  color: ${({ theme }) => theme.colors.darkGray};
`;

const LoadingMessage = styled.div`
  padding: 3rem 1rem;
  text-align: center;
  color: ${({ theme }) => theme.colors.lightGray};
  font-size: 1.1rem;
`;

const ErrorMessage = styled.div`
  padding: 2rem 1rem;
  text-align: center;
  color: ${({ theme }) => theme.colors.error};
  background-color: rgba(239, 68, 68, 0.1);
  border-radius: ${({ theme }) => theme.borderRadius.default};
  border: 1px solid rgba(239, 68, 68, 0.3);
`;

const NoFilesMessage = styled.div`
  padding: 3rem 1rem;
  text-align: center;
  color: ${({ theme }) => theme.colors.lightGray};
  background-color: rgba(255, 255, 255, 0.02);
  border-radius: ${({ theme }) => theme.borderRadius.default};
  border: 1px dashed rgba(255, 255, 255, 0.2);
`;

const RefreshButton = styled.button`
  padding: 0.5rem 1rem;
  background-color: ${({ theme }) => theme.colors.primary};
  color: white;
  border: none;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  cursor: pointer;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  
  &:hover {
    background-color: #1e56c9;
  }
`;

export default FileList;