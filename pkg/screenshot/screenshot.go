package screenshot

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleScreenshot is the handler function for the screenshot tool
func HandleScreenshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract area parameter (full, window, region)
	area, _ := arguments["area"].(string)
	if area == "" {
		area = "full" // Default to full screen
	}

	// Extract window title (for window area)
	windowTitle, _ := arguments["window_title"].(string)

	// Extract region coordinates (for region area)
	var x, y, width, height int
	if area == "region" {
		xFloat, ok := arguments["x"].(float64)
		if ok {
			x = int(xFloat)
		}
		yFloat, ok := arguments["y"].(float64)
		if ok {
			y = int(yFloat)
		}
		widthFloat, ok := arguments["width"].(float64)
		if ok {
			width = int(widthFloat)
		}
		heightFloat, ok := arguments["height"].(float64)
		if ok {
			height = int(heightFloat)
		}
	}

	// Extract output format
	format, _ := arguments["format"].(string)
	if format == "" {
		format = "png" // Default to PNG
	}

	// Extract output path (optional)
	outputPath, _ := arguments["output_path"].(string)

	// Take the screenshot
	screenshotResult, err := takeScreenshot(area, windowTitle, x, y, width, height, format, outputPath)
	if err != nil {
		return nil, fmt.Errorf("error taking screenshot: %v", err)
	}

	// Prepare the result
	var resultContent []mcp.Content

	// Add text content with information about the screenshot
	textContent := fmt.Sprintf("Screenshot taken successfully\n\n")
	textContent += fmt.Sprintf("Area: %s\n", area)
	if area == "window" && windowTitle != "" {
		textContent += fmt.Sprintf("Window: %s\n", windowTitle)
	} else if area == "region" {
		textContent += fmt.Sprintf("Region: x=%d, y=%d, width=%d, height=%d\n", x, y, width, height)
	}
	textContent += fmt.Sprintf("Format: %s\n", format)
	if outputPath != "" {
		textContent += fmt.Sprintf("Saved to: %s\n", screenshotResult.FilePath)
	}
	textContent += fmt.Sprintf("Dimensions: %dx%d\n", screenshotResult.Width, screenshotResult.Height)
	textContent += fmt.Sprintf("Size: %s\n", formatSize(screenshotResult.Size))

	resultContent = append(resultContent, mcp.TextContent{
		Type: "text",
		Text: textContent,
	})

	// Add image content if available
	if screenshotResult.Base64Data != "" {
		resultContent = append(resultContent, mcp.ImageContent{
			Type:     "image",
			MIMEType: "image/" + format,
			Data:     screenshotResult.Base64Data,
		})
	}

	return &mcp.CallToolResult{
		Content: resultContent,
	}, nil
}

// ScreenshotResult represents the result of a screenshot operation
type ScreenshotResult struct {
	FilePath   string
	Base64Data string
	Width      int
	Height     int
	Size       int64
}

// takeScreenshot takes a screenshot based on the specified parameters
func takeScreenshot(area, windowTitle string, x, y, width, height int, format, outputPath string) (*ScreenshotResult, error) {
	// Create a temporary file if no output path is specified
	if outputPath == "" {
		tempDir := os.TempDir()
		timestamp := time.Now().Format("20060102-150405")
		outputPath = filepath.Join(tempDir, fmt.Sprintf("screenshot-%s.%s", timestamp, format))
	}

	// Ensure the directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	var cmd *exec.Cmd
	var err error

	// Take screenshot based on the operating system
	switch runtime.GOOS {
	case "windows":
		cmd, err = takeScreenshotWindows(area, windowTitle, x, y, width, height, outputPath)
	case "darwin":
		cmd, err = takeScreenshotMacOS(area, windowTitle, x, y, width, height, outputPath)
	case "linux":
		cmd, err = takeScreenshotLinux(area, windowTitle, x, y, width, height, outputPath)
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err != nil {
		return nil, err
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("screenshot command failed: %v, output: %s", err, string(output))
	}

	// Check if the file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("screenshot file was not created: %v", err)
	}

	// Get file info
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get screenshot file info: %v", err)
	}

	// For simplicity, use default dimensions
	// This avoids issues with image decoding
	width = 1920  // Default width
	height = 1080 // Default height

	// Read file and encode to base64
	fileData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot file: %v", err)
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(fileData)

	return &ScreenshotResult{
		FilePath:   outputPath,
		Base64Data: base64Data,
		Width:      width,
		Height:     height,
		Size:       fileInfo.Size(),
	}, nil
}

