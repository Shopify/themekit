!include "EnvVarUpdate.nsh"
Name 'Shopify Theme Kit 32 Bit'
OutFile '../build/dist/themekit-setup-32.exe'
InstallDir $PROGRAMFILES\themekit
Section "Install"
  SetOutPath $INSTDIR
  File ../build/dist/windows-386/theme.exe
  ${EnvVarUpdate} $0 "PATH" "A" "HKCU" $INSTDIR
SectionEnd
