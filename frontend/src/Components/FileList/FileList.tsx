import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { GetFilesInFolder, ExportFiles, ExportZipFolders, DeleteFiles } from '../../../wailsjs/go/app/App';
import { Dialog } from '../Dialog/Dialog';
import { LoadingDialog } from '../Dialog/LoadingDialog';
import { SuccessToast } from '../Toast/SuccessToast';
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
  
  const folderId = propFolderId || (paramFolderId ? parseInt(paramFolderId, 10) : null);
  
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [folderInfo, setFolderInfo] = useState<FolderInfo | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedFiles, setSelectedFiles] = useState<Set<number>>(new Set());

  // Export dialog states
  const [showExportDialog, setShowExportDialog] = useState<boolean>(false);
  const [showExportZipDialog, setShowExportZipDialog] = useState<boolean>(false);
  const [showExportLoading, setShowExportLoading] = useState<boolean>(false);

  const [showDeleteDialog, setShowDeleteDialog] = useState<boolean>(false);
  const [showDeleteLoading, setShowDeleteLoading] = useState<boolean>(false);

  // Toast states
  const [showSuccessToast, setShowSuccessToast] = useState<boolean>(false);
  const [successMessage, setSuccessMessage] = useState<string>('');
  const [isExporting, setIsExporting] = useState<boolean>(false);
  const [isDeleting, setIsDeleting] = useState<boolean>(false);

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

  const handleExportClick = () => {
    if (selectedFiles.size === 0) return;
    setShowExportDialog(true);
  };

  const handleExportZipClick = () => {
    if (selectedFiles.size <= 1) return;
    setShowExportZipDialog(true);
  };

  const handleExportZipConfirm = async () => {
    if (selectedFiles.size <= 1 || !folderId) return;
    
    setIsExporting(true);
    setShowExportZipDialog(false);
    setShowExportLoading(true);
    
    try {
      const fileIds = Array.from(selectedFiles);
      
      const exportPaths = await ExportZipFolders([folderId], fileIds);
      
      setSuccessMessage(`ZIP file created successfully: ${exportPaths[0]}`);
      
      setSelectedFiles(new Set());
      setShowSuccessToast(true);
      
    } catch (error) {
      console.error('ZIP export failed:', error);
      setSuccessMessage('ZIP export failed. Please try again.');
      setShowSuccessToast(true);
    } finally {
      setIsExporting(false);
      setShowExportLoading(false);
    }
  };

  const handleExportConfirm = async () => {
    if (selectedFiles.size === 0) return;
    
    setIsExporting(true);
    
    setShowExportDialog(false);
    setShowExportZipDialog(false);
    setShowExportLoading(true);
    
    try {
      const fileIds = Array.from(selectedFiles);
      
      const exportPaths = await ExportFiles(fileIds);
      
      if (fileIds.length === 1) {
        setSuccessMessage(`File exported successfully to: ${exportPaths[0]}`);
      } else {
        const exportDir = exportPaths[0].substring(0, exportPaths[0].lastIndexOf('/'));
        setSuccessMessage(`${exportPaths.length} files exported successfully to: ${exportDir}`);
      }
      
      setSelectedFiles(new Set());
      setShowSuccessToast(true);
      
    } catch (error) {
      console.error('Export failed:', error);
      setSuccessMessage('Export failed. Please try again.');
      setShowSuccessToast(true);
    } finally {
      setIsExporting(false);
      setShowExportLoading(false);
    }
  };

  const handleDialogCancel = () => {
    setShowExportDialog(false);
    setShowExportZipDialog(false);
    setShowDeleteDialog(false);
  };

  const handleLoadingCancel = () => {
    if (isExporting) {
      setShowExportLoading(false);
      setIsExporting(false);
    }
    if (isDeleting) {
      setShowDeleteLoading(false);
      setIsDeleting(false);
    }
  };

  const handleDeleteClick = () => {
    if (selectedFiles.size === 0) return;
    setShowDeleteDialog(true);
  };

  const handleDeleteConfirm = async () => {
    if (selectedFiles.size === 0) return;
    
    setIsDeleting(true);
    setShowDeleteDialog(false);
    setShowDeleteLoading(true);
    
    try {
      const fileIds = Array.from(selectedFiles);
      
      // Call the delete API
      await DeleteFiles(fileIds);
      
      // Remove deleted files from local state
      setFiles(prevFiles => prevFiles.filter(file => !selectedFiles.has(file.id)));
      
      // Show success message
      if (fileIds.length === 1) {
        setSuccessMessage('File deleted successfully');
      } else {
        setSuccessMessage(`${fileIds.length} files deleted successfully`);
      }
      
      setSelectedFiles(new Set());
      setShowSuccessToast(true);
      
    } catch (error) {
      console.error('Delete failed:', error);
      setSuccessMessage('Delete failed. Please try again.');
      setShowSuccessToast(true);
    } finally {
      setIsDeleting(false);
      setShowDeleteLoading(false);
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
          <ExportButton onClick={handleExportClick}>
            <ExportFileIcon />
            EXPORT
          </ExportButton>
          {selectedFiles.size > 1 && (
            <ExportZipButton onClick={handleExportZipClick}>
              <ExportIcon />
              EXPORT AS ZIP
            </ExportZipButton>
          )}
          <DeleteButton onClick={handleDeleteClick}>
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

      <Dialog
        isOpen={showExportDialog}
        onClose={handleDialogCancel}
        onConfirm={handleExportConfirm}
        confirmButtonText="EXPORT"
        title={selectedFiles.size === 1 ? "Export file?" : `Export ${selectedFiles.size} files?`}
      >
        <p>
          Exporting {selectedFiles.size === 1 ? 'this file' : `these ${selectedFiles.size} files`} will create {selectedFiles.size === 1 ? 'a copy' : 'copies'} that {selectedFiles.size === 1 ? 'is' : 'are'} accessible, 
          unencrypted, outside of Tella.
        </p>
        <p>
          Remember that for now, it is not possible to re-import files 
          from your computer into Tella Desktop.
        </p>
      </Dialog>

      <Dialog
        isOpen={showExportZipDialog}
        onClose={handleDialogCancel}
        onConfirm={handleExportZipConfirm}
        confirmButtonText="EXPORT"
        title="Export files as ZIP?"
      >
        <p>
          Exporting these {selectedFiles.size} files will create a ZIP archive 
          that is accessible, unencrypted, outside of Tella.
        </p>
        <p>
          Remember that for now, it is not possible to re-import files 
          from your computer into Tella Desktop.
        </p>
      </Dialog>

      <Dialog
        isOpen={showDeleteDialog}
        onClose={handleDialogCancel}
        onConfirm={handleDeleteConfirm}
        title={selectedFiles.size === 1 ? "Delete file?" : `Delete ${selectedFiles.size} files?`}
        confirmButtonText="DELETE"
      >
        <p>
          Deleting {selectedFiles.size === 1 ? 'this file' : `these ${selectedFiles.size} files`} will delete {selectedFiles.size === 1 ? 'it' : 'them'} from 
          Tella Desktop and the system. This action is permanent and cannot be reversed.
        </p>
      </Dialog>

      <LoadingDialog
        isOpen={showExportLoading}
        onCancel={handleDialogCancel}
        title="Your files are exporting"
        message="Please wait while your files are exporting. Do not close Tella Desktop or the export may fail."
      />

      <LoadingDialog
        isOpen={showDeleteLoading}
        onCancel={handleLoadingCancel}
        title="Deleting files"
        message="Please wait while your files are being permanently deleted. Do not close Tella Desktop or the deletion may fail."
      />

      <SuccessToast
        isVisible={showSuccessToast}
        message={successMessage}
        onClose={() => setShowSuccessToast(false)}
      />
    </Container>
  );
}

export default FileList;