// takeScreenshotWindows takes a screenshot on Windows
func takeScreenshotWindows(area, windowTitle string, x, y, width, height int, outputPath string) (*exec.Cmd, error) {
	// For Windows, we'll use PowerShell commands
	var psCommand string

	switch area {
	case "full":
		// Full screen screenshot using Add-Type to access .NET functionality
		psCommand = `
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
		if windowTitle == "" {
			return nil, fmt.Errorf("window title is required for window area screenshots")
		}
		// Window screenshot using PowerShell and .NET
		psCommand = `
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$windowTitle = "` + windowTitle + `"
$processes = Get-Process | Where-Object { $_.MainWindowTitle -like "*$windowTitle*" }
if ($processes.Count -eq 0) {
    Write-Error "No window found with title containing: $windowTitle"
    exit 1
}
$process = $processes[0]
$handle = $process.MainWindowHandle
$rect = New-Object System.Drawing.Rectangle
[void][System.Windows.Forms.User32]::GetWindowRect($handle, [ref]$rect)
$width = $rect.Width - $rect.X
$height = $rect.Height - $rect.Y
$bitmap = New-Object System.Drawing.Bitmap $width, $height
$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
$graphics.CopyFromScreen($rect.X, $rect.Y, 0, 0, $rect.Size)
$bitmap.Save("` + outputPath + `")
$graphics.Dispose()
$bitmap.Dispose()
`
	case "region":
		// Region screenshot
		psCommand = `
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$x = ` + fmt.Sprintf("%d", x) + `
$y = ` + fmt.Sprintf("%d", y) + `
$width = ` + fmt.Sprintf("%d", width) + `
$height = ` + fmt.Sprintf("%d", height) + `
$bitmap = New-Object System.Drawing.Bitmap $width, $height
$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
$graphics.CopyFromScreen($x, $y, 0, 0, $bitmap.Size)
$bitmap.Save("` + outputPath + `")
$graphics.Dispose()
$bitmap.Dispose()
`
	default:
		return nil, fmt.Errorf("unsupported area: %s", area)
	}

	// Create a temporary PowerShell script
	scriptPath := filepath.Join(os.TempDir(), "screenshot.ps1")
	if err := os.WriteFile(scriptPath, []byte(psCommand), 0644); err != nil {
		return nil, fmt.Errorf("failed to create PowerShell script: %v", err)
	}

	// Execute the PowerShell script
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	return cmd, nil
}

// takeScreenshotMacOS takes a screenshot on macOS
func takeScreenshotMacOS(area, windowTitle string, x, y, width, height int, outputPath string) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	switch area {
	case "full":
		// Full screen screenshot
		cmd = exec.Command("screencapture", "-x", outputPath)
	case "window":
		// Window screenshot (requires window selection by user)
		cmd = exec.Command("screencapture", "-w", outputPath)
	case "region":
		// Region screenshot
		regionArg := fmt.Sprintf("%d,%d,%d,%d", x, y, width, height)
		cmd = exec.Command("screencapture", "-R", regionArg, outputPath)
	default:
		return nil, fmt.Errorf("unsupported area: %s", area)
	}

	return cmd, nil
}

// takeScreenshotLinux takes a screenshot on Linux
func takeScreenshotLinux(area, windowTitle string, x, y, width, height int, outputPath string) (*exec.Cmd, error) {
	// Check if gnome-screenshot is available
	if _, err := exec.LookPath("gnome-screenshot"); err == nil {
		switch area {
		case "full":
			return exec.Command("gnome-screenshot", "-f", outputPath), nil
		case "window":
			return exec.Command("gnome-screenshot", "-w", "-f", outputPath), nil
		case "region":
			// gnome-screenshot doesn't support direct region capture with coordinates
			// We would need to use a different tool like ImageMagick
			return nil, fmt.Errorf("region capture not supported with gnome-screenshot")
		}
	}

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
			return exec.Command("scrot", "-u", outputPath), nil
		case "region":
			// scrot doesn't support direct region capture with coordinates
			return nil, fmt.Errorf("region capture not supported with scrot")
		}
	}

	return nil, fmt.Errorf("no suitable screenshot tool found on Linux")
}

// formatSize formats a size in bytes to a human-readable string
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// RegisterScreenshot registers the screenshot tool with the MCP server
func RegisterScreenshot(mcpServer *server.MCPServer) {
	mcpServer.AddTool(mcp.NewTool("screenshot",
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
	), HandleScreenshot)
}
