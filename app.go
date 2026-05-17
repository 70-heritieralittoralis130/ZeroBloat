package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// AppInfo represents information about an app
type AppInfo struct {
	PackageName string `json:"packageName"`
	AppName     string `json:"appName"`
	Description string `json:"description"`
	IsBloat     bool   `json:"isBloat"`
	IsSystem    bool   `json:"isSystem"`
}

// BloatPackage represents a known bloatware package
type BloatPackage struct {
	Package     string `json:"package"`
	AppName     string `json:"app_name"`
	Description string `json:"description"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.pollDevices()
}

func (a *App) pollDevices() {
	var lastDevices []string
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
			devices, _ := a.GetDevices()
			if !slicesEqual(devices, lastDevices) {
				runtime.EventsEmit(a.ctx, "devices_updated", devices)
				lastDevices = devices
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CheckAdb checks if adb is installed
func (a *App) CheckAdb() bool {
	_, err := exec.LookPath("adb")
	return err == nil
}

// GetDevices returns a list of connected device serials
func (a *App) GetDevices() ([]string, error) {
	cmd := execCommand("adb", "devices")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	devices := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "device" {
			devices = append(devices, parts[0])
		}
	}
	return devices, nil
}

// GetDeviceInfo returns details about a device
func (a *App) GetDeviceInfo(serial string) (map[string]string, error) {
	info := make(map[string]string)
	
	// Model
	cmd := execCommand("adb", "-s", serial, "shell", "getprop", "ro.product.model")
	output, _ := cmd.Output()
	info["model"] = strings.TrimSpace(string(output))

	// Android Version
	cmd = execCommand("adb", "-s", serial, "shell", "getprop", "ro.build.version.release")
	output, _ = cmd.Output()
	info["version"] = strings.TrimSpace(string(output))

	// Manufacturer
	cmd = execCommand("adb", "-s", serial, "shell", "getprop", "ro.product.manufacturer")
	output, _ = cmd.Output()
	info["manufacturer"] = strings.TrimSpace(string(output))

	return info, nil
}

// Known packages map
var knownPackages = map[string]string{
	"com.facebook.katana": "Facebook",
	"com.facebook.system": "Facebook System",
	"com.facebook.appmanager": "Facebook App Manager",
	"com.facebook.services": "Facebook Services",
	"com.samsung.android.bixby.agent": "Bixby Voice",
	"com.samsung.android.bixby.wakeup": "Bixby Wakeup",
	"com.samsung.android.app.spage": "Bixby Home",
	"com.sec.android.app.sbrowser": "Samsung Internet",
	"com.samsung.android.messaging": "Samsung Messaging",
	"com.miui.analytics": "MIUI Analytics",
	"com.miui.hybrid": "Quick Apps",
	"com.miui.msa.global": "MSA (Ads)",
	"com.miui.securitycenter": "Security",
	"com.google.android.youtube": "YouTube",
	"com.google.android.apps.photos": "Google Photos",
	"com.android.chrome": "Chrome",
	"com.google.android.apps.maps": "Google Maps",
	"com.google.android.gm": "Gmail",
	"com.google.android.music": "YouTube Music",
}

// GetPackages returns a list of packages on the device
func (a *App) GetPackages(serial string) ([]AppInfo, error) {
	cmd := execCommand("adb", "-s", serial, "shell", "pm", "list", "packages")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	apps := []AppInfo{}
	
	// Get manufacturer to check bloat
	info, _ := a.GetDeviceInfo(serial)
	manufacturer := info["manufacturer"]
	bloatList := a.GetBloatwareList(manufacturer)

	// Get third-party packages
	cmd3 := execCommand("adb", "-s", serial, "shell", "pm", "list", "packages", "-3")
	output3, _ := cmd3.Output()
	thirdPartyList := strings.Split(string(output3), "\n")
	isThirdParty := make(map[string]bool)
	for _, l := range thirdPartyList {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "package:") {
			isThirdParty[strings.TrimPrefix(l, "package:")] = true
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package:") {
			pkg := strings.TrimPrefix(line, "package:")
			
			appName := pkg
			if name, ok := knownPackages[pkg]; ok {
				appName = name
			} else {
				// Try to make a readable name from package
				parts := strings.Split(pkg, ".")
				if len(parts) >= 2 {
					appName = strings.Title(parts[len(parts)-1])
				}
			}

			isBloat := false
			description := ""
			for _, b := range bloatList {
				if b.Package == pkg {
					isBloat = true
					description = b.Description
					if b.AppName != "" {
						appName = b.AppName
					}
					break
				}
			}

			apps = append(apps, AppInfo{
				PackageName: pkg,
				AppName:     appName,
				Description: description,
				IsBloat:     isBloat,
				IsSystem:    !isThirdParty[pkg],
			})
		}
	}
	return apps, nil
}

// UninstallPackage uninstalls a package for user 0
func (a *App) UninstallPackage(serial string, packageName string) error {
	cmd := execCommand("adb", "-s", serial, "shell", "pm", "uninstall", "--user", "0", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall: %s, output: %s", err, string(output))
	}
	return nil
}

// GetBloatwareList returns a list of known bloatware for a manufacturer
func (a *App) GetBloatwareList(manufacturer string) []BloatPackage {
	manufacturer = strings.ToLower(manufacturer)

	xiaomiPocoList := []BloatPackage{
		{
			Package:     "com.miui.analytics",
			AppName:     "MIUI Analytics",
			Description: "Tracks usage data and serves background telemetry. Highly recommended to remove.",
		},
		{
			Package:     "com.miui.msa.global",
			AppName:     "MIUI System Ads",
			Description: "Responsible for serving advertisements across system applications.",
		},
		{
			Package:     "com.mi.globalbrowser",
			AppName:     "Mi Browser",
			Description: "Stock browser, famous for push notifications and bloat.",
		},
		{
			Package:     "com.miui.videoplayer",
			AppName:     "Mi Video",
			Description: "Stock video player bundled with online streaming ads.",
		},
		{
			Package:     "com.miui.player",
			AppName:     "Mi Music",
			Description: "Stock audio player featuring heavy ad integration.",
		},
		{
			Package:     "com.xiaomi.glance.internet",
			AppName:     "Glance for Mi",
			Description: "Lock screen wallpaper carousel that consumes extensive battery and data.",
		},
		{
			Package:     "com.xiaomi.midrop",
			AppName:     "ShareMe",
			Description: "Xiaomi's file-sharing utility. Safe to replace with Google Quick Share.",
		},
		{
			Package:     "com.miui.hybrid",
			AppName:     "Quick Apps",
			Description: "App framework used to run instant applications without installation.",
		},
	}

	oppoRealmeList := []BloatPackage{
		{
			Package:     "com.heytap.browser",
			AppName:     "HeyTap Browser",
			Description: "Stock ColorOS/RealmeUI browser app.",
		},
		{
			Package:     "com.heytap.habit",
			AppName:     "HeyTap Habit",
			Description: "Wellness tracker native to ColorOS platforms.",
		},
		{
			Package:     "com.heytap.market",
			AppName:     "App Market",
			Description: "Secondary app store. Safe to drop if using Google Play Store.",
		},
		{
			Package:     "com.coloros.gamespace",
			AppName:     "Game Space",
			Description: "Oppo's in-game dashboard utility.",
		},
		{
			Package:     "com.coloros.video",
			AppName:     "ColorOS Video",
			Description: "Default localized video player application.",
		},
		{
			Package:     "com.nearme.atlas",
			AppName:     "NearMe Atlas",
			Description: "Telemetry and system metrics service.",
		},
		{
			Package:     "com.nearme.statistics.rom",
			AppName:     "NearMe Statistics",
			Description: "Background analytics engine for Oppo/Realme devices.",
		},
	}

	bloatware := map[string][]BloatPackage{
		"samsung": {
			{
				Package:     "com.samsung.android.bixby.agent",
				AppName:     "Bixby Voice",
				Description: "Samsung's voice assistant. Safe to remove if you use Google Assistant or no assistant.",
			},
			{
				Package:     "com.samsung.android.kidshome",
				AppName:     "Samsung Kids",
				Description: "Sandbox environment for children. Safe to remove if not needed.",
			},
			{
				Package:     "com.sec.android.app.sbrowser",
				AppName:     "Samsung Internet Browser",
				Description: "Stock Samsung web browser. Safe to remove if using Chrome, Firefox, etc.",
			},
			{
				Package:     "com.samsung.android.app.spage",
				AppName:     "Bixby Home / Samsung Daily",
				Description: "The left-most home screen feed.",
			},
			{
				Package:     "com.samsung.android.messaging",
				AppName:     "Samsung Messages",
				Description: "Stock SMS client. Safe if Google Messages or another SMS app is installed.",
			},
			{
				Package:     "com.samsung.android.spay",
				AppName:     "Samsung Pay",
				Description: "Samsung's payment framework. Safe if you use Google Wallet or don't use NFC payments.",
			},
			{
				Package:     "com.samsung.android.app.watchmanager",
				AppName:     "Galaxy Wearable",
				Description: "Manages Samsung smartwatches/buds. Safe if you don't own Samsung wearables.",
			},
		},
		"xiaomi": xiaomiPocoList,
		"poco":   xiaomiPocoList,
		"oppo":   oppoRealmeList,
		"realme": oppoRealmeList,
		"vivo": {
			{
				Package:     "com.vivo.browser",
				AppName:     "Vivo Browser",
				Description: "Default FuntouchOS browser engine.",
			},
			{
				Package:     "com.vivo.appstore",
				AppName:     "V-Appstore",
				Description: "Vivo's independent alternative app repository.",
			},
			{
				Package:     "com.vivo.gamecube",
				AppName:     "Ultra Game Mode",
				Description: "Gaming performance overlay suite.",
			},
			{
				Package:     "com.bbk.theme",
				AppName:     "iTheme",
				Description: "Vivo theme marketplace application.",
			},
			{
				Package:     "com.vivo.globalsearch",
				AppName:     "Global Search",
				Description: "On-device search engine launcher shortcut.",
			},
			{
				Package:     "com.bbk.cloud",
				AppName:     "vivoCloud",
				Description: "Proprietary cloud backup services.",
			},
		},
		"motorola": {
			{
				Package:     "com.motorola.moto",
				AppName:     "Moto App",
				Description: "The central hub for Moto gestures. Only remove if you explicitly do not use chop-chop flashlight or twist camera triggers.",
			},
			{
				Package:     "com.motorola.gamemode",
				AppName:     "Moto Gametime",
				Description: "In-game sidebar control module.",
			},
			{
				Package:     "com.motorola.userhub",
				AppName:     "Motorola Notifications",
				Description: "Sends promotional offers and feedback surveys.",
			},
			{
				Package:     "com.motorola.motoitv",
				AppName:     "Ready For",
				Description: "Desktop environment projection suite. Safe to remove if you do not dock to monitors/TVs.",
			},
		},
	}

	commonBloat := []BloatPackage{
		{
			Package:     "com.facebook.appmanager",
			AppName:     "Facebook App Manager",
			Description: "Silently updates Facebook ecosystem services in the background.",
		},
		{
			Package:     "com.facebook.services",
			AppName:     "Facebook Services",
			Description: "Background system daemon utilized by Facebook infrastructure.",
		},
		{
			Package:     "com.facebook.system",
			AppName:     "Facebook System Installer",
			Description: "Allows pre-loaded Facebook services to run as system elements.",
		},
		{
			Package:     "com.netflix.mediaclient",
			AppName:     "Netflix",
			Description: "Pre-installed streaming consumer client.",
		},
		{
			Package:     "com.netflix.partner.activation",
			AppName:     "Netflix Partner Activation",
			Description: "Account provisioning hook for system-level activation.",
		},
	}

	result := commonBloat
	if list, ok := bloatware[manufacturer]; ok {
		result = append(result, list...)
	}

	return result
}

// GetAppIcon fetches the app icon URL from Google Play Store
func (a *App) GetAppIcon(packageName string) string {
	url := fmt.Sprintf("https://play.google.com/store/apps/details?id=%s", packageName)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	html := string(body)
	searchStr := `property="og:image" content="`
	idx := strings.Index(html, searchStr)
	if idx == -1 {
		return ""
	}

	start := idx + len(searchStr)
	end := strings.Index(html[start:], `"`)
	if end == -1 {
		return ""
	}

	iconUrl := html[start : start+end]
	return iconUrl
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// execCommand wraps exec.Command to allow platform-specific configuration
func execCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	prepareCmd(cmd)
	return cmd
}
