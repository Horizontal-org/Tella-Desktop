import { createContext, useContext, useState, useCallback, useRef, ReactNode } from 'react';
import { StartServer, StopServer, GetDefaultPort } from '../../wailsjs/go/app/App';
import { log } from "../util/util"

interface ServerContextValue {
  isRunning: boolean;
  isStarting: boolean;
  error: string | null;
  port: number;
  
  startServer: () => Promise<boolean>;
  stopServer: () => Promise<boolean>;
  clearError: () => void;

  getServerStatus: () => 'stopped' | 'starting' | 'running' | 'error';
}

const ServerContext = createContext<ServerContextValue | undefined>(undefined);

interface ServerProviderProps {
  children: ReactNode;
}

export function ServerProvider({ children }: ServerProviderProps) {
  const [isRunning, setIsRunning] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [serverPort, setServerPort] = useState(-1)
  
  const serverStateRef = useRef({
    isRunning: false,
    isStarting: false
  });

  const startServer = useCallback(async (): Promise<boolean> => {
    const defaultPort = await GetDefaultPort();
    setServerPort(defaultPort)
    if (serverStateRef.current.isRunning || serverStateRef.current.isStarting) {
      if (serverStateRef.current.isRunning) {
          log("Server is already running");
      } else if (serverStateRef.current.isStarting) {
          log("Server is starting up");
      }
      return serverStateRef.current.isRunning || serverStateRef.current.isStarting;
    }

    try {
      setIsStarting(true);
      setError(null);
      serverStateRef.current.isStarting = true;
      
      await StartServer(defaultPort);
      
      setIsRunning(true);
      serverStateRef.current.isRunning = true;
      serverStateRef.current.isStarting = false;
      
      return true;
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to start server';
      
      setError(errorMessage);
      serverStateRef.current.isRunning = false;
      serverStateRef.current.isStarting = false;
      
      return false;
    } finally {
      setIsStarting(false);
    }
  }, []);

  const stopServer = useCallback(async (): Promise<boolean> => {
    if (!serverStateRef.current.isRunning) {
      log('Server not running, nothing to stop');
      return true;
    }

    try {
      await StopServer();
      
      setIsRunning(false);
      setError(null);
      serverStateRef.current.isRunning = false;
      serverStateRef.current.isStarting = false;
      
      return true;
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to stop server';
      
      setError(errorMessage);
      return false;
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const getServerStatus = useCallback((): 'stopped' | 'starting' | 'running' | 'error' => {
    if (error) return 'error';
    if (isStarting) return 'starting';
    if (isRunning) return 'running';
    return 'stopped';
  }, [isRunning, isStarting, error]);

  const value: ServerContextValue = {
    isRunning,
    isStarting,
    error,
    port: serverPort,

    startServer,
    stopServer,
    clearError,
    
    getServerStatus
  };

  return (
    <ServerContext.Provider value={value}>
      {children}
    </ServerContext.Provider>
  );
}

export function useServer() {
  const context = useContext(ServerContext);
  if (context === undefined) {
    throw new Error('useServer must be used within a ServerProvider');
  }
  return context;
}

export function useServerState() {
  const { isRunning, isStarting, error, port, getServerStatus } = useServer();
  return { isRunning, isStarting, error, port, getServerStatus };
}

export function useServerActions() {
  const { startServer, stopServer, clearError } = useServer();
  return { startServer, stopServer, clearError };
}
