package instance

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const socketName = "fusion-core.sock"

func getSocketPath() string {
	tmpDir := os.TempDir()
	return filepath.Join(tmpDir, socketName)
}

// TryConnect attempts to connect to an existing instance and send the nxm URL
func TryConnect(nxmURL string) bool {
	socketPath := getSocketPath()
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return false
	}
	defer conn.Close()

	if nxmURL != "" {
		conn.Write([]byte(nxmURL))
	}
	return true
}

// StartServer starts listening for new instance connections
func StartServer(onNxmURL func(string)) error {
	socketPath := getSocketPath()
	
	// Remove existing socket if it exists
	os.Remove(socketPath)
	
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket: %v", err)
	}

	go func() {
		defer listener.Close()
		defer os.Remove(socketPath)
		
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				n, err := c.Read(buf)
				if err != nil || n == 0 {
					return
				}
				
				nxmURL := string(buf[:n])
				if nxmURL != "" {
					onNxmURL(nxmURL)
				}
			}(conn)
		}
	}()
	
	return nil
}