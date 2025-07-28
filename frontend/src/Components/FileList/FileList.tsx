import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { GetFilesInFolder } from '../../../wailsjs/go/app/App';
import {
  Container,
  Header,
  HeaderTitle,
  ToolbarContainer,
  ToolbarActions,
  ExportButton,
  ExportZipButton,
  DeleteButton,
  ExportFileIcon,
  ExportIcon,
  DeleteIcon,
  TableContainer,
  Table,
  TableHeader,
  TableBody,
  HeaderRow,
  TableRow,
  CheckboxCell,
  NameCell,
  SizeCell,
  DateCell,
  NameHeader,
  SizeHeader,
  DateHeader,
  FileIcon,
  FileName,
  Checkbox,
  LoadingMessage,
  ErrorMessage,
  NoItemsMessage
} from '../../styles/TableStyles';

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
    return date.toLocaleDateString('en-US', {
      day: 'numeric',
      month: 'short',
      year: 'numeric'
    });
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
  const [selectedFiles, setSelectedFiles] = useState<Set<number>>(new Set());

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

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setSelectedFiles(new Set());
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleFileClick = (fileId: number, event: React.MouseEvent) => {
    if (selectedFiles.has(fileId) && selectedFiles.size === 1) {
      setSelectedFiles(new Set());
    } else {
      setSelectedFiles(new Set([fileId]));
    }
  };

  const handleCheckboxChange = (fileId: number, checked: boolean) => {
    const newSelected = new Set(selectedFiles);
    if (checked) {
      newSelected.add(fileId);
    } else {
      newSelected.delete(fileId);
    }
    setSelectedFiles(newSelected);
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedFiles(new Set(files.map(f => f.id)));
    } else {
      setSelectedFiles(new Set());
    }
  };

  const isAllSelected = files.length > 0 && selectedFiles.size === files.length;
  const isIndeterminate = selectedFiles.size > 0 && selectedFiles.size < files.length;

  if (loading) {
    return (
      <Container>
        <Header>
          <HeaderTitle>Loading...</HeaderTitle>
        </Header>
        <LoadingMessage>Loading files...</LoadingMessage>
      </Container>
    );
  }

  if (error) {
    return (
      <Container>
        <Header>
          <HeaderTitle>Error</HeaderTitle>
        </Header>
        <ErrorMessage>{error}</ErrorMessage>
      </Container>
    );
  }

  if (!files || files.length === 0) {
    return (
      <Container>
        <Header>
          <HeaderTitle>Received &gt; {folderInfo?.name || 'Folder'}</HeaderTitle>
        </Header>
        <NoItemsMessage>
          No files found in this folder.
        </NoItemsMessage>
      </Container>
    );
  }

  return (
    <Container>
      <Header>
        <HeaderTitle>Received &gt; {folderInfo?.name || 'Folder'}</HeaderTitle>
      </Header>
      
      <ToolbarContainer $isVisible={selectedFiles.size > 0}>
        <ToolbarActions>
          <ExportButton>
            <ExportFileIcon />
            EXPORT
          </ExportButton>
          {selectedFiles.size > 1 && (
            <ExportZipButton>
              <ExportIcon />
              EXPORT AS ZIP
            </ExportZipButton>
          )}
          <DeleteButton>
            <DeleteIcon />
            DELETE
          </DeleteButton>
        </ToolbarActions>
      </ToolbarContainer>
      
      <TableContainer>
        <Table>
          <TableHeader>
            <HeaderRow>
              <CheckboxCell>
                <Checkbox
                  type="checkbox"
                  checked={isAllSelected}
                  ref={(input) => {
                    if (input) input.indeterminate = isIndeterminate;
                  }}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                />
              </CheckboxCell>
              <NameHeader>Name</NameHeader>
              <SizeHeader>File size</SizeHeader>
              <DateHeader>Date received</DateHeader>
            </HeaderRow>
          </TableHeader>
          <TableBody>
            {files.map((file) => (
              <TableRow
                key={file.id}
                $isSelected={selectedFiles.has(file.id)}
                onClick={(e) => handleFileClick(file.id, e)}
              >
                <CheckboxCell>
                  <Checkbox
                    type="checkbox"
                    checked={selectedFiles.has(file.id)}
                    onChange={(e) => {
                      e.stopPropagation();
                      handleCheckboxChange(file.id, e.target.checked);
                    }}
                  />
                </CheckboxCell>
                <NameCell>
                  <FileIcon>{getFileIcon(file.mimeType)}</FileIcon>
                  <FileName>{file.name}</FileName>
                </NameCell>
                <SizeCell>{formatFileSize(file.size)}</SizeCell>
                <DateCell>{formatTimestamp(file.timestamp)}</DateCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Container>
  );
}

export default FileList;