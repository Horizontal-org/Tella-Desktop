import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { GetStoredFolders, ExportZipFolders, DeleteFolders } from '../../../wailsjs/go/app/App';
import {
  Container,
  Header,
  HeaderTitle,
  ToolbarContainer,
  ToolbarActions,
  ExportButton,
  DeleteButton,
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
  FilesCell,
  DateCell,
  NameHeader,
  FilesHeader,
  DateHeader,
  FolderIcon,
  FolderName,
  Checkbox,
  NoItemsMessage,
} from '../../styles/TableStyles';
import { Dialog } from '../Dialog/Dialog';
import { LoadingDialog } from '../Dialog/LoadingDialog';
import { SuccessToast } from '../Toast/SuccessToast';

interface FolderInfo {
  id: number
  name: string
  timestamp: string
  fileCount: number
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
    return timestamp
  }
};

export function FolderList() {
  const navigate = useNavigate();
  const [folders, setFolders] = useState<FolderInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedFolders, setSelectedFolders] = useState<Set<number>>(new Set());

  // Export dialog states
  const [showExportDialog, setShowExportDialog] = useState<boolean>(false);
  const [showExportLoading, setShowExportLoading] = useState<boolean>(false);
  const [showSuccessToast, setShowSuccessToast] = useState<boolean>(false);
  const [successMessage, setSuccessMessage] = useState<string>('');
  const [isExporting, setIsExporting] = useState<boolean>(false);

  const [showDeleteDialog, setShowDeleteDialog] = useState<boolean>(false);
  const [showDeleteLoading, setShowDeleteLoading] = useState<boolean>(false);
  const [isDeleting, setIsDeleting] = useState<boolean>(false);
  
  const fetchFolders = async () => {
    try {
      setLoading(true);
      setError(null);
      const foldersData = await GetStoredFolders();
      setFolders(foldersData || []); // Handle null case explicitly
    } catch (err) {
      console.error('Failed to fetch folders:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFolders();
  }, []);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setSelectedFolders(new Set());
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleFolderClick = (folderId: number, event: React.MouseEvent) => {
    if (selectedFolders.has(folderId) && selectedFolders.size === 1) {
      setSelectedFolders(new Set());
    } else {
      setSelectedFolders(new Set([folderId]));
    }
  };

  const handleFolderDoubleClick = (folderId: number) => {
    navigate(`/folder/${folderId}`);
  };

  const handleCheckboxChange = (folderId: number, checked: boolean) => {
    const newSelected = new Set(selectedFolders);
    if (checked) {
      newSelected.add(folderId);
    } else {
      newSelected.delete(folderId);
    }
    setSelectedFolders(newSelected);
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedFolders(new Set(folders.map(f => f.id)));
    } else {
      setSelectedFolders(new Set());
    }
  };

  const handleExportClick = () => {
    if (selectedFolders.size === 0) return;
    setShowExportDialog(true);
  };

  const handleExportConfirm = async () => {
    if (selectedFolders.size === 0) return;
    
    setIsExporting(true);
    setShowExportDialog(false);
    setShowExportLoading(true);
    
    try {
      const folderIds = Array.from(selectedFolders);
      
      // Export entire folders as ZIP (empty selectedFileIDs array)
      const exportPaths = await ExportZipFolders(folderIds, []);
      
      if (folderIds.length === 1) {
        setSuccessMessage(`Folder exported as ZIP: ${exportPaths[0]}`);
      } else {
        const exportDir = exportPaths[0].substring(0, exportPaths[0].lastIndexOf('/'));
        setSuccessMessage(`${exportPaths.length} folders exported as ZIP files to: ${exportDir}`);
      }
      
      // Reset selection and show success
      setSelectedFolders(new Set());
      setShowSuccessToast(true);
      
    } catch (error) {
      console.error('Folder ZIP export failed:', error);
      setSuccessMessage('Folder export failed. Please try again.');
      setShowSuccessToast(true);
    } finally {
      setIsExporting(false);
      setShowExportLoading(false);
    }
  };

  const handleExportCancel = () => {
    if (isExporting) {
      setShowExportLoading(false);
      setIsExporting(false);
    } else {
      setShowExportDialog(false);
    }
  };

  const handleDeleteClick = () => {
    if (selectedFolders.size === 0) return;
    setShowDeleteDialog(true);
  };

  const handleDeleteConfirm = async () => {
    if (selectedFolders.size === 0) return;
    
    setIsDeleting(true);
    setShowDeleteDialog(false);
    setShowDeleteLoading(true);
    
    try {
      const folderIds = Array.from(selectedFolders);
      
      // Call the delete API
      await DeleteFolders(folderIds);
      
      // Remove deleted folders from local state
      setFolders(prevFolders => prevFolders.filter(folder => !selectedFolders.has(folder.id)));
      
      // Show success message
      if (folderIds.length === 1) {
        setSuccessMessage('Folder deleted successfully');
      } else {
        setSuccessMessage(`${folderIds.length} folders deleted successfully`);
      }
      
      setSelectedFolders(new Set());
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

  // Update handleExportCancel to handle delete loading too
  const handleDialogCancel = () => {
    if (isExporting) {
      setShowExportLoading(false);
      setIsExporting(false);
    }
    if (isDeleting) {
      setShowDeleteLoading(false);
      setIsDeleting(false);
    }
    setShowExportDialog(false);
    setShowDeleteDialog(false);
  };


  const isAllSelected = folders.length > 0 && selectedFolders.size === folders.length;
  const isIndeterminate = selectedFolders.size > 0 && selectedFolders.size < folders.length;

  if (loading) {
    return <Container>Loading folders...</Container>;
  }

  if (error) {
    return <Container>{error}</Container>;
  }
  
  if (!folders || folders.length === 0) {
    return (
      <Container>
        <NoItemsMessage>
          This is where the files you receive will be displayed. Click on “Nearby Sharing” to get started.
        </NoItemsMessage>
      </Container>
    );
  }

  return (
    <Container>
      <Header>
        <HeaderTitle>Received</HeaderTitle>
      </Header>
      
      <ToolbarContainer $isVisible={selectedFolders.size > 0}>
        <ToolbarActions>
          <ExportButton onClick={handleExportClick}>
            <ExportIcon/>
            EXPORT AS ZIP
          </ExportButton>
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
              <FilesHeader>Files</FilesHeader>
              <DateHeader>Date received</DateHeader>
            </HeaderRow>
          </TableHeader>
          <TableBody>
            {folders.map((folder) => (
              <TableRow
                key={folder.id}
                $isSelected={selectedFolders.has(folder.id)}
                onClick={(e) => handleFolderClick(folder.id, e)}
                onDoubleClick={() => handleFolderDoubleClick(folder.id)}
              >
                <CheckboxCell>
                  <Checkbox
                    type="checkbox"
                    checked={selectedFolders.has(folder.id)}
                    onChange={(e) => {
                      e.stopPropagation();
                      handleCheckboxChange(folder.id, e.target.checked);
                    }}
                  />
                </CheckboxCell>
                <NameCell>
                  <FolderIcon />
                  <FolderName>{folder.name}</FolderName>
                </NameCell>
                <FilesCell>{folder.fileCount} files</FilesCell>
                <DateCell>{formatTimestamp(folder.timestamp)}</DateCell>
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
        title={selectedFolders.size === 1 ? "Export folder as ZIP?" : `Export ${selectedFolders.size} folders as ZIP?`}
      >
        <p>
          Exporting {selectedFolders.size === 1 ? 'this folder' : `these ${selectedFolders.size} folders`} will create ZIP {selectedFolders.size === 1 ? 'archive' : 'archives'} that {selectedFolders.size === 1 ? 'is' : 'are'} accessible, 
          unencrypted, outside of Tella.
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
        title={selectedFolders.size === 1 ? "Delete folder?" : `Delete ${selectedFolders.size} folders?`}
        confirmButtonText="DELETE"
      >
        <p>
          Deleting {selectedFolders.size === 1 ? 'this folder' : `these ${selectedFolders.size} folders`} will permanently delete {selectedFolders.size === 1 ? 'it' : 'them'} and all files inside from 
          Tella Desktop. This action cannot be reversed.
        </p>
      </Dialog>

      <LoadingDialog
        isOpen={showExportLoading}
        onCancel={handleDialogCancel}
        title="Your folders are exporting"
        message="Please wait while your folders are being exported as ZIP files. Do not close Tella Desktop or the export may fail."
      />

      <LoadingDialog
        isOpen={showDeleteLoading}
        onCancel={handleDialogCancel}
        title="Deleting folders"
        message="Please wait while your folders and files are being permanently deleted. Do not close Tella Desktop or the deletion may fail."
      />

      <SuccessToast
        isVisible={showSuccessToast}
        message={successMessage}
        onClose={() => setShowSuccessToast(false)}
      />
    </Container>
  );
}

export default FolderList;