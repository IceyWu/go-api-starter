$target = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\bin\server.exe'))

Get-CimInstance Win32_Process -Filter "Name = 'server.exe'" |
    Where-Object { $_.ExecutablePath -eq $target } |
    ForEach-Object { Stop-Process -Id $_.ProcessId -Force }
