@echo off
echo Running unit tests with coverage...
go test -coverprofile cover.out ./...
echo.
echo Coverage summary:
go tool cover -func cover.out
echo.
echo To view detailed coverage report, run: go tool cover -html=cover.out
pause