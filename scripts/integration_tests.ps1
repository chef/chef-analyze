ECHO "-- Installing Golang"
hab pkg install -b afiune/go

ECHO "-- Building chef-analyze EXE"
go build -o bin\chef-analyze_windows.exe

ECHO "-- Running integration tests (github.com/chef/chef-analyze/integration)"
$Env:CHEF_ANALYZE_BIN = (Get-PSDrive Habitat).Root + "\src\bin\chef-analyze_windows.exe"
go test github.com/chef/chef-analyze/integration -v
