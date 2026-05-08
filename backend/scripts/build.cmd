@echo off
echo Building Schematics2...
echo.
echo Step 1: Building Frontend (npm)...
cd ..\frontend
call npm run build
if errorlevel 1 (
    echo Error building Frontend!
    exit /b 1
)
cd ..\backend
echo Frontend build complete!
echo.
echo Step 2: Generating TLS certificate...
go run ./cmd/gencert
if errorlevel 1 (
    echo Error generating TLS certificate!
    exit /b 1
)
echo TLS certificate generation complete!
echo.
echo Step 3: Building Go binaries...
go build -ldflags="-s -w" -o ./bin/schematic2.exe ./cmd/server
if errorlevel 1 (
    echo Error building server!
    exit /b 1
)
echo.
echo Build complete!