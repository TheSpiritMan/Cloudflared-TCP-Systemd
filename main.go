package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	serviceName = "cloudflared-tcp"
	scriptPath  = "/usr/local/bin/cloudflared-tcp.sh"
	servicePath = "/etc/systemd/system/cloudflared-tcp.service"
	configPath  = "/etc/cloudflared-tcp.conf"
)

// get current user
func getUser() string {
	user := os.Getenv("USER")
	if user == "" {
		out, err := exec.Command("logname").Output()
		if err != nil {
			return "root"
		}
		user = strings.TrimSpace(string(out))
	}
	return user
}

// write file using sudo without printing file content
func writeFileWithSudo(path, content string) {
	cmd := exec.Command("sudo", "tee", path)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = nil       // suppress stdout
	cmd.Stderr = os.Stderr // show errors only
	cmd.Run()
}

// check if cloudflared is installed
func checkBinary() {
	_, err := exec.LookPath("cloudflared")
	if err != nil {
		fmt.Println("❌ cloudflared is not installed. Please install it first.")
		os.Exit(1)
	}
}

// setup config, script, and systemd service
func setup() {
	checkBinary()

	fmt.Println("👉 Creating config file...")
	config := `# Format:
# <hostname> <local_port>

postgres-db.example.com 15432
`
	writeFileWithSudo(configPath, config)
	fmt.Println("👉 Config file created")

	fmt.Println("👉 Creating runner script...")
	script := `#!/bin/bash

CONFIG_FILE="/etc/cloudflared-tcp.conf"

while read -r host port; do
  [[ "$host" =~ ^#.*$ || -z "$host" ]] && continue
  echo "Starting tunnel: $host -> localhost:$port"
  cloudflared access tcp --hostname "$host" --url "localhost:$port" &
done < "$CONFIG_FILE"

wait
`
	writeFileWithSudo(scriptPath, script)
	exec.Command("sudo", "chmod", "+x", scriptPath).Run()
	fmt.Println("👉 Runner script created")

	fmt.Println("👉 Creating systemd service...")
	user := getUser()
	service := fmt.Sprintf(`[Unit]
Description=Cloudflared TCP Access Tunnel
After=network.target

[Service]
Type=simple
ExecStart=%s
Restart=always
RestartSec=5
User=%s

[Install]
WantedBy=multi-user.target
`, scriptPath, user)
	writeFileWithSudo(servicePath, service)
	fmt.Println("👉 Systemd service file created")

	fmt.Println("👉 Reloading systemd...")
	exec.Command("sudo", "systemctl", "daemon-reexec").Run()
	exec.Command("sudo", "systemctl", "daemon-reload").Run()

	fmt.Println("👉 Enabling + starting service...")
	exec.Command("sudo", "systemctl", "enable", serviceName).Run()
	exec.Command("sudo", "systemctl", "start", serviceName).Run()

	fmt.Println("✅ Setup complete!")
}

// restart systemd service
func restartService() {
	fmt.Println("🔄 Restarting service...")
	exec.Command("sudo", "systemctl", "restart", serviceName).Run()
}

// show service status
func status() {
	cmd := exec.Command("sudo", "systemctl", "status", serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// show live logs
func logs() {
	cmd := exec.Command("journalctl", "-u", serviceName, "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// list forwarded services
func listServices() {
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Printf("⚠️ Config file not found: %s\n", configPath)
		return
	}
	defer file.Close()

	fmt.Printf("\n%-30s | %-10s\n", "HOSTNAME", "LOCAL PORT")
	fmt.Println("-----------------------------------------------")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 {
			fmt.Printf("%-30s | %-10s\n", parts[0], parts[1])
		}
	}
	fmt.Println("")
}

// cleanup service and files
func cleanup() {
	fmt.Println("⚠️ Stopping service...")
	exec.Command("sudo", "systemctl", "stop", serviceName).Run()

	fmt.Println("👉 Disabling service...")
	exec.Command("sudo", "systemctl", "disable", serviceName).Run()

	fmt.Println("🧹 Removing files...")
	exec.Command("sudo", "rm", "-f", servicePath).Run()
	exec.Command("sudo", "rm", "-f", scriptPath).Run()
	exec.Command("sudo", "rm", "-f", configPath).Run()

	fmt.Println("👉 Reloading systemd...")
	exec.Command("sudo", "systemctl", "daemon-reload").Run()
	fmt.Println("✅ Cleanup complete")
}

// show help
func help() {
	fmt.Println(`
Cloudflared TCP Manager CLI

Usage:
  setup       → Setup systemd service and start tunnels
  restart     → Restart the systemd service
  status      → Show systemd service status
  logs        → Show live logs of the service
  list        → List all forwarded services (local + remote ports)
  cleanup     → Stop and remove systemd service and scripts
  help        → Show this help message

Config file:
  /etc/cloudflared-tcp.conf
  Format: <hostname> <local_port>

Example:
  postgres-db.example.com 15432
  redis-db.example.com    6379
`)
}

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}
	switch os.Args[1] {
	case "setup":
		setup()
	case "restart":
		restartService()
	case "status":
		status()
	case "logs":
		logs()
	case "list":
		listServices()
	case "cleanup":
		cleanup()
	case "help", "-h", "--help":
		help()
	default:
		fmt.Println("Usage: cloudflared-tcp {setup|restart|status|logs|cleanup|help}")
	}
}