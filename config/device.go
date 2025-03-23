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
	defaultMaxUses = 5 // 默认使用5次后更新
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
		OSVersion:   generateOSVersion(deviceType),
		SystemType:  getSystemType(deviceType),
		UseCount:    0,
		MaxUses:     defaultMaxUses + rand.Intn(3),
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
func generateOSVersion(deviceType string) string {
	switch deviceType {
	case "windows":
		versions := []string{
			"Microsoft Windows 11 专业版",
			"Microsoft Windows 11 企业版",
			"Microsoft Windows 10 专业版",
			"Microsoft Windows 10 企业版",
		}
		return versions[rand.Intn(len(versions))]
	case "macos":
		versions := []string{
			"macOS 14.3.1",
			"macOS 14.2.1",
			"macOS 13.6.4",
			"macOS 13.5.2",
		}
		return versions[rand.Intn(len(versions))]
	case "linux":
		versions := []string{
			"Ubuntu 22.04 LTS",
			"Ubuntu 20.04 LTS",
			"CentOS 7.9",
			"Debian 11",
		}
		return versions[rand.Intn(len(versions))]
	default:
		return "Microsoft Windows 10 专业版"
	}
}

// 生成随机设备类型
func generateDeviceType() string {
	// 增加 mac 和 linux 的概率
	weights := []struct {
		os     string
		weight int
	}{
		{"macos", 45},   // 45% 概率
		{"linux", 45},   // 45% 概率
		{"windows", 10}, // 10% 概率
	}

	// 计算总权重
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w.weight
	}

	// 生成随机数
	r := rand.Intn(totalWeight)

	// 根据权重选择系统类型
	current := 0
	for _, w := range weights {
		current += w.weight
		if r < current {
			return w.os
		}
	}

	return "linux" // 默认返回 linux
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
