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
	T_MIN     = 45.0
	T_FULL    = 65.0
	S_MIN     = 20
	CHECK_INT = 3
	// Коэффициент сглаживания (от 0.0 до 1.0)
	// Чем меньше число, тем медленнее меняется скорость
	SMOOTHING = 0.2
)

var (
	lastNetBytes  uint64
	lastDiskBytes uint64
	currentSpeed  float64 // Текущая сглаженная скорость
)

func getTemp() float64 {
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}
	tempInt, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return float64(tempInt) / 1000.0
}

func getSystemLoad() float64 {
	cpuPerc, _ := cpu.Percent(0, false)

	netStats, _ := net.IOCounters(true)
	var currentNetBytes uint64
	for _, n := range netStats {
		if n.Name == "wlan0" || n.Name == "eth0" {
			currentNetBytes += n.BytesSent + n.BytesRecv
		}
	}
	netDiff := float64(currentNetBytes-lastNetBytes) / float64(CHECK_INT)
	lastNetBytes = currentNetBytes

	diskStats, _ := disk.IOCounters()
	var currentDiskBytes uint64
	for _, d := range diskStats {
		currentDiskBytes += d.ReadBytes + d.WriteBytes
	}
	diskDiff := float64(currentDiskBytes-lastDiskBytes) / float64(CHECK_INT)
	lastDiskBytes = currentDiskBytes

	var loadWeight float64
	if netDiff > 5*1024*1024 {
		loadWeight += 5.0
	}
	if diskDiff > 10*1024*1024 {
		loadWeight += 5.0
	}
	if len(cpuPerc) > 0 && cpuPerc[0] > 70 {
		loadWeight += 10.0
	}

	return loadWeight
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
		effTemp := realTemp + extraLoad

		// 1. Вычисляем "желаемую" скорость
		var targetSpeed float64
		if effTemp < T_MIN {
			targetSpeed = 0
		} else if effTemp >= T_FULL {
			targetSpeed = 100
		} else {
			targetSpeed = float64(S_MIN) + (effTemp-T_MIN)*(100-float64(S_MIN))/(T_FULL-T_MIN)
		}

		// 2. Экспоненциальное сглаживание
		// Новая скорость = Старая скорость + Сглаживание * (Целевая - Старая)
		currentSpeed = currentSpeed + SMOOTHING*(targetSpeed-currentSpeed)

		// 3. Установка ШИМ
		duty := uint32((currentSpeed * float64(PWM_MAX)) / 100)
		pin.DutyCycle(duty, PWM_MAX)

		log.Printf("[LOG] Temp: %.1f°C | Target: %.0f%% | Smooth Speed: %.1f%%",
			realTemp, targetSpeed, currentSpeed)

		time.Sleep(CHECK_INT * time.Second)
	}
}
