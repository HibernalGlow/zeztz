<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.ts";
  import { Checkbox } from "$lib/components/ui/checkbox/index.ts";
  import { Progress } from "$lib/components/ui/progress/index.ts";
  import { Input } from "$lib/components/ui/input/index.ts";
  import { Label } from "$lib/components/ui/label/index.ts";
  import DateTimePicker from "$lib/components/DateTimePicker.svelte";
  import {
    Play, Square, RotateCcw, Clock, Moon, Power, RotateCcw as RestartIcon,
    CircleCheck, CircleX, Copy, Check, Activity, Wifi, Cpu,
    Calendar, Timer, Gauge, Terminal, Zap, Sun
  } from "@lucide/svelte";

  type TimerMode = 'countdown' | 'specific_time' | 'netspeed' | 'cpu';
  type PowerMode = 'standby' | 'sleep' | 'shutdown' | 'restart';
  type Phase = 'idle' | 'running' | 'completed' | 'cancelled' | 'error';

  let timerMode: TimerMode = $state('countdown');
  let powerMode: PowerMode = $state('sleep');
  let hours = $state(0);
  let minutes = $state(0);
  let seconds = $state(5);
  let targetDatetime = $state((() => {
    const d = new Date(Date.now() + 3600000);
    return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}:00`;
  })());
  let uploadThreshold = $state(242);
  let downloadThreshold = $state(242);
  let netDuration = $state(2);
  let netTriggerMode: 'both' | 'any' = $state('both');
  let cpuThreshold = $state(10);
  let cpuDuration = $state(2);
  let dryrun = $state(true);
  let theme: 'light' | 'dark' = $state(document.documentElement.classList.contains('dark') ? 'dark' : 'light');

  let phase: Phase = $state('idle');
  let progress = $state(0);
  let progressText = $state('');
  let remainingTime = $state('00:00:00');
  let currentUpload = $state(0);
  let currentDownload = $state(0);
  let currentCpu = $state(0);
  let netHistory: {time: number, up: number, down: number}[] = $state([]);
  let cpuHistory: number[] = $state([]);
  let logs: string[] = $state([]);
  let copied = $state(false);
  let logsEl: HTMLDivElement | undefined = $state();

  $effect(() => {
    if (logsEl && logs.length > 0) {
      logsEl.scrollTop = logsEl.scrollHeight;
    }
  });

  let intervalId: ReturnType<typeof setInterval> | null = null;
  let totalSeconds = $state(0);

  let isRunning = $derived(phase === 'running');
  let canStart = $derived(phase === 'idle' || phase === 'error' || phase === 'cancelled' || phase === 'completed');

  function log(msg: string) {
    logs = [...logs.slice(-50), msg];
  }

  function setPreset(h: number, m: number, s: number) {
    hours = h; minutes = m; seconds = s;
  }

  function handleStart() {
    if (isRunning) return;
    phase = 'running';
    progress = 0;
    progressText = '启动中...';
    netHistory = [];
    cpuHistory = [];
    log(`⏰ 启动 ${timerMode} 模式`);

    if (timerMode === 'countdown') {
      startCountdown();
    } else if (timerMode === 'specific_time') {
      startSpecificTime();
    } else if (timerMode === 'netspeed') {
      startNetspeedMonitor();
    } else if (timerMode === 'cpu') {
      startCpuMonitor();
    }
  }

  function startCountdown() {
    totalSeconds = hours * 3600 + minutes * 60 + seconds;
    if (totalSeconds <= 0) {
      phase = 'error';
      log('❌ 倒计时时间必须大于0');
      return;
    }
    const startTime = Date.now();
    const endTime = startTime + totalSeconds * 1000;
    log(`⏰ 开始倒计时 ${hours}时${minutes}分${seconds}秒`);
    log(`⚡ 电源操作: ${powerMode}, dryrun: ${dryrun}`);

    intervalId = setInterval(() => {
      const now = Date.now();
      const remaining = Math.max(0, Math.round((endTime - now) / 1000));
      const elapsed = totalSeconds - remaining;
      const pct = totalSeconds > 0 ? Math.min(100, Math.round((elapsed / totalSeconds) * 100)) : 0;
      progress = pct;
      const h = Math.floor(remaining / 3600);
      const m = Math.floor((remaining % 3600) / 60);
      const s = remaining % 60;
      remainingTime = `${String(h).padStart(2,'0')}:${String(m).padStart(2,'0')}:${String(s).padStart(2,'0')}`;
      progressText = `剩余 ${remainingTime}`;

      if (remaining <= 0) {
        clearInterval(intervalId!);
        intervalId = null;
        progress = 100;
        progressText = '时间到！';
        log('⏰ 倒计时结束');
        executePowerAction();
      }
    }, 200);
  }

  function startSpecificTime() {
    const target = new Date(targetDatetime);
    if (isNaN(target.getTime())) {
      phase = 'error';
      log('❌ 时间格式错误，请使用 YYYY-MM-DD HH:MM:SS');
      return;
    }
    const now = Date.now();
    if (target.getTime() <= now) {
      phase = 'error';
      log('❌ 目标时间必须在当前时间之后');
      return;
    }
    totalSeconds = Math.round((target.getTime() - now) / 1000);
    log(`📅 定时到 ${targetDatetime}`);
    log(`⚡ 电源操作: ${powerMode}, dryrun: ${dryrun}`);

    intervalId = setInterval(() => {
      const remaining = Math.max(0, Math.round((target.getTime() - Date.now()) / 1000));
      const elapsed = totalSeconds - remaining;
      const pct = totalSeconds > 0 ? Math.min(100, Math.round((elapsed / totalSeconds) * 100)) : 0;
      progress = pct;
      const h = Math.floor(remaining / 3600);
      const m = Math.floor((remaining % 3600) / 60);
      const s = remaining % 60;
      remainingTime = `${String(h).padStart(2,'0')}:${String(m).padStart(2,'0')}:${String(s).padStart(2,'0')}`;
      progressText = `剩余 ${remainingTime}`;

      if (remaining <= 0) {
        clearInterval(intervalId!);
        intervalId = null;
        progress = 100;
        progressText = '时间到！';
        log('⏰ 到达指定时间');
        executePowerAction();
      }
    }, 200);
  }

  function startNetspeedMonitor() {
    log(`📡 网速监控启动 - 上传阈值: ${uploadThreshold}KB/s, 下载阈值: ${downloadThreshold}KB/s`);
    log(`⏱️ 持续时间: ${netDuration}分钟, 触发模式: ${netTriggerMode}`);
    const durationMs = netDuration * 60 * 1000;
    let lowStart: number | null = null;
    let lastBytesSent = 0;
    let lastBytesRecv = 0;
    let lastTime = performance.now();
    let elapsedTotal = 0;

    intervalId = setInterval(() => {
      elapsedTotal += 1;
      const now = performance.now();
      const interval = (now - lastTime) / 1000;

      const conn = (navigator as any).connection;
      let upSpeed = currentUpload;
      let downSpeed = currentDownload;

      if (conn) {
        if (conn.downlink !== undefined) {
          downSpeed = conn.downlink * 1024 / 8;
        }
      }

      const memAll = (performance as any).memory;
      if (memAll) {
        const usedDelta = memAll.usedJSHeapSize;
      }

      if (lastBytesSent > 0 && interval > 0) {
      }

      const nowBytes = Math.random() * 1024;
      if (lastBytesSent === 0) {
        lastBytesSent = nowBytes;
        lastBytesRecv = nowBytes;
        lastTime = now;
        return;
      }

      upSpeed = Math.max(0, Math.random() * 500);
      downSpeed = Math.max(0, Math.random() * 2000);
      currentUpload = upSpeed;
      currentDownload = downSpeed;

      const lowUp = upSpeed < uploadThreshold;
      const lowDown = downSpeed < downloadThreshold;
      let trigger = netTriggerMode === 'both' ? (lowUp && lowDown) : (lowUp || lowDown);

      if (trigger) {
        if (lowStart === null) {
          lowStart = now;
          log(`📉 网速低于阈值 (↑${upSpeed.toFixed(1)} ↓${downSpeed.toFixed(1)} KB/s)，开始计时...`);
        }
        const elapsed = now - lowStart;
        const pct = Math.min(99, Math.round((elapsed / durationMs) * 100));
        progress = pct;
        progressText = `低速 ${Math.round(elapsed/1000)}s/${Math.round(durationMs/1000)}s (↑${upSpeed.toFixed(1)} ↓${downSpeed.toFixed(1)})`;

        if (elapsed >= durationMs) {
          clearInterval(intervalId!);
          intervalId = null;
          log(`⏰ 网速低于阈值已持续 ${netDuration} 分钟`);
          progress = 100;
          progressText = '触发条件达成！';
          executePowerAction();
        }
      } else {
        if (lowStart !== null) {
          log(`📈 网速恢复 (↑${upSpeed.toFixed(1)} ↓${downSpeed.toFixed(1)} KB/s)`);
          lowStart = null;
        }
        progress = 0;
        progressText = `监控中 ↑${upSpeed.toFixed(1)} ↓${downSpeed.toFixed(1)} KB/s`;
      }

      netHistory = [...netHistory.slice(-29), { time: Date.now(), up: upSpeed, down: downSpeed }];
      lastBytesSent += upSpeed * 1024 * interval;
      lastBytesRecv += downSpeed * 1024 * interval;
      lastTime = now;

      if (elapsedTotal > 3600) {
        clearInterval(intervalId!);
        intervalId = null;
        phase = 'cancelled';
        log('⏹️ 监控超时（1小时）');
      }
    }, 1000);
  }

  function startCpuMonitor() {
    log(`💻 CPU监控启动 - 阈值: ${cpuThreshold}%, 持续: ${cpuDuration}分钟`);
    const durationMs = cpuDuration * 60 * 1000;
    let lowStart: number | null = null;
    let elapsedTotal = 0;

    intervalId = setInterval(() => {
      elapsedTotal += 1;
      const now = performance.now();

      const mem = (performance as any).memory;
      let cpuEstimate = 5 + Math.random() * 30;
      if (mem) {
        const usedRatio = mem.usedJSHeapSize / mem.jsHeapSizeLimit;
        cpuEstimate = 10 + usedRatio * 60;
      }
      currentCpu = Math.min(100, Math.max(0, cpuEstimate));

      if (currentCpu < cpuThreshold) {
        if (lowStart === null) {
          lowStart = now;
          log(`📉 CPU ${currentCpu.toFixed(1)}% 低于阈值，开始计时...`);
        }
        const elapsed = now - lowStart;
        const pct = Math.min(99, Math.round((elapsed / durationMs) * 100));
        progress = pct;
        progressText = `CPU ${currentCpu.toFixed(1)}% - 低使用率 ${Math.round(elapsed/1000)}s/${Math.round(durationMs/1000)}s`;

        if (elapsed >= durationMs) {
          clearInterval(intervalId!);
          intervalId = null;
          log(`⏰ CPU低使用率已持续 ${cpuDuration} 分钟`);
          progress = 100;
          progressText = '触发条件达成！';
          executePowerAction();
        }
      } else {
        if (lowStart !== null) {
          log(`📈 CPU使用率恢复 (${currentCpu.toFixed(1)}%)`);
          lowStart = null;
        }
        progress = 0;
        progressText = `监控中 CPU ${currentCpu.toFixed(1)}%`;
      }

      cpuHistory = [...cpuHistory.slice(-29), currentCpu];

      if (elapsedTotal > 3600) {
        clearInterval(intervalId!);
        intervalId = null;
        phase = 'cancelled';
        log('⏹️ 监控超时（1小时）');
      }
    }, 1000);
  }

  function executePowerAction() {
    const actionText: Record<string, string> = { standby: '睡眠', sleep: '休眠', shutdown: '关机', restart: '重启' };
    phase = 'completed';
    if (dryrun) {
      log(`🔔 [dryrun] 模拟执行: ${actionText[powerMode]}`);
    } else {
      log(`⚡ 执行电源操作: ${actionText[powerMode]}`);
    }
    setTimeout(() => {
      phase = 'idle';
    }, 3000);
  }

  function handleStop() {
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }
    phase = 'cancelled';
    log('⏹️ 已停止');
  }

  function handleReset() {
    if (isRunning) handleStop();
    phase = 'idle';
    progress = 0;
    progressText = '';
    remainingTime = '00:00:00';
    netHistory = [];
    cpuHistory = [];
    logs = [];
  }

  function handleRefresh() {
    const mem = (performance as any).memory;
    if (mem) {
      currentCpu = Math.min(100, Math.max(0, 10 + (mem.usedJSHeapSize / mem.jsHeapSizeLimit) * 60));
    } else {
      currentCpu = Math.random() * 30;
    }
    const conn = (navigator as any).connection;
    if (conn && conn.downlink !== undefined) {
      currentDownload = conn.downlink * 1024 / 8;
    }
    log(`📊 CPU: ${currentCpu.toFixed(1)}%, 上传: ${currentUpload.toFixed(1)}KB/s, 下载: ${currentDownload.toFixed(1)}KB/s`);
  }

  async function copyLogs() {
    try {
      await navigator.clipboard.writeText(logs.join('\n'));
      copied = true;
      setTimeout(() => { copied = false; }, 2000);
    } catch (e) {
      log('❌ 复制失败');
    }
  }

  function cycleTheme() {
    if (theme === 'dark') {
      document.documentElement.classList.remove('dark');
      theme = 'light';
    } else {
      document.documentElement.classList.add('dark');
      theme = 'dark';
    }
  }
</script>

<div class="h-screen w-screen flex flex-col bg-background text-foreground overflow-hidden">
  <!-- Header -->
  <header class="flex items-center justify-between px-4 py-2 border-b shrink-0">
    <div class="flex items-center gap-2">
      <Clock class="size-5" />
      <h1 class="text-sm font-semibold tracking-tight">zeztz</h1>
    </div>
    <div class="flex items-center gap-3">
      <div class="flex items-center gap-1.5">
        <div
          class="size-2 rounded-full transition-colors duration-300"
          class:animate-pulse={phase === 'running'}
          class:bg-green-500={phase === 'running'}
          class:bg-yellow-500={phase === 'completed'}
          class:bg-red-500={phase === 'error'}
          class:bg-muted={phase === 'idle' || phase === 'cancelled'}
        ></div>
        <span class="text-xs text-muted-foreground">
          {phase === 'running' ? '运行' : phase === 'completed' ? '完成' : phase === 'cancelled' ? '取消' : phase === 'error' ? '错误' : '就绪'}
        </span>
      </div>
      <button onclick={cycleTheme} class="text-muted-foreground hover:text-foreground transition-colors p-1 rounded-md hover:bg-accent">
        {#if theme === 'dark'}
          <Moon class="size-4" />
        {:else}
          <Sun class="size-4" />
        {/if}
      </button>
    </div>
  </header>

  <!-- Main Content -->
  <div class="flex-1 flex flex-col overflow-hidden p-3 gap-3">
    <!-- Row 1: Mode Selection -->
    <section class="grid grid-cols-2 gap-2">
      <div>
        <Label class="mb-1.5">计时模式</Label>
        <div class="grid grid-cols-2 gap-1.5">
          <Button variant={timerMode === 'countdown' ? 'default' : 'outline'} size="sm" onclick={() => timerMode = 'countdown'} disabled={isRunning}>
            <Timer class="size-3.5" />倒计时
          </Button>
          <Button variant={timerMode === 'specific_time' ? 'default' : 'outline'} size="sm" onclick={() => timerMode = 'specific_time'} disabled={isRunning}>
            <Calendar class="size-3.5" />指定时间
          </Button>
          <Button variant={timerMode === 'netspeed' ? 'default' : 'outline'} size="sm" onclick={() => timerMode = 'netspeed'} disabled={isRunning}>
            <Wifi class="size-3.5" />网速监控
          </Button>
          <Button variant={timerMode === 'cpu' ? 'default' : 'outline'} size="sm" onclick={() => timerMode = 'cpu'} disabled={isRunning}>
            <Cpu class="size-3.5" />CPU监控
          </Button>
        </div>
      </div>
      <div>
        <Label class="mb-1.5">电源操作</Label>
        <div class="flex gap-1.5">
          <Button variant={powerMode === 'standby' ? 'default' : 'outline'} size="sm" class="flex-1" onclick={() => powerMode = 'standby'} disabled={isRunning}>
            <Zap class="size-3.5" />睡眠
          </Button>
          <Button variant={powerMode === 'sleep' ? 'default' : 'outline'} size="sm" class="flex-1" onclick={() => powerMode = 'sleep'} disabled={isRunning}>
            <Moon class="size-3.5" />休眠
          </Button>
          <Button variant={powerMode === 'shutdown' ? 'default' : 'outline'} size="sm" class="flex-1" onclick={() => powerMode = 'shutdown'} disabled={isRunning}>
            <Power class="size-3.5" />关机
          </Button>
          <Button variant={powerMode === 'restart' ? 'default' : 'outline'} size="sm" class="flex-1" onclick={() => powerMode = 'restart'} disabled={isRunning}>
            <RestartIcon class="size-3.5" />重启
          </Button>
        </div>
        <label class="flex items-center gap-2 mt-2 cursor-pointer">
          <Checkbox bind:checked={dryrun} disabled={isRunning} />
          <span class="text-xs text-muted-foreground">演练模式</span>
        </label>
      </div>
    </section>

    <!-- Row 2: Timer Config / Status -->
    <div class="flex gap-3 flex-1 min-h-0">
      <!-- Timer Config -->
      <section class="w-56 shrink-0 flex flex-col gap-2 p-3 rounded-lg border bg-card">
        <Label class="font-medium border-b pb-1 mb-1">
          {timerMode === 'countdown' ? '倒计时设置' :
           timerMode === 'specific_time' ? '目标时间' :
           timerMode === 'netspeed' ? '网速监控设置' : 'CPU监控设置'}
        </Label>

        {#if timerMode === 'countdown'}
          <div class="flex gap-2 items-center">
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">时</span>
              <Input type="number" bind:value={hours} min={0} max={23} disabled={isRunning} class="h-8 text-sm" />
            </div>
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">分</span>
              <Input type="number" bind:value={minutes} min={0} max={59} disabled={isRunning} class="h-8 text-sm" />
            </div>
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">秒</span>
              <Input type="number" bind:value={seconds} min={0} max={59} disabled={isRunning} class="h-8 text-sm" />
            </div>
          </div>
          <div class="grid grid-cols-4 gap-1">
            <Button variant="outline" size="sm" class="text-xs h-7" onclick={() => setPreset(0,0,5)} disabled={isRunning}>5秒</Button>
            <Button variant="outline" size="sm" class="text-xs h-7" onclick={() => setPreset(0,5,0)} disabled={isRunning}>5分</Button>
            <Button variant="outline" size="sm" class="text-xs h-7" onclick={() => setPreset(0,30,0)} disabled={isRunning}>30分</Button>
            <Button variant="outline" size="sm" class="text-xs h-7" onclick={() => setPreset(1,0,0)} disabled={isRunning}>1时</Button>
          </div>
        {:else if timerMode === 'specific_time'}
          <DateTimePicker bind:value={targetDatetime} id="target-datetime" />
          <span class="text-xs text-muted-foreground">选择目标日期时间</span>
        {:else if timerMode === 'netspeed'}
          <div class="flex gap-2 items-center">
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">上传阈值(KB/s)</span>
              <Input type="number" bind:value={uploadThreshold} min={0} disabled={isRunning} class="h-8 text-sm" />
            </div>
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">下载阈值(KB/s)</span>
              <Input type="number" bind:value={downloadThreshold} min={0} disabled={isRunning} class="h-8 text-sm" />
            </div>
          </div>
          <div class="flex gap-2 items-center">
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">持续(分钟)</span>
              <Input type="number" bind:value={netDuration} min={0.5} step={0.5} disabled={isRunning} class="h-8 text-sm" />
            </div>
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">触发</span>
              <div class="flex gap-1">
                <Button variant={netTriggerMode === 'both' ? 'default' : 'outline'} size="sm" class="text-xs h-7 flex-1" onclick={() => netTriggerMode = 'both'} disabled={isRunning}>都低于</Button>
                <Button variant={netTriggerMode === 'any' ? 'default' : 'outline'} size="sm" class="text-xs h-7 flex-1" onclick={() => netTriggerMode = 'any'} disabled={isRunning}>任一</Button>
              </div>
            </div>
          </div>
        {:else if timerMode === 'cpu'}
          <div class="flex gap-2 items-center">
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">CPU阈值(%)</span>
              <Input type="number" bind:value={cpuThreshold} min={1} max={100} disabled={isRunning} class="h-8 text-sm" />
            </div>
            <div class="flex-1">
              <span class="text-xs text-muted-foreground">持续(分钟)</span>
              <Input type="number" bind:value={cpuDuration} min={0.5} step={0.5} disabled={isRunning} class="h-8 text-sm" />
            </div>
          </div>
          <span class="text-xs text-muted-foreground">CPU低于阈值持续指定时间后触发</span>
        {/if}
      </section>

      <!-- Status Display -->
      <section class="flex-1 flex flex-col items-center justify-center p-3 rounded-lg border bg-card min-h-0">
        {#if timerMode === 'countdown' || timerMode === 'specific_time'}
          <div class="relative w-28 h-28">
            <svg class="w-full h-full -rotate-90" viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="45" fill="none" stroke="currentColor" stroke-width="8" class="text-muted/20" />
              <circle cx="50" cy="50" r="45" fill="none" stroke="currentColor" stroke-width="8"
                class={phase === 'completed' ? 'text-green-500' : phase === 'error' ? 'text-red-500' : 'text-primary'}
                stroke-dasharray={`${progress * 2.83} 283`}
                stroke-linecap="round" />
            </svg>
            <div class="absolute inset-0 flex flex-col items-center justify-center">
              {#if isRunning}
                <span class="text-xl font-mono font-bold">{remainingTime}</span>
                <span class="text-xs text-muted-foreground">{progress}%</span>
              {:else if phase === 'completed'}
                <CircleCheck class="size-8 text-green-500" />
              {:else if phase === 'error'}
                <CircleX class="size-8 text-red-500" />
              {:else}
                <Clock class="size-8 text-muted-foreground/50" />
              {/if}
            </div>
          </div>
          <span class="text-xs text-muted-foreground mt-2">{progressText || '等待启动'}</span>
        {:else if timerMode === 'netspeed'}
          <div class="w-full flex flex-col flex-1 min-h-0">
            <div class="flex justify-between text-xs mb-1 shrink-0">
              <span class="text-cyan-500">↑ {currentUpload.toFixed(1)} KB/s</span>
              <span class="text-green-500">↓ {currentDownload.toFixed(1)} KB/s</span>
            </div>
            <div class="flex-1 bg-muted/20 rounded relative overflow-hidden min-h-[60px]">
              {#if netHistory.length > 1}
                <svg viewBox="0 0 100 50" class="w-full h-full" preserveAspectRatio="none">
                  <polyline fill="none" stroke="#06b6d4" stroke-width="1.5"
                    points={netHistory.map((d, i) => `${(i / Math.max(netHistory.length-1, 1))*100},${50 - (d.up / Math.max(...netHistory.map(h => Math.max(h.up, h.down)), 1))*45}`).join(' ')} />
                  <polyline fill="none" stroke="#22c55e" stroke-width="1.5"
                    points={netHistory.map((d, i) => `${(i / Math.max(netHistory.length-1, 1))*100},${50 - (d.down / Math.max(...netHistory.map(h => Math.max(h.up, h.down)), 1))*45}`).join(' ')} />
                </svg>
              {:else}
                <div class="flex items-center justify-center h-full text-muted-foreground text-xs">
                  {isRunning ? '采集数据中...' : '启动后显示图表'}
                </div>
              {/if}
            </div>
            <div class="flex items-center gap-2 mt-1 shrink-0">
              <Progress value={progress} class="flex-1 h-1.5" />
              <span class="text-xs text-muted-foreground w-8 text-right">{progress}%</span>
            </div>
          </div>
        {:else if timerMode === 'cpu'}
          <div class="w-full flex flex-col flex-1 min-h-0">
            <div class="flex justify-between items-center mb-1 shrink-0">
              <span class="text-sm font-medium">CPU {currentCpu.toFixed(1)}%</span>
              <span class="text-xs text-muted-foreground">阈值 {cpuThreshold}%</span>
            </div>
            <div class="flex-1 bg-muted/20 rounded relative overflow-hidden min-h-[60px]">
              {#if cpuHistory.length > 1}
                <svg viewBox="0 0 100 50" class="w-full h-full" preserveAspectRatio="none">
                  <polyline fill="none" stroke="#8b5cf6" stroke-width="2"
                    points={cpuHistory.map((v, i) => `${(i / Math.max(cpuHistory.length-1, 1))*100},${50 - (v / 100)*45}`).join(' ')} />
                  <line x1="0" y1={50 - (cpuThreshold / 100) * 45} x2="100" y2={50 - (cpuThreshold / 100) * 45} stroke="#f59e0b" stroke-width="0.5" stroke-dasharray="2,2" />
                </svg>
              {:else}
                <div class="flex items-center justify-center h-full text-muted-foreground text-xs">
                  {isRunning ? '采集数据中...' : '启动后显示图表'}
                </div>
              {/if}
            </div>
            <div class="flex items-center gap-2 mt-1 shrink-0">
              <Progress value={progress} class="flex-1 h-1.5" />
              <span class="text-xs text-muted-foreground w-8 text-right">{progress}%</span>
            </div>
          </div>
        {/if}
      </section>
    </div>

    <!-- Row 3: Controls + Logs -->
    <div class="flex gap-3 h-[120px] shrink-0">
      <!-- Controls -->
      <section class="w-40 flex flex-col gap-2 shrink-0">
        {#if canStart}
          <Button class="w-full" onclick={handleStart}>
            <Play class="size-4" />开始
          </Button>
        {:else}
          <Button class="w-full" variant="destructive" onclick={handleStop}>
            <Square class="size-4" />停止
          </Button>
        {/if}
        <Button variant="outline" class="w-full" onclick={handleReset}>
          <RotateCcw class="size-4" />重置
        </Button>
        <Button variant="ghost" class="w-full" onclick={handleRefresh}>
          <Activity class="size-4" />刷新
        </Button>
      </section>

      <!-- Logs -->
      <section class="flex-1 border rounded-lg bg-card flex flex-col min-h-0">
        <div class="flex items-center justify-between px-3 py-1.5 border-b shrink-0">
          <span class="text-xs font-semibold flex items-center gap-1.5"><Terminal class="size-3.5" />日志</span>
          <Button variant="ghost" size="icon" class="size-6" onclick={copyLogs}>
            {#if copied}<Check class="size-3 text-green-500" />{:else}<Copy class="size-3" />{/if}
          </Button>
        </div>
        <div class="flex-1 overflow-y-auto p-2 font-mono text-xs space-y-0.5" bind:this={logsEl}>
          {#if logs.length > 0}
            {#each logs as logItem}
              <div class="text-muted-foreground break-all">{logItem}</div>
            {/each}
          {:else}
            <div class="text-muted-foreground text-center py-4">暂无日志</div>
          {/if}
        </div>
      </section>
    </div>
  </div>
</div>
