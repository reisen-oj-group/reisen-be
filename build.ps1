<#
.SYNOPSIS
    ����Զ������ű����� Windows11 �� CentOS7
.DESCRIPTION
    �˽ű����Զ����� Go ��ĿΪ Linux (CentOS) ��ִ���ļ�
    ��Ҫ��ǰ��װ Go 1.16+ ����
.NOTES
    File Name      : build.ps1
    Prerequisite   : PowerShell 5.1+, Go 1.16+
#>

# ���ñ���
$ProjectName = "reisen-be"   # ��Ŀ����
$OutputDir = ".\bin"         # ���Ŀ¼

# ��� Go ����
function Check-GoEnv {
    try {
        $goVersion = (go version) -split " " | Select-Object -Index 2
        Write-Host "��⵽ Go �汾: $goVersion" -ForegroundColor Green
    } catch {
        Write-Host "δ��⵽ Go ���������Ȱ�װ Go" -ForegroundColor Red
        exit 1
    }
}

# �������Ŀ¼
function Create-OutputDirs {
    if (-not (Test-Path -Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir | Out-Null
    }
    Write-Host "���Ŀ¼�Ѵ���: $OutputDir" -ForegroundColor Green
}

# ����ɹ���
function Clean-OldBuilds {
    Remove-Item "$OutputDir\*" -Force -ErrorAction SilentlyContinue
    Write-Host "������ɹ����ļ�" -ForegroundColor Green
}

# ���� Linux �汾
function Build-ForLinux {
    Write-Host "��ʼ���� Linux (CentOS) �汾..." -ForegroundColor Cyan
    
    # ���� Linux ���뻷������
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    
    # ����������
    $outputFile = "$OutputDir\$ProjectName"
    go build -o $outputFile -ldflags="-s -w" ./cmd/server
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "����ɹ�: $outputFile" -ForegroundColor Green
    } else {
        Write-Host "����ʧ��" -ForegroundColor Red
        exit 1
    }
}

# ���������ļ�
function Copy-ConfigFiles {
    $configFiles = @("config.yaml", ".env")
    
    foreach ($file in $configFiles) {
        if (Test-Path -Path $file) {
            Copy-Item $file -Destination $OutputDir
            Write-Host "�Ѹ��������ļ�: $file" -ForegroundColor Green
        }
    }
}

# ��ִ������
Check-GoEnv
Create-OutputDirs
Clean-OldBuilds
Build-ForLinux
Copy-ConfigFiles

Write-Host "������ɣ�" -ForegroundColor Magenta
