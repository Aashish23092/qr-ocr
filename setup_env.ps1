# Enable CGO
$env:CGO_ENABLED = "1"

# ---------------- PATH ----------------
$RequiredPaths = @(
    "C:\tools\opencv\build\x64\vc16\bin",
    "C:\msys64\mingw64\bin",
    "C:\ProgramData\mingw64\mingw64\bin"
)

$CurrentPath = $env:PATH
foreach ($path in $RequiredPaths) {
    if ($CurrentPath -notlike "*$path*") {
        $CurrentPath = "$path;$CurrentPath"
    }
}
$env:PATH = $CurrentPath

# ---------------- CFLAGS ----------------
$env:CGO_CFLAGS = @(
    "-IC:/tools/opencv/build/include",
    "-I""C:/Program Files (x86)/ZBar/include""",
    "-IC:/ProgramData/mingw64/mingw64/include"
) -join " "

# ---------------- LDFLAGS ----------------
$env:CGO_LDFLAGS = @(
    "-L""C:/Program Files (x86)/ZBar/lib"" -lzbar",
    "-LC:/ProgramData/mingw64/mingw64/lib -lquirc",
    "-LC:/tools/opencv/build/x64/vc16/lib -lopencv_world4110"
) -join " "

Write-Host "CGO ENABLED: $env:CGO_ENABLED"
Write-Host "CGO_CFLAGS:  $env:CGO_CFLAGS"
Write-Host "CGO_LDFLAGS: $env:CGO_LDFLAGS"
Write-Host "PATH updated."

Write-Host "Run this next:"
Write-Host "go build -v -x -o aadhaar-qr-service.exe"
