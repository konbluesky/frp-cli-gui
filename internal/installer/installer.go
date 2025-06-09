package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Installer FRP 安装管理器
type Installer struct {
	installDir string
	version    string
	baseURL    string
}

// InstallStatus 安装状态
type InstallStatus struct {
	IsInstalled   bool
	Version       string
	FrpsPath      string
	FrpcPath      string
	InstallDir    string
	NeedsUpdate   bool
	LatestVersion string
}

// NewInstaller 创建新的安装管理器
func NewInstaller(installDir string) *Installer {
	if installDir == "" {
		// 默认安装到用户目录下的 .frp-manager 文件夹，与配置目录保持一致
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			installDir = filepath.Join(homeDir, ".frp-manager")
		} else {
			installDir = ".frp-manager"
		}
	}

	return &Installer{
		installDir: installDir,
		version:    "0.52.3", // 当前稳定版本
		baseURL:    "https://github.com/fatedier/frp/releases/download",
	}
}

// CheckInstallation 检查 FRP 安装状态
func (i *Installer) CheckInstallation() (*InstallStatus, error) {
	status := &InstallStatus{
		InstallDir: i.installDir,
	}

	// 检查安装目录是否存在
	if _, err := os.Stat(i.installDir); os.IsNotExist(err) {
		return status, nil
	}

	// 检查 frps 和 frpc 是否存在
	frpsPath := filepath.Join(i.installDir, "frps")
	frpcPath := filepath.Join(i.installDir, "frpc")

	if runtime.GOOS == "windows" {
		frpsPath += ".exe"
		frpcPath += ".exe"
	}

	frpsExists := i.fileExists(frpsPath)
	frpcExists := i.fileExists(frpcPath)

	if frpsExists && frpcExists {
		status.IsInstalled = true
		status.FrpsPath = frpsPath
		status.FrpcPath = frpcPath

		// 尝试获取版本信息
		if version, err := i.getInstalledVersion(frpsPath); err == nil {
			status.Version = version
			// 检查是否需要更新
			status.NeedsUpdate = i.needsUpdate(version)
			status.LatestVersion = i.version
		}
	}

	return status, nil
}

// InstallFRP 安装 FRP
func (i *Installer) InstallFRP() error {
	// 创建安装目录
	if err := os.MkdirAll(i.installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 获取下载 URL
	downloadURL, filename, err := i.getDownloadURL()
	if err != nil {
		return fmt.Errorf("获取下载链接失败: %w", err)
	}

	// 下载文件
	tempFile := filepath.Join(os.TempDir(), filename)
	if err := i.downloadFile(downloadURL, tempFile); err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer os.Remove(tempFile)

	// 解压文件
	if err := i.extractFile(tempFile, i.installDir); err != nil {
		return fmt.Errorf("解压文件失败: %w", err)
	}

	// 设置执行权限 (Unix 系统)
	if runtime.GOOS != "windows" {
		frpsPath := filepath.Join(i.installDir, "frps")
		frpcPath := filepath.Join(i.installDir, "frpc")

		os.Chmod(frpsPath, 0755)
		os.Chmod(frpcPath, 0755)
	}

	return nil
}

// getDownloadURL 获取下载链接
func (i *Installer) getDownloadURL() (string, string, error) {
	var arch, osName, ext string

	// 确定操作系统
	switch runtime.GOOS {
	case "linux":
		osName = "linux"
		ext = "tar.gz"
	case "darwin":
		osName = "darwin"
		ext = "tar.gz"
	case "windows":
		osName = "windows"
		ext = "zip"
	default:
		return "", "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	// 确定架构
	switch runtime.GOARCH {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "386"
	case "arm":
		arch = "arm"
	default:
		return "", "", fmt.Errorf("不支持的架构: %s", runtime.GOARCH)
	}

	filename := fmt.Sprintf("frp_%s_%s_%s.%s", i.version, osName, arch, ext)
	url := fmt.Sprintf("%s/v%s/%s", i.baseURL, i.version, filename)

	return url, filename, nil
}

// downloadFile 下载文件
func (i *Installer) downloadFile(url, filepath string) error {
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Minute, // 30分钟超时
	}

	// 发送请求
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	// 复制数据
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// extractFile 解压文件
func (i *Installer) extractFile(src, dest string) error {
	if strings.HasSuffix(src, ".tar.gz") {
		return i.extractTarGz(src, dest)
	} else if strings.HasSuffix(src, ".zip") {
		return i.extractZip(src, dest)
	}
	return fmt.Errorf("不支持的文件格式")
}

// extractTarGz 解压 tar.gz 文件
func (i *Installer) extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 只提取 frps 和 frpc 文件
		filename := filepath.Base(header.Name)
		if filename != "frps" && filename != "frpc" {
			continue
		}

		target := filepath.Join(dest, filename)

		switch header.Typeflag {
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}

	return nil
}

// extractZip 解压 zip 文件
func (i *Installer) extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// 只提取 frps.exe 和 frpc.exe 文件
		filename := filepath.Base(f.Name)
		if filename != "frps.exe" && filename != "frpc.exe" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		target := filepath.Join(dest, filename)
		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

// fileExists 检查文件是否存在
func (i *Installer) fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// getInstalledVersion 获取已安装的版本
func (i *Installer) getInstalledVersion(execPath string) (string, error) {
	// 这里可以通过执行 frps --version 来获取版本
	// 为了简化，我们返回配置的版本
	return i.version, nil
}

// needsUpdate 检查是否需要更新
func (i *Installer) needsUpdate(currentVersion string) bool {
	// 简单的版本比较，实际应该使用语义化版本比较
	return currentVersion != i.version
}

// GetInstallDir 获取安装目录
func (i *Installer) GetInstallDir() string {
	return i.installDir
}

// SetVersion 设置要安装的版本
func (i *Installer) SetVersion(version string) {
	i.version = version
}

// GetVersion 获取当前设置的版本
func (i *Installer) GetVersion() string {
	return i.version
}

// Uninstall 卸载 FRP
func (i *Installer) Uninstall() error {
	if _, err := os.Stat(i.installDir); os.IsNotExist(err) {
		return nil // 已经不存在
	}

	return os.RemoveAll(i.installDir)
}

// UpdateFRP 更新 FRP
func (i *Installer) UpdateFRP() error {
	// 备份当前安装
	backupDir := i.installDir + ".backup"
	if err := os.Rename(i.installDir, backupDir); err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}

	// 尝试安装新版本
	if err := i.InstallFRP(); err != nil {
		// 安装失败，恢复备份
		os.RemoveAll(i.installDir)
		os.Rename(backupDir, i.installDir)
		return fmt.Errorf("更新失败: %w", err)
	}

	// 删除备份
	os.RemoveAll(backupDir)
	return nil
}
