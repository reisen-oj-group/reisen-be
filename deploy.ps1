# 配置参数
$REMOTE_USER = "root"                     # 远程服务器用户名
$REMOTE_HOST = "se.nanani-fan.club"       # 远程服务器IP/域名
$REMOTE_DIR = "/var/www/reisen/backend"   # 远程目标目录
$LOCAL_DIR = "bin"                        # 本地 bin 目录
$ZIP_NAME = "deploy_temp.zip"             # 临时压缩包名称

# 0. 检查本地目录是否存在
if (-not (Test-Path $LOCAL_DIR)) {
    Write-Output "错误：本地目录 $LOCAL_DIR 不存在！"
    exit 1
}

# 1. 在本地创建压缩包
try {
    # 使用Compress-Archive压缩文件
    Compress-Archive -Path "$LOCAL_DIR\*" -DestinationPath $ZIP_NAME -Force
    Write-Output "本地文件已压缩为 $ZIP_NAME"
}
catch {
    Write-Output "压缩文件时出错: $_"
    exit 1
}

# 2. 清空远程目录并确保目录存在
ssh ${REMOTE_USER}@${REMOTE_HOST} "rm -rf ${REMOTE_DIR}/* && mkdir -p ${REMOTE_DIR}"

# 3. 上传压缩包到远程服务器
scp -P 22 $ZIP_NAME ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/

# 4. 在远程服务器上解压并删除压缩包
ssh ${REMOTE_USER}@${REMOTE_HOST} "unzip -o ${REMOTE_DIR}/${ZIP_NAME} -d ${REMOTE_DIR}"
ssh ${REMOTE_USER}@${REMOTE_HOST} "rm ${REMOTE_DIR}/${ZIP_NAME}"

# 5. 赋予可执行程序权限
ssh ${REMOTE_USER}@${REMOTE_HOST} "chmod 777 ${REMOTE_DIR}/reisen-be"

# 6. 删除本地临时压缩包
Remove-Item $ZIP_NAME -Force

# 检查结果
if ($LASTEXITCODE -eq 0) {
    Write-Output "部署成功！"
} else {
    Write-Output "部署失败！"
}