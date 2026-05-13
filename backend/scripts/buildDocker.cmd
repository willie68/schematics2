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
docker build -f ./build/package/Dockerfile ../ -t mcs/schematics2:latest
echo.
echo Step 3: Tagging image for Docker registry...
docker tag mcs/schematics2:latest 192.168.178.14:5000/mcs/schematics2:latest
echo.
echo Step 4: Pushing image to Docker registry (192.168.178.14:5000)...
docker push 192.168.178.14:5000/mcs/schematics2:latest
echo.
echo Docker image successfully pushed to 192.168.178.14:5000/mcs/schematics2:latest
echo.
echo To run the container, execute:
echo docker run -p 8080:8080 -p 8443:8443 -v %%cd%%\configs:/app/configs mcs/schematics2:latest
pause