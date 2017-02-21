!include "EnvVarUpdate.nsh"
Name 'Shopify Theme Kit 64 Bit'
OutFile '../build/dist/themekit-setup-64.exe'
InstallDir $PROGRAMFILES\themekit
Section "Install"
  SetOutPath $INSTDIR
  File ../build/dist/windows-amd64/theme.exe
  ${EnvVarUpdate} $0 "PATH" "A" "HKCU" $INSTDIR
SectionEnd
