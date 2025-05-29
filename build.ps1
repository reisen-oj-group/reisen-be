<#
.SYNOPSIS
    后端自动构建脚本，从 Windows11 到 CentOS7
.DESCRIPTION
    此脚本会自动编译 Go 项目为 Linux (CentOS) 可执行文件
    需要提前安装 Go 1.16+ 环境
.NOTES
    File Name      : build.ps1
    Prerequisite   : PowerShell 5.1+, Go 1.16+
#>

# 设置变量
$ProjectName = "reisen-be"   # 项目名称
$OutputDir = ".\bin"         # 输出目录

# 检查 Go 环境
function Check-GoEnv {
    try {
        $goVersion = (go version) -split " " | Select-Object -Index 2
        Write-Host "检测到 Go 版本: $goVersion" -ForegroundColor Green
    } catch {
        Write-Host "未检测到 Go 环境，请先安装 Go" -ForegroundColor Red
        exit 1
    }
}

# 创建输出目录
function Create-OutputDirs {
    if (-not (Test-Path -Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir | Out-Null
    }
    Write-Host "输出目录已创建: $OutputDir" -ForegroundColor Green
}

# 清理旧构建
function Clean-OldBuilds {
    Remove-Item "$OutputDir\*" -Force -ErrorAction SilentlyContinue
    Write-Host "已清理旧构建文件" -ForegroundColor Green
}

# 编译 Linux 版本
function Build-ForLinux {
    Write-Host "开始编译 Linux (CentOS) 版本..." -ForegroundColor Cyan
    
    # 设置 Linux 编译环境变量
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    
    # 编译主程序
    $outputFile = "$OutputDir\$ProjectName"
    go build -o $outputFile -ldflags="-s -w" ./cmd/server
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "编译成功: $outputFile" -ForegroundColor Green
    } else {
        Write-Host "编译失败" -ForegroundColor Red
        exit 1
    }
}

# 复制配置文件
function Copy-ConfigFiles {
    $configFiles = @("config.yaml", ".env")
    
    foreach ($file in $configFiles) {
        if (Test-Path -Path $file) {
            Copy-Item $file -Destination $OutputDir
            Write-Host "已复制配置文件: $file" -ForegroundColor Green
        }
    }
}

# 主执行流程
Check-GoEnv
Create-OutputDirs
Clean-OldBuilds
Build-ForLinux
Copy-ConfigFiles

Write-Host "构建完成！" -ForegroundColor Magenta
