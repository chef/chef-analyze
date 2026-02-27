# Stop script execution when a non-terminating error occurs
$ErrorActionPreference = "Stop"

# Enable the use of Go vendor/ directory
$Env:GOFLAGS = "-mod=vendor"

ECHO "-- Installing Golang"
hab pkg install -b core/go26
If ($LastExitCode -ne 0) { Exit $LastExitCode }

ECHO "-- Building chef-analyze EXE"
go build -o bin\chef-analyze_windows.exe
If ($LastExitCode -ne 0) { Exit $LastExitCode }

ECHO "-- Running integration tests (github.com/chef/chef-analyze/integration)"
$Env:CHEF_ANALYZE_BIN = (Get-PSDrive Habitat).Root + "\src\bin\chef-analyze_windows.exe"
go test github.com/chef/chef-analyze/integration -v
If ($LastExitCode -ne 0) { Exit $LastExitCode }
