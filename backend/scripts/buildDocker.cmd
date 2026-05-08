@echo off
echo Building Schematic2 Docker image...
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
docker build -f ./build/package/Dockerfile ../ -t schematic2:latest
echo.
echo To run the container, execute:
echo docker run -p 8080:8080 -p 8443:8443 -v %%cd%%\configs:/app/configs schematic2:latest
pause