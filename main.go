package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/stianeikeland/go-rpio/v4"
)

const (
	GPIO_PIN  = 14
	PWM_MAX   = 1000
	T_MIN     = 45.0 // Градусы Цельсия
	T_FULL    = 65.0
	S_MIN     = 20 // Минимальные обороты %
	CHECK_INT = 3  // Интервал проверки (сек)
)

// Структура для хранения предыдущих состояний (для расчета скорости)
var (
	lastNetBytes  uint64
	lastDiskBytes uint64
)

func getSystemLoad() float64 {
	// Нагрузка на CPU (%)
	cpuPerc, _ := cpu.Percent(0, false)

	// Нагрузка на Сеть (wlan0 + eth0)
	netStats, _ := net.IOCounters(true)
	var currentNetBytes uint64
	for _, n := range netStats {
		if n.Name == "wlan0" || n.Name == "eth0" {
			currentNetBytes += n.BytesSent + n.BytesRecv
		}
	}
	netDiff := (currentNetBytes - lastNetBytes) / uint64(CHECK_INT)
	lastNetBytes = currentNetBytes

	// Нагрузка на Диск
	diskStats, _ := disk.IOCounters()
	var currentDiskBytes uint64
	for _, d := range diskStats {
		currentDiskBytes += d.ReadBytes + d.WriteBytes
	}
	diskDiff := (currentDiskBytes - lastDiskBytes) / uint64(CHECK_INT)
	lastDiskBytes = currentDiskBytes

	// Если трафик > 5MB/s или диск > 10MB/s, добавляем "вес" к нагрузке
	var loadWeight float64
	if netDiff > 5*1024*1024 {
		loadWeight += 5.0
	}
	if diskDiff > 10*1024*1024 {
		loadWeight += 5.0
	}
	if cpuPerc[0] > 70 {
		loadWeight += 10.0
	}

	return loadWeight
}

func getTemp() float64 {
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}
	tempInt, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return float64(tempInt) / 1000.0
}

func main() {
	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
	defer rpio.Close()

	pin := rpio.Pin(GPIO_PIN)
	pin.Pwm()
	pin.Freq(25000)

	fmt.Println("RPi Fan Controller started...")

	for {
		realTemp := getTemp()
		extraLoad := getSystemLoad()

		// Итоговая температура для расчета скорости
		effTemp := realTemp + extraLoad

		var speed int
		if effTemp < T_MIN {
			speed = 0
		} else if effTemp >= T_FULL {
			speed = 100
		} else {
			// Линейная интерполяция
			speed = int(float64(S_MIN) + (effTemp-T_MIN)*(100-float64(S_MIN))/(T_FULL-T_MIN))
		}

		duty := uint32((speed * PWM_MAX) / 100)
		pin.DutyCycle(duty, PWM_MAX)

		log.Printf("[INFO] Temp: %.1f°C  Fan: %d%%", realTemp, speed)

		time.Sleep(CHECK_INT * time.Second)
	}
}
