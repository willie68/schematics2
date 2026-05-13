@echo off
setlocal enabledelayedexpansion

REM Accept optional BASE_PATH argument for reverse-proxy deployment
REM Usage: buildDocker.cmd [BASE_PATH]
REM   buildDocker.cmd            builds for reverse-proxy at /schematics2 (default)
REM   buildDocker.cmd /client    builds for direct container access

if "%~1"=="" (
    set BASE_PATH=/schematics2
    echo Building Schematic2 Docker image for reverse-proxy at /schematics2 ^(default^)...
) else (
    set BASE_PATH=%~1
    echo Building Schematic2 Docker image with BASE_PATH=!BASE_PATH!...
)

echo.

REM Get build information for ldflags
for /f "tokens=*" %%i in ('powershell -Command "[System.DateTime]::UtcNow.ToString('o')"') do set BUILD_TIME=%%i
for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set VCS_REF=%%i
if "!VCS_REF!"=="" set VCS_REF=unknown

echo Build Information:
echo   BUILD_TIME: !BUILD_TIME!
echo   VCS_REF (Commit): !VCS_REF!
echo   BASE_PATH: !BASE_PATH!
echo.
echo Step 1: Generating TLS certificate...
go run ./cmd/gencert
if errorlevel 1 (
    echo Error generating TLS certificate!
    pause
    exit /b 1
)
echo TLS certificate generation complete!
echo.
echo Step 2: Building Docker image...
if not exist .\build\package\Dockerfile (
    echo Warning: Dockerfile not found at .\build\package\Dockerfile
    echo Please create a Dockerfile for schematic2
    pause
    exit /b 1
)
docker build -f ./build/package/Dockerfile ^
    --build-arg BASE_PATH=!BASE_PATH! ^
    --build-arg BUILD_TIME=!BUILD_TIME! ^
    --build-arg VCS_REF=!VCS_REF! ^
    ../ -t mcs/schematics2:latest
if errorlevel 1 (
    echo Error building Docker image!
    pause
    exit /b 1
)
echo Docker image build complete!
echo.
echo Step 3: Tagging image for Docker registry...
docker tag mcs/schematics2:latest 192.168.178.14:5000/mcs/schematics2:latest
echo.
echo Step 4: Pushing image to Docker registry (192.168.178.14:5000)...
docker push 192.168.178.14:5000/mcs/schematics2:latest
echo.
echo Docker image successfully pushed to 192.168.178.14:5000/mcs/schematics2:latest
echo.
echo On Ubuntu Server (192.168.178.14), pull the image with:
echo docker pull 192.168.178.14:5000/mcs/schematics2:latest
echo.
echo Or use docker-compose or Kubernetes to deploy from the registry.
echo.
echo To run the container locally, execute:
echo docker run -p 8080:8080 -p 8443:8443 -v %%cd%%\configs:/app/configs mcs/schematics2:latest
