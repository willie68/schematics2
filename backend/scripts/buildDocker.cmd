@echo off
echo Building Schematic2 Docker image with deployment config...
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
echo Step 2: Building Docker image (with BASE_PATH=/schematics2)...
if not exist .\build\package\Dockerfile (
    echo Warning: Dockerfile not found at .\build\package\Dockerfile
    echo Please create a Dockerfile for schematic2
    pause
    exit /b 1
)
docker build -f ./build/package/Dockerfile --build-arg BASE_PATH=/schematics2 ../ -t mcs/schematics2:latest
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
pause