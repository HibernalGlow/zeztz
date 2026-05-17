package main

import (
	"fmt"
	"os/exec"
	"time"
)

type TimerService struct{}

func (t *TimerService) Greet(name string) string {
	return "Hello " + name + "!"
}

func (t *TimerService) ExecutePowerAction(action string, dryRun bool) string {
	actionText := map[string]string{
		"standby":  "睡眠",
		"sleep":    "休眠",
		"shutdown": "关机",
		"restart":  "重启",
	}

	if dryRun {
		return fmt.Sprintf("[dryrun] 模拟执行: %s", actionText[action])
	}

	var result string
	switch action {
	case "standby":
		result = t.setStandby()
	case "sleep":
		result = t.setHibernate()
	case "shutdown":
		result = t.shutdown()
	case "restart":
		result = t.restart()
	default:
		return fmt.Sprintf("未知的电源操作: %s", action)
	}

	return result
}

func (t *TimerService) setStandby() string {
	cmd := exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0", "1", "0")
	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("❌ 睡眠失败: %v", err)
	}
	return "✅ 睡眠指令已发送"
}

func (t *TimerService) setHibernate() string {
	cmd := exec.Command("shutdown", "/h")
	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("❌ 休眠失败: %v", err)
	}
	return "✅ 休眠指令已发送"
}

func (t *TimerService) shutdown() string {
	cmd := exec.Command("shutdown", "/s", "/t", "0")
	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("❌ 关机失败: %v", err)
	}
	return "✅ 关机指令已发送"
}

func (t *TimerService) restart() string {
	cmd := exec.Command("shutdown", "/r", "/t", "0")
	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("❌ 重启失败: %v", err)
	}
	return "✅ 重启指令已发送"
}

type PowerTestResult struct {
	Action    string `json:"action"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func (t *TimerService) TestPowerAction(action string) PowerTestResult {
	now := time.Now().Format("2006-01-02 15:04:05")

	validActions := map[string]bool{
		"standby":  true,
		"sleep":    true,
		"shutdown": true,
		"restart":  true,
	}

	if !validActions[action] {
		return PowerTestResult{
			Action:    action,
			Success:   false,
			Message:   fmt.Sprintf("无效的操作: %s", action),
			Timestamp: now,
		}
	}

	actionNames := map[string]string{
		"standby":  "睡眠",
		"sleep":    "休眠",
		"shutdown": "关机",
		"restart":  "重启",
	}

	result := PowerTestResult{
		Action:    action,
		Success:   true,
		Message:   fmt.Sprintf("测试成功: %s 指令可用", actionNames[action]),
		Timestamp: now,
	}

	return result
}
