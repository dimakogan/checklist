set BROWSER="C:\Program Files\Mozilla Firefox\firefox.exe"
# Location of a Firefox profile that needs a safebrowsing update
set PROFILE=ff-profile

set TMP=%TEMP%\%RANDOM%\
del /s /q %TMP%
mkdir %TMP%
copy /y %PROFILE% %TMP%

echo "Using profile in %TMP%"
%BROWSER% --jsconsole --profile %TMP%
