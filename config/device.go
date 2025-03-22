package config

import (
	"fmt"
	"math/rand"
	"sync"
)

type DeviceInfo struct {
	DeviceCPU     string
	DeviceID      string
	MachineID     string
	DeviceBrand   string
	DeviceType    string
	OSVersion     string
	SystemType    string
	WorkspacePath string
	UseCount      int
	MaxUses       int
}

var (
	currentDevice  *DeviceInfo
	deviceMutex    sync.RWMutex
	defaultMaxUses = 8 // 默认使用8次后更新
)

// GetCurrentDevice 获取当前设备信息，如果超过使用次数限制则生成新的
func GetCurrentDevice() *DeviceInfo {
	deviceMutex.Lock()
	defer deviceMutex.Unlock()

	if currentDevice == nil || currentDevice.UseCount >= currentDevice.MaxUses {
		// 生成新的设备信息
		currentDevice = generateNewDevice()
	}

	// 增加使用计数
	currentDevice.UseCount++
	return currentDevice
}

// 生成新的设备信息
func generateNewDevice() *DeviceInfo {
	deviceType := generateDeviceType()
	return &DeviceInfo{
		DeviceCPU:   generateCPUBrand(),
		DeviceID:    generateDeviceID(),
		MachineID:   generateMachineID(),
		DeviceBrand: generateDeviceBrand(),
		DeviceType:  deviceType,
		OSVersion:   generateOSVersion(),
		SystemType:  getSystemType(deviceType),
		UseCount:    0,
		MaxUses:     defaultMaxUses + rand.Intn(3), // 8-10次之间随机
	}
}

// 生成随机设备ID
func generateDeviceID() string {
	return fmt.Sprintf("%d", rand.Int63())
}

// 生成随机机器ID (64位十六进制)
func generateMachineID() string {
	bytes := make([]byte, 32)
	for i := range bytes {
		bytes[i] = byte(rand.Intn(16))
	}
	return fmt.Sprintf("%x", bytes)
}

// 生成随机设备品牌
func generateDeviceBrand() string {
	brands := []string{"82L5", "X1C9", "T14S", "P15V", "E15G4"}
	return brands[rand.Intn(len(brands))]
}

// 生成随机CPU品牌
func generateCPUBrand() string {
	cpus := []string{"AMD", "Intel", "ARM"}
	return cpus[rand.Intn(len(cpus))]
}

// 生成随机操作系统版本
func generateOSVersion() string {
	versions := []string{
		"Microsoft Windows 11 专业版",
		"Microsoft Windows 10 企业版",
		"macOS 14.3.1",
		"macOS 15.2.1",
	}
	return versions[rand.Intn(len(versions))]
}

// 生成随机设备类型
func generateDeviceType() string {
	types := []string{"windows", "linux", "macos"}
	return types[rand.Intn(len(types))]
}

// 获取系统类型
func getSystemType(deviceType string) string {
	switch deviceType {
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	case "macos":
		return "macOS"
	default:
		return "Windows"
	}
}
