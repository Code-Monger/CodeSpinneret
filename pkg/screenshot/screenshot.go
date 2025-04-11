package screenshot

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleScreenshot is the handler function for the screenshot tool
func HandleScreenshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract area
	area, _ := arguments["area"].(string)
	if area == "" {
		area = "full" // Default to full screen
	}

	// Extract window title (for window area)
	windowTitle, _ := arguments["window_title"].(string)

	// Extract region coordinates (for region area)
	var x, y, width, height float64
	if xFloat, ok := arguments["x"].(float64); ok {
		x = xFloat
	}
	if yFloat, ok := arguments["y"].(float64); ok {
		y = yFloat
	}
	if widthFloat, ok := arguments["width"].(float64); ok {
		width = widthFloat
	}
	if heightFloat, ok := arguments["height"].(float64); ok {
		height = heightFloat
	}

	// Extract format
	format, _ := arguments["format"].(string)
	if format == "" {
		format = "png" // Default to PNG
	}

	// Extract output path
	outputPath, _ := arguments["output_path"].(string)
	if outputPath == "" {
		// Generate a temporary file path
		tempDir := os.TempDir()
		timestamp := time.Now().Format("20060102-150405")
		outputPath = filepath.Join(tempDir, fmt.Sprintf("screenshot-%s.%s", timestamp, format))
	}

	// Take the screenshot
	screenshotPath, err := takeScreenshot(area, windowTitle, int(x), int(y), int(width), int(height), format, outputPath)
	if err != nil {
		return nil, fmt.Errorf("error taking screenshot: %v", err)
	}

	// Read the screenshot file
	screenshotData, err := os.ReadFile(screenshotPath)
	if err != nil {
		return nil, fmt.Errorf("error reading screenshot file: %v", err)
	}

	// Encode the screenshot as base64
	base64Data := base64.StdEncoding.EncodeToString(screenshotData)

	// Determine the MIME type
	mimeType := "image/png"
	if format == "jpg" || format == "jpeg" {
		mimeType = "image/jpeg"
	}

	// Create the result
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Screenshot taken successfully and saved to %s", screenshotPath),
			},
			mcp.ImageContent{
				Type:     "image",
				MIMEType: mimeType,
				Data:     base64Data,
			},
		},
	}

	return result, nil
}

// takeScreenshot takes a screenshot based on the specified parameters
func takeScreenshot(area, windowTitle string, x, y, width, height int, format, outputPath string) (string, error) {
	// Create the command to take the screenshot
	cmd, err := createScreenshotCommand(area, windowTitle, x, y, width, height, format, outputPath)
	if err != nil {
		return "", err
	}

	// Execute the command
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing screenshot command: %v", err)
	}

	return outputPath, nil
}

