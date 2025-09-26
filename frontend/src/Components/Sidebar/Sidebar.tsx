import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import tellaIcon from '../../assets/images/icons/tella-icon.svg'
import nearbySharingIcon from '../../assets/images/icons/nearby-sharing-icon.svg'
import receivedIcon from '../../assets/images/icons/received-icon.svg'
import lockIcon from '../../assets/images/icons/lock-icon.svg'
import { useServerState, useServerActions } from '../../Contexts/ServerContext';
import { Dialog } from '../Dialog/Dialog';
import { LockApp } from '../../../wailsjs/go/app/App';

interface SidebarProps {
  className?: string;
  onLock?: () => void;
}

export function Sidebar({ className, onLock }: SidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();

  const { isRunning: serverRunning } = useServerState()
  const { stopServer } = useServerActions()

  const [showExitDialog, setShowExitDialog] = useState(false);
  const [pendingNavigation, setPendingNavigation] = useState<string | null>(null);
  const [isExiting, setIsExiting] = useState(false);

  const isOnNearbySharing = location.pathname === '/nearby-sharing';
  const shouldGuard = isOnNearbySharing && serverRunning;

  const handleNavigation = (path: string) => {
    if (shouldGuard && path !== '/nearby-sharing') {
      setPendingNavigation(path);
      setShowExitDialog(true);
    } else {
      navigate(path);
    }
  };

  const isActive = (path: string) => {
    if (path === '/' && (location.pathname === '/' || location.pathname.startsWith('/folder'))) return true;
    if (path !== '/' && location.pathname.startsWith(path)) return true;
    return false;
  };

  const handleLock = async () => {
    try {
      await LockApp();
      if (onLock) {
        onLock();
      }
    } catch (error) {
      console.error('Failed to lock app:', error);
    }
  }

  const handleExitConfirm = async () => {
    if (!pendingNavigation) return;

    try {
      setIsExiting(true);
      
      // Stop the server
      await stopServer();    
      navigate(pendingNavigation);  
    } catch (error) {
      console.error('Error stopping server:', error);
    } finally {
      setIsExiting(false);
      setShowExitDialog(false);
      setPendingNavigation(null);
    }
  };

  const handleExitCancel = () => {
    setShowExitDialog(false);
    setPendingNavigation(null);
  };
  return (
    <SidebarContainer className={className}>
      <SidebarHeader>
        <Icon icon='tella' />
      </SidebarHeader>
      
      <Navigation>
        <NavItem 
          $isActive={isActive('/')} 
          onClick={() => handleNavigation('/')}
        >
          <Icon icon='received' />
          <NavText $isActive={isActive('/')} >Received</NavText>
        </NavItem>
        
        <NavItem 
          $isActive={isActive('/nearby-sharing')} 
          onClick={() => handleNavigation('/nearby-sharing')}
        >
          <Icon icon='nearbySharing' />
          <NavText $isActive={isActive('/nearby-sharing')} >Nearby Sharing</NavText>
        </NavItem>
      </Navigation>

      <LockButton onClick={handleLock}>
        <Icon icon='lock' />
        <LockText>Lock</LockText>
      </LockButton>

      <Dialog 
        isOpen={showExitDialog}
        onClose={handleExitCancel}
        onConfirm={handleExitConfirm}
        title='Exit Nearby Sharing?'
        cancelButtonText='CONTINUE NEARBY SHARING'
        confirmButtonText='EXIT' 
      >
        <p>If you exit Nearby Sharing, you will have to restart the process from the beginning. </p>
      </Dialog>
    </SidebarContainer>
  );
}

const SidebarContainer = styled.div`
  width: 250px;
  border-right: 1px solid #e9ecef;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  padding: 0 1rem;
`;

const SidebarHeader = styled.div`
  padding: 2rem 1.5rem 1.5rem;
`;

const iconMap = {
  tella: tellaIcon,
  nearbySharing: nearbySharingIcon,
  received: receivedIcon,
  lock: lockIcon,
} as const;

type IconName = keyof typeof iconMap;

const Icon = styled.div<{ 
  icon: IconName;
  size?: string;
}>`
  width: ${({ size }) => size || '1.5rem'};
  height: ${({ size }) => size || '1.5rem'};
  flex-shrink: 0;
  background-image: url(${({ icon }) => iconMap[icon]});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
`;


const Navigation = styled.nav`
  padding: 1rem 0;
  flex: 1;
`;

const NavItem = styled.div<{ $isActive: boolean }>`
  display: flex;
  align-items: center;
  padding: 1rem;
  gap: 0.8rem;
  cursor: pointer;
  transition: all 0.2s ease;
  border-radius: 0;
  margin: 0 0.75rem;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  
  background-color: ${({ $isActive }) => $isActive ? '#E9F2FF' : 'transparent'};
  color: ${({ $isActive }) => $isActive ? '#065485' : '#404040'};
`;

const NavText = styled.span<{ $isActive: boolean }>`
  font-size: 1rem;
  font-weight: 700;
  border-bottom:  ${({ $isActive }) => $isActive ? '2px solid #065485' : 'none'};
`;

const LockButton = styled.button`
  display: flex;
  align-items: center;
  padding: 1rem;
  gap: 0.8rem;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 2px solid #CFCFCF;
  background: transparent;
  margin: 0 0.75rem 1rem;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  color: #595959;
  
  &:hover {
    background-color: #f8f9fa;
  }
`;

const LockText = styled.span`
  font-size: 1rem;
  font-weight: 700;
`;