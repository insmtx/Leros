$ErrorActionPreference = 'Stop'
. "$PSScriptRoot\shared.ps1"
Import-DevEnvFile

$root = Get-LerosRepoRoot
$runtimeState = Initialize-DevRuntimeState
$dbPath = Get-WorkerRecoveryDbPath -RepoRoot $root
$workspaceRoot = Get-WorkerWorkspaceRoot -RepoRoot $root

Stop-DevProcessesByPorts -Ports @([int]$runtimeState.serverPort, [int]$runtimeState.workerPort)

Write-Host '[Leros] Stopping remaining backend processes...' -ForegroundColor Cyan
Get-Process leros -ErrorAction SilentlyContinue | Stop-Process -Force

if (-not (Test-Path $dbPath)) {
    throw "Worker recovery database not found: $dbPath"
}

$sqliteExe = Get-Sqlite3Exe

# 只清理问答任务主题的恢复游标，避免旧的终态记录继续误判新消息。
Write-Host '[Leros] Clearing worker task recovery state for chat stuck issue...' -ForegroundColor Cyan
& $sqliteExe $dbPath "delete from task_seq where topic='org.1.worker.1.task';"
if ($LASTEXITCODE -ne 0) {
    throw 'Failed to clear task_seq recovery state.'
}

# 这里额外打印剩余记录数，便于快速确认清理是否生效。
$remainingCount = & $sqliteExe $dbPath "select count(*) from task_seq where topic='org.1.worker.1.task';"
Write-Host "[Leros] Remaining recovery rows for org.1.worker.1.task: $remainingCount" -ForegroundColor Green

if (-not (Test-Path "$root\bundles\leros.exe")) {
    & "$PSScriptRoot\rebuild-backend.ps1"
}

Write-Host '[Leros] Restarting server and worker...' -ForegroundColor Cyan
Start-Process powershell.exe -ArgumentList '-NoExit', '-ExecutionPolicy', 'Bypass', '-File', "$PSScriptRoot\run-server-dev.ps1" | Out-Null
Start-Sleep -Seconds 2
Start-Process powershell.exe -ArgumentList '-NoExit', '-ExecutionPolicy', 'Bypass', '-File', "$PSScriptRoot\run-worker-dev.ps1" | Out-Null

Write-Host ''
Write-Host '[Leros] Chat stuck recovery completed.' -ForegroundColor Green
Write-Host "[Leros] Workspace root: $workspaceRoot" -ForegroundColor DarkGray
Write-Host '[Leros] If the page is still generating forever, refresh once and retry the question.' -ForegroundColor Green
Read-Host 'Press Enter to exit'