// createScreenshotCommand creates the command to take a screenshot based on the OS and available tools
func createScreenshotCommand(area, windowTitle string, x, y, width, height int, format, outputPath string) (*exec.Cmd, error) {
	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory: %v", err)
	}

	// Create the command based on the OS
	switch runtime.GOOS {
	case "windows":
		return createWindowsScreenshotCommand(area, windowTitle, x, y, width, height, format, outputPath)
	case "darwin":
		return createMacScreenshotCommand(area, windowTitle, x, y, width, height, format, outputPath)
	case "linux":
		return createLinuxScreenshotCommand(area, windowTitle, x, y, width, height, format, outputPath)
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// createWindowsScreenshotCommand creates a screenshot command for Windows
func createWindowsScreenshotCommand(area, windowTitle string, x, y, width, height int, format, outputPath string) (*exec.Cmd, error) {
	// Check if PowerShell is available
	if _, err := exec.LookPath("powershell.exe"); err == nil {
		// PowerShell script to take a screenshot
		script := ""
		switch area {
		case "full":
			script = `
				Add-Type -AssemblyName System.Windows.Forms
				Add-Type -AssemblyName System.Drawing
				$screen = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
				$bitmap = New-Object System.Drawing.Bitmap $screen.Width, $screen.Height
				$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
				$graphics.CopyFromScreen($screen.X, $screen.Y, 0, 0, $screen.Size)
				$bitmap.Save("` + outputPath + `")
				$graphics.Dispose()
				$bitmap.Dispose()
			`
		case "window":
			// This is a simplification; in practice, you'd need to find the window by title
			return nil, fmt.Errorf("capturing specific window by title not implemented on Windows")
		case "region":
			script = fmt.Sprintf(`
				Add-Type -AssemblyName System.Windows.Forms
				Add-Type -AssemblyName System.Drawing
				$bitmap = New-Object System.Drawing.Bitmap %d, %d
				$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
				$graphics.CopyFromScreen(%d, %d, 0, 0, $bitmap.Size)
				$bitmap.Save("%s")
				$graphics.Dispose()
				$bitmap.Dispose()
			`, width, height, x, y, outputPath)
		}
		return exec.Command("powershell.exe", "-Command", script), nil
	}

	return nil, fmt.Errorf("no suitable screenshot tool found on Windows")
}

// createMacScreenshotCommand creates a screenshot command for macOS
func createMacScreenshotCommand(area, windowTitle string, x, y, width, height int, format, outputPath string) (*exec.Cmd, error) {
	switch area {
	case "full":
		return exec.Command("screencapture", "-t", format, outputPath), nil
	case "window":
		if windowTitle == "" {
			// Capture active window
			return exec.Command("screencapture", "-t", format, "-W", outputPath), nil
		}
		// This is a simplification; in practice, you'd need to find the window by title
		return nil, fmt.Errorf("capturing specific window by title not implemented on macOS")
	case "region":
		return exec.Command("screencapture", "-t", format, "-R", fmt.Sprintf("%d,%d,%d,%d", x, y, width, height), outputPath), nil
	}

	return nil, fmt.Errorf("invalid area: %s", area)
}

// createLinuxScreenshotCommand creates a screenshot command for Linux
func createLinuxScreenshotCommand(area, windowTitle string, x, y, width, height int, format, outputPath string) (*exec.Cmd, error) {
	// Check if import (ImageMagick) is available
	if _, err := exec.LookPath("import"); err == nil {
		switch area {
		case "full":
			return exec.Command("import", "-window", "root", outputPath), nil
		case "window":
			if windowTitle == "" {
				// Capture active window
				return exec.Command("import", "-window", "$(xdotool getactivewindow)", outputPath), nil
			}
			// This is a simplification; in practice, you'd need to find the window ID by title
			return nil, fmt.Errorf("capturing specific window by title not implemented")
		case "region":
			geometry := fmt.Sprintf("%dx%d+%d+%d", width, height, x, y)
			return exec.Command("import", "-window", "root", "-crop", geometry, outputPath), nil
		}
	}

	// Check if scrot is available
	if _, err := exec.LookPath("scrot"); err == nil {
		switch area {
		case "full":
			return exec.Command("scrot", outputPath), nil
		case "window":
			if windowTitle == "" {
				// Capture active window
				return exec.Command("scrot", "-u", outputPath), nil
			}
			// This is a simplification; in practice, you'd need to find the window by title
			return nil, fmt.Errorf("capturing specific window by title not implemented")
		case "region":
			// scrot doesn't support direct region capture, would need to use other tools
			return nil, fmt.Errorf("region capture not implemented with scrot")
		}
	}

	// Check if gnome-screenshot is available
	if _, err := exec.LookPath("gnome-screenshot"); err == nil {
		switch area {
		case "full":
			return exec.Command("gnome-screenshot", "-f", outputPath), nil
		case "window":
			return exec.Command("gnome-screenshot", "-w", "-f", outputPath), nil
		case "region":
			// gnome-screenshot doesn't support direct region capture via command line
			return nil, fmt.Errorf("region capture not implemented with gnome-screenshot")
		}
	}

	return nil, fmt.Errorf("no suitable screenshot tool found on Linux")
}

// RegisterScreenshot registers the screenshot tool with the MCP server
func RegisterScreenshot(mcpServer *server.MCPServer) {
	// Create the tool definition
	screenshotTool := mcp.NewTool("screenshot",
		mcp.WithDescription("Take screenshots of the screen, windows, or specific regions"),
		mcp.WithString("area",
			mcp.Description("Area to capture: 'full' (entire screen), 'window' (specific window), or 'region' (specific region)"),
		),
		mcp.WithString("window_title",
			mcp.Description("Window title to capture (for 'window' area)"),
		),
		mcp.WithNumber("x",
			mcp.Description("X coordinate of the region to capture (for 'region' area)"),
		),
		mcp.WithNumber("y",
			mcp.Description("Y coordinate of the region to capture (for 'region' area)"),
		),
		mcp.WithNumber("width",
			mcp.Description("Width of the region to capture (for 'region' area)"),
		),
		mcp.WithNumber("height",
			mcp.Description("Height of the region to capture (for 'region' area)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (png, jpg)"),
		),
		mcp.WithString("output_path",
			mcp.Description("Path to save the screenshot (optional)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("screenshot", HandleScreenshot)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(screenshotTool, wrappedHandler)

	// Log the registration
	log.Printf("[Screenshot] Registered screenshot tool")
}
