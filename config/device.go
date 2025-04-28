package config

import (
	"fmt"
	"math/rand"
	"sync"
)

// DeviceInfo 存储设备相关信息
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
	defaultMaxUses = 3 // 默认使用3次后更新
)

// 生成新的设备信息
func generateNewDevice() *DeviceInfo {
	return &DeviceInfo{
		DeviceCPU:   "AMD", // 固定为 AMD
		DeviceID:    generateDeviceID(),
		MachineID:   generateMachineID(),
		DeviceBrand: generateDeviceBrand(),
		DeviceType:  "windows", // 固定为 windows
		SystemType:  "Windows", // 固定为 Windows
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
	brands := []string{"92L3", "91C9", "814S", "8P15V", "35G4", "65G4", "55G4"}
	return brands[rand.Intn(len(brands))]
}

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